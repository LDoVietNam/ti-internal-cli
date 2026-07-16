package prediction

import "errors"

type TaskType string

const (
	TaskUnknown       TaskType = "unknown"
	TaskBugFix        TaskType = "bug_fix"
	TaskFeature       TaskType = "implement_feature"
	TaskRefactor      TaskType = "refactor"
	TaskCodeReview    TaskType = "code_review"
	TaskInvestigation TaskType = "investigate"
)

var ErrNoEligibleCandidate = errors.New("no eligible prediction candidate")

type TaskProfile struct {
	Type           TaskType
	Intent         []string
	Language       string
	Framework      string
	RepositoryTags []string
}

type Request struct {
	Task                 TaskProfile
	RequiredCapabilities []string
	MinimumSourceTrust    float64
}

type Candidate struct {
	ID                 string
	WorkflowID         string
	WorkerID           string
	ProviderID         string
	ModelID            string
	PromptIDs          []string
	RequiredTools      []string
	RequiredContext    []string
	VerificationPlan   []string
	StopConditions     []string
	Capabilities       []string
	TaskTypes          []TaskType
	Languages          []string
	Frameworks         []string
	RepositoryTags     []string
	PolicyAllowed      bool
	Enabled            bool
	Available          bool
	HistoricalSuccess  float64
	VerificationSuccess float64
	SourceTrust        float64
	Quality            float64
	Freshness          float64
	RiskPenalty        float64
	RedundancyPenalty  float64
}

type Reason struct {
	Factor string
	Value  float64
	Weight float64
	Score  float64
}

type PredictionCandidate struct {
	CandidateID string
	WorkflowID  string
	WorkerID    string
	ProviderID  string
	ModelID     string
	Score       float64
}

type Prediction struct {
	CandidateID      string
	TaskType         TaskType
	Intent           []string
	RiskLevel        string
	WorkflowID       string
	WorkerID         string
	ProviderID       string
	ModelID          string
	PromptIDs        []string
	RequiredTools    []string
	RequiredContext  []string
	VerificationPlan []string
	StopConditions   []string
	Confidence       float64
	Reasons          []Reason
	Fallbacks        []PredictionCandidate
}

type Weights struct {
	TaskMatch           float64
	RepositoryMatch     float64
	LanguageFramework   float64
	CapabilityMatch     float64
	HistoricalSuccess   float64
	VerificationSuccess float64
	SourceTrust         float64
	Quality             float64
	Freshness           float64
}

func DefaultWeights() Weights {
	return Weights{
		TaskMatch:           0.25,
		RepositoryMatch:     0.15,
		LanguageFramework:   0.12,
		CapabilityMatch:     0.12,
		HistoricalSuccess:   0.12,
		VerificationSuccess: 0.08,
		SourceTrust:         0.06,
		Quality:             0.05,
		Freshness:           0.05,
	}
}
