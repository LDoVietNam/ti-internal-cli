package task

import "time"

type State string

const (
	Created   State = "created"
	Profiled  State = "profiled"
	Planned   State = "planned"
	Executing State = "executing"
	Verifying State = "verifying"
	Completed State = "completed"
	Failed    State = "failed"
)

type Task struct {
	ID        string
	Type      string
	Workspace string
	Workflow  string
	State     State
	CreatedAt time.Time
}
