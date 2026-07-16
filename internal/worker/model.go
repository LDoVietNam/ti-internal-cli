package worker

type Result struct {
	Success bool
	Output  string
}

type Worker interface {
	ID() string
	Capabilities() []string
	Execute(string) Result
}
