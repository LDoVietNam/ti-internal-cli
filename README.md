# Ti Internal CLI v0.5.0 Agent Runtime

Functional architecture layer:

- Task Runtime
- Workflow Engine
- Memory System
- Context Engine contracts
- Agent Loop contracts
- Worker abstraction
- Prompt Intelligence integration boundary
- Evaluation loop

This release keeps providers and tools decoupled.

## Prediction Intelligence Phase 1

The deterministic prediction core selects an execution strategy, not only a prompt.

Implemented in `internal/prediction`:

- task profiling for bug fixes, features, refactors, code reviews, and investigations
- hard capability, policy, availability, and source-trust filters
- explainable weighted ranking
- deterministic fallback ordering
- built-in workflow recommendations
- typed verification plans and stop conditions

The engine is intentionally independent of embeddings and external services. DocsBot and other public prompt sources will enter through a later normalized, policy-checked prompt supply chain.

```go
engine := prediction.NewEngine(prediction.DefaultWeights())
result, err := engine.Predict(request, candidates)
```

Verification:

```bash
go test ./...
go vet ./...
go build ./...
```
