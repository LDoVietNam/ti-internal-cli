package capability

import (
	"errors"
	"fmt"
	"sync"
)

type SupportLevel string

const (
	SupportNative   SupportLevel = "native"
	SupportEmulated SupportLevel = "emulated"
	SupportNone     SupportLevel = "none"
)

var ErrUnsupportedCapability = errors.New("unsupported provider capability")

type Capabilities struct {
	ProviderID  string
	Tools       SupportLevel
	Streaming   SupportLevel
	JSONMode    SupportLevel
	Attachments SupportLevel
	Notes       []string
}

type RequestFeatures struct {
	NeedsTools       bool
	NeedsStreaming   bool
	NeedsJSONMode    bool
	NeedsAttachments bool
}

type Registry struct {
	mu        sync.RWMutex
	providers map[string]Capabilities
}

func NewRegistry(values []Capabilities) *Registry {
	providers := make(map[string]Capabilities, len(values))
	for _, value := range values {
		providers[value.ProviderID] = clone(value)
	}
	return &Registry{providers: providers}
}

func (r *Registry) Get(providerID string) (Capabilities, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	value, ok := r.providers[providerID]
	return clone(value), ok
}

func (r *Registry) Validate(providerID string, request RequestFeatures) error {
	caps, ok := r.Get(providerID)
	if !ok {
		return fmt.Errorf("%w: unknown provider %s", ErrUnsupportedCapability, providerID)
	}
	checks := []struct {
		needed bool
		level  SupportLevel
		name   string
	}{
		{request.NeedsTools, caps.Tools, "tools"},
		{request.NeedsStreaming, caps.Streaming, "streaming"},
		{request.NeedsJSONMode, caps.JSONMode, "json_mode"},
		{request.NeedsAttachments, caps.Attachments, "attachments"},
	}
	for _, check := range checks {
		if check.needed && check.level == SupportNone {
			return fmt.Errorf("%w: provider %s does not support %s", ErrUnsupportedCapability, providerID, check.name)
		}
	}
	return nil
}

func clone(value Capabilities) Capabilities {
	value.Notes = append([]string(nil), value.Notes...)
	return value
}
