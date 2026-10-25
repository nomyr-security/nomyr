# Phase 0 decision and validation register

This register tracks decisions and evidence required by Phase 0. It is a planning record, not a
claim that the listed systems or controls are implemented. ADR proposals remain proposals until
their named owner accepts them. Planned work and unresolved evidence remain explicitly open.

## Architecture decision status

| ADR | Topic | Status | Decision owner | Acceptance or validation still required |
|---|---|---|---|---|
| [0001](../adr/0001-backend-language.md) | Backend language | Proposed | `@nomyr-ai` (owner-designated account; acceptance pending) | Review partial spike evidence; connector enumeration and hosted/runtime/platform gaps remain open. |
| 0002 (local design record; not published) | Durable execution | Accepted (project owner, 2026-09-19) | Project owner | Track Phase 5 deployment and restore considerations; Temporal is not required for Phases 1–4. |
| [0003](../adr/0003-monorepo-and-module-boundaries.md) | Monorepo and module boundaries | Proposed | `@nomyr-ai` (owner-designated account; acceptance pending) | Implement and test forbidden-import matrix and artifact boundaries. |
| [0004](../adr/0004-license-model.md) | License model | Proposed | `@nomyr-ai` (owner-designated account; acceptance pending) | Confirm public license promise and complete pinned dependency reviews; do not infer compatibility from process boundaries. |
| [0005](../adr/0005-tenant-isolation.md) | Tenant isolation | Proposed | `@nomyr-ai` (owner-designated account; acceptance pending) | Design review and cross-tenant negative tests for API, jobs, aggregates, and caches. |
| [0006](../adr/0006-bitemporal-assertion-ledger.md) | Assertion ledger | Proposed | `@nomyr-ai` (owner-designated account; acceptance pending) | `pgx` data-model/write-path spike, temporal edge cases, and retention/privacy review. |
| [0007](../adr/0007-graph-projection.md) | Graph projection | Proposed | `@nomyr-ai` (owner-designated account; acceptance pending) | Representative PostgreSQL traversal benchmark; review Neo4j edition, license, and fit before adoption. |
| [0008](../adr/0008-authentication-and-authorization.md) | Authentication and authorization | Proposed | `@nomyr-ai` (owner-designated account; acceptance pending) | OIDC and policy-engine security review, protocol spike, and deny/unknown/tenant-scope tests. |
| [0009](../adr/0009-event-bus-and-outbox.md) | Event bus and outbox | Proposed | `@nomyr-ai` (owner-designated account; acceptance pending) | Relay/consumer prototype, crash/retry and duplicate tests, tenant routing review. |
| [0010](../adr/0010-connector-transport.md) | Connector transport | Proposed | `@nomyr-ai` (owner-designated account; acceptance pending) | Interoperability and compatibility spike; boundary tests for identity, sanitization, and credential exclusion. |
| [0011](../adr/0011-signing-key-purposes.md) | Signing-key purposes | Proposed | `@nomyr-ai` (owner-designated account; acceptance pending) | Cryptographic and key-lifecycle review; negative verification tests. |
| [0012](../adr/0012-single-binary-embedded-ui-and-lite-mode.md) | Embedded UI and lite mode | Proposed | `@nomyr-ai` (owner-designated account; acceptance pending) | Static identity route and cold-start sample exist; embedded-PostgreSQL platform matrix remains open. No lite capability is implied implemented. |

Issue #18 is the evidence workstream for closing the remaining Phase 0 architecture validation
gaps and supporting final ratification under issue #2. Foundational design and interface work
tracked by issues #4, #5, and #8 may proceed using these proposals as documented working choices;
that dependency does not change their proposed status or count as owner acceptance.

## Phase 0 validation and unknowns

| Seed / validation area | Current state | Evidence needed | Owner / tracking |
|---|---|---|---|
| Temporal Go worker, saga, and reconcile-before-retry (ADR-0001/0002) | Partial; isolated SDK workflow test | Server persistence, worker failover, and deployment evidence remain required. See [spike evidence](../spikes/phase-0-architecture.md). | Issue #18; owner to assign |
| Embedded OPA bundle signing and verification (ADR-0001/0008) | Partial; trusted, tampered, and untrusted-key cases in the isolated Go SDK test | Policy lifecycle, key rotation, authorization integration, and runtime status behavior remain open. See [spike evidence](../spikes/phase-0-architecture.md). | Issue #18; owner to assign |
| `pgx` bitemporal assertion write path (ADR-0001/0006) | Partial; hosted disposable PostgreSQL test passed valid-time/record-time boundaries | Retention, correction, tenancy, performance, and schema review remain open. See [spike evidence](../spikes/phase-0-architecture.md). | Issue #18; owner to assign |
| AWS, Entra, and Kubernetes enumeration through connector contract (ADR-0001/0010) | Blocked by current contract: `Describe` only | Define enumeration contract and sanitized fixture schema before provider-specific adapters or claims. See [spike evidence](../spikes/phase-0-architecture.md). | Issue #18; owner to assign |
| Embedded UI identity detail route and SPA fallback (ADR-0012) | Local and hosted build/export and Go-embed runtime evidence exists for one static identity route | No dynamic retrieval or authorization; arbitrary identity routes remain unproven. See [spike evidence](../spikes/phase-0-architecture.md). | Issue #18; owner to assign |
| `nomyr demo` cold-start budget (ADR-0012) | One local sample measured; no budget established | Agree a runner and repeated measurement method. See [spike evidence](../spikes/phase-0-architecture.md). | Issue #18; owner to assign |
| Embedded PostgreSQL platform behavior (ADR-0012) | Unknown; no distribution selected or run | Maintained distribution, provenance, offline and external-URL behavior, and macOS/Linux/Windows × arm64/amd64 native evidence remain required. See [spike evidence](../spikes/phase-0-architecture.md). | Issue #18; owner to assign |
| Provider API permissions and edition limits | Unknown | Per-provider operation-to-permission/edition matrix and unsupported cases. | Owner and issue to assign |
| Source event retention | Unknown | Retention limits by provider and resulting coverage gaps. | Owner and issue to assign |
| Effective-permission semantics and coverage | Unknown | Provider-specific semantics, evaluation limits, and explicit partial/unknown outcomes. | Owner and issue to assign |
| Non-overlapping rotation strategies | Unknown | Provider-specific overlap, rollback, and reconciliation evidence before Phase 6 implementation. | Owner and issue to assign |
| Legacy health checks | Unknown | Supported health signal, freshness, and unavailable/unknown behavior. | Owner and issue to assign |
| Directory and host access | Unknown | Supported sources, permission boundaries, and collection constraints. | Owner and issue to assign |
| Third-party write support | Unknown | Provider/API coverage, safety constraints, and explicit unsupported operations. | Owner and issue to assign |
| Regional deployment rules | Unknown | Data residency, processing, and cross-region constraints. | Owner and issue to assign |
| Runtime revocation limits | Unknown | Revocation propagation and residual-access measurements for planned runtime controls. | Owner and issue to assign |
| Neo4j edition and license | Unknown; no selection made | Version-specific technical and license review plus benchmark showing need; see ADR-0007. | Owner and issue to assign |
| Reused-component license compatibility | Unknown; no blanket conclusion | Review exact version, license, obligations, distribution, technical fit, and residual risk for each adopted component. | Owner and issue to assign |

## Reused-component review queue

The Phase 0 plan names the components below for review. This queue intentionally records no
version, license conclusion, or technical-fit result where the repository has not established
one. Complete a review against an exact candidate version before adoption or distribution.

| Component | Version | License / obligations | Technical fit | Unresolved risk | Status |
|---|---|---|---|---|---|
| Cartography | Not pinned | Not reviewed | Not reviewed | Compatibility and integration boundary unknown | Pending |
| Prowler | Not pinned | Not reviewed | Not reviewed | Compatibility and integration boundary unknown | Pending |
| TruffleHog | Not pinned | Not reviewed | Not reviewed | Plan flags AGPL-3.0; compatibility and distribution review required | Pending |
| BloodHound CE | Not pinned | Not reviewed | Not reviewed | Edition and integration boundary unknown | Pending |
| PMapper | Not pinned | Not reviewed | Not reviewed | Exact component/version and obligations unknown | Pending |
| OpenBao | Not pinned | Not reviewed | Not reviewed | Version-specific obligations and adapter fit unknown | Pending |
| SPIRE | Not pinned | Not reviewed | Not reviewed | Version-specific obligations and integration fit unknown | Pending |
| Warden | Not pinned | Not reviewed | Not reviewed | Exact project/version, obligations, and fit must be identified | Pending |
| Neo4j (only if reconsidered) | Not selected | Edition/license review required | Workload need unproven | Community/enterprise terms and operational fit unknown | Deferred / pending evidence |
| Other adopted dependencies | Not pinned here | Not reviewed here | Not reviewed here | Complete inventory and notice obligations required | Pending |

## Register maintenance

For each evidence item, record the test or review date, exact versions and environment, result,
known limits, and a durable evidence link. A failed or incomplete test remains failed or unknown;
do not convert missing evidence into acceptance. Keep credentials, reusable secret values,
customer data, and private infrastructure details out of this register and linked public evidence.
