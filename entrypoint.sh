#!/bin/sh
set -e

mode="${1:-bench-all}"
shift || true

case "$mode" in
  rest-server|grpc-server|bench-client)
    exec "$mode" "$@"
    ;;
  bench-all)
    ;;
  *)
    echo "unknown mode: $mode" >&2
    exit 64
    ;;
esac

grpc-server &
grpc_pid=$!
rest-server &
rest_pid=$!

cleanup() {
  kill "$grpc_pid" "$rest_pid" 2>/dev/null || true
  wait "$grpc_pid" "$rest_pid" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

attempt=0
until wget -q -O /dev/null http://127.0.0.1:8080/healthz; do
  attempt=$((attempt + 1))
  if [ "$attempt" -ge 50 ]; then
    echo "REST server did not become ready" >&2
    exit 1
  fi
  sleep 0.1
done

bench-client -rest 127.0.0.1:8080 -grpc 127.0.0.1:50051 "$@"
