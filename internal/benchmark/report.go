package benchmark

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type EnvInfo struct {
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	GoVersion   string `json:"go_version"`
	CPUs        int    `json:"num_cpu"`
	Hostname    string `json:"hostname,omitempty"`
	DockerImage string `json:"docker_image,omitempty"`
}

type BenchmarkReport struct {
	Project    string           `json:"project"`
	Claim      string           `json:"claim"`
	Metric     string           `json:"primary_metric"`
	Unit       string           `json:"unit"`
	Timestamp  string           `json:"timestamp"`
	Env        EnvInfo          `json:"environment"`
	Command    string           `json:"command"`
	Results    []ProtocolResult `json:"results"`
	Comparison map[string]float64 `json:"comparison,omitempty"`
}

func NewReport(command string) *BenchmarkReport {
	host, _ := os.Hostname()
	return &BenchmarkReport{
		Project: "grpc-vs-rest-bench",
		Claim:   "comparacao REST vs gRPC",
		Metric:  "latency_ms_by_protocol",
		Unit:    "milliseconds",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Env: EnvInfo{
			OS:        runtime.GOOS,
			Arch:      runtime.GOARCH,
			GoVersion: runtime.Version(),
			CPUs:      runtime.NumCPU(),
			Hostname:  host,
		},
		Command: command,
		Results: []ProtocolResult{},
	}
}

func (r *BenchmarkReport) AddResult(res ProtocolResult) {
	r.Results = append(r.Results, res)
	r.computeComparison()
}

func (r *BenchmarkReport) computeComparison() {
	var restResult *ProtocolResult
	var grpcResult *ProtocolResult
	for i := range r.Results {
		switch r.Results[i].Protocol {
		case "REST":
			restResult = &r.Results[i]
		case "gRPC":
			grpcResult = &r.Results[i]
		}
	}
	if restResult != nil && grpcResult != nil {
		r.Comparison = map[string]float64{
			"rest_mean_ms":     float64(restResult.MeanLatency) / 1e6,
			"grpc_mean_ms":     float64(grpcResult.MeanLatency) / 1e6,
			"speedup_factor":   float64(restResult.MeanLatency) / float64(grpcResult.MeanLatency),
			"rest_p99_ms":      float64(restResult.P99Latency) / 1e6,
			"grpc_p99_ms":      float64(grpcResult.P99Latency) / 1e6,
			"rest_throughput":  restResult.Throughput,
			"grpc_throughput":  grpcResult.Throughput,
		}
	}
}

func (r *BenchmarkReport) Save(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	path := filepath.Join(dir, "benchmark-result.json")
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

func (r *BenchmarkReport) Print() {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(r)
}
