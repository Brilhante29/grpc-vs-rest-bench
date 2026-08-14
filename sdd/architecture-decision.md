# Architecture Decision

## Status

Accepted and measured.

## Problem Forces

- low domain complexity: one Echo use case
- high comparison integrity: both protocols must execute identical logic
- medium throughput pressure: measurement must avoid framework-heavy noise
- no persistence, messaging, cloud, UI state, or independent capability deployment

## Decision

Use a modular monolith with three binaries (`rest-server`, `grpc-server`, and `bench-client`) and one transport-independent `internal` package.

The application contract is `Echoer`. REST and gRPC are adapters that translate wire data to domain-owned request/response structures. Composition roots under `cmd/` select concrete adapters; dependencies point inward.

## Principles

- SRP: each binary owns one process concern; the benchmark package owns measurement/evidence.
- OCP: another transport can implement the same contract without changing `EchoLogic`.
- LSP: each client is substitutable under the benchmark `Client` interface and must satisfy semantic parity.
- ISP: `Echoer` and benchmark `Client` expose one operation only.
- DIP: transport packages depend on the use-case contract, not the reverse.
- KISS/DRY/YAGNI: one use case, one implementation, no unrelated infrastructure.

## Rejected Alternatives

| Alternative | Reason |
|---|---|
| microservices | independent deployment would add network and orchestration variables without a second business capability |
| full clean-architecture rings | extra entity/repository/use-case layers would not protect any additional policy in this single-operation domain |
| one handler implementation per protocol | duplicate business logic would invalidate semantic and performance comparison |
| in-process protocol mocks | they would not measure real HTTP/1.1, HTTP/2, JSON, and Protobuf paths |

## Consequences

The benchmark isolates transport cost while retaining real sockets and serialization. Results still apply only to this unary, same-host workload. Persistence or streaming would introduce different forces and require a new architecture decision and comparability key.
