package promptintel

import (
	"testing"

	"github.com/ti-system/ti-internal-cli/internal/docsbot"
)

func testRegistry(t *testing.T) *docsbot.Registry {
	t.Helper()
	skills := []docsbot.Skill{
		{ID: "docsbot-programming-general", Source: "docsbot", Generic: true, Domains: []string{"programming"}, Phases: []string{"observe", "plan", "implement", "verify", "finalize"}},
		{ID: "docsbot-go-testing", Source: "docsbot", Domains: []string{"golang"}, TaskTypes: []string{"bug_fix", "feature"}, Languages: []string{"go"}, Phases: []string{"verify"}, HistoricalSuccess: .8},
		{ID: "docsbot-root-cause", Source: "docsbot", Domains: []string{"debugging"}, TaskTypes: []string{"bug_fix"}, Phases: []string{"observe", "diagnose"}, HistoricalSuccess: .9},
		{ID: "docsbot-code-review", Source: "docsbot", Domains: []string{"programming"}, Phases: []string{"verify"}, HistoricalSuccess: .7},
	}
	registry, err := docsbot.NewRegistry(skills, "docsbot-programming-general")
	if err != nil {
		t.Fatal(err)
	}
	return registry
}

func TestMandatoryResolverAlwaysReturnsDocsBotSkill(t *testing.T) {
	resolver := NewResolver(testRegistry(t), Config{Mode: ModeMandatory, MinimumSkillsPerTask: 1})
	stack, err := resolver.Resolve(Profile{Domain: "rare-domain", TaskType: "unknown", Phase: "implement"})
	if err != nil {
		t.Fatal(err)
	}
	if len(stack.Skills) < 1 {
		t.Fatal("expected at least one skill")
	}
	if stack.Skills[0].Source != "docsbot" {
		t.Fatalf("source = %q, want docsbot", stack.Skills[0].Source)
	}
}

func TestLowScoreUsesGenericFallback(t *testing.T) {
	resolver := NewResolver(testRegistry(t), Config{Mode: ModeMandatory, MinimumSkillsPerTask: 1})
	stack, err := resolver.Resolve(Profile{Domain: "rare-domain", TaskType: "unknown", Language: "zig", Phase: "implement"})
	if err != nil {
		t.Fatal(err)
	}
	if stack.Skills[0].ID != "docsbot-programming-general" {
		t.Fatalf("first skill = %q", stack.Skills[0].ID)
	}
}

func TestAggressiveModeIncludesPhaseAndVerificationSkills(t *testing.T) {
	resolver := NewResolver(testRegistry(t), Config{Mode: ModeAggressive, MinimumSkillsPerTask: 3, IncludePhaseSkills: true, IncludeVerificationSkill: true})
	stack, err := resolver.Resolve(Profile{Domain: "debugging", TaskType: "bug_fix", Language: "go", Phase: "observe"})
	if err != nil {
		t.Fatal(err)
	}
	if len(stack.Skills) < 3 {
		t.Fatalf("skills = %d, want >= 3", len(stack.Skills))
	}
	if !containsSkill(stack.Skills, "docsbot-go-testing") && !containsSkill(stack.Skills, "docsbot-code-review") {
		t.Fatal("expected a verification skill")
	}
}

func containsSkill(skills []docsbot.Skill, id string) bool {
	for _, skill := range skills {
		if skill.ID == id {
			return true
		}
	}
	return false
}
