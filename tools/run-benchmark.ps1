param(
  [ValidateRange(1, 10000000)][int]$Requests = 1000,
  [ValidateRange(1, 16777216)][int]$PayloadBytes = 256,
  [ValidateRange(1, 10000)][int]$Concurrency = 10,
  [ValidateRange(1, 1000000)][int]$WarmupRequests = 100,
  [ValidateRange(3, 100)][int]$Repetitions = 3,
  [string]$ResultsDir = ""
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot

Push-Location -LiteralPath $root
try {
  $sourceCommit = (git rev-parse HEAD).Trim()
  if ($LASTEXITCODE -ne 0 -or $sourceCommit -notmatch "^[0-9a-f]{40}$") {
    throw "Cannot resolve an exact source commit"
  }
  if (git status --porcelain) {
    throw "Benchmark provenance requires a clean worktree; commit implementation changes first"
  }

  $shortCommit = $sourceCommit.Substring(0, 12)
  $env:SOURCE_COMMIT = $sourceCommit
  $env:IMAGE_REF = "grpc-vs-rest-bench:$shortCommit"
  $env:GO_SUM_SHA256 = (Get-FileHash -Algorithm SHA256 -LiteralPath "go.sum").Hash.ToLowerInvariant()
  $env:BENCHMARK_REQUESTS = $Requests.ToString()
  $env:BENCHMARK_PAYLOAD_BYTES = $PayloadBytes.ToString()
  $env:BENCHMARK_CONCURRENCY = $Concurrency.ToString()
  $env:BENCHMARK_WARMUP_REQUESTS = $WarmupRequests.ToString()
  $env:BENCHMARK_REPETITIONS = $Repetitions.ToString()
  $env:BENCHMARK_COMMAND = "pwsh ./tools/run-benchmark.ps1 -Requests $Requests -PayloadBytes $PayloadBytes -Concurrency $Concurrency -WarmupRequests $WarmupRequests -Repetitions $Repetitions"
  if ([string]::IsNullOrWhiteSpace($ResultsDir)) {
    $ResultsDir = Join-Path $root "benchmarks/results"
  } elseif (-not [System.IO.Path]::IsPathRooted($ResultsDir)) {
    $ResultsDir = Join-Path $root $ResultsDir
  }
  New-Item -ItemType Directory -Force -Path $ResultsDir | Out-Null
  $env:BENCHMARK_RESULTS_DIR = (Resolve-Path -LiteralPath $ResultsDir).Path

  docker compose build benchmark
  if ($LASTEXITCODE -ne 0) { throw "Docker image build failed" }

  $env:BENCHMARK_IMAGE_DIGEST = (docker image inspect $env:IMAGE_REF --format "{{.Id}}").Trim()
  if ($LASTEXITCODE -ne 0 -or $env:BENCHMARK_IMAGE_DIGEST -notmatch "^sha256:[0-9a-f]{64}$") {
    throw "Cannot resolve the benchmark image digest"
  }

  docker compose up --abort-on-container-exit --exit-code-from benchmark benchmark
  $benchmarkExit = $LASTEXITCODE
  if ($benchmarkExit -ne 0) { throw "Benchmark failed with exit code $benchmarkExit" }

  docker run --rm -v "${env:BENCHMARK_RESULTS_DIR}:/results:ro" $env:IMAGE_REF bench-client -validate /results/benchmark-result.json
  if ($LASTEXITCODE -ne 0) { throw "Benchmark report validation failed" }
} finally {
  $cleanupPreference = $ErrorActionPreference
  $ErrorActionPreference = "SilentlyContinue"
  docker compose down --remove-orphans *> $null
  $ErrorActionPreference = $cleanupPreference
  $global:LASTEXITCODE = 0
  Pop-Location
}
