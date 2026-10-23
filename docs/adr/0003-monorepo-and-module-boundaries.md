# ADR-0003 — Monorepo and module boundaries

- **Status:** Proposed (owner acceptance and boundary checks pending)
- **Date:** 2026-10-23
- **Decision owner:** Project owner (acceptance pending)
- **Scope:** Phase 0 repository and build boundaries; does not authorize later-phase capabilities

## Context

The Phase 0 plan places contracts, services, connectors, actions, SDKs, content, and the web
application in one repository. Their licensing, permissions, and release needs differ. The four
runtime planes must remain separate even when their source is colocated. The friendly `nomyr`
binary must not link `internal/custody` or `actions/*`.

## Proposed decision

Keep a monorepo with explicit package, binary, and licensing boundaries. Enforce dependency
boundaries with Go package visibility, a CI import-policy matrix, and separate build artifacts
for independently permissioned processes. Treat the matrix as a required guardrail, not proof
that it is already implemented. Public contracts and SDKs remain separate from server
implementation directories.

## Consequences

- Shared contracts and coordinated changes are easier to review and version.
- A monorepo alone does not provide runtime isolation; process identities, deployment policy,
  package checks, and artifact review remain necessary.
- Until the import-policy check is implemented and tested, compile-time plane isolation is an
  unresolved risk.

## Validation evidence

No dependency-matrix implementation or isolation test is recorded as evidence for this ADR.
The repository layout and Phase 0 plan are design inputs, not validation. Add an import-matrix
check and verify each binary's dependency graph before acceptance.

## Revisit when

The repository structure prevents independent licensing or release obligations from being
enforced, or the build system cannot reliably reject forbidden cross-plane imports.
