package capability

import "testing"

func TestRegistryRejectsUnsupportedToolRequests(t *testing.T) {
	registry := NewRegistry([]Capabilities{{ProviderID: "plain-web", Tools: SupportNone}})
	err := registry.Validate("plain-web", RequestFeatures{NeedsTools: true})
	if err == nil {
		t.Fatal("expected unsupported tools error")
	}
}

func TestRegistryAllowsEmulatedTools(t *testing.T) {
	registry := NewRegistry([]Capabilities{{ProviderID: "gemini-web", Tools: SupportEmulated}})
	if err := registry.Validate("gemini-web", RequestFeatures{NeedsTools: true}); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestRegistryReturnsImmutableSnapshot(t *testing.T) {
	registry := NewRegistry([]Capabilities{{ProviderID: "gemini-web", Tools: SupportEmulated, Notes: []string{"prompt shim"}}})
	caps, ok := registry.Get("gemini-web")
	if !ok {
		t.Fatal("provider not found")
	}
	caps.Notes[0] = "mutated"
	again, _ := registry.Get("gemini-web")
	if again.Notes[0] != "prompt shim" {
		t.Fatalf("registry snapshot mutated: %q", again.Notes[0])
	}
}
