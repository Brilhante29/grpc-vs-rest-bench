# #15 grpc-vs-rest-bench

**Measured result:** REST median p95 was `3.239 ms` and gRPC median p95 was `3.796 ms` for the same 256-byte unary Echo contract. Median throughput was `7,752.31 req/s` for REST and `6,036.66 req/s` for gRPC, with zero request or semantic-parity failures.

**Claim:** protocol choice must follow the workload. This repository compares REST/HTTP 1.1 + JSON and gRPC/HTTP 2 + Protobuf behind one transport-independent Go use case.

**Stack:** Go 1.26, chi, gRPC, Protobuf, Docker Compose, PowerShell benchmark harness.

## Run

```bash
docker build -t grpc-vs-rest-bench .
docker run --rm grpc-vs-rest-bench
```

The container starts both servers, warms both protocols, runs three measured repetitions, prints the V2 report, and exits. To persist a result with exact Git and image provenance:

```powershell
pwsh ./tools/run-benchmark.ps1
```

Output: `benchmarks/results/benchmark-result.json`.

## Benchmark V2

Workload: 1,000 requests per protocol per repetition, 100 warmup requests per protocol, three repetitions, concurrency 10, and a deterministic 256-byte payload. Protocol order alternates to reduce order bias.

| Metric | REST | gRPC |
|---|---:|---:|
| median p50 latency | 0.858 ms | 1.152 ms |
| median p95 latency | 3.239 ms | 3.796 ms |
| median p99 latency | 5.812 ms | 7.264 ms |
| median throughput | 7,752.31 req/s | 6,036.66 req/s |
| request/parity failures | 0 | 0 |

The paired median `rest_over_grpc_p95_ratio` was `0.740`; REST's paired p95 was about 26% lower in this run. One of three paired repetitions still favored gRPC for p95 and throughput, which is why the report preserves samples instead of reducing the result to a universal winner.

Evidence source: clean commit `7e91e873967ab098e3a5dd04e9a71eed6ade8050`, Go `1.26.5`, Linux/amd64, six Docker CPUs. The committed JSON includes all samples, workload/config digests, image digest, dependency-lock digest, executable digest, run UUID, and comparability key.

## Architecture

```text
benchmark client
  |-- REST adapter (HTTP/1.1 + JSON) ----|
  `-- gRPC adapter (HTTP/2 + Protobuf) --+--> Echoer port --> EchoLogic
```

`internal.Echoer` is the narrow application port. Both transport adapters map their wire formats to the same `EchoRequest` and `EchoResponse`; the use case imports neither chi, `net/http`, gRPC, nor generated Protobuf types.

- SRP: transport, use case, workload runner, and evidence writer have separate ownership.
- DIP/ISP: servers and benchmark clients depend on the `Echoer`/`Client` capabilities they use.
- LSP: REST and gRPC clients pass the same parity checks through the `Client` contract.
- KISS/YAGNI: no database, broker, auth, streaming, or cloud dependency obscures protocol cost.

## Interpretation

Choose REST when browser/tool interoperability, cache semantics, and operational simplicity dominate. Consider gRPC for typed service contracts, streaming, code generation, and multiplexed internal communication. Re-run this harness with representative payloads, TLS, network distance, and streaming before making a production decision.

## Limits

- One unary echo operation is not a universal protocol ranking.
- Both servers and the client run on one Docker host without TLS, proxies, persistence, or cross-region latency.
- JSON and Protobuf represent the same logical payload but produce different wire sizes.
- Compare artifacts only when `comparability_key` and host class are compatible.

## Verify

```powershell
pwsh ./tools/validate-project.ps1
```

The gate checks tests, vet, build, the committed common V2 artifact, Docker build/Compose configuration, SDD completeness, and forbidden legacy content. Sources and reuse attribution are in `REFERENCES.md`.
