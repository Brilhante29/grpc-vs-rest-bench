package benchmark

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"sort"
	"time"

	"github.com/Brilhante29/grpc-vs-rest-bench/internal"
)

var (
	SourceCommit = "unknown"
	ImageRef     = "unknown"
	GoSumSHA256  = "unknown"
)

type Methodology struct {
	ContractVersion string `json:"contract_version"`
	RequestsPerRep  int    `json:"requests_per_repetition"`
	WarmupRequests  int    `json:"warmup_requests_per_protocol"`
	Repetitions     int    `json:"repetitions"`
	Concurrency     int    `json:"concurrency"`
	PayloadBytes    int    `json:"payload_bytes"`
	PayloadSHA256   string `json:"payload_sha256"`
	ExecutionOrder  string `json:"execution_order"`
	ConnectionModel string `json:"connection_model"`
}

type Dependency struct {
	Path    string `json:"path"`
	Version string `json:"version"`
	Sum     string `json:"sum,omitempty"`
}

type Provenance struct {
	SourceCommit string       `json:"source_commit"`
	ImageRef     string       `json:"image_ref"`
	ImageDigest  string       `json:"image_digest"`
	GoVersion    string       `json:"go_version"`
	GoSumSHA256  string       `json:"go_sum_sha256"`
	Dependencies []Dependency `json:"dependencies"`
}

type Environment struct {
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	CPUs     int    `json:"num_cpu"`
	Hostname string `json:"hostname"`
}

type Report struct {
	SchemaVersion    string             `json:"schema_version"`
	Project          string             `json:"project"`
	Claim            string             `json:"claim"`
	PrimaryMetric    string             `json:"primary_metric"`
	Unit             string             `json:"unit"`
	GeneratedAt      string             `json:"generated_at"`
	Command          string             `json:"command"`
	ComparabilityKey string             `json:"comparability_key"`
	Methodology      Methodology        `json:"methodology"`
	Provenance       Provenance         `json:"provenance"`
	Environment      Environment        `json:"environment"`
	Results          []ProtocolResult   `json:"results"`
	Comparison       map[string]float64 `json:"comparison"`
	Limitations      []string           `json:"limitations"`
}

func NewReport(cfg Config, payload, command string, results []ProtocolResult) Report {
	payloadDigest := fmt.Sprintf("%x", sha256.Sum256([]byte(payload)))
	keyInput := fmt.Sprintf("%s|rest=http1.1+json|grpc=http2+protobuf|requests=%d|warmup=%d|repetitions=%d|concurrency=%d|payload=%d:%s|order=alternating-sequential",
		internal.ContractVersion, cfg.Requests, cfg.WarmupRequests, cfg.Repetitions, cfg.Concurrency, cfg.PayloadBytes, payloadDigest)
	key := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(keyInput)))
	hostname, _ := os.Hostname()
	report := Report{
		SchemaVersion: "benchmark-report/v2", Project: "grpc-vs-rest-bench",
		Claim:         "Compare REST/HTTP+JSON and gRPC/HTTP2+Protobuf for one unary echo contract",
		PrimaryMetric: "p95_latency_ms_by_protocol", Unit: "milliseconds",
		GeneratedAt: time.Now().UTC().Format(time.RFC3339), Command: command, ComparabilityKey: key,
		Methodology: Methodology{ContractVersion: internal.ContractVersion, RequestsPerRep: cfg.Requests,
			WarmupRequests: cfg.WarmupRequests, Repetitions: cfg.Repetitions, Concurrency: cfg.Concurrency,
			PayloadBytes: cfg.PayloadBytes, PayloadSHA256: payloadDigest,
			ExecutionOrder:  "sequential protocols; first protocol alternates per repetition",
			ConnectionModel: "REST keep-alive pool; one multiplexed gRPC channel; plaintext loopback TCP"},
		Provenance: collectProvenance(), Environment: Environment{OS: runtime.GOOS, Arch: runtime.GOARCH, CPUs: runtime.NumCPU(), Hostname: hostname},
		Results: results,
		Limitations: []string{
			"This isolates one unary in-process echo workload; it is not a universal protocol ranking.",
			"JSON and Protobuf encode the same logical UTF-8 payload but have different wire sizes.",
			"Client and servers share one container host and run without TLS, proxies, persistence, or cross-region latency.",
			"A matching comparability key does not replace matching host resources and Docker runtime conditions.",
		},
	}
	report.computeComparison()
	return report
}

func collectProvenance() Provenance {
	dependencies := []Dependency{}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, module := range info.Deps {
			dependencies = append(dependencies, Dependency{Path: module.Path, Version: module.Version, Sum: module.Sum})
		}
	}
	sort.Slice(dependencies, func(i, j int) bool { return dependencies[i].Path < dependencies[j].Path })
	return Provenance{SourceCommit: SourceCommit, ImageRef: ImageRef,
		ImageDigest: os.Getenv("BENCHMARK_IMAGE_DIGEST"), GoVersion: runtime.Version(),
		GoSumSHA256: GoSumSHA256, Dependencies: dependencies}
}

func (r *Report) computeComparison() {
	if len(r.Results) != 2 {
		return
	}
	byName := map[string]AggregateResult{}
	for _, result := range r.Results {
		byName[result.Protocol] = result.Aggregate
	}
	rest, restOK := byName["REST"]
	grpc, grpcOK := byName["gRPC"]
	if !restOK || !grpcOK || grpc.P95MS == 0 || rest.ThroughputReqPerS == 0 {
		return
	}
	r.Comparison = map[string]float64{
		"rest_over_grpc_p95_ratio":        rest.P95MS / grpc.P95MS,
		"grpc_over_rest_throughput_ratio": grpc.ThroughputReqPerS / rest.ThroughputReqPerS,
	}
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
	if report.SchemaVersion != "benchmark-report/v2" {
		issues = append(issues, "schema_version must be benchmark-report/v2")
	}
	if report.Methodology.Repetitions < 3 || report.Methodology.WarmupRequests <= 0 {
		issues = append(issues, "V2 requires at least 3 repetitions and positive warmup")
	}
	if len(report.Results) != 2 {
		issues = append(issues, "exactly REST and gRPC results are required")
	}
	seenProtocols := map[string]bool{}
	for _, result := range report.Results {
		seenProtocols[result.Protocol] = true
		if len(result.Repetitions) != report.Methodology.Repetitions {
			issues = append(issues, result.Protocol+" repetition count does not match methodology")
		}
		if result.Aggregate.Attempts != result.Aggregate.Successes+result.Aggregate.Failures {
			issues = append(issues, result.Protocol+" attempt accounting is inconsistent")
		}
		if result.Aggregate.P50MS > result.Aggregate.P95MS || result.Aggregate.P95MS > result.Aggregate.P99MS {
			issues = append(issues, result.Protocol+" percentiles are not monotonic")
		}
		if result.Aggregate.Failures != 0 {
			issues = append(issues, result.Protocol+" benchmark must have zero failures")
		}
	}
	if !seenProtocols["REST"] || !seenProtocols["gRPC"] {
		issues = append(issues, "results must include REST and gRPC")
	}
	hex40 := regexp.MustCompile(`^[0-9a-f]{40}$`)
	hex64 := regexp.MustCompile(`^(sha256:)?[0-9a-f]{64}$`)
	if !regexp.MustCompile(`^sha256:[0-9a-f]{64}$`).MatchString(report.ComparabilityKey) {
		issues = append(issues, "comparability_key must be a SHA-256 key")
	}
	if exactProvenance {
		if !hex40.MatchString(report.Provenance.SourceCommit) {
			issues = append(issues, "source_commit must be an exact 40-character Git SHA")
		}
		if report.Provenance.ImageRef == "" || report.Provenance.ImageRef == "unknown" {
			issues = append(issues, "image_ref is required")
		}
		if !hex64.MatchString(report.Provenance.ImageDigest) || len(report.Provenance.ImageDigest) != 71 {
			issues = append(issues, "image_digest must be an exact sha256 image ID")
		}
		if !hex64.MatchString(report.Provenance.GoSumSHA256) {
			issues = append(issues, "go_sum_sha256 is required")
		}
		if len(report.Provenance.Dependencies) == 0 {
			issues = append(issues, "dependency provenance is empty")
		}
	}
	return issues
}
