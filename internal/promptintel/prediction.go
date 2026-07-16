package promptintel

import "github.com/ti-system/ti-internal-cli/internal/prediction"

func AttachToPrediction(value prediction.Prediction, stack SkillStack) prediction.Prediction {
	ids := make([]string, 0, len(stack.Skills))
	confidence := 0.0
	for _, selection := range stack.Selections {
		ids = append(ids, selection.Skill.ID)
		if selection.Score > confidence {
			confidence = selection.Score
		}
	}
	if len(ids) == 0 {
		for _, skill := range stack.Skills {
			ids = append(ids, skill.ID)
		}
	}
	value.DocsBotSkills = prediction.SkillPrediction{
		SkillIDs:   ids,
		Mode:       string(stack.Mode),
		Confidence: confidence,
		Fallback:   stack.Fallback,
	}
	return value
}
