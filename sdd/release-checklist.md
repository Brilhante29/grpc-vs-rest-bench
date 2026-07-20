# Release Checklist

## Pre-release

- [x] go build ./... passes
- [x] go test -v ./... passes
- [x] Docker build passes
- [x] Benchmark JSON written to benchmarks/results/
- [x] README has project number and benchmark table
- [x] REFERENCES.md complete
- [x] project.yaml status: benchmarked
- [x] All SDD documents filled
- [x] No secrets committed
- [x] .gitignore correct

## Post-release

- [ ] Tag release with v0.1.0
- [ ] Push to GitHub
- [ ] Verify CI passes
