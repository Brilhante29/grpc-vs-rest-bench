package benchmark

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/Brilhante29/grpc-vs-rest-bench/internal"
)

type Config struct {
	Requests       int
	WarmupRequests int
	Repetitions    int
	Concurrency    int
	PayloadBytes   int
	RequestTimeout time.Duration
}

type RepetitionResult struct {
	Index             int     `json:"index"`
	Attempts          int     `json:"attempts"`
	Successes         int     `json:"successes"`
	Failures          int     `json:"failures"`
	WallTimeMS        float64 `json:"wall_time_ms"`
	P50MS             float64 `json:"p50_ms"`
	P95MS             float64 `json:"p95_ms"`
	P99MS             float64 `json:"p99_ms"`
	ThroughputReqPerS float64 `json:"throughput_req_per_sec"`
}

type AggregateResult struct {
	Attempts          int     `json:"attempts"`
	Successes         int     `json:"successes"`
	Failures          int     `json:"failures"`
	FailureRate       float64 `json:"failure_rate"`
	P50MS             float64 `json:"p50_ms"`
	P95MS             float64 `json:"p95_ms"`
	P99MS             float64 `json:"p99_ms"`
	ThroughputReqPerS float64 `json:"throughput_req_per_sec"`
}

type ProtocolResult struct {
	Protocol      string             `json:"protocol"`
	Transport     string             `json:"transport"`
	Serialization string             `json:"serialization"`
	Repetitions   []RepetitionResult `json:"repetitions"`
	Aggregate     AggregateResult    `json:"aggregate"`
}

type Client interface {
	Echo(context.Context, internal.EchoRequest) (internal.EchoResponse, error)
}

type protocolSpec struct {
	name, transport, serialization string
	client                         Client
	repetitions                    []RepetitionResult
	latencies                      []time.Duration
	wall                           time.Duration
}

func (c Config) Validate() error {
	if c.Requests <= 0 || c.WarmupRequests <= 0 || c.Repetitions < 3 || c.Concurrency <= 0 || c.PayloadBytes <= 0 {
		return fmt.Errorf("requests, warmup, concurrency and payload must be positive; repetitions must be at least 3")
	}
	if c.RequestTimeout <= 0 {
		return fmt.Errorf("request timeout must be positive")
	}
	return nil
}

func RunPair(ctx context.Context, cfg Config, payload string, rest, grpc Client) ([]ProtocolResult, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if len(payload) != cfg.PayloadBytes {
		return nil, fmt.Errorf("payload length %d does not match configured %d bytes", len(payload), cfg.PayloadBytes)
	}
	protocols := []*protocolSpec{
		{name: "REST", transport: "HTTP/1.1", serialization: "JSON", client: rest},
		{name: "gRPC", transport: "HTTP/2", serialization: "Protobuf", client: grpc},
	}
	for _, spec := range protocols {
		warmup := cfg
		warmup.Requests = cfg.WarmupRequests
		result, _ := run(ctx, spec.client, spec.name, 0, warmup, payload)
		if result.Failures > 0 {
			return nil, fmt.Errorf("%s warmup failed: %d/%d requests", spec.name, result.Failures, result.Attempts)
		}
	}

	for repetition := 1; repetition <= cfg.Repetitions; repetition++ {
		order := protocols
		if repetition%2 == 0 {
			order = []*protocolSpec{protocols[1], protocols[0]}
		}
		for _, spec := range order {
			result, latencies := run(ctx, spec.client, spec.name, repetition, cfg, payload)
			spec.repetitions = append(spec.repetitions, result)
			spec.latencies = append(spec.latencies, latencies...)
			spec.wall += time.Duration(result.WallTimeMS * float64(time.Millisecond))
		}
	}

	results := make([]ProtocolResult, 0, len(protocols))
	for _, spec := range protocols {
		results = append(results, summarize(spec))
	}
	return results, nil
}

func run(ctx context.Context, client Client, protocol string, repetition int, cfg Config, payload string) (RepetitionResult, []time.Duration) {
	jobs := make(chan int)
	latencies := make([]time.Duration, 0, cfg.Requests)
	failures := 0
	var mu sync.Mutex
	var wg sync.WaitGroup
	workers := cfg.Concurrency
	if workers > cfg.Requests {
		workers = cfg.Requests
	}
	started := time.Now()
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range jobs {
				request := internal.EchoRequest{RequestID: fmt.Sprintf("r%02d-%06d", repetition, index), Payload: payload}
				callCtx, cancel := context.WithTimeout(ctx, cfg.RequestTimeout)
				callStarted := time.Now()
				response, err := client.Echo(callCtx, request)
				elapsed := time.Since(callStarted)
				cancel()
				if err == nil {
					digest := sha256.Sum256([]byte(payload))
					expectedHash := fmt.Sprintf("%x", digest)
					if response.RequestID != request.RequestID || response.Payload != payload || response.PayloadSHA256 != expectedHash {
						err = fmt.Errorf("semantic parity failure")
					}
				}
				mu.Lock()
				if err != nil {
					failures++
				} else {
					latencies = append(latencies, elapsed)
				}
				mu.Unlock()
			}
		}()
	}
	for index := 0; index < cfg.Requests; index++ {
		jobs <- index
	}
	close(jobs)
	wg.Wait()
	wall := time.Since(started)
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	return RepetitionResult{
		Index: repetition, Attempts: cfg.Requests, Successes: len(latencies), Failures: failures,
		WallTimeMS: ms(wall), P50MS: percentile(latencies, 0.50), P95MS: percentile(latencies, 0.95),
		P99MS: percentile(latencies, 0.99), ThroughputReqPerS: rate(len(latencies), wall),
	}, latencies
}

func summarize(spec *protocolSpec) ProtocolResult {
	aggregate := AggregateResult{}
	for _, repetition := range spec.repetitions {
		aggregate.Attempts += repetition.Attempts
		aggregate.Successes += repetition.Successes
		aggregate.Failures += repetition.Failures
	}
	sort.Slice(spec.latencies, func(i, j int) bool { return spec.latencies[i] < spec.latencies[j] })
	if aggregate.Attempts > 0 {
		aggregate.FailureRate = float64(aggregate.Failures) / float64(aggregate.Attempts)
	}
	aggregate.P50MS = percentile(spec.latencies, 0.50)
	aggregate.P95MS = percentile(spec.latencies, 0.95)
	aggregate.P99MS = percentile(spec.latencies, 0.99)
	aggregate.ThroughputReqPerS = rate(aggregate.Successes, spec.wall)
	return ProtocolResult{Protocol: spec.name, Transport: spec.transport, Serialization: spec.serialization, Repetitions: spec.repetitions, Aggregate: aggregate}
}

func percentile(sorted []time.Duration, quantile float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	index := int(float64(len(sorted)-1) * quantile)
	return ms(sorted[index])
}

func ms(duration time.Duration) float64 { return float64(duration) / float64(time.Millisecond) }

func rate(successes int, elapsed time.Duration) float64 {
	if elapsed <= 0 {
		return 0
	}
	return float64(successes) / elapsed.Seconds()
}
