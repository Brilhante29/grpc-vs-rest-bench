# Security Runtime Refresh

Decision: Go 1.26.8, gRPC 1.83.2, x/net 0.59.0 and x/text 0.42.0 replace vulnerable standard-library and HTTP/2 dependencies. Protocol, fixtures and measurement logic are unchanged.

The existing benchmark JSON is historical evidence for its recorded source commit and image digest. It is preserved byte-for-byte, is not a measurement of the refreshed runtime, and must not block security upgrades. A new publication needs a clean source commit, new image identity, and new benchmark evidence; do not overwrite the old result or claim comparability without remeasurement.

Old untracked .portfolio-control/security reports are retained as historical scans, not current-source attestation. Fresh review outputs and residual findings are recorded in the parent go-security-handoff.md. No commit or push is performed by this reviewer.

Version sources: https://go.dev/dl/?mode=json, https://proxy.golang.org/, upstream GitHub releases, and the Terraform registry where applicable.
