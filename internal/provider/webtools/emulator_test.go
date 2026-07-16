package webtools

import (
	"strings"
	"testing"
)

func TestPrepareInjectsAllowedToolContracts(t *testing.T) {
	emulator := NewEmulator(Config{MaxArgumentBytes: 1024})
	messages, err := emulator.Prepare([]Message{{Role: "user", Content: "inspect repository"}}, []ToolDefinition{{Name: "filesystem.read", Description: "read a file"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || !strings.Contains(messages[0].Content, "filesystem.read") {
		t.Fatalf("tool contract not injected: %#v", messages)
	}
}

func TestParseReturnsMultipleCallsAndPreservesText(t *testing.T) {
	emulator := NewEmulator(Config{MaxArgumentBytes: 1024})
	input := "I will inspect files.\n<tool_call name=\"filesystem.read\">{\"path\":\"a.go\"}</tool_call>\n<tool_call name=\"test.run\">{\"command\":\"go test ./...\"}</tool_call>\nDone."
	result, err := emulator.Parse(input, []ToolDefinition{{Name: "filesystem.read"}, {Name: "test.run"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.ToolCalls) != 2 {
		t.Fatalf("tool calls = %d, want 2", len(result.ToolCalls))
	}
	if !strings.Contains(result.Text, "I will inspect files.") || !strings.Contains(result.Text, "Done.") {
		t.Fatalf("text not preserved: %q", result.Text)
	}
}

func TestParseRejectsUnknownTool(t *testing.T) {
	emulator := NewEmulator(Config{MaxArgumentBytes: 1024})
	_, err := emulator.Parse(`<tool_call name="shell.delete">{"path":"/"}</tool_call>`, []ToolDefinition{{Name: "filesystem.read"}})
	if err == nil {
		t.Fatal("expected unknown tool error")
	}
}

func TestParseRejectsMalformedArguments(t *testing.T) {
	emulator := NewEmulator(Config{MaxArgumentBytes: 1024})
	_, err := emulator.Parse(`<tool_call name="filesystem.read">{bad}</tool_call>`, []ToolDefinition{{Name: "filesystem.read"}})
	if err == nil {
		t.Fatal("expected malformed arguments error")
	}
}

func TestParseDoesNotTreatIncidentalWordsAsToolCalls(t *testing.T) {
	emulator := NewEmulator(Config{MaxArgumentBytes: 1024})
	result, err := emulator.Parse("fetch latest docs", []ToolDefinition{{Name: "test.run"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.ToolCalls) != 0 || result.Text != "fetch latest docs" {
		t.Fatalf("unexpected parse result: %#v", result)
	}
}
