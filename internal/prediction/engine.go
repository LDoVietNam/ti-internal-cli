package prediction

import (
	"sort"
	"strings"
)

type Engine struct {
	weights Weights
}

func NewEngine(weights Weights) Engine {
	return Engine{weights: weights}
}

type scoredCandidate struct {
	candidate Candidate
	score     float64
	reasons   []Reason
}

func (e Engine) Predict(request Request, candidates []Candidate) (Prediction, error) {
	eligible := make([]scoredCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if !isEligible(request, candidate) {
			continue
		}
		score, reasons := e.score(request, candidate)
		eligible = append(eligible, scoredCandidate{candidate: candidate, score: score, reasons: reasons})
	}
	if len(eligible) == 0 {
		return Prediction{}, ErrNoEligibleCandidate
	}

	sort.SliceStable(eligible, func(i, j int) bool {
		if eligible[i].score == eligible[j].score {
			return eligible[i].candidate.ID < eligible[j].candidate.ID
		}
		return eligible[i].score > eligible[j].score
	})

	best := eligible[0]
	fallbacks := make([]PredictionCandidate, 0, len(eligible)-1)
	for _, item := range eligible[1:] {
		fallbacks = append(fallbacks, PredictionCandidate{
			CandidateID: item.candidate.ID,
			WorkflowID:  item.candidate.WorkflowID,
			WorkerID:    item.candidate.WorkerID,
			ProviderID:  item.candidate.ProviderID,
			ModelID:     item.candidate.ModelID,
			Score:       item.score,
		})
	}

	return Prediction{
		CandidateID:      best.candidate.ID,
		TaskType:         request.Task.Type,
		Intent:           append([]string(nil), request.Task.Intent...),
		WorkflowID:       best.candidate.WorkflowID,
		WorkerID:         best.candidate.WorkerID,
		ProviderID:       best.candidate.ProviderID,
		ModelID:          best.candidate.ModelID,
		PromptIDs:        append([]string(nil), best.candidate.PromptIDs...),
		RequiredTools:    append([]string(nil), best.candidate.RequiredTools...),
		RequiredContext:  append([]string(nil), best.candidate.RequiredContext...),
		VerificationPlan: append([]string(nil), best.candidate.VerificationPlan...),
		StopConditions:   append([]string(nil), best.candidate.StopConditions...),
		Confidence:       clamp(best.score),
		Reasons:          best.reasons,
		Fallbacks:        fallbacks,
	}, nil
}

func isEligible(request Request, candidate Candidate) bool {
	if !candidate.Enabled || !candidate.Available || !candidate.PolicyAllowed {
		return false
	}
	if candidate.SourceTrust < request.MinimumSourceTrust {
		return false
	}
	return coverage(request.RequiredCapabilities, candidate.Capabilities) == 1
}

func (e Engine) score(request Request, candidate Candidate) (float64, []Reason) {
	values := []struct {
		name   string
		value  float64
		weight float64
	}{
		{"task_match", taskMatch(request.Task.Type, candidate.TaskTypes), e.weights.TaskMatch},
		{"repository_match", coverage(request.Task.RepositoryTags, candidate.RepositoryTags), e.weights.RepositoryMatch},
		{"language_framework_match", languageFrameworkMatch(request.Task, candidate), e.weights.LanguageFramework},
		{"capability_match", coverage(request.RequiredCapabilities, candidate.Capabilities), e.weights.CapabilityMatch},
		{"historical_success", clamp(candidate.HistoricalSuccess), e.weights.HistoricalSuccess},
		{"verification_success", clamp(candidate.VerificationSuccess), e.weights.VerificationSuccess},
		{"source_trust", clamp(candidate.SourceTrust), e.weights.SourceTrust},
		{"quality", clamp(candidate.Quality), e.weights.Quality},
		{"freshness", clamp(candidate.Freshness), e.weights.Freshness},
	}

	reasons := make([]Reason, 0, len(values))
	total := 0.0
	for _, value := range values {
		score := value.value * value.weight
		reasons = append(reasons, Reason{Factor: value.name, Value: value.value, Weight: value.weight, Score: score})
		total += score
	}
	total -= clamp(candidate.RiskPenalty)
	total -= clamp(candidate.RedundancyPenalty)
	return clamp(total), reasons
}

func taskMatch(task TaskType, supported []TaskType) float64 {
	if task == TaskUnknown || len(supported) == 0 {
		return 0
	}
	for _, item := range supported {
		if item == task {
			return 1
		}
	}
	return 0
}

func languageFrameworkMatch(task TaskProfile, candidate Candidate) float64 {
	parts := make([]float64, 0, 2)
	if task.Language != "" {
		parts = append(parts, stringMatch(task.Language, candidate.Languages))
	}
	if task.Framework != "" {
		parts = append(parts, stringMatch(task.Framework, candidate.Frameworks))
	}
	if len(parts) == 0 {
		return 1
	}
	total := 0.0
	for _, part := range parts {
		total += part
	}
	return total / float64(len(parts))
}

func stringMatch(value string, values []string) float64 {
	for _, item := range values {
		if strings.EqualFold(value, item) {
			return 1
		}
	}
	return 0
}

func coverage(required, available []string) float64 {
	if len(required) == 0 {
		return 1
	}
	set := make(map[string]struct{}, len(available))
	for _, item := range available {
		set[strings.ToLower(item)] = struct{}{}
	}
	matched := 0
	for _, item := range required {
		if _, ok := set[strings.ToLower(item)]; ok {
			matched++
		}
	}
	return float64(matched) / float64(len(required))
}

func clamp(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
