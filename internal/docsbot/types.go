package docsbot

type Instructions struct {
	Objective    []string
	Execution    []string
	Verification []string
}

type Skill struct {
	ID                string
	Version           int
	Source            string
	SourceURL         string
	ContentHash       string
	Generic           bool
	Domains           []string
	TaskTypes         []string
	Languages         []string
	Frameworks        []string
	Phases            []string
	RepositoryTags    []string
	Priority          float64
	HistoricalSuccess float64
	UserPreference    float64
	Instructions      Instructions
}
