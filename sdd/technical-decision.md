# Technical Decision

## Status

Accepted

## Decision Type

stack, api-style, library

## Context

Project: grpc-vs-rest-bench
Problem: Compare latency of REST (HTTP/JSON) vs gRPC (HTTP/2 + Protobuf) for identical request/response workloads
Portfolio program: backend-reliability-platform
Public signal: Demonstrates Go gRPC/REST implementation and benchmark methodology
Benchmark: latency_ms_by_protocol

## Selected Option

Selected: Go + chi/grpc dual-server modular monolith

Reason:

Go compiles to static binaries, has native gRPC support via google.golang.org/grpc, and chi provides minimal-overhead HTTP routing. The dual-server pattern uses the same EchoLogic for both transports, isolating the protocol as the only variable.

## Decision Brain Fields

- Stack profile: go-backend
- API style: rest-http+grpc
- Messaging: none
- Cloud mode: none
- Database/runtime: none / docker-container
- Library policy: Minimal deps — chi for REST routing, grpc+protobuf for gRPC, stdlib for everything else

## Engineering Principles

Coupling boundary:

Domain/use cases must not depend on framework, DB, broker, cloud SDK, transport, or UI. EchoLogic is a pure Go struct with no imports from grpc, chi, or net/http.

SOLID application:

- SRP: cmd/ owns transport; internal/ owns domain; benchmark/ owns measurement
- OCP: Add a new transport by writing a new cmd/ binary; internal/ stays unchanged
- LSP: EchoLogic returns *pb.EchoResponse regardless of transport
- ISP: EchoLogic.Echo takes a single request type and returns a single response type
- DIP: cmd/ depends on internal/; no reverse dependency

Simplicity:

- KISS: One Echo RPC. No streaming, no auth, no DB.
- YAGNI: No middleware, no interceptors beyond what grpc/chi require, no circuit breakers.
- DRY: EchoLogic is the single source of truth for echo behavior.

Testability evidence:

- EchoLogic tested without gRPC or HTTP transport (internal/service_test.go)
- Report tested without running servers (internal/benchmark/report_test.go)

## Rejected Options

| Option | Why rejected |
|---|---|
| Python/FastAPI | Higher runtime overhead would confound benchmark results |
| Java/Spring | JVM warmup adds non-determinism to latency measurement |
| Rust | Excessive build complexity for a simple Echo benchmark |

## API Contract

Contract artifact: protobuf (proto/benchmark/v1/service.proto)

GraphQL controls, when applicable: N/A

## Cloud Local-First

Local provider: none

Real provider target: none

Config switch: N/A — no cloud dependencies

Unsupported local behaviors: N/A

## Benchmark Impact

Expected impact:

- gRPC should show 2-5x lower latency than REST for the same payload due to binary serialization and HTTP/2 multiplexing
- The speedup factor is the primary result

Validation command:

```bash
docker build -t grpc-vs-rest-bench . && docker run --rm grpc-vs-rest-bench
```

## Operational Cost

- Docker services added: none (single image)
- Local demo complexity: low (one docker command)
- Failure case required: no

## Follow-up

If gRPC is not faster, investigate: protobuf serialization overhead, HTTP connection pooling, or Go net/http vs grpc internals.
