# Architecture Decision

## Status

Proposed

## Context

Project: `<project-name>`
Claim: `<measurable claim>`
Benchmark: `<primary metric>`

Problem forces:

- Domain complexity: `<low|medium|high>`
- Integration pressure: `<low|medium|high>`
- UI state complexity: `<low|medium|high|none>`
- Data/ML reproducibility: `<low|medium|high>`
- Auditability/event history: `<low|medium|high>`
- Throughput/async pressure: `<low|medium|high>`
- Independent deployability need: `<low|medium|high>`

## Decision

Chosen architecture: `<style>`

Reason:

`<Explain why this architecture fits the actual problem and benchmark.>`

Dependency rule:

`<Example: domain/application do not depend on infra; adapters depend inward through ports.>`

## Rejected Alternatives

| Alternative | Why rejected |
|---|---|
| `<style>` | `<reason>` |
| `<style>` | `<reason>` |

## Folder Layout

```txt
src/
  <folders>
test/
benchmarks/
```

## Testing Strategy

- Unit tests: `<what is isolated>`
- Integration tests: `<what is wired>`
- Benchmark: `<what proves the claim>`

## Consequences

Positive:

- `<benefit>`

Tradeoffs:

- `<cost>`

Migration path:

- `<how to evolve if the problem grows>`
