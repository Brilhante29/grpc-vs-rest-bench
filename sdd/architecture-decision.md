# Architecture Decision

## Status

Accepted

## Context

Project: grpc-vs-rest-bench
Claim: comparacao REST vs gRPC
Benchmark: latency_ms_by_protocol

Problem forces:

- Domain complexity: low
- Integration pressure: low
- UI state complexity: none
- Data/ML reproducibility: low
- Auditability/event history: low
- Throughput/async pressure: medium
- Independent deployability need: low

## Decision

Chosen architecture: modular-monolith

Reason:

The benchmark compares two transport protocols (REST/HTTP vs gRPC) executing identical business logic. A modular monolith with two server binaries sharing the same internal package is the simplest architecture that proves the claim. Splitting into microservices would add orchestration overhead without benchmark benefit.

Dependency rule:

cmd/ depends on internal/ — internal/ never imports cmd/.

## Rejected Alternatives

| Alternative | Why rejected |
|---|---|
| microservices | Added deploy and orchestration complexity irrelevant to transport comparison |
| hexagonal/ports-adapters | Overkill for a single Echo RPC; the adapter layer is the cmd/ boundary itself |

## Folder Layout

```
cmd/
  grpc-server/      -- gRPC transport adapter
  rest-server/      -- REST/HTTP transport adapter
  bench-client/     -- benchmark harness entrypoint
internal/
  proto/benchmark/v1/  -- generated protobuf types
  service.go           -- shared EchoLogic
  benchmark/
    suite.go           -- benchmark execution harness
    report.go          -- JSON result output
proto/benchmark/v1/    -- proto source definition
```

## Testing Strategy

- Unit tests: EchoLogic (pure Go), Report serialization
- Integration tests: gRPC and REST server start + request/response verify
- Benchmark: bench-client sends N requests to both, outputs JSON

## Consequences

Positive:

- Single Docker image with all three entrypoints
- Shared logic guaranteed identical between transports
- Simple to understand and reproduce

Tradeoffs:

- Two server processes need coordination in benchmark (solved by starting both in bench-client orchestration)

Migration path:

If the benchmark needs database-backed workloads, extract server binaries into independent services.
