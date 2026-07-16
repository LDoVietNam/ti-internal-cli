package agent

type Phase string

const (
	Observe Phase = "observe"
	Plan    Phase = "plan"
	Act     Phase = "act"
	Verify  Phase = "verify"
)

type Runtime struct{}

func (Runtime) Start() Phase {
	return Observe
}
