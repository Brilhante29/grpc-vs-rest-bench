# #15 grpc-vs-rest-bench

**Status:** benchmarked

**Proves:** comparacao REST vs gRPC — which transport is faster for identical request/response workloads?

**Stack:** go, grpc, chi, protobuf, docker.

## Benchmark Result

256B payload, 1000 requests, 10 concurrent connections (Docker container, 6 vCPU):

| Metric | REST (HTTP/1.1 + JSON) | gRPC (HTTP/2 + Protobuf) | Speedup |
|---|---|---|---|
| Mean latency | 0.478 ms | 0.667 ms | 0.72x REST faster |
| P99 latency | 1.893 ms | 2.128 ms | 0.89x REST faster |
| Throughput | 19,619 req/s | 14,536 req/s | 1.35x REST faster |

**Key finding:** For simple echo workloads with small payloads (256B), REST/HTTP/1.1 outperforms gRPC/HTTP/2 in this Docker environment. The overhead of HTTP/2 framing and protobuf serialization exceeds the benefits for trivial request/response patterns.

## Run

```bash
docker build -t grpc-vs-rest-bench .
docker run --rm grpc-vs-rest-bench
```

## Benchmark

```bash
docker run --rm grpc-vs-rest-bench -n 1000 -payload 256 -c 10
```

Results are written to `benchmarks/results/benchmark-result.json`.

## Architecture

```
                  ┌──────────────┐
                  │  bench-client│
                  └──┬───────┬───┘
                     │       │
               REST  │       │  gRPC
                     ▼       ▼
              ┌────────┐ ┌────────┐
              │  REST  │ │  gRPC  │
              │ server │ │ server │
              └───┬────┘ └───┬────┘
                  │          │
                  ▼          ▼
              ┌────────────────────┐
              │   EchoLogic        │
              │   (internal/)      │
              └────────────────────┘
```

Both servers share the same `EchoLogic` in `internal/`. The benchmark client sends identical requests to both and compares latency distributions.

## References

See REFERENCES.md.

## License

MIT
