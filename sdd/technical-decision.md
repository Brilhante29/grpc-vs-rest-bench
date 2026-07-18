# Technical Decision

## Status

Proposed

## Decision Type

`<stack|api-style|cloud|messaging|database|library|runtime|framework>`

## Context

Project: `<project-name>`
Problem: `<problem to solve>`
Portfolio program: `<program>`
Public signal: `<GitHub/LinkedIn proficiency signal>`
Benchmark: `<metric>`

## Selected Option

Selected: `<option>`

Reason:

`<Why this option fits the problem, benchmark, and public signal.>`

## Decision Brain Fields

- Stack profile: `<spring-kotlin-backend|fastapi-backend|go-backend|node-typescript-backend|angular|nextjs|python-ml|terraform>`
- API style: `<rest-http|graphql|grpc|websocket|sse|cli>`
- Messaging: `<none|outbox-only|rabbitmq|kafka|redis-streams|nats>`
- Cloud mode: `<none|kumo-local-first|adapter-fake|real-cloud-required>`
- Database/runtime: `<selection>`
- Library policy: `<selection>`

## Engineering Principles

Coupling boundary:

`<Domain/use cases must not depend on framework, DB, broker, cloud SDK, transport, or UI.>`

SOLID application:

- SRP: `<how responsibilities are split>`
- OCP: `<how behavior extends without rewriting stable policy>`
- LSP: `<how adapters/fakes/reals stay substitutable>`
- ISP: `<small ports/interfaces used>`
- DIP: `<high-level policy depends on abstractions>`

Simplicity:

- KISS: `<simplest design that proves the claim>`
- YAGNI: `<future abstraction intentionally not added>`
- DRY: `<duplicated business knowledge removed without premature abstraction>`

Testability evidence:

- `<use case test without transport/infrastructure>`
- `<adapter or contract test>`
## Rejected Options

| Option | Why rejected |
|---|---|
| `<option>` | `<reason>` |
| `<option>` | `<reason>` |

## API Contract

Contract artifact:

`<OpenAPI|GraphQL schema|protobuf|event contract|CLI output schema|none>`

GraphQL controls, when applicable:

- Query complexity/depth limit: `<yes|no|not applicable>`
- N+1 prevention: `<DataLoader/batching plan|not applicable>`
- Field-level auth rule: `<yes|no|not applicable>`

## Cloud Local-First

Local provider:

`<kumo|none|adapter fake>`

Real provider target:

`<aws|none|other>`

Config switch:

```txt
CLOUD_PROVIDER=<kumo|aws|none>
CLOUD_ENDPOINT=http://localhost:4566
```

Unsupported local behaviors:

- `<behavior or none>`

## Benchmark Impact

Expected impact:

- `<metric/result this decision should improve or clarify>`

Validation command:

```powershell
<command>
```

## Operational Cost

- Docker services added: `<none|kumo|postgres|redis|rabbitmq|redpanda|...>`
- Local demo complexity: `<low|medium|high>`
- Failure case required: `<yes|no>`

## Follow-up

- `<what must be revisited if benchmark fails>`
