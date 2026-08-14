# Benchmark Plan: grpc-vs-rest-bench

## Question

How do REST/HTTP 1.1 + JSON and gRPC/HTTP 2 + Protobuf compare for one identical unary Echo workload on a local Docker host?

This is a measurement question, not a prediction that one protocol always wins.

## Canonical Command

```powershell
pwsh ./tools/run-benchmark.ps1 -Requests 1000 -PayloadBytes 256 -Concurrency 10 -WarmupRequests 100 -Repetitions 3
```

The runner refuses a dirty worktree, builds an image tagged from the exact source commit, resolves the image digest, runs the real REST and gRPC servers, validates the report, and tears down only this Compose project.

## Controlled Inputs

- contract: `echo.v1`
- payload: deterministic 256-byte ASCII string, recorded by SHA-256
- measured requests: 1,000 per protocol and repetition
- warmup: 100 requests per protocol
- repetitions: 3
- concurrency: 10
- execution order: alternating first protocol per repetition
- REST connection model: HTTP keep-alive pool
- gRPC connection model: one multiplexed channel

## Metrics

| Metric family | Unit | Interpretation |
|---|---|---|
| `{rest,grpc}_p{50,95,99}_latency_ms` | milliseconds | lower is better within the controlled workload |
| `{rest,grpc}_throughput_rps` | requests/second | higher is better within the controlled workload |
| `rest_over_grpc_p95_ratio` | ratio | paired p95 comparison |
| `grpc_over_rest_throughput_ratio` | ratio | paired throughput comparison |
| `request_failures` | count | publication target is zero |

Each value is the median of three repetition-level samples. The raw samples and min/median/max/mean summaries remain in the JSON.

## Observed Baseline

- REST p95: `3.238946 ms`
- gRPC p95: `3.796319 ms`
- REST throughput: `7,752.311 req/s`
- gRPC throughput: `6,036.664 req/s`
- failures: `0`

REST won median p50, p95, p99, and throughput in this small unary workload. One of three paired repetitions favored gRPC for p95 and throughput, so the post must explain variance and workload limits instead of claiming a universal winner.

## Result Contract

The artifact conforms to `.portfolio/contracts/benchmark-result-v2.schema.json`: integer schema version 2, UUID, workload digests, repetition samples, execution metadata, environment, exact provenance, and comparability key.

CI writes smoke evidence to the runner temporary directory so it never replaces the committed publication baseline.
