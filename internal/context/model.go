package context

type Item struct {
	Source   string
	Content  string
	Priority float64
}

type Builder struct{}

func (Builder) Select(items []Item, limit int) []Item {
	if limit >= len(items) {
		return items
	}
	return items[:limit]
}
