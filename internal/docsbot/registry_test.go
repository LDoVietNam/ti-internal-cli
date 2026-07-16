package docsbot

import "testing"

func TestRegistryRequiresGenericFallback(t *testing.T) {
	_, err := NewRegistry(nil, "docsbot-programming-general")
	if err == nil {
		t.Fatal("expected missing fallback error")
	}

	fallback := Skill{ID: "docsbot-programming-general", Source: "docsbot", Generic: true}
	registry, err := NewRegistry([]Skill{fallback}, fallback.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := registry.Get(fallback.ID); !ok || got.ID != fallback.ID {
		t.Fatalf("fallback lookup = %#v, %v", got, ok)
	}
}

func TestRegistryRejectsNonDocsBotSkills(t *testing.T) {
	_, err := NewRegistry([]Skill{{ID: "internal-go", Source: "internal", Generic: true}}, "internal-go")
	if err == nil {
		t.Fatal("expected non-docsbot source error")
	}
}
