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
	report, err := NewReport(cfg, payload, "test command", results, time.Now(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if report.SchemaVersion != 2 || len(report.Metrics) != 11 {
		t.Fatalf("expected common V2 report with 11 metrics, got schema %d and %d metrics", report.SchemaVersion, len(report.Metrics))
	}
	if issues := Validate(report, false); len(issues) > 0 {
		t.Fatalf("unexpected validation issues: %v", issues)
	}
}

func TestReportRoundTrip(t *testing.T) {
	cfg, payload, results := reportFixture()
	report, err := NewReport(cfg, payload, "test", results, time.Now(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := report.Save(dir); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(dir + "/benchmark-result.json")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ComparabilityKey != report.ComparabilityKey || loaded.RunID != report.RunID {
		t.Fatal("identity changed after JSON round trip")
	}
	if _, err := os.Stat(dir + "/benchmark-result.json"); err != nil {
		t.Fatal(err)
	}
}

func TestValidateRejectsFailureAndMissingProtocol(t *testing.T) {
	cfg, payload, results := reportFixture()
	results[0].Repetitions[0].Failures = 1
	results[0].Aggregate.Failures = 1
	results[1].Protocol = "HTTP"
	report, err := NewReport(cfg, payload, "test", results, time.Now(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	issues := strings.Join(Validate(report, false), "; ")
	if !strings.Contains(issues, "zero failures") || !strings.Contains(issues, "missing required metric: grpc") {
		t.Fatalf("expected failure and protocol issues, got %q", issues)
	}
}

func reportFixture() (Config, string, []ProtocolResult) {
	cfg := Config{Requests: 10, WarmupRequests: 2, Repetitions: 3, Concurrency: 2, PayloadBytes: 1, RequestTimeout: time.Second}
	makeRepetitions := func(base float64) []RepetitionResult {
		return []RepetitionResult{
			{Index: 1, Attempts: 10, Successes: 10, P50MS: base, P95MS: base + 1, P99MS: base + 2, ThroughputReqPerS: 1000 / base},
			{Index: 2, Attempts: 10, Successes: 10, P50MS: base + 0.1, P95MS: base + 1.1, P99MS: base + 2.1, ThroughputReqPerS: 990 / base},
			{Index: 3, Attempts: 10, Successes: 10, P50MS: base + 0.2, P95MS: base + 1.2, P99MS: base + 2.2, ThroughputReqPerS: 980 / base},
		}
	}
	return cfg, "A", []ProtocolResult{
		{Protocol: "REST", Repetitions: makeRepetitions(1), Aggregate: AggregateResult{Attempts: 30, Successes: 30}},
		{Protocol: "gRPC", Repetitions: makeRepetitions(2), Aggregate: AggregateResult{Attempts: 30, Successes: 30}},
	}
}
