FROM golang:1.23-alpine AS build

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
    CGO_ENABLED=0 go build -o /build/bench-client ./cmd/bench-client

FROM alpine:3.21
RUN apk add --no-cache ca-certificates bash
WORKDIR /app
COPY --from=build /build/ /usr/local/bin/
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

EXPOSE 50051 8080

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["-n", "1000", "-payload", "256", "-c", "10"]
