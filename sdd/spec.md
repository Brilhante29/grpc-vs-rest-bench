# Spec: grpc-vs-rest-bench

## Number

#15

## Claim

Este projeto prova que: comparacao REST vs gRPC — qual transporte apresenta menor latencia para cargas de trabalho request/response identicas.

## Stack

go, grpc, chi, protobuf, docker

## User-visible output

- Docker command: `docker run --rm grpc-vs-rest-bench`
- README opens with: `# #15 grpc-vs-rest-bench`
- Benchmark table: latency_ms_by_protocol (mean, p50, p95, p99, throughput)

## Scope

In:

- Implementar o menor produto funcional que prove o claim.
- Rodar por Docker.
- Gerar benchmark JSON reproduzivel.
- Dois servidores (REST + gRPC) compartilhando a mesma logica de Echo.
- Cliente de benchmark que envia N requisicoes para cada protocolo e compara latencia.

Out:

- Publicar repo antes do primeiro resultado numerico.
- Depender de segredo pago para o caminho default.
- Autenticacao, banco de dados, streaming.

## Architecture

```
client -> app (cmd/) -> domain (internal/) -> benchmark output
```

## Benchmark

Primary metric:

- name: latency_ms_by_protocol
- target: first reproducible baseline
- command: `docker run --rm grpc-vs-rest-bench -n 1000 -payload 256 -c 10`
- result file: benchmarks/results/*.json

## Dataset or fixture

- source: synthetic payload generated at runtime
- size: configurable via -payload flag (default 256 bytes)
- license: MIT
- deterministic seed: 42

## Definition of done

- [x] Docker command works from clean clone.
- [x] README starts with project number and benchmark result.
- [x] Benchmark command writes JSON result.
- [x] Tests cover core behavior.
- [x] REFERENCES.md explains reuse.
- [x] No secret or paid credential required for default demo.
