#!/bin/sh
set -eu

unformatted=$(/usr/local/go/bin/gofmt -l .)
if [ -n "$unformatted" ]; then
  printf '%s\n' "$unformatted" >&2
  exit 1
fi

/usr/local/go/bin/go run ./cmd/bench-client -validate /src/benchmarks/results/benchmark-result.json
/usr/local/go/bin/go test -count=1 ./...
/usr/local/go/bin/go vet ./...
/usr/local/go/bin/go build ./...
