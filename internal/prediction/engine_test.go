package prediction

import "testing"

func TestEngineRejectsCandidatesMissingRequiredCapabilities(t *testing.T) {
	engine := NewEngine(DefaultWeights())
	request := Request{
		Task:                 TaskProfile{Type: TaskBugFix, Language: "go"},
		RequiredCapabilities: []string{"filesystem.read", "test.run"},
	}
	candidates := []Candidate{
		{
			ID:            "incomplete",
			WorkflowID:    "fix_bug",
			WorkerID:      "worker-a",
			ProviderID:    "provider-a",
			ModelID:       "model-a",
			Capabilities:  []string{"filesystem.read"},
			PolicyAllowed: true,
			Enabled:       true,
			Available:     true,
		},
		{
			ID:                  "eligible",
			WorkflowID:          "fix_bug",
			WorkerID:            "worker-b",
			ProviderID:          "provider-b",
			ModelID:             "model-b",
			Capabilities:        []string{"filesystem.read", "test.run"},
			TaskTypes:           []TaskType{TaskBugFix},
			Languages:           []string{"go"},
			PolicyAllowed:       true,
			Enabled:             true,
			Available:           true,
			HistoricalSuccess:   0.8,
			VerificationSuccess: 0.9,
			SourceTrust:         1,
			Quality:             0.8,
			Freshness:           1,
		},
	}

	got, err := engine.Predict(request, candidates)
	if err != nil {
		t.Fatal(err)
	}
	if got.CandidateID != "eligible" {
		t.Fatalf("candidate = %q, want eligible", got.CandidateID)
	}
	if len(got.Fallbacks) != 0 {
		t.Fatalf("fallbacks = %d, want 0", len(got.Fallbacks))
	}
}

func TestEngineRanksCandidatesAndExplainsEveryFactor(t *testing.T) {
	engine := NewEngine(DefaultWeights())
	request := Request{
		Task: TaskProfile{
			Type:           TaskBugFix,
			Language:       "go",
			Framework:      "net/http",
			RepositoryTags: []string{"router", "backend"},
		},
		RequiredCapabilities: []string{"filesystem.read"},
	}
	candidates := []Candidate{
		{
			ID:                  "generic",
			WorkflowID:          "investigate",
			Capabilities:        []string{"filesystem.read"},
			TaskTypes:           []TaskType{TaskInvestigation},
			Languages:           []string{"python"},
			PolicyAllowed:       true,
			Enabled:             true,
			Available:           true,
			HistoricalSuccess:   0.5,
			VerificationSuccess: 0.5,
			SourceTrust:         0.6,
			Quality:             0.6,
			Freshness:           0.8,
		},
		{
			ID:                  "specialized",
			WorkflowID:          "fix_bug",
			Capabilities:        []string{"filesystem.read"},
			TaskTypes:           []TaskType{TaskBugFix},
			Languages:           []string{"go"},
			Frameworks:          []string{"net/http"},
			RepositoryTags:      []string{"router"},
			PolicyAllowed:       true,
			Enabled:             true,
			Available:           true,
			HistoricalSuccess:   0.9,
			VerificationSuccess: 0.9,
			SourceTrust:         1,
			Quality:             0.9,
			Freshness:           1,
		},
	}

	got, err := engine.Predict(request, candidates)
	if err != nil {
		t.Fatal(err)
	}
	if got.CandidateID != "specialized" {
		t.Fatalf("candidate = %q, want specialized", got.CandidateID)
	}
	if got.Confidence <= 0 || got.Confidence > 1 {
		t.Fatalf("confidence = %v, want (0,1]", got.Confidence)
	}
	if len(got.Reasons) != 9 {
		t.Fatalf("reasons = %d, want 9", len(got.Reasons))
	}
	if len(got.Fallbacks) != 1 || got.Fallbacks[0].CandidateID != "generic" {
		t.Fatalf("fallbacks = %#v, want generic", got.Fallbacks)
	}
}

func TestProfilerClassifiesCommonCodingTasks(t *testing.T) {
	profiler := NewProfiler()
	cases := []struct {
		input string
		want  TaskType
	}{
		{"Fix timeout in router", TaskBugFix},
		{"Implement OAuth login", TaskFeature},
		{"Refactor provider registry", TaskRefactor},
		{"Review this pull request", TaskCodeReview},
		{"Investigate intermittent latency", TaskInvestigation},
	}
	for _, tc := range cases {
		got := profiler.Profile(tc.input)
		if got.Type != tc.want {
			t.Errorf("Profile(%q) = %q, want %q", tc.input, got.Type, tc.want)
		}
	}
}

func TestWorkflowRecommendationIsDeterministic(t *testing.T) {
	registry := DefaultWorkflowRegistry()
	first, err := registry.Recommend(TaskProfile{Type: TaskBugFix})
	if err != nil {
		t.Fatal(err)
	}
	second, err := registry.Recommend(TaskProfile{Type: TaskBugFix})
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != "fix_bug" || second.ID != first.ID {
		t.Fatalf("recommendations = %q and %q, want fix_bug", first.ID, second.ID)
	}
	if len(first.Steps) != 5 {
		t.Fatalf("steps = %d, want 5", len(first.Steps))
	}
}
