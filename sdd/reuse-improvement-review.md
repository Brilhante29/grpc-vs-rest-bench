# Reuse Improvement Review

## Project

#15 grpc-vs-rest-bench

## Date

2026-07-20

## Findings

| Finding | Decision | Patch |
|---|---|---|
| CI benchmark smoke could overwrite a stable publication artifact | patch_now | Kit skills and this workflow now require a runner-temporary smoke path |
| Custom reports can call themselves V2 while violating the shared schema | patch_now | Replaced the custom envelope with the shared V2 contract and strict semantic validation |
| Protocol comparison needs paired ratio metrics and parity failures | patch_now | Added paired p95/throughput ratios and a zero-failure publication gate |

## Final Gate

- [x] Reusable improvements were patched or recorded.
- [x] Project-specific implementation was not moved into the kit.
- [x] Validation reflects the project.yaml status and benchmark result path.
