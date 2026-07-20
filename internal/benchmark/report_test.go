package benchmark

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestReportAddResult(t *testing.T) {
	r := NewReport("test-command")
	r.AddResult(ProtocolResult{
		Protocol: "REST", Requests: 100, PayloadSize: 256,
		MeanLatency: 5 * time.Millisecond, P50Latency: 4 * time.Millisecond,
		P95Latency: 10 * time.Millisecond, P99Latency: 15 * time.Millisecond,
		MinLatency: 1 * time.Millisecond, MaxLatency: 20 * time.Millisecond,
		Throughput: 1000,
	})
	r.AddResult(ProtocolResult{
		Protocol: "gRPC", Requests: 100, PayloadSize: 256,
		MeanLatency: 2 * time.Millisecond, P50Latency: 1 * time.Millisecond,
		P95Latency: 5 * time.Millisecond, P99Latency: 8 * time.Millisecond,
		MinLatency: 500 * time.Microsecond, MaxLatency: 10 * time.Millisecond,
		Throughput: 2000,
	})
	if len(r.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(r.Results))
	}
	if r.Comparison == nil {
		t.Fatal("expected comparison to be computed")
	}
	if r.Comparison["speedup_factor"] != 2.5 {
		t.Fatalf("expected speedup 2.5, got %v", r.Comparison["speedup_factor"])
	}
}

func TestReportSave(t *testing.T) {
	r := NewReport("go test")
	r.AddResult(ProtocolResult{
		Protocol: "REST", Requests: 10, PayloadSize: 100,
		MeanLatency: time.Millisecond, P50Latency: time.Millisecond,
		P95Latency: 2 * time.Millisecond, P99Latency: 3 * time.Millisecond,
		MinLatency: 500 * time.Microsecond, MaxLatency: 5 * time.Millisecond,
		Throughput: 500,
	})
	dir := t.TempDir()
	if err := r.Save(dir); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(dir + "/benchmark-result.json")
	if err != nil {
		t.Fatal(err)
	}
	var loaded BenchmarkReport
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatal(err)
	}
	if loaded.Project != "grpc-vs-rest-bench" {
		t.Fatalf("expected project grpc-vs-rest-bench, got %s", loaded.Project)
	}
}

func TestReportPrint(t *testing.T) {
	r := NewReport("print-test")
	r.AddResult(ProtocolResult{
		Protocol: "gRPC", Requests: 50, PayloadSize: 256,
		MeanLatency: time.Millisecond,
	})
	r.Print()
}
