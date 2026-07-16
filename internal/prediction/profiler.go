package prediction

import "strings"

type Profiler struct{}

func NewProfiler() Profiler { return Profiler{} }

func (Profiler) Profile(input string) TaskProfile {
	text := strings.ToLower(strings.TrimSpace(input))
	profile := TaskProfile{Type: TaskUnknown}

	switch {
	case containsAny(text, "fix", "bug", "broken", "timeout", "error", "failure", "crash"):
		profile.Type = TaskBugFix
		profile.Intent = []string{"diagnose", "repair", "verify"}
	case containsAny(text, "implement", "add", "create", "build", "feature"):
		profile.Type = TaskFeature
		profile.Intent = []string{"design", "implement", "verify"}
	case containsAny(text, "refactor", "restructure", "cleanup", "simplify"):
		profile.Type = TaskRefactor
		profile.Intent = []string{"analyze_impact", "transform", "verify_compatibility"}
	case containsAny(text, "review", "pull request", "diff"):
		profile.Type = TaskCodeReview
		profile.Intent = []string{"inspect_diff", "identify_risk", "report_findings"}
	case containsAny(text, "investigate", "analyze", "research", "diagnose"):
		profile.Type = TaskInvestigation
		profile.Intent = []string{"collect_evidence", "test_hypotheses", "report"}
	}

	profile.Language = detectLanguage(text)
	profile.Framework = detectFramework(text)
	profile.RepositoryTags = detectTags(text)
	return profile
}

func containsAny(text string, values ...string) bool {
	for _, value := range values {
		if strings.Contains(text, value) {
			return true
		}
	}
	return false
}

func detectLanguage(text string) string {
	switch {
	case containsAny(text, "golang", " go ", "go code", "go module"):
		return "go"
	case containsAny(text, "typescript", ".ts", " ts "):
		return "typescript"
	case containsAny(text, "javascript", ".js", " node "):
		return "javascript"
	case containsAny(text, "python", ".py"):
		return "python"
	case containsAny(text, "rust", ".rs"):
		return "rust"
	default:
		return ""
	}
}

func detectFramework(text string) string {
	frameworks := []string{"net/http", "gin", "echo", "fiber", "react", "next.js", "vue", "django", "fastapi"}
	for _, framework := range frameworks {
		if strings.Contains(text, framework) {
			return framework
		}
	}
	return ""
}

func detectTags(text string) []string {
	known := []string{"router", "backend", "frontend", "provider", "mcp", "memory", "workflow", "cli"}
	var tags []string
	for _, tag := range known {
		if strings.Contains(text, tag) {
			tags = append(tags, tag)
		}
	}
	return tags
}
