# ADR-0005 — Tenant isolation

- **Status:** Proposed (security review and isolation tests pending)
- **Date:** 2026-10-23
- **Decision owner:** Project owner (acceptance pending)
- **Scope:** Planned multi-tenant control-plane data access

## Context

The architecture plan identifies PostgreSQL as authoritative state and requires tenant isolation
for API, jobs, projections, and caches. A missed tenant predicate can expose one tenant's data to
another. Application filtering alone is not a sufficient defense for shared relational storage.

## Proposed decision

Bind an authenticated tenant identity to request and job context, require tenant scope in data
access APIs and cache keys, and use PostgreSQL Row-Level Security as a second enforcement layer.
Database roles and transaction setup must fail closed when tenant context is absent or invalid.
Administrative cross-tenant operations require a separately authorized and audited path.

## Consequences

- The context and database policy provide defense in depth but require consistent propagation
  through background work and connection pooling.
- Migrations, maintenance tasks, and privileged roles can bypass ordinary RLS behavior and need
  explicit controls.
- Isolation remains a design proposal until negative tests cover reads, writes, aggregates,
  caches, and background jobs.

## Validation evidence

The plan specifies tenant-bound context and RLS; no implemented database policies or cross-tenant
test results are supplied by this ADR. Acceptance requires a threat review and automated negative
tests, including absent-context and privileged-role cases.

## Revisit when

Deployment tenancy changes, database access patterns cannot preserve tenant context, or tests
show that RLS and application authorization do not provide the required isolation.
