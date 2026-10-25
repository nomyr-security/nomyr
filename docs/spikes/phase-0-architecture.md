# Phase 0 architecture spike evidence

**Evidence status:** Partial; not an ADR ratification or product implementation.

**Run date:** 2026-10-25
**Reproduction entrypoint:** [`../../tools/spikes/README.md`](../../tools/spikes/README.md)

The isolated module pins Temporal Go SDK `v1.48.0`, OPA `v1.20.2`, pgx `v5.10.0`, NATS Go client
`v1.54.0`, and NATS server `v2.14.7`. NATS server v2.14.7 declares Go 1.26.0 in its module and
its upstream release notes identify Go 1.26.8 as its build version; the spike compiler is pinned at
Go 1.26.10. These versions were checked against upstream release/tag metadata and module manifests.
OPA tests use an ephemeral RSA signing key and public-only verification; runtime fixtures and the
unsafe secret-canary negative case generate values during each run. No reusable secret is stored in
source, fixtures, logs, or docs.

Direct dependency license review: [Temporal Go SDK](https://github.com/temporalio/sdk-go) is MIT,
[OPA](https://github.com/open-policy-agent/opa) is Apache-2.0,
[pgx](https://github.com/jackc/pgx) is MIT, and
[NATS Go](https://github.com/nats-io/nats.go) and
[NATS Server](https://github.com/nats-io/nats-server/blob/v2.14.7/LICENSE) are Apache-2.0. The
spike module does not vendor, copy, or redistribute dependency source. `go mod verify` and
`govulncheck -test ./...` cover the isolated module graph, including test imports; license and
notice obligations must be revisited if a dependency is copied
or distributed in a product artifact.

| Area | Reproducible evidence | Observed result | Remaining gap |
|---|---|---|---|
| Temporal reconcile-before-retry | `cd tools/spikes && go test ./temporal` | SDK workflow test environment exercises reconciliation after a StartToClose timeout error and before retry; payload safety negative case rejects a generated canary. | Synthetic SDK timeout error only; no Temporal server, persistence, worker-loss, failover, or deployment test. |
| Embedded OPA signed bundle | `cd tools/spikes && go test ./opa` | Embedded evaluation; an ephemeral RSA private key signs, while verification receives only its public key. The reader accepts a trusted signature and rejects tampering and a different public key. | No policy lifecycle, key rotation, authorization integration, or server status-path test. |
| pgx bitemporal SQL | Set `NOMYR_SPIKE_POSTGRES_URL` to a disposable local PostgreSQL database and run `go test ./pgx`. | When run, checks valid-time inclusive/exclusive boundaries and record-time cutoff using actual SQL. | Not run unless the environment variable is present; no retention, correction, tenant isolation, throughput, or production schema review. |
| AWS, Entra, Kubernetes enumeration | Contract inspection: `proto/nhi/connector/v1/connector.proto`. | Contract currently contains only `ConnectorService.Describe`; no enumeration request, response, pagination, identity model, or sanitized fixture schema exists. | Blocked pending a contract change and owner-approved provider fixture boundaries. No provider APIs or accounts were accessed. |
| PostgreSQL / NATS platform behavior | `tools/spikes` workflow starts digest-pinned PostgreSQL 18.6 published only on `127.0.0.1`; the test starts pinned NATS server `v2.14.7` in-process on loopback with an ephemeral port and temporary JetStream store. | Hosted PR run [40188374485](https://github.com/nomyr-security/nomyr/actions/runs/40188374485) succeeded at code SHA `ee8dc3d5417deef563bd8eea324b9f068b932cef`: actual pgx interval-boundary SQL passed against the disposable service; embedded NATS publish/ack/read passed; nested module compiler, audit, and test jobs passed. Run `40187790385` was the earlier root-level npm failure and is superseded. |
| Embedded PostgreSQL OS/architecture matrix | No executable selected distribution. | Not run. No EOL database binary was downloaded or executed. | Maintained distribution, provenance, supported OS/architecture, offline path, and all six native runs remain unknown. |
| Static identity route and Go embedding | `go test -count=1 -v ./staticui` after `make build`; launches the built binary on loopback with a temporary telemetry-disabled config. | Local and hosted (`40188374485`) build-and-route tests passed. `/identities/demo-service-account/?as_of=...` returned HTTP 200 and the expected static route content from embedded assets. No dynamic retrieval or authorization. | One generated static identity path only; no arbitrary identity routing. |
| Demo cold start | `go test -count=1 -v ./staticui` after `make build`. | One local process-start-to-first-HTTP-200 sample was **15.0 ms** on macOS arm64 (Go 1.26.10). | Single sample, not a benchmark or budget; standard CI runner, repeated sampling method, and target remain unspecified. |

## Validation record

Observed on 2026-10-25 with Go 1.26.10 on darwin/arm64:

- `make web-install` — passed; npm reported 0 vulnerabilities in 354 audited packages.
- `make build` — passed; Next.js exported `/identities/demo-service-account/` and all five Go binaries built.
- `make test` — passed, including root Go tests, web tests/typecheck/lint, race tests, root lint/vulnerability checks, and Buf lint. Root tests do not include the nested spike module; it was tested and audited separately below.
- `(cd tools/spikes && go test -count=1 -v ./...)` — passed locally with Go 1.26.10; embedded NATS, asymmetric OPA, Temporal, and static-route tests passed; PostgreSQL skipped locally because no disposable URL was configured. Hosted run `40188374485` then passed PostgreSQL boundary subtests against its disposable service.
- `(cd tools/spikes && go vet ./... && go mod verify && govulncheck -test ./...)` — passed locally with Go 1.26.10; hosted run `40188374485` also verified modules and reported no vulnerabilities with test imports included. The earlier `govulncheck ./...` invocation did not include test-only imports and is not evidence for those imports.
- Static route request — passed through the built Go binary on loopback; the latest single cold-start-to-health response measured at 15.0 ms.
- `git diff --check` — passed.
- Docker daemon was unavailable locally; no local PostgreSQL result is claimed. The initial hosted run `40187790385` failed at root-level `npm ci`; corrected PR-head run `40188374485` at SHA `ee8dc3d5417deef563bd8eea324b9f068b932cef` passed the build, disposable PostgreSQL, nested compiler/vulnerability audit, and all spike tests.
- Connector enumeration, embedded PostgreSQL platform tests, and the six-platform native matrix remain unrun and unresolved.

Hosted PR run `40188374485` is evidence for that PR head only, not for an integrated `main` commit.
The workflow now also runs on pushes to `main`; main-branch evidence remains pending integration.
Skipped local PostgreSQL tests and unsupported/unavailable platforms remain labeled as such.

## Decision status

ADR-0001 and ADR-0012 remain proposed. The results provide partial evidence for review, but the
connector, embedded database platform matrix, cold-start measurement, and hosted service evidence
remain open. No Phase 1 prerequisite or issue is marked complete by this report.
