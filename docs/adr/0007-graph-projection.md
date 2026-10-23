# ADR-0007 — Graph projection

- **Status:** Proposed (workload and compatibility validation pending)
- **Date:** 2026-10-23
- **Decision owner:** Project owner (acceptance pending)
- **Scope:** Rebuildable graph query and projection design

## Context

Identity relationships support ownership, entitlement, and path analysis. The specification
describes graph capability but the technology plan defers a graph-specific database until
multi-hop workload evidence justifies it. Assertions remain authoritative; a graph is derived.

## Proposed decision

Define a versioned graph query contract and begin with a PostgreSQL-backed, rebuildable
projection. Keep projection inputs traceable to authoritative assertions and expose incomplete
coverage explicitly. Consider a Neo4j-compatible implementation only after measured multi-hop
workloads justify the operational and licensing cost. Do not make a graph database authoritative.

## Consequences

- Initial deployments avoid another required database and preserve the option to replace the
  projection behind a stable contract.
- PostgreSQL traversal performance and projection freshness may limit workload size and must be
  measured.
- Neo4j edition, license, deployment, and compatibility questions remain open until a specific
  product/version review and workload benchmark exist.

## Validation evidence

The architecture and technology plans propose PostgreSQL-first and defer Neo4j. No benchmark,
query contract conformance result, or edition/license review is recorded here. Acceptance requires
representative traversal measurements and documented component review before selecting another
backend.

## Revisit when

Representative multi-hop workloads exceed the PostgreSQL projection's measured service goals, or
projection semantics cannot meet freshness and explainability requirements.
