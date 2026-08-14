# Release Checklist

- [x] `go test -race -count=1 ./...` is part of CI.
- [x] `go vet ./...` and `go build ./...` pass locally in Docker.
- [x] Real REST and gRPC paths pass semantic parity tests.
- [x] Docker build and Compose benchmark pass.
- [x] Common V2 artifact has three samples and exact provenance.
- [x] README opens with measured numbers and limitations.
- [x] CI smoke writes outside the canonical publication path.
- [x] SDD, references, reuse review, and no-secret default are complete.

The external release controller must confirm the CI run belongs to the exact pushed head before counting the repository as published.
