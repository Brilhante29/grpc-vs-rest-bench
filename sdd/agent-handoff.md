# Agent Handoff

## Current State

Project #15 is implementation- and evidence-complete. The canonical V2 result was generated from clean source commit `4442f1be7961830dde793c980a3dad56f7841349` and must remain separate from CI smoke output.

## Verified Behavior

- one pure `EchoLogic` serves real REST and gRPC adapters
- clients assert request ID, payload, and SHA-256 parity
- three measured repetitions alternate protocol order after warmup
- common V2 report contains UUID, samples, digests, exact source/image/artifact provenance, and zero failures
- PowerShell cleanup does not turn Docker stderr into a false benchmark failure
- CI writes regenerated evidence under `RUNNER_TEMP`

## Invariants For The Next Agent

1. Do not overwrite `benchmarks/results/benchmark-result.json` in CI.
2. Do not claim one protocol universally wins.
3. Keep `internal.EchoLogic` independent from transport and generated types.
4. Change the comparability key whenever workload or connection semantics change.
5. Feed reusable harness improvements back to `portfolio-reuse-kit`.

## Verification

```powershell
pwsh ./tools/validate-project.ps1
```

For publication, verify GitHub Actions against the exact pushed head; do not infer success from an older green run.
