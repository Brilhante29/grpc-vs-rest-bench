package benchmark

import (
	"context"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Brilhante29/grpc-vs-rest-bench/internal"
)

type fakeClient struct{ failures atomic.Int64 }

func (f *fakeClient) Echo(_ context.Context, request internal.EchoRequest) (internal.EchoResponse, error) {
	response, err := internal.NewEchoLogic().Echo(context.Background(), request)
	return response, err
}

func TestRunPairProducesThreeComparableRepetitions(t *testing.T) {
	cfg := Config{Requests: 20, WarmupRequests: 2, Repetitions: 3, Concurrency: 4, PayloadBytes: 32, RequestTimeout: time.Second}
	payload := strings.Repeat("A", cfg.PayloadBytes)
	results, err := RunPair(context.Background(), cfg, payload, &fakeClient{}, &fakeClient{})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected two protocols, got %d", len(results))
	}
	for _, result := range results {
		if len(result.Repetitions) != 3 || result.Aggregate.Attempts != 60 || result.Aggregate.Failures != 0 {
			t.Fatalf("unexpected aggregate for %s: %#v", result.Protocol, result.Aggregate)
		}
		if result.Aggregate.P50MS > result.Aggregate.P95MS || result.Aggregate.P95MS > result.Aggregate.P99MS {
			t.Fatalf("non-monotonic percentiles for %s", result.Protocol)
		}
	}
	report := NewReport(cfg, payload, "test command", results)
	if issues := Validate(report, false); len(issues) > 0 {
		t.Fatalf("unexpected validation issues: %v", issues)
	}
}

func TestReportRoundTrip(t *testing.T) {
	cfg := Config{Requests: 1, WarmupRequests: 1, Repetitions: 3, Concurrency: 1, PayloadBytes: 1, RequestTimeout: time.Second}
	report := NewReport(cfg, "A", "test", []ProtocolResult{{Protocol: "REST", Repetitions: make([]RepetitionResult, 3)}, {Protocol: "gRPC", Repetitions: make([]RepetitionResult, 3)}})
	dir := t.TempDir()
	if err := report.Save(dir); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(dir + "/benchmark-result.json")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ComparabilityKey != report.ComparabilityKey {
		t.Fatal("comparability key changed after JSON round trip")
	}
	if _, err := os.Stat(dir + "/benchmark-result.json"); err != nil {
		t.Fatal(err)
	}
}

func TestValidateRejectsFailureAndMissingProtocol(t *testing.T) {
	cfg := Config{Requests: 1, WarmupRequests: 1, Repetitions: 3, Concurrency: 1, PayloadBytes: 1, RequestTimeout: time.Second}
	repetitions := make([]RepetitionResult, 3)
	report := NewReport(cfg, "A", "test", []ProtocolResult{
		{Protocol: "REST", Repetitions: repetitions, Aggregate: AggregateResult{Attempts: 1, Failures: 1}},
		{Protocol: "HTTP", Repetitions: repetitions},
	})
	issues := strings.Join(Validate(report, false), "; ")
	if !strings.Contains(issues, "zero failures") || !strings.Contains(issues, "REST and gRPC") {
		t.Fatalf("expected failure and protocol issues, got %q", issues)
	}
}
