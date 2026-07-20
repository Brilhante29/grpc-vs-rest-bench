#!/bin/sh
set -e

# Start both servers in background
grpc-server &
GRPC_PID=$!

rest-server &
REST_PID=$!

# Wait for servers to be ready
sleep 2

# Run benchmark
bench-client "$@"
BENCH_EXIT=$?

# Cleanup
kill $GRPC_PID $REST_PID 2>/dev/null || true
wait $GRPC_PID $REST_PID 2>/dev/null || true

exit $BENCH_EXIT
