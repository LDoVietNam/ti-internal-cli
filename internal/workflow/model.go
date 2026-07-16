package workflow

type Step struct {
	ID     string
	Action string
}

type Definition struct {
	Name  string
	Steps []Step
}

type Engine struct{}

func (Engine) Validate(w Definition) bool {
	return w.Name != "" && len(w.Steps) > 0
}
