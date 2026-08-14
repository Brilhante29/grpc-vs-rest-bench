param(
  [switch]$SkipDocker
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$failures = New-Object System.Collections.Generic.List[string]

function Add-Failure {
  param([string]$Message)
  $script:failures.Add($Message)
}

function Require-File {
  param([string]$RelativePath)
  $path = Join-Path $root $RelativePath
  if (-not (Test-Path -LiteralPath $path -PathType Leaf)) {
    Add-Failure "Missing file: $RelativePath"
  }
}

function Invoke-Checked {
  param(
    [string]$Label,
    [scriptblock]$Command
  )
  & $Command
  $exitCode = $LASTEXITCODE
  if ($exitCode -ne 0) {
    Add-Failure "$Label failed with exit code $exitCode"
  }
  $global:LASTEXITCODE = 0
}

$requiredFiles = @(
  "README.md",
  "compose.yaml",
  "project.yaml",
  "REFERENCES.md",
  "AGENTS.md",
  "sdd/spec.md",
  "sdd/benchmark-plan.md",
  "sdd/architecture-decision.md",
  "sdd/technical-decision.md",
  "sdd/agent-handoff.md",
  "sdd/reuse-improvement-review.md"
  "tools/validate-go-container.sh"
)
foreach ($file in $requiredFiles) { Require-File $file }

$readmePath = Join-Path $root "README.md"
if (Test-Path -LiteralPath $readmePath -PathType Leaf) {
  $readme = Get-Content -Raw -LiteralPath $readmePath
  if ($readme -notmatch '^# #15 grpc-vs-rest-bench') {
    Add-Failure "README must open with project number and name"
  }
  foreach ($evidence in @("3.239 ms", "3.796 ms", "7,752.31 req/s", "6,036.66 req/s", "rest_over_grpc_p95_ratio")) {
    if ($readme -notmatch [regex]::Escape($evidence)) {
      Add-Failure "README is missing publication evidence: $evidence"
    }
  }
}

$projectPath = Join-Path $root "project.yaml"
if (Test-Path -LiteralPath $projectPath -PathType Leaf) {
  $project = Get-Content -Raw -LiteralPath $projectPath
  if ($project -notmatch '(?m)^status: published\r?$') {
    Add-Failure "project.yaml status must be published"
  }
  if ($project -notmatch '(?m)^  primary_metric: rest_over_grpc_p95_ratio\r?$') {
    Add-Failure "project.yaml primary metric must match the V2 comparison artifact"
  }
}

$workflowPath = Join-Path $root ".github/workflows/ci.yml"
if (Test-Path -LiteralPath $workflowPath -PathType Leaf) {
  $workflow = Get-Content -Raw -LiteralPath $workflowPath
  if ($workflow -notmatch 'RUNNER_TEMP' -or $workflow -notmatch 'ResultsDir') {
    Add-Failure "CI benchmark smoke must write outside the canonical publication path"
  }
}

$reuseReviewPath = Join-Path $root "sdd/reuse-improvement-review.md"
if (Test-Path -LiteralPath $reuseReviewPath -PathType Leaf) {
  $reuseReview = Get-Content -Raw -LiteralPath $reuseReviewPath
  if ($reuseReview -match "<id>|<project-name>") {
    Add-Failure "Reuse improvement review still contains template placeholders"
  }
  if ($reuseReview.Contains('|  | `patch_now|backlog|reject` |')) {
    Add-Failure "Reuse improvement review still contains the blank template finding row"
  }
  $requiredFinalGatePatterns = @(
    "(?m)^- \[x\] Reusable improvements were patched or recorded\.\r?$",
    "(?m)^- \[x\] Project-specific implementation was not moved into the kit\.\r?$",
    "(?m)^- \[x\] Validation reflects .+\.\r?$"
  )
  foreach ($pattern in $requiredFinalGatePatterns) {
    if ($reuseReview -notmatch $pattern) {
      Add-Failure "Reuse improvement review final gate is incomplete: $pattern"
    }
  }
}

$benchmarkFiles = @()
$benchmarkDir = Join-Path $root "benchmarks/results"
if (Test-Path -LiteralPath $benchmarkDir -PathType Container) {
  $benchmarkFiles = @(Get-ChildItem -LiteralPath $benchmarkDir -Filter *.json -File)
}
if ($benchmarkFiles.Count -eq 0) {
  Add-Failure "Missing benchmark JSON under benchmarks/results"
}

Push-Location -LiteralPath $root
try {
  $go = Get-Command go -ErrorAction SilentlyContinue
  if ($go) {
    foreach ($file in $benchmarkFiles) {
      Invoke-Checked "benchmark JSON validation: $($file.Name)" { go run ./cmd/bench-client -validate $file.FullName }
    }
    Invoke-Checked "Go format" {
      $unformatted = @(gofmt -l .)
      if ($unformatted.Count -gt 0) { Write-Host ($unformatted -join [Environment]::NewLine); $global:LASTEXITCODE = 1 }
    }
    Invoke-Checked "Go tests" { go test ./... }
    Invoke-Checked "Go vet" { go vet ./... }
    Invoke-Checked "Go build" { go build ./... }
  } elseif (-not $SkipDocker) {
    Invoke-Checked "containerized Go validation" { docker run --rm -v "${root}:/src" -w /src golang:1.26.5-alpine3.24 sh /src/tools/validate-go-container.sh }
  } else {
    Add-Failure "Go toolchain is unavailable and Docker validation was skipped"
  }

  if (Test-Path -LiteralPath (Join-Path $root "src") -PathType Container) {
    $previousPythonPath = $env:PYTHONPATH
    $srcPath = Join-Path $root "src"
    if ($previousPythonPath) {
      $env:PYTHONPATH = $srcPath + [System.IO.Path]::PathSeparator + $previousPythonPath
    } else {
      $env:PYTHONPATH = $srcPath
    }
    Invoke-Checked "python compile src" { python -m compileall -q (Join-Path $root "src") }
    if (Test-Path -LiteralPath (Join-Path $root "tests") -PathType Container) {
      Invoke-Checked "python compile tests" { python -m compileall -q (Join-Path $root "tests") }
      Invoke-Checked "python unittest" { python -m unittest discover -s (Join-Path $root "tests") -v }
    }
    $env:PYTHONPATH = $previousPythonPath
  }
} finally {
  Pop-Location
}

$legacy = ("ro" + "che" + "do")
$patterns = @($legacy, ($legacy.Substring(0,1).ToUpper() + $legacy.Substring(1)))
$searchFiles = Get-ChildItem -Path $root -Recurse -File | Where-Object {
  $normalized = $_.FullName -replace "\\", "/"
  $normalized -notmatch "/.git/" -and
  $normalized -notmatch "/data/runtime/" -and
  $_.Extension -in @(".md", ".yaml", ".yml", ".json", ".ps1", ".py", ".js", ".ts", ".tsx", ".go", ".kt", ".java")
}
$forbidden = Select-String -Path $searchFiles.FullName -Pattern $patterns -SimpleMatch -ErrorAction SilentlyContinue
if ($forbidden) {
  Add-Failure "Forbidden legacy project nickname found"
}
$mojibakePatterns = @(
  [string][char]0x00C3,
  ([string][char]0x00E2 + [string][char]0x20AC)
)
$mojibake = Select-String -Path $searchFiles.FullName -Pattern $mojibakePatterns -SimpleMatch -ErrorAction SilentlyContinue
if ($mojibake) {
  Add-Failure "Mojibake found in publication files"
}

if (-not $SkipDocker -and (Test-Path -LiteralPath (Join-Path $root "Dockerfile") -PathType Leaf)) {
  $imageName = (Split-Path -Leaf $root).ToLowerInvariant()
  Invoke-Checked "docker build" { docker build -t $imageName $root | Out-Null }
  Push-Location -LiteralPath $root
  try {
    Invoke-Checked "Docker Compose config" { docker compose config --quiet }
  } finally {
    Pop-Location
  }
}

if ($failures.Count -gt 0) {
  # Write-Error is a terminating error while $ErrorActionPreference is "Stop",
  # so emitting the list through it aborts on the first entry and hides every
  # remaining failure. Report the complete list on the success stream instead.
  Write-Host "portfolio project validation failed with $($failures.Count) issue(s):"
  foreach ($failure in $failures) {
    Write-Host "  - $failure"
    if ($env:GITHUB_ACTIONS -eq "true") {
      Write-Host "::error::$failure"
    }
  }
  exit 1
}

Write-Host "portfolio project validation passed"
