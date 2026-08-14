FROM golang:1.26.5-alpine3.24 AS build

ARG SOURCE_COMMIT=unknown
ARG IMAGE_REF=grpc-vs-rest-bench:local
ARG GO_SUM_SHA256=unknown

RUN apk add --no-cache protoc protobuf-dev
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

WORKDIR /src
COPY . .

RUN rm -rf internal/proto && \
    export PATH="$PATH:$(go env GOPATH)/bin" && \
    protoc --go_out=internal --go_opt=paths=source_relative \
    --go-grpc_out=internal --go-grpc_opt=paths=source_relative \
    -I. proto/benchmark/v1/service.proto && \
    go mod tidy && \
    CGO_ENABLED=0 go build -o /build/grpc-server ./cmd/grpc-server && \
    CGO_ENABLED=0 go build -o /build/rest-server ./cmd/rest-server && \
    CGO_ENABLED=0 go build \
      -ldflags "-X github.com/Brilhante29/grpc-vs-rest-bench/internal/benchmark.SourceCommit=${SOURCE_COMMIT} -X github.com/Brilhante29/grpc-vs-rest-bench/internal/benchmark.ImageRef=${IMAGE_REF} -X github.com/Brilhante29/grpc-vs-rest-bench/internal/benchmark.GoSumSHA256=${GO_SUM_SHA256}" \
      -o /build/bench-client ./cmd/bench-client

FROM alpine:3.24
ARG SOURCE_COMMIT=unknown
ARG IMAGE_REF=grpc-vs-rest-bench:local
LABEL org.opencontainers.image.revision=$SOURCE_COMMIT \
      org.opencontainers.image.title="grpc-vs-rest-bench" \
      org.opencontainers.image.ref.name=$IMAGE_REF
RUN apk add --no-cache ca-certificates bash
WORKDIR /app
COPY --from=build /build/ /usr/local/bin/
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

EXPOSE 50051 8080

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["bench-all", "-n", "1000", "-payload", "256", "-c", "10", "-warmup", "100", "-repetitions", "3", "-out", "/results"]
