package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Brilhante29/grpc-vs-rest-bench/internal/benchmark"
	pb "github.com/Brilhante29/grpc-vs-rest-bench/internal/proto/benchmark/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	requests := flag.Int("n", 1000, "number of requests per protocol")
	payloadBytes := flag.Int("payload", 256, "payload size in bytes")
	concurrency := flag.Int("c", 10, "concurrency level")
	restAddr := flag.String("rest", "localhost:8080", "REST server address")
	grpcAddr := flag.String("grpc", "localhost:50051", "gRPC server address")
	resultsDir := flag.String("out", "benchmarks/results", "output directory for results")
	flag.Parse()

	suite := benchmark.NewSuite()
	report := benchmark.NewReport(fmt.Sprintf("./bench-client -n %d -payload %d -c %d", *requests, *payloadBytes, *concurrency))

	log.Printf("Running benchmark: %d requests, %dB payload, %d concurrency", *requests, *payloadBytes, *concurrency)

	restResult, err := suite.Run("REST", *requests, *payloadBytes, *concurrency, func(ctx context.Context) (time.Duration, error) {
		return doREST(ctx, *restAddr, *payloadBytes)
	})
	if err != nil {
		log.Fatalf("REST benchmark failed: %v", err)
	}
	log.Printf("REST: mean=%v p50=%v p95=%v p99=%v throughput=%.0f/s",
		restResult.MeanLatency, restResult.P50Latency, restResult.P95Latency, restResult.P99Latency, restResult.Throughput)
	report.AddResult(restResult)

	grpcResult, err := suite.Run("gRPC", *requests, *payloadBytes, *concurrency, func(ctx context.Context) (time.Duration, error) {
		return doGRPC(ctx, *grpcAddr, *payloadBytes)
	})
	if err != nil {
		log.Fatalf("gRPC benchmark failed: %v", err)
	}
	log.Printf("gRPC: mean=%v p50=%v p95=%v p99=%v throughput=%.0f/s",
		grpcResult.MeanLatency, grpcResult.P50Latency, grpcResult.P95Latency, grpcResult.P99Latency, grpcResult.Throughput)
	report.AddResult(grpcResult)

	if err := report.Save(*resultsDir); err != nil {
		log.Fatalf("failed to save report: %v", err)
	}
	fmt.Println()
	report.Print()

	fmt.Printf("\nSpeedup (REST mean / gRPC mean): %.2fx\n",
		float64(restResult.MeanLatency)/float64(grpcResult.MeanLatency))
}

var httpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	},
	Timeout: 30 * time.Second,
}

func doREST(ctx context.Context, addr string, payloadBytes int) (time.Duration, error) {
	body := map[string]interface{}{
		"message":       "benchmark",
		"payload_bytes": payloadBytes,
	}
	data, _ := json.Marshal(body)
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("http://%s/api/v1/echo", addr), bytes.NewReader(data))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return time.Since(start), nil
}

var grpcConn *grpc.ClientConn
var grpcClient pb.BenchmarkServiceClient

func doGRPC(ctx context.Context, addr string, payloadBytes int) (time.Duration, error) {
	if grpcConn == nil {
		conn, err := grpc.Dial(addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return 0, fmt.Errorf("grpc dial: %w", err)
		}
		grpcConn = conn
		grpcClient = pb.NewBenchmarkServiceClient(conn)
	}
	start := time.Now()
	_, err := grpcClient.Echo(ctx, &pb.EchoRequest{
		Message:      "benchmark",
		PayloadBytes: int32(payloadBytes),
	})
	if err != nil {
		return 0, err
	}
	return time.Since(start), nil
}

func init() {
	log.SetFlags(0)
	log.SetOutput(os.Stderr)
}
