# Technical Decision

## Status

Accepted and measured.

## Selected Stack

- Go 1.26 for low-overhead servers, clients, and one static runtime image
- chi for minimal REST routing on standard `net/http`
- official gRPC-Go and Protobuf implementations
- Docker Compose for real multi-process orchestration
- PowerShell for the Windows-first publication harness

## API Decision

Expose the same `echo.v1` behavior through REST and gRPC because protocol comparison is the product. GraphQL is rejected here: field selection and resolver execution would introduce a different query model rather than a transport-only comparison.

## Messaging, Data, and Cloud

No broker, database, cache, or cloud provider is used. The request is synchronous, stateless, and has no durability requirement. Kumo is therefore not started; adding AWS semantics would be unrelated to the claim.

## Library Policy

Use standard-library capabilities unless the protocol requires an official implementation. Domain/use-case code imports no router, HTTP, gRPC, generated Protobuf, Docker, or cloud package.

## Rejected Stack Options

| Option | Reason |
|---|---|
| Python/FastAPI | runtime and serialization costs would change the protocol-focused Go question |
| Spring/Kotlin | JVM warmup would require a different experimental design |
| Fiber | a non-standard HTTP API adds framework behavior without needed features |
| Connect | the experiment explicitly compares native gRPC framing with REST/JSON |
| Kafka or RabbitMQ | asynchronous delivery is outside a unary request/response protocol comparison |

## Observed Impact

The original expectation that gRPC would automatically win was rejected by evidence. REST produced lower median p95 and higher median throughput; gRPC produced lower median p99. The technical conclusion is to benchmark the representative workload rather than select a protocol from reputation.

Validation command:

```powershell
pwsh ./tools/validate-project.ps1
```
