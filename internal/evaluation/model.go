package evaluation

type Result struct {
	Success      bool
	TestsPassed  bool
	HumanRewrite bool
}

func Reward(r Result) float64 {
	score := 0.0
	if r.Success {
		score += 0.4
	}
	if r.TestsPassed {
		score += 0.4
	}
	if !r.HumanRewrite {
		score += 0.2
	}
	return score
}
