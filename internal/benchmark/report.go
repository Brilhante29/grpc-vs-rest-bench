package benchmark

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/Brilhante29/grpc-vs-rest-bench/internal"
)

var (
	SourceCommit = "unknown"
	ImageRef     = "unknown"
	GoSumSHA256  = "unknown"
)

type Workload struct {
	Version            string `json:"version"`
	FixtureDigest      string `json:"fixture_digest"`
	ConfigDigest       string `json:"config_digest"`
	WarmupIterations   int    `json:"warmup_iterations"`
	MeasuredIterations int    `json:"measured_iterations"`
	Concurrency        int    `json:"concurrency"`
}

type Metric struct {
	Name      string             `json:"name"`
	Value     float64            `json:"value"`
	Unit      string             `json:"unit"`
	Direction string             `json:"direction"`
	Samples   []float64          `json:"samples"`
	Failures  int                `json:"failures"`
	Summary   map[string]float64 `json:"summary"`
}

type Execution struct {
	Command         string  `json:"command"`
	StartedAt       string  `json:"started_at"`
	DurationSeconds float64 `json:"duration_seconds"`
	ExitCode        int     `json:"exit_code"`
	Repeat          int     `json:"repeat"`
}

type Environment struct {
	Runtime       string `json:"runtime"`
	Architecture  string `json:"architecture"`
	HardwareClass string `json:"hardware_class"`
	OS            string `json:"os"`
	CPUCount      int    `json:"cpu_count"`
	Topology      string `json:"topology"`
}

type Provenance struct {
	SourceCommit         string `json:"source_commit"`
	CleanTree            bool   `json:"clean_tree"`
	ImageRef             string `json:"image_ref"`
	ImageDigest          string `json:"image_digest"`
	DependencyLockDigest string `json:"dependency_lock_digest"`
	Producer             string `json:"producer"`
	CIRunURL             string `json:"ci_run_url,omitempty"`
	ArtifactDigest       string `json:"artifact_digest"`
}

type Report struct {
	SchemaVersion    int         `json:"schema_version"`
	RunID            string      `json:"run_id"`
	Project          string      `json:"project"`
	BenchmarkID      string      `json:"benchmark_id"`
	Workload         Workload    `json:"workload"`
	Metrics          []Metric    `json:"metrics"`
	Execution        Execution   `json:"execution"`
	Environment      Environment `json:"environment"`
	Provenance       Provenance  `json:"provenance"`
	ComparabilityKey string      `json:"comparability_key"`
}

func NewReport(cfg Config, payload, command string, results []ProtocolResult, startedAt time.Time, duration time.Duration) (Report, error) {
	payloadDigest := digestBytes([]byte(payload))
	configInput := fmt.Sprintf("contract=%s|requests=%d|warmup=%d|repetitions=%d|concurrency=%d|payload_bytes=%d|order=alternating-sequential|rest=http1.1-json|grpc=http2-protobuf",
		internal.ContractVersion, cfg.Requests, cfg.WarmupRequests, cfg.Repetitions, cfg.Concurrency, cfg.PayloadBytes)
	configDigest := digestBytes([]byte(configInput))
	runID, err := newUUID()
	if err != nil {
		return Report{}, fmt.Errorf("create benchmark run id: %w", err)
	}
	artifactDigest, err := executableDigest()
	if err != nil {
		return Report{}, fmt.Errorf("hash benchmark executable: %w", err)
	}
	producer := os.Getenv("BENCHMARK_PRODUCER")
	if producer == "" {
		producer = "local"
	}
	hardwareClass := os.Getenv("BENCHMARK_HARDWARE_CLASS")
	if hardwareClass == "" {
		hardwareClass = fmt.Sprintf("docker-%d-vcpu", runtime.NumCPU())
	}

	return Report{
		SchemaVersion: 2,
		RunID:         runID,
		Project:       "grpc-vs-rest-bench",
		BenchmarkID:   "rest-grpc-unary-echo",
		Workload: Workload{
			Version:            fmt.Sprintf("echo-%s-n%d-p%d", internal.ContractVersion, cfg.Requests, cfg.PayloadBytes),
			FixtureDigest:      payloadDigest,
			ConfigDigest:       configDigest,
			WarmupIterations:   cfg.WarmupRequests,
			MeasuredIterations: cfg.Requests,
			Concurrency:        cfg.Concurrency,
		},
		Metrics: buildMetrics(results),
		Execution: Execution{
			Command:         command,
			StartedAt:       startedAt.UTC().Format(time.RFC3339Nano),
			DurationSeconds: duration.Seconds(),
			ExitCode:        0,
			Repeat:          cfg.Repetitions,
		},
		Environment: Environment{
			Runtime:       runtime.Version(),
			Architecture:  runtime.GOARCH,
			HardwareClass: hardwareClass,
			OS:            runtime.GOOS,
			CPUCount:      runtime.NumCPU(),
			Topology:      "one-client-one-rest-server-one-grpc-server-loopback",
		},
		Provenance: Provenance{
			SourceCommit:         SourceCommit,
			CleanTree:            true,
			ImageRef:             ImageRef,
			ImageDigest:          os.Getenv("BENCHMARK_IMAGE_DIGEST"),
			DependencyLockDigest: normalizeDigest(GoSumSHA256),
			Producer:             producer,
			CIRunURL:             os.Getenv("BENCHMARK_CI_RUN_URL"),
			ArtifactDigest:       artifactDigest,
		},
		ComparabilityKey: configDigest,
	}, nil
}

func buildMetrics(results []ProtocolResult) []Metric {
	metrics := make([]Metric, 0, 11)
	byProtocol := make(map[string]ProtocolResult, len(results))
	for _, result := range results {
		key := strings.ToLower(result.Protocol)
		byProtocol[key] = result
		failures := result.Aggregate.Failures
		metrics = append(metrics,
			metricFromRepetitions(key+"_p50_latency_ms", "milliseconds", "lower_is_better", failures, result.Repetitions, func(r RepetitionResult) float64 { return r.P50MS }),
			metricFromRepetitions(key+"_p95_latency_ms", "milliseconds", "lower_is_better", failures, result.Repetitions, func(r RepetitionResult) float64 { return r.P95MS }),
			metricFromRepetitions(key+"_p99_latency_ms", "milliseconds", "lower_is_better", failures, result.Repetitions, func(r RepetitionResult) float64 { return r.P99MS }),
			metricFromRepetitions(key+"_throughput_rps", "requests_per_second", "higher_is_better", failures, result.Repetitions, func(r RepetitionResult) float64 { return r.ThroughputReqPerS }),
		)
	}

	rest, restOK := byProtocol["rest"]
	grpc, grpcOK := byProtocol["grpc"]
	if restOK && grpcOK {
		failureSamples := pairSamples(rest.Repetitions, grpc.Repetitions, func(rest, grpc RepetitionResult) float64 {
			return float64(rest.Failures + grpc.Failures)
		})
		metrics = append(metrics, newMetric("request_failures", "count", "target", failureSamples, sumIntSamples(failureSamples)))

		p95Ratio := pairSamples(rest.Repetitions, grpc.Repetitions, func(rest, grpc RepetitionResult) float64 {
			if grpc.P95MS == 0 {
				return 0
			}
			return rest.P95MS / grpc.P95MS
		})
		metrics = append(metrics, newMetric("rest_over_grpc_p95_ratio", "ratio", "target", p95Ratio, rest.Aggregate.Failures+grpc.Aggregate.Failures))

		throughputRatio := pairSamples(rest.Repetitions, grpc.Repetitions, func(rest, grpc RepetitionResult) float64 {
			if rest.ThroughputReqPerS == 0 {
				return 0
			}
			return grpc.ThroughputReqPerS / rest.ThroughputReqPerS
		})
		metrics = append(metrics, newMetric("grpc_over_rest_throughput_ratio", "ratio", "target", throughputRatio, rest.Aggregate.Failures+grpc.Aggregate.Failures))
	}
	return metrics
}

func metricFromRepetitions(name, unit, direction string, failures int, repetitions []RepetitionResult, value func(RepetitionResult) float64) Metric {
	samples := make([]float64, 0, len(repetitions))
	for _, repetition := range repetitions {
		samples = append(samples, value(repetition))
	}
	return newMetric(name, unit, direction, samples, failures)
}

func pairSamples(left, right []RepetitionResult, value func(RepetitionResult, RepetitionResult) float64) []float64 {
	count := len(left)
	if len(right) < count {
		count = len(right)
	}
	samples := make([]float64, 0, count)
	for index := 0; index < count; index++ {
		samples = append(samples, value(left[index], right[index]))
	}
	return samples
}

func newMetric(name, unit, direction string, samples []float64, failures int) Metric {
	return Metric{Name: name, Value: median(samples), Unit: unit, Direction: direction, Samples: samples, Failures: failures, Summary: summarizeSamples(samples)}
}

func summarizeSamples(samples []float64) map[string]float64 {
	if len(samples) == 0 {
		return map[string]float64{"min": 0, "median": 0, "max": 0, "mean": 0}
	}
	ordered := append([]float64(nil), samples...)
	sort.Float64s(ordered)
	total := 0.0
	for _, sample := range ordered {
		total += sample
	}
	return map[string]float64{
		"min": ordered[0], "median": median(ordered), "max": ordered[len(ordered)-1], "mean": total / float64(len(ordered)),
	}
}

func median(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	ordered := append([]float64(nil), samples...)
	sort.Float64s(ordered)
	middle := len(ordered) / 2
	if len(ordered)%2 == 0 {
		return (ordered[middle-1] + ordered[middle]) / 2
	}
	return ordered[middle]
}

func sumIntSamples(samples []float64) int {
	total := 0
	for _, sample := range samples {
		total += int(sample)
	}
	return total
}

func newUUID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	encoded := hex.EncodeToString(value)
	return fmt.Sprintf("%s-%s-%s-%s-%s", encoded[0:8], encoded[8:12], encoded[12:16], encoded[16:20], encoded[20:32]), nil
}

func executableDigest() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return digestBytes(data), nil
}

func digestBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func normalizeDigest(value string) string {
	if strings.HasPrefix(value, "sha256:") {
		return value
	}
	if len(value) == 64 {
		return "sha256:" + value
	}
	return value
}

func (r Report) Save(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create result directory: %w", err)
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	return os.WriteFile(filepath.Join(dir, "benchmark-result.json"), append(data, '\n'), 0o644)
}

func Load(path string) (Report, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Report{}, err
	}
	var report Report
	if err := json.Unmarshal(data, &report); err != nil {
		return Report{}, err
	}
	return report, nil
}

func Validate(report Report, exactProvenance bool) []string {
	issues := []string{}
	sha256Pattern := regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	if report.SchemaVersion != 2 {
		issues = append(issues, "schema_version must be integer 2")
	}
	if !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(report.RunID) {
		issues = append(issues, "run_id must be a UUID v4")
	}
	if report.Project != "grpc-vs-rest-bench" || report.BenchmarkID != "rest-grpc-unary-echo" {
		issues = append(issues, "project and benchmark_id must identify the unary echo comparison")
	}
	if report.Workload.WarmupIterations <= 0 || report.Workload.MeasuredIterations <= 0 || report.Workload.Concurrency <= 0 {
		issues = append(issues, "workload requires positive warmup, measured iterations and concurrency")
	}
	if !sha256Pattern.MatchString(report.Workload.FixtureDigest) || !sha256Pattern.MatchString(report.Workload.ConfigDigest) {
		issues = append(issues, "workload fixture/config digests must be SHA-256 values")
	}
	if report.Execution.Command == "" || report.Execution.ExitCode != 0 || report.Execution.Repeat < 3 || report.Execution.DurationSeconds < 0 {
		issues = append(issues, "execution metadata is incomplete or inconsistent")
	}
	if _, err := time.Parse(time.RFC3339Nano, report.Execution.StartedAt); err != nil {
		issues = append(issues, "execution started_at must be RFC3339")
	}
	if report.Environment.Runtime == "" || report.Environment.Architecture == "" || report.Environment.HardwareClass == "" {
		issues = append(issues, "runtime, architecture and hardware_class are required")
	}

	requiredMetrics := map[string]bool{
		"rest_p50_latency_ms": false, "rest_p95_latency_ms": false, "rest_p99_latency_ms": false, "rest_throughput_rps": false,
		"grpc_p50_latency_ms": false, "grpc_p95_latency_ms": false, "grpc_p99_latency_ms": false, "grpc_throughput_rps": false,
		"request_failures": false, "rest_over_grpc_p95_ratio": false, "grpc_over_rest_throughput_ratio": false,
	}
	metricByName := make(map[string]Metric, len(report.Metrics))
	for _, metric := range report.Metrics {
		metricByName[metric.Name] = metric
		if _, required := requiredMetrics[metric.Name]; required {
			requiredMetrics[metric.Name] = true
		}
		if metric.Name == "" || metric.Unit == "" || len(metric.Samples) != report.Execution.Repeat {
			issues = append(issues, metric.Name+" metric metadata/sample count is invalid")
		}
		if metric.Direction != "higher_is_better" && metric.Direction != "lower_is_better" && metric.Direction != "target" {
			issues = append(issues, metric.Name+" direction is invalid")
		}
		if metric.Failures != 0 {
			issues = append(issues, metric.Name+" must report zero failures for publication")
		}
		for _, sample := range metric.Samples {
			if math.IsNaN(sample) || math.IsInf(sample, 0) || sample < 0 {
				issues = append(issues, metric.Name+" contains an invalid sample")
				break
			}
		}
	}
	for name, present := range requiredMetrics {
		if !present {
			issues = append(issues, "missing required metric: "+name)
		}
	}
	for index := 0; index < report.Execution.Repeat; index++ {
		for _, protocol := range []string{"rest", "grpc"} {
			p50, ok50 := metricByName[protocol+"_p50_latency_ms"]
			p95, ok95 := metricByName[protocol+"_p95_latency_ms"]
			p99, ok99 := metricByName[protocol+"_p99_latency_ms"]
			if ok50 && ok95 && ok99 && len(p50.Samples) > index && len(p95.Samples) > index && len(p99.Samples) > index &&
				(p50.Samples[index] > p95.Samples[index] || p95.Samples[index] > p99.Samples[index]) {
				issues = append(issues, fmt.Sprintf("%s latency percentiles are not monotonic in repetition %d", protocol, index+1))
			}
		}
	}
	if failureMetric, ok := metricByName["request_failures"]; ok && failureMetric.Value != 0 {
		issues = append(issues, "request_failures must be zero")
	}
	if !sha256Pattern.MatchString(report.ComparabilityKey) {
		issues = append(issues, "comparability_key must be a SHA-256 key")
	}
	if exactProvenance {
		if !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(report.Provenance.SourceCommit) {
			issues = append(issues, "source_commit must be an exact 40-character Git SHA")
		}
		if !report.Provenance.CleanTree {
			issues = append(issues, "clean_tree must be true")
		}
		if report.Provenance.ImageRef == "" || report.Provenance.ImageRef == "unknown" {
			issues = append(issues, "image_ref is required")
		}
		if !sha256Pattern.MatchString(report.Provenance.ImageDigest) || !sha256Pattern.MatchString(report.Provenance.DependencyLockDigest) || !sha256Pattern.MatchString(report.Provenance.ArtifactDigest) {
			issues = append(issues, "image, dependency lock and artifact digests must be exact SHA-256 values")
		}
		if report.Provenance.Producer != "local" && report.Provenance.Producer != "github-actions" && report.Provenance.Producer != "other-ci" {
			issues = append(issues, "producer must be local, github-actions or other-ci")
		}
	}
	return issues
}
