# ADR-0001 — Backend language

- **Status:** Proposed; owner acceptance pending
- **Date:** 2026-10-25
- **Decision owner:** `@nomyr-ai` (owner-designated maintainer account)
- **Scope:** Backend implementation language for planned Nomyr services

## Context

The project needs a backend language that can support the planned API, connector boundaries,
durable workflows, policy evaluation, and a distributable command-line experience. The current
Phase 0 repository uses Go, but that implementation history does not itself constitute an accepted
architecture decision.

## Proposed decision

Continue evaluating Go as the backend language. Keep the proposed choice open until the language
spikes and their remaining acceptance gaps are reviewed by the decision owner. Keep the Go spike
module under `tools/spikes`; it is not a runtime dependency or a product capability.

## Evidence and limitations

The isolated evidence and its reproducible commands are maintained in
[`../spikes/phase-0-architecture.md`](../spikes/phase-0-architecture.md). Temporal's official Go
SDK test environment, embedded OPA bundle verification with an ephemeral RSA signer and public-key
only verification, and in-process NATS JetStream tests have local test coverage. A `pgx` PostgreSQL
SQL test is authored but requires a disposable database; the hosted workflow provisions one on
loopback. The connector protobuf does not define enumeration; cloud-provider execution was not
attempted. These results do not establish provider support, production availability, or the complete
suitability of Go. See the spike report for exact workflow outcomes and the test-inclusive
vulnerability scan.

## Consequences

- The experiments add dependencies only to a nested Go module and do not change the root module or
  any Nomyr binary.
- Further evaluation is required for cross-platform distribution, connector interoperability,
  production deployment, and the explicit gaps recorded in the spike report.
- This ADR remains proposed until the owner records a decision.

## Revisit when

The remaining spikes, build targets, or operational requirements show material limitations in the
proposed Go direction.
