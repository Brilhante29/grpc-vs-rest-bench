# Spec: grpc-vs-rest-bench

## Identity

- portfolio number: 15
- macro project: Backend Reliability and Architecture Platform
- claim: compare REST and gRPC with the same use case and publish reproducible latency, throughput, and parity evidence
- stack: Go, chi, gRPC, Protobuf, Docker Compose

## Functional Scope

- expose `echo.v1` through REST/JSON and gRPC/Protobuf
- keep `EchoLogic` independent from both transports
- reject a missing request ID consistently
- verify identical request ID, payload, and SHA-256 semantics
- warm both protocols and alternate measured order
- emit the common benchmark-result V2 artifact

## Out of Scope

- streaming, TLS, authentication, persistence, service mesh, browser UI, and cross-region transport
- a universal REST-versus-gRPC recommendation
- paid credentials or managed cloud dependencies

## Public Evidence

- primary metric: `rest_over_grpc_p95_ratio`
- baseline: `0.739759`
- secondary evidence: REST/gRPC p50, p95, p99, throughput, and zero failures
- command: `pwsh ./tools/run-benchmark.ps1`
- artifact: `benchmarks/results/benchmark-result.json`

## Definition of Done

- [x] Both real transports implement one shared domain contract.
- [x] Unit, integration, race, vet, and build checks exist.
- [x] Docker runs without paid credentials.
- [x] Three-repetition V2 evidence contains exact provenance.
- [x] README opens with the measured number and limitations.
- [x] CI smoke evidence cannot overwrite the canonical result.
- [x] Reuse findings are recorded in the kit loop.
