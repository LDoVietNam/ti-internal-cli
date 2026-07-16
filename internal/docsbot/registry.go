package docsbot

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

var ErrFallbackRequired = errors.New("docsbot fallback skill is required")

type Registry struct {
	mu         sync.RWMutex
	skills     map[string]Skill
	fallbackID string
}

func NewRegistry(skills []Skill, fallbackID string) (*Registry, error) {
	registry := &Registry{skills: make(map[string]Skill, len(skills)), fallbackID: fallbackID}
	for _, skill := range skills {
		if skill.ID == "" {
			return nil, errors.New("docsbot skill id is required")
		}
		if skill.Source != "docsbot" {
			return nil, fmt.Errorf("skill %s has unsupported source %q", skill.ID, skill.Source)
		}
		if _, exists := registry.skills[skill.ID]; exists {
			return nil, fmt.Errorf("duplicate docsbot skill %s", skill.ID)
		}
		registry.skills[skill.ID] = skill
	}
	fallback, ok := registry.skills[fallbackID]
	if fallbackID == "" || !ok || !fallback.Generic {
		return nil, ErrFallbackRequired
	}
	return registry, nil
}

func (r *Registry) Get(id string) (Skill, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	skill, ok := r.skills[id]
	return skill, ok
}

func (r *Registry) Fallback() Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.skills[r.fallbackID]
}

func (r *Registry) All() []Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Skill, 0, len(r.skills))
	for _, skill := range r.skills {
		out = append(out, skill)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
