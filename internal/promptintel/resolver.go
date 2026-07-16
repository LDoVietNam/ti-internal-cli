package promptintel

import (
	"errors"
	"sort"
	"strings"

	"github.com/ti-system/ti-internal-cli/internal/docsbot"
)

type Mode string

const (
	ModeMandatory  Mode = "mandatory"
	ModeAggressive Mode = "aggressive"
)

type Config struct {
	Mode                     Mode
	MinimumSkillsPerTask     int
	IncludePhaseSkills       bool
	IncludeVerificationSkill bool
}

type Profile struct {
	Domain         string
	TaskType       string
	Language       string
	Framework      string
	Phase          string
	RepositoryTags []string
}

type SkillReason struct {
	Factor string
	Score  float64
}

type SkillSelection struct {
	Skill   docsbot.Skill
	Score   float64
	Reasons []SkillReason
}

type SkillStack struct {
	Mode       Mode
	Skills     []docsbot.Skill
	Selections []SkillSelection
	Fallback   bool
}

type Resolver struct {
	registry *docsbot.Registry
	config   Config
}

func NewResolver(registry *docsbot.Registry, config Config) *Resolver {
	if config.Mode == "" {
		config.Mode = ModeMandatory
	}
	if config.MinimumSkillsPerTask < 1 {
		config.MinimumSkillsPerTask = 1
	}
	return &Resolver{registry: registry, config: config}
}

func (r *Resolver) Resolve(profile Profile) (SkillStack, error) {
	if r.registry == nil {
		return SkillStack{}, errors.New("docsbot registry is required")
	}

	selections := make([]SkillSelection, 0)
	for _, skill := range r.registry.All() {
		if skill.Generic {
			continue
		}
		score, reasons := scoreSkill(profile, skill)
		if hasContextMatch(reasons) {
			selections = append(selections, SkillSelection{Skill: skill, Score: score, Reasons: reasons})
		}
	}
	sort.SliceStable(selections, func(i, j int) bool {
		if selections[i].Score == selections[j].Score {
			return selections[i].Skill.ID < selections[j].Skill.ID
		}
		return selections[i].Score > selections[j].Score
	})

	minimum := r.config.MinimumSkillsPerTask
	if r.config.Mode == ModeAggressive && minimum < 3 {
		minimum = 3
	}
	selected := make([]SkillSelection, 0, minimum)
	seen := map[string]bool{}
	add := func(selection SkillSelection) {
		if seen[selection.Skill.ID] {
			return
		}
		seen[selection.Skill.ID] = true
		selected = append(selected, selection)
	}

	for _, selection := range selections {
		if len(selected) >= minimum {
			break
		}
		add(selection)
	}

	if r.config.Mode == ModeAggressive && r.config.IncludeVerificationSkill {
		for _, selection := range selections {
			if containsFold(selection.Skill.Phases, "verify") {
				add(selection)
				break
			}
		}
	}

	fallback := false
	if len(selected) < minimum {
		fallback = true
		skill := r.registry.Fallback()
		add(SkillSelection{Skill: skill, Score: 0, Reasons: []SkillReason{{Factor: "generic_fallback", Score: 1}}})
	}
	if len(selected) == 0 {
		fallback = true
		skill := r.registry.Fallback()
		selected = append(selected, SkillSelection{Skill: skill, Reasons: []SkillReason{{Factor: "generic_fallback", Score: 1}}})
	}

	skills := make([]docsbot.Skill, 0, len(selected))
	for _, selection := range selected {
		skills = append(skills, selection.Skill)
	}
	return SkillStack{Mode: r.config.Mode, Skills: skills, Selections: selected, Fallback: fallback}, nil
}

func scoreSkill(profile Profile, skill docsbot.Skill) (float64, []SkillReason) {
	factors := []SkillReason{
		{Factor: "domain_match", Score: overlapOne(profile.Domain, skill.Domains)},
		{Factor: "task_type_match", Score: overlapOne(profile.TaskType, skill.TaskTypes)},
		{Factor: "language_match", Score: overlapOne(profile.Language, skill.Languages)},
		{Factor: "framework_match", Score: overlapOne(profile.Framework, skill.Frameworks)},
		{Factor: "phase_match", Score: overlapOne(profile.Phase, skill.Phases)},
		{Factor: "repository_match", Score: overlap(profile.RepositoryTags, skill.RepositoryTags)},
		{Factor: "historical_success", Score: clamp(skill.HistoricalSuccess)},
		{Factor: "user_preference", Score: clamp(skill.UserPreference)},
		{Factor: "priority", Score: clamp(skill.Priority / 100)},
	}
	weights := []float64{.22, .18, .14, .08, .16, .08, .07, .04, .03}
	total := 0.0
	for i := range factors {
		total += factors[i].Score * weights[i]
	}
	return total, factors
}

func hasContextMatch(reasons []SkillReason) bool {
	for _, reason := range reasons {
		switch reason.Factor {
		case "domain_match", "task_type_match", "language_match", "framework_match", "phase_match", "repository_match":
			if reason.Score > 0 {
				return true
			}
		}
	}
	return false
}

func overlapOne(value string, values []string) float64 {
	if value == "" || len(values) == 0 {
		return 0
	}
	for _, item := range values {
		if strings.EqualFold(value, item) {
			return 1
		}
	}
	return 0
}

func overlap(left, right []string) float64 {
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	matched := 0
	for _, item := range left {
		if containsFold(right, item) {
			matched++
		}
	}
	return float64(matched) / float64(len(left))
}

func containsFold(values []string, value string) bool {
	for _, item := range values {
		if strings.EqualFold(item, value) {
			return true
		}
	}
	return false
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
