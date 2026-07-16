package prediction

import "fmt"

type WorkflowStep struct {
	ID     string
	Action string
}

type WorkflowDefinition struct {
	ID    string
	Steps []WorkflowStep
}

type WorkflowRegistry struct {
	byTask map[TaskType]WorkflowDefinition
}

func DefaultWorkflowRegistry() WorkflowRegistry {
	return WorkflowRegistry{byTask: map[TaskType]WorkflowDefinition{
		TaskBugFix: {
			ID: "fix_bug",
			Steps: []WorkflowStep{
				{ID: "inspect", Action: "repository.inspect"},
				{ID: "locate", Action: "code.locate"},
				{ID: "patch", Action: "agent.patch"},
				{ID: "test", Action: "verification.test"},
				{ID: "review", Action: "verification.review"},
			},
		},
		TaskFeature: {
			ID: "implement_feature",
			Steps: []WorkflowStep{
				{ID: "clarify", Action: "task.clarify"},
				{ID: "design", Action: "agent.design"},
				{ID: "implement", Action: "agent.implement"},
				{ID: "test", Action: "verification.test"},
				{ID: "review", Action: "verification.review"},
			},
		},
		TaskRefactor: {
			ID: "refactor",
			Steps: []WorkflowStep{
				{ID: "impact", Action: "repository.impact"},
				{ID: "baseline", Action: "verification.baseline"},
				{ID: "transform", Action: "agent.refactor"},
				{ID: "verify", Action: "verification.compatibility"},
			},
		},
		TaskCodeReview: {
			ID: "code_review",
			Steps: []WorkflowStep{
				{ID: "diff", Action: "repository.diff"},
				{ID: "risk", Action: "verification.risk_scan"},
				{ID: "evidence", Action: "verification.collect"},
				{ID: "findings", Action: "verification.report"},
			},
		},
		TaskInvestigation: {
			ID: "investigate",
			Steps: []WorkflowStep{
				{ID: "evidence", Action: "repository.collect_evidence"},
				{ID: "hypotheses", Action: "agent.form_hypotheses"},
				{ID: "test", Action: "agent.test_hypotheses"},
				{ID: "report", Action: "agent.report"},
			},
		},
	}}
}

func (r WorkflowRegistry) Recommend(profile TaskProfile) (WorkflowDefinition, error) {
	workflow, ok := r.byTask[profile.Type]
	if !ok {
		return WorkflowDefinition{}, fmt.Errorf("no workflow for task type %q", profile.Type)
	}
	return workflow, nil
}
