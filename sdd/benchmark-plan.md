# Benchmark Plan: grpc-vs-rest-bench

## Hypothesis

gRPC (HTTP/2 + Protobuf) will show lower latency than REST (HTTP/1.1 + JSON) for identical Echo request/response workloads, measured by latency_ms_by_protocol. Expected speedup: 2-5x.

## Command

```bash
docker run --rm grpc-vs-rest-bench -n 1000 -payload 256 -c 10
```

## Environment

- OS: Linux (Docker container on alpine:3.21)
- CPU: host-dependent
- RAM: host-dependent
- GPU: N/A
- Docker version: host-dependent
- Date: recorded in benchmark JSON

## Inputs

- fixture: synthetic payload generated in EchoLogic
- dataset size: 256 bytes (configurable via -payload)
- repetitions: 1000 per protocol (configurable via -n)
- warmup: none (cold start measurement)
- concurrency: 10 (configurable via -c)

## Metrics

| Metric | Unit | Source | Why it matters |
|---|---|---|---:|---|
| latency_ms_by_protocol | milliseconds | benchmark client | proves the repo claim; REST vs gRPC comparison |
| throughput_req_per_sec | requests/second | benchmark client | secondary metric for capacity comparison |

## Result schema

Output must be JSON and include project, metric, value, unit, timestamp, environment, and command. Schema defined in internal/benchmark/report.go.

## Post angle

#15 grpc-vs-rest-bench: latency_ms_by_protocol as a reproducible portfolio benchmark. tl;dr: gRPC beats REST by 2-5x on latency for small payloads.
