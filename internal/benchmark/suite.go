package benchmark

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type ProtocolResult struct {
	Protocol    string        `json:"protocol"`
	Requests    int           `json:"requests"`
	PayloadSize int           `json:"payload_bytes"`
	TotalTime   time.Duration `json:"total_time_ns"`
	MeanLatency time.Duration `json:"mean_latency_ns"`
	P50Latency  time.Duration `json:"p50_latency_ns"`
	P95Latency  time.Duration `json:"p95_latency_ns"`
	P99Latency  time.Duration `json:"p99_latency_ns"`
	MinLatency  time.Duration `json:"min_latency_ns"`
	MaxLatency  time.Duration `json:"max_latency_ns"`
	Throughput  float64       `json:"throughput_req_per_sec"`
}

type RunnerFunc func(ctx context.Context) (time.Duration, error)

type Suite struct {
	Results []ProtocolResult
	mu      sync.Mutex
}

func NewSuite() *Suite {
	return &Suite{}
}

func (s *Suite) Run(protocol string, requests int, payloadBytes int, concurrency int, fn RunnerFunc) (ProtocolResult, error) {
	latencies := make([]time.Duration, requests)
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for i := 0; i < requests; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int) {
			defer wg.Done()
			defer func() { <-sem }()
			dur, err := fn(ctx)
			if err != nil {
				mu.Lock()
				latencies[idx] = 0
				mu.Unlock()
				return
			}
			mu.Lock()
			latencies[idx] = dur
			mu.Unlock()
		}(i)
	}
	wg.Wait()
	totalTime := time.Since(start)

	valid := make([]time.Duration, 0, requests)
	for _, l := range latencies {
		if l > 0 {
			valid = append(valid, l)
		}
	}

	if len(valid) == 0 {
		return ProtocolResult{}, fmt.Errorf("no successful requests for %s", protocol)
	}

	sortDurations(valid)
	totalNs := int64(0)
	for _, l := range valid {
		totalNs += l.Nanoseconds()
	}
	mean := time.Duration(totalNs / int64(len(valid)))

	result := ProtocolResult{
		Protocol:    protocol,
		Requests:    len(valid),
		PayloadSize: payloadBytes,
		TotalTime:   totalTime,
		MeanLatency: mean,
		P50Latency:  valid[len(valid)*50/100],
		P95Latency:  valid[len(valid)*95/100],
		P99Latency:  valid[len(valid)*99/100],
		MinLatency:  valid[0],
		MaxLatency:  valid[len(valid)-1],
		Throughput:  float64(len(valid)) / totalTime.Seconds(),
	}

	s.mu.Lock()
	s.Results = append(s.Results, result)
	s.mu.Unlock()

	return result, nil
}

func sortDurations(d []time.Duration) {
	n := len(d)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if d[i] > d[j] {
				d[i], d[j] = d[j], d[i]
			}
		}
	}
}
