package memory

import "time"

type Type string

const (
	Working  Type = "working"
	Session  Type = "session"
	Project  Type = "project"
	Decision Type = "decision"
	Skill    Type = "skill"
)

type Memory struct {
	ID         string
	Type       Type
	Scope      string
	Content    string
	Importance float64
	Confidence float64
	CreatedAt  time.Time
}
