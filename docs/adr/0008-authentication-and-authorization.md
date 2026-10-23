# ADR-0008 — Authentication and authorization

- **Status:** Proposed (security design and protocol validation pending)
- **Date:** 2026-10-23
- **Decision owner:** Project owner (acceptance pending)
- **Scope:** Planned human authentication and product authorization

## Context

The product plans a browser console and API with tenant-scoped access. UI filtering cannot enforce
authorization. The plans name OIDC for authentication and OPA/Rego for policy evaluation, while
the current repository scaffold does not implement these product services.

## Proposed decision

Use a standards-based OIDC integration for human authentication and centralize authorization in
the API and shared policy enforcement points. Evaluate policy with an embedded OPA-compatible
engine and versioned policy bundles. Require deny-by-default behavior, tenant and resource scope
checks, and auditable policy decisions. The browser may hide unavailable actions for usability but
must never be the authorization boundary.

## Consequences

- A shared authorization path can reduce inconsistent checks across routes and jobs, but must be
  explicitly propagated to exports, caches, graph traversals, and background tasks.
- Identity-provider configuration, session security, policy distribution, and emergency access
  create operational requirements.
- No OIDC flow, policy bundle lifecycle, or authorization enforcement is implied to exist today.

## Validation evidence

Plans identify OIDC and embedded OPA as intended choices. No threat-model review, OIDC protocol
spike, policy evaluation test, or cross-tenant authorization result is recorded here. Acceptance
requires those reviews and failure-path tests, including unknown policy and unavailable identity
provider behavior.

## Revisit when

Protocol or policy requirements cannot be met safely by the proposed components, or measured
availability and latency fail service goals.
