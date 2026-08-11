package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Brilhante29/grpc-vs-rest-bench/internal"
	"github.com/Brilhante29/grpc-vs-rest-bench/internal/benchmark"
)

func main() {
	requests := flag.Int("n", 1000, "requests per repetition and protocol")
	payloadBytes := flag.Int("payload", 256, "logical UTF-8 payload size in bytes")
	concurrency := flag.Int("c", 10, "concurrent clients")
	repetitions := flag.Int("repetitions", 3, "measured repetitions")
	warmup := flag.Int("warmup", 100, "warmup requests per protocol")
	restAddr := flag.String("rest", "localhost:8080", "REST server address")
	grpcAddr := flag.String("grpc", "localhost:50051", "gRPC server address")
	resultsDir := flag.String("out", "benchmarks/results", "result directory")
	validatePath := flag.String("validate", "", "validate an existing V2 report and exit")
	flag.Parse()

	if *validatePath != "" {
		validate(*validatePath)
		return
	}

	cfg := benchmark.Config{Requests: *requests, WarmupRequests: *warmup, Repetitions: *repetitions,
		Concurrency: *concurrency, PayloadBytes: *payloadBytes, RequestTimeout: 10 * time.Second}
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
	payload := strings.Repeat("A", cfg.PayloadBytes)
	restClient := internal.NewRESTClient(*restAddr, cfg.Concurrency, cfg.RequestTimeout)
	grpcClient, err := internal.NewGRPCClient(*grpcAddr, cfg.RequestTimeout)
	if err != nil {
		log.Fatal(err)
	}
	defer grpcClient.Close()

	command := fmt.Sprintf("bench-client -n %d -payload %d -c %d -warmup %d -repetitions %d",
		cfg.Requests, cfg.PayloadBytes, cfg.Concurrency, cfg.WarmupRequests, cfg.Repetitions)
	log.Printf("V2 benchmark: %d requests x %d repetitions, %dB payload, concurrency %d", cfg.Requests, cfg.Repetitions, cfg.PayloadBytes, cfg.Concurrency)
	results, err := benchmark.RunPair(context.Background(), cfg, payload, restClient, grpcClient)
	if err != nil {
		log.Fatal(err)
	}
	report := benchmark.NewReport(cfg, payload, command, results)
	if err := report.Save(*resultsDir); err != nil {
		log.Fatal(err)
	}
	data, _ := json.MarshalIndent(report, "", "  ")
	fmt.Println(string(data))
	for _, result := range results {
		if result.Aggregate.Failures > 0 {
			log.Fatalf("%s recorded %d failures", result.Protocol, result.Aggregate.Failures)
		}
	}
}

func validate(path string) {
	report, err := benchmark.Load(path)
	if err != nil {
		log.Fatal(err)
	}
	issues := benchmark.Validate(report, true)
	if len(issues) > 0 {
		for _, issue := range issues {
			fmt.Fprintln(os.Stderr, "-", issue)
		}
		os.Exit(1)
	}
	fmt.Println("benchmark-report/v2 validation passed")
}

func init() {
	log.SetFlags(0)
	log.SetOutput(os.Stderr)
}
