package geminiweb

import (
	"errors"

	"github.com/ti-system/ti-internal-cli/internal/provider/webtools"
)

type Transport interface {
	Send(messages []webtools.Message) (string, error)
}

type Request struct {
	Messages []webtools.Message
	Tools    []webtools.ToolDefinition
}

type Response struct {
	Content   string
	ToolCalls []webtools.ToolCall
}

type Adapter struct {
	transport Transport
	emulator  *webtools.Emulator
}

func NewAdapter(transport Transport, emulator *webtools.Emulator) *Adapter {
	return &Adapter{transport: transport, emulator: emulator}
}

func (a *Adapter) Execute(request Request) (Response, error) {
	if a.transport == nil {
		return Response{}, errors.New("gemini web transport is required")
	}
	if a.emulator == nil {
		return Response{}, errors.New("web tool emulator is required")
	}
	prepared, err := a.emulator.Prepare(request.Messages, request.Tools)
	if err != nil {
		return Response{}, err
	}
	content, err := a.transport.Send(prepared)
	if err != nil {
		return Response{}, err
	}
	parsed, err := a.emulator.Parse(content, request.Tools)
	if err != nil {
		return Response{}, err
	}
	return Response{Content: parsed.Text, ToolCalls: parsed.ToolCalls}, nil
}
