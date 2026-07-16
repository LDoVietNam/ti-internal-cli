package webtools

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"regexp"
	"sort"
	"strings"
)

type Message struct {
	Role    string
	Content string
}

type ToolDefinition struct {
	Name        string
	Description string
	Schema      json.RawMessage
}

type ToolCall struct {
	ID        string
	Name      string
	Arguments json.RawMessage
}

type ParsedResponse struct {
	Text      string
	ToolCalls []ToolCall
}

type Config struct {
	MaxArgumentBytes int
}

type Emulator struct {
	config Config
}

var toolCallPattern = regexp.MustCompile(`(?s)<tool_call\s+name="([^"]+)">\s*(.*?)\s*</tool_call>`)

func NewEmulator(config Config) *Emulator {
	if config.MaxArgumentBytes <= 0 {
		config.MaxArgumentBytes = 64 * 1024
	}
	return &Emulator{config: config}
}

func (e *Emulator) Prepare(messages []Message, tools []ToolDefinition) ([]Message, error) {
	if len(tools) == 0 {
		return append([]Message(nil), messages...), nil
	}
	seen := map[string]bool{}
	sorted := append([]ToolDefinition(nil), tools...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
	var builder strings.Builder
	builder.WriteString("Use only the tools listed below. Emit exact XML tool calls and no invented tool names.\n")
	for _, tool := range sorted {
		name := strings.TrimSpace(tool.Name)
		if name == "" || seen[name] {
			return nil, errors.New("tool names must be non-empty and unique")
		}
		seen[name] = true
		builder.WriteString("<tool name=\"")
		builder.WriteString(html.EscapeString(name))
		builder.WriteString("\" description=\"")
		builder.WriteString(html.EscapeString(tool.Description))
		builder.WriteString("\" />\n")
	}
	prepared := make([]Message, 0, len(messages)+1)
	prepared = append(prepared, Message{Role: "system", Content: builder.String()})
	prepared = append(prepared, messages...)
	return prepared, nil
}

func (e *Emulator) Parse(content string, allowed []ToolDefinition) (ParsedResponse, error) {
	allowedNames := make(map[string]bool, len(allowed))
	for _, tool := range allowed {
		allowedNames[tool.Name] = true
	}
	matches := toolCallPattern.FindAllStringSubmatchIndex(content, -1)
	calls := make([]ToolCall, 0, len(matches))
	var text strings.Builder
	cursor := 0
	for index, match := range matches {
		text.WriteString(content[cursor:match[0]])
		name := content[match[2]:match[3]]
		arguments := strings.TrimSpace(content[match[4]:match[5]])
		if !allowedNames[name] {
			return ParsedResponse{}, fmt.Errorf("unknown tool %q", name)
		}
		if len(arguments) > e.config.MaxArgumentBytes {
			return ParsedResponse{}, fmt.Errorf("tool arguments exceed %d bytes", e.config.MaxArgumentBytes)
		}
		var value any
		if err := json.Unmarshal([]byte(arguments), &value); err != nil {
			return ParsedResponse{}, fmt.Errorf("tool %s arguments: %w", name, err)
		}
		if _, ok := value.(map[string]any); !ok {
			return ParsedResponse{}, fmt.Errorf("tool %s arguments must be a JSON object", name)
		}
		calls = append(calls, ToolCall{ID: fmt.Sprintf("call_%d", index+1), Name: name, Arguments: json.RawMessage(arguments)})
		cursor = match[1]
	}
	text.WriteString(content[cursor:])
	return ParsedResponse{Text: strings.TrimSpace(text.String()), ToolCalls: calls}, nil
}
