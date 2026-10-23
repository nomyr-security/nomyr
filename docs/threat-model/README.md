<!-- SPDX-FileCopyrightText: 2026 Nomyr technologies -->
<!-- SPDX-License-Identifier: AGPL-3.0-only -->

# Threat model

This is a source document for the planned Nomyr product threat model, seeded from
specification §12.1. It records risks and verification work; it does not claim
that planned product controls are implemented. The Phase 0 build and workflow
risks are marked as current security work where applicable. Human owners have
not been assigned. The listed accountable roles and issue links are routing
targets, not named approvals.

| Threat | Planned product control | Verification requirement | Accountable role | Tracking |
|---|---|---|---|---|
| Compromised collection credential | Minimal source scopes, independent identity, revocation, read/write separation | Stolen collector identity cannot execute actions or enroll a runner | [Collection-plane maintainers (unassigned)](#accountable-roles) | [#17](https://github.com/nomyr-security/nomyr/issues/17); future collection-plane security issue required |
| Cross-tenant object or graph access | Tenant binding throughout query, cache, export, search, and async jobs | Negative tests cover IDs, aggregates, paths, and evidence downloads | [Control-plane maintainers (unassigned)](#accountable-roles) | Future tenant-isolation issue required |
| Malicious source content or prompt injection | Treat names, logs, documents, and prompts as data without tool authority | Seeded instructions cannot alter approvals, endpoints, or policy | [Product security (unassigned)](#accountable-roles) | Future ingestion and agent security issues required |
| Forged or replayed approval | Strong approver identity, plan hash/version, scope, expiry, and nonce | Modified, expired, or replayed approval is rejected | [Governance maintainers (unassigned)](#accountable-roles) | Future approval security issue required |
| Provider timeout after successful write | Reconciliation before retry, adapter idempotency, and target locks | Unknown outcomes do not cause duplicate identities or key deletion | [Action-plane maintainers (unassigned)](#accountable-roles) | Future action-plane security issue required |
| Credential leakage through diagnostics | Redaction, vault references, safe errors, and crash/log controls | Seeded secret values are absent from APIs, storage, traces, workflow history, and exports | [Security engineering (unassigned)](#accountable-roles) | [#17](https://github.com/nomyr-security/nomyr/issues/17) covers repository scanning; runtime tests remain future work |
| Malicious connector endpoint or SSRF | Endpoint allowlists, destination validation, redirect/DNS controls, and scoped egress | Metadata services and unauthorized internal destinations are blocked | [Connector maintainers (unassigned)](#accountable-roles) | Future connector security issue required |
| Malicious or compromised plugin | Signed/pinned artifacts, SBOM, privilege manifest, sandbox, and staged rollout | Plugin cannot exceed declared network, filesystem, or credential scope | [Supply-chain security (unassigned)](#accountable-roles) | [#17](https://github.com/nomyr-security/nomyr/issues/17) covers repository artifacts; plugin controls are future work |
| Privileged insider bypass | Separation of duties, scoped approvals, strong reauthentication, and independent audit export | Administrator cannot silently act as another approver | [Governance maintainers (unassigned)](#accountable-roles) | Future governance security issue required |
| False ownership or poisoned metadata | Candidate state, authoritative evidence, reviewer confirmation, and history | Spoofed tags or CODEOWNERS cannot grant action rights | [Identity graph maintainers (unassigned)](#accountable-roles) | Future ownership security issue required |
| Runtime confused deputy | Bind subject, tenant, audience, resource, and delegation; authorize downstream | Token for one service cannot authorize another | [Runtime-access maintainers (unassigned)](#accountable-roles) | Future runtime-access security issue required |
| Denial of service or noisy tenant | Rate, size, and depth budgets, per-tenant queues, and backpressure | Expensive graph/export work does not starve access decisions | [Reliability engineering (unassigned)](#accountable-roles) | Future scale/security issue required |
| Stale policy or restored state | Versioned signed policy, expiry, reconciliation, and execution freeze | Restore cannot replay an old approved destructive plan | [Policy and execution maintainers (unassigned)](#accountable-roles) | Future restore and execution issues required |
| Bypassed access gateway | Resource-side controls and alternate-path testing | Protected status is removed when uncontrolled access remains | [Runtime-access maintainers (unassigned)](#accountable-roles) | Future gateway security issue required |
| Data loss or evidence alteration | Encrypted backups, retention, signed export manifests, and independent copies | Restore and integrity checks detect missing or changed evidence | [Storage maintainers (unassigned)](#accountable-roles) | Future resilience/security issue required |
| Malicious contribution or compromised build dependency | Read-only untrusted PR workflows, pinned automation dependencies, secret/dependency scanning, isolated builds, and trusted-context attestations | Untrusted PR code receives no signing or write token; artifact provenance identifies the trusted workflow | [Repository security maintainers (unassigned)](#accountable-roles) | [#17](https://github.com/nomyr-security/nomyr/issues/17) |
| Runtime plane accidentally gains cross-plane authority | Separate binaries and permission identities with explicit dependency boundaries | Dependency/build checks show that experience and collection artifacts do not contain action or custody implementations | [Architecture maintainers (unassigned)](#accountable-roles) | [#17](https://github.com/nomyr-security/nomyr/issues/17) container boundaries; future architecture test issue required |

## Status and ownership

Roles are intentionally unassigned until maintainers establish human accountability.
No product control in this table should be read as implemented without matching
source, tests, and an accepted issue. Track verification gaps as separate issues
when a concrete phase and owner are known. Revisit this document as phase work
adds control implementation and tests.

## Phase 0 artifact boundary

Container images built by issue #17 are test artifacts for the current
entrypoints. The API, worker, collector, and runner entrypoints still report
`not-implemented`; their image definitions do not add provider access,
PostgreSQL persistence, or action behavior. Different non-root UIDs express
container defaults for the four planes but do not replace deployment identity,
network policy, or runtime permission configuration. Pull-request builds receive
read-only repository permissions; keyless signing and GitHub artifact
attestations run only for the default branch. Availability of GitHub OIDC and
attestation account features must be verified in the repository settings and CI.

## Accountable roles

Every role named in the table is an accountability placeholder, not an assigned
person. The repository's [CODEOWNERS](../../.github/CODEOWNERS) routes review to
`@nomyr-ai`; it does not establish human ownership of these product controls.
