package promptintel

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/ti-system/ti-internal-cli/internal/docsbot"
)

type CompilerConfig struct {
	TokenBudget int
}

type CompileRequest struct {
	RuntimePolicy       string
	ProjectInstructions string
	Task                string
	Stack               SkillStack
	ToolContracts       map[string]string
}

type Section struct {
	Name    string
	Content string
}

type Provenance struct {
	SkillID   string
	Source     string
	SourceURL  string
	ContentHash string
}

type CompiledPrompt struct {
	Sections   []Section
	Provenance []Provenance
	Text       string
}

type Compiler struct {
	config CompilerConfig
}

func NewCompiler(config CompilerConfig) *Compiler {
	if config.TokenBudget <= 0 {
		config.TokenBudget = 2048
	}
	return &Compiler{config: config}
}

var pseudoToolPattern = regexp.MustCompile(`\[TOOL_EXECUTE:\s*([^\]]+)\]`)

func (c *Compiler) Compile(request CompileRequest) (CompiledPrompt, error) {
	if len(request.Stack.Skills) == 0 {
		return CompiledPrompt{}, errors.New("at least one DocsBot skill is required")
	}

	guidance := make([]string, 0)
	provenance := make([]Provenance, 0, len(request.Stack.Skills))
	seen := map[string]bool{}
	for _, skill := range request.Stack.Skills {
		if skill.Source != "docsbot" {
			return CompiledPrompt{}, fmt.Errorf("skill %s is not sourced from DocsBot", skill.ID)
		}
		for _, instruction := range flatten(skill.Instructions) {
			compiled, err := compilePseudoTools(instruction, request.ToolContracts)
			if err != nil {
				return CompiledPrompt{}, fmt.Errorf("skill %s: %w", skill.ID, err)
			}
			key := strings.ToLower(strings.TrimSpace(compiled))
			if key == "" || seen[key] {
				continue
			}
			seen[key] = true
			guidance = append(guidance, compiled)
		}
		provenance = append(provenance, Provenance{
			SkillID: skill.ID, Source: skill.Source, SourceURL: skill.SourceURL, ContentHash: skill.ContentHash,
		})
	}

	sort.SliceStable(provenance, func(i, j int) bool { return provenance[i].SkillID < provenance[j].SkillID })
	sections := []Section{
		{Name: "runtime_policy", Content: request.RuntimePolicy},
		{Name: "project_instructions", Content: request.ProjectInstructions},
		{Name: "task_requirements", Content: request.Task},
		{Name: "docsbot_skills", Content: strings.Join(guidance, "\n")},
	}
	text := renderSections(sections)
	text = truncateByApproxTokens(text, c.config.TokenBudget)
	return CompiledPrompt{Sections: sections, Provenance: provenance, Text: text}, nil
}

func flatten(instructions docsbot.Instructions) []string {
	out := make([]string, 0, len(instructions.Objective)+len(instructions.Execution)+len(instructions.Verification))
	out = append(out, instructions.Objective...)
	out = append(out, instructions.Execution...)
	out = append(out, instructions.Verification...)
	return out
}

func compilePseudoTools(instruction string, contracts map[string]string) (string, error) {
	matches := pseudoToolPattern.FindAllStringSubmatch(instruction, -1)
	compiled := instruction
	for _, match := range matches {
		request := strings.ToLower(strings.TrimSpace(match[1]))
		capability := inferCapability(request)
		contract, ok := contracts[capability]
		if !ok {
			return "", fmt.Errorf("unknown pseudo-tool request %q", match[1])
		}
		compiled = strings.ReplaceAll(compiled, match[0], contract)
	}
	return compiled, nil
}

func inferCapability(request string) string {
	switch {
	case strings.Contains(request, "test"):
		return "test"
	case strings.Contains(request, "build"):
		return "build"
	case strings.Contains(request, "search"), strings.Contains(request, "grep"):
		return "search"
	case strings.Contains(request, "read"), strings.Contains(request, "inspect"):
		return "read"
	case strings.Contains(request, "patch"), strings.Contains(request, "edit"):
		return "edit"
	default:
		return ""
	}
}

func renderSections(sections []Section) string {
	var builder strings.Builder
	for _, section := range sections {
		if strings.TrimSpace(section.Content) == "" {
			continue
		}
		builder.WriteString("[")
		builder.WriteString(strings.ToUpper(section.Name))
		builder.WriteString("]\n")
		builder.WriteString(strings.TrimSpace(section.Content))
		builder.WriteString("\n\n")
	}
	return strings.TrimSpace(builder.String())
}

func truncateByApproxTokens(text string, budget int) string {
	if budget <= 0 {
		return ""
	}
	maxRunes := budget * 4
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}
	return string(runes[:maxRunes])
}
