package geminiweb

import (
	"testing"

	"github.com/ti-system/ti-internal-cli/internal/provider/webtools"
)

type fakeTransport struct {
	response string
	prompt   []webtools.Message
}

func (f *fakeTransport) Send(messages []webtools.Message) (string, error) {
	f.prompt = messages
	return f.response, nil
}

func TestAdapterTranslatesToolsToOpenAICompatibleCalls(t *testing.T) {
	transport := &fakeTransport{response: `<tool_call name="filesystem.read">{"path":"main.go"}</tool_call>`}
	adapter := NewAdapter(transport, webtools.NewEmulator(webtools.Config{MaxArgumentBytes: 1024}))
	response, err := adapter.Execute(Request{
		Messages: []webtools.Message{{Role: "user", Content: "read main.go"}},
		Tools:    []webtools.ToolDefinition{{Name: "filesystem.read", Description: "read file"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(transport.prompt) != 2 {
		t.Fatalf("prepared messages = %d, want 2", len(transport.prompt))
	}
	if len(response.ToolCalls) != 1 || response.ToolCalls[0].Name != "filesystem.read" {
		t.Fatalf("response = %#v", response)
	}
}

func TestAdapterPreservesNormalChatWithoutTools(t *testing.T) {
	transport := &fakeTransport{response: "normal response"}
	adapter := NewAdapter(transport, webtools.NewEmulator(webtools.Config{MaxArgumentBytes: 1024}))
	response, err := adapter.Execute(Request{Messages: []webtools.Message{{Role: "user", Content: "hello"}}})
	if err != nil {
		t.Fatal(err)
	}
	if response.Content != "normal response" || len(response.ToolCalls) != 0 {
		t.Fatalf("response = %#v", response)
	}
}
