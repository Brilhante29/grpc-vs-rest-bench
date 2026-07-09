# #15 grpc-vs-rest-bench

**Status:** scaffold

**Proves:** comparacao REST vs gRPC.

**Benchmark target:** latency_ms_by_protocol.

**Stack:** go, grpc, chi, ghz, k6, docker.

## Next milestone

Implement the smallest Docker-runnable version and produce the first JSON benchmark under enchmarks/results/.

## Run

`ash
docker build -t grpc-vs-rest-bench .
docker run --rm grpc-vs-rest-bench
`

## Benchmark

`ash
docker run --rm grpc-vs-rest-bench benchmark
`

| Metric | Value | Unit |
|---|---:|---|
| latency_ms_by_protocol | pending | pending |

## Architecture

Defined in sdd/spec.md before implementation.

## References

See REFERENCES.md.