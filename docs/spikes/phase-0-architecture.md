# Phase 0 architecture spike evidence

**Evidence status:** Partial; not an ADR ratification or product implementation.

**Run date:** 2026-10-25
**Reproduction entrypoint:** [`../../tools/spikes/README.md`](../../tools/spikes/README.md)

The isolated module pins Temporal Go SDK `v1.48.0`, OPA `v1.20.2`, pgx `v5.10.0`, and NATS Go
client `v1.54.0`. These versions were checked against their upstream GitHub release/tag metadata
and module manifests. The OPA bundle test uses an ephemeral in-memory signing key. Runtime fixtures
and the unsafe secret-canary negative case generate values during each run; no reusable secret is
stored in source, fixtures, logs, or docs.

Direct dependency license review: [Temporal Go SDK](https://github.com/temporalio/sdk-go) is MIT,
[OPA](https://github.com/open-policy-agent/opa) is Apache-2.0,
[pgx](https://github.com/jackc/pgx) is MIT, and
[NATS Go](https://github.com/nats-io/nats.go) is Apache-2.0. The spike module does not vendor,
copy, or redistribute dependency source. `go mod verify` and `govulncheck ./...` cover the
isolated module graph; license and notice obligations must be revisited if a dependency is copied
or distributed in a product artifact.

| Area | Reproducible evidence | Observed result | Remaining gap |
|---|---|---|---|
| Temporal reconcile-before-retry | `cd tools/spikes && go test ./temporal` | SDK workflow test environment exercises reconciliation after a StartToClose timeout error and before retry; payload safety negative case rejects a generated canary. | Synthetic SDK timeout error only; no Temporal server, persistence, worker-loss, failover, or deployment test. |
| Embedded OPA signed bundle | `cd tools/spikes && go test ./opa` | Embedded evaluation; actual OPA bundle reader accepts a trusted signature and rejects tampering and a different verification key. | No policy lifecycle, key rotation, authorization integration, or server status-path test. |
| pgx bitemporal SQL | Set `NOMYR_SPIKE_POSTGRES_URL` to a disposable local PostgreSQL database and run `go test ./pgx`. | When run, checks valid-time inclusive/exclusive boundaries and record-time cutoff using actual SQL. | Not run unless the environment variable is present; no retention, correction, tenant isolation, throughput, or production schema review. |
| AWS, Entra, Kubernetes enumeration | Contract inspection: `proto/nhi/connector/v1/connector.proto`. | Contract currently contains only `ConnectorService.Describe`; no enumeration request, response, pagination, identity model, or sanitized fixture schema exists. | Blocked pending a contract change and owner-approved provider fixture boundaries. No provider APIs or accounts were accessed. |
| PostgreSQL / NATS platform behavior | `tools/spikes` workflow starts digest-pinned PostgreSQL 18.6 and NATS 2.15.1 JetStream services, then runs actual `pgx` SQL and JetStream publish/ack/read tests. | Docker daemon was unavailable locally; these service tests skipped locally. The new hosted workflow has not yet run, so hosted outcomes are pending. NATS is a container service, not an embedded server. |
| Embedded PostgreSQL OS/architecture matrix | No executable selected distribution. | Not run. No EOL database binary was downloaded or executed. | Maintained distribution, provenance, supported OS/architecture, offline path, and all six native runs remain unknown. |
| Static identity route and Go embedding | `go test -count=1 -v ./staticui` after `make build`; launches the built binary on loopback with a temporary telemetry-disabled config. | Static export and all Go binaries built. `/identities/demo-service-account/?as_of=...` returned HTTP 200 and the expected static route content from embedded assets. No dynamic retrieval or authorization. | One generated static identity path only; no arbitrary identity routing. Hosted build-and-route run is pending. |
| Demo cold start | `go test -count=1 -v ./staticui` after `make build`. | One local process-start-to-first-HTTP-200 sample was **15.1 ms** on macOS arm64 (Go 1.26.10). | Single sample, not a benchmark or budget; standard CI runner, repeated sampling method, and target remain unspecified. |

## Validation record

Observed on 2026-10-25 with Go 1.26.10 on darwin/arm64:

- `make web-install` — passed; npm reported 0 vulnerabilities in 354 audited packages.
- `make build` — passed; Next.js exported `/identities/demo-service-account/` and all five Go binaries built.
- `make test` — passed, including root Go tests, web tests/typecheck/lint, race tests, root lint/vulnerability checks, and Buf lint. Root tests do not include the nested spike module; it was tested and audited separately below.
- `(cd tools/spikes && go test -count=1 -v ./...)` — Temporal, OPA, and static-route tests passed; PostgreSQL and NATS tests skipped because no explicit local service URLs were set.
- `(cd tools/spikes && go vet ./... && go mod verify && govulncheck ./...)` — passed; all modules verified and govulncheck reported no vulnerabilities in the isolated graph.
- Static route request — passed through the built Go binary on loopback; the latest single cold-start-to-health response measured at 15.1 ms.
- `git diff --cached --check` — passed.
- Docker daemon was unavailable locally; no PostgreSQL or NATS integration result is claimed. No hosted workflow result exists yet.
- Connector enumeration, embedded PostgreSQL platform tests, and the six-platform native matrix remain unrun and unresolved.

CI workflow configuration is not itself evidence that a hosted run completed successfully. Skipped
service tests and unavailable platforms remain labeled as such.

## Decision status

ADR-0001 and ADR-0012 remain proposed. The results provide partial evidence for review, but the
connector, embedded database platform matrix, cold-start measurement, and hosted service evidence
remain open. No Phase 1 prerequisite or issue is marked complete by this report.
