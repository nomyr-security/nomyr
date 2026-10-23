# ADR-0006 — Bitemporal assertion ledger

- **Status:** Proposed (data-model spike and retention review pending)
- **Date:** 2026-10-23
- **Decision owner:** Project owner (acceptance pending)
- **Scope:** Planned authoritative record of observed identity-security assertions

## Context

Provider observations can be incomplete, delayed, revised, or contradictory. The specification
calls for evidence-backed assertions with both valid time and record time. Replacing prior rows
with a single current-state value would lose the history needed to explain what was known and
when.

## Proposed decision

Represent observations as immutable, tenant-scoped assertions with source, evidence reference,
collection mode, validity interval, and record timestamp. Keep valid time (when the claim applies)
distinct from transaction/record time (when Nomyr recorded it). Derive current views from
assertions; corrections append new records and preserve provenance. Preserve explicit unknown,
partial, and not-applicable states rather than converting missing evidence into a negative fact.

## Consequences

- Historical and as-of queries become possible, at the cost of storage, retention, and query
  complexity.
- Evidence references must not embed reusable secret values or customer payloads in inappropriate
  control-plane records.
- Retention, correction semantics, deduplication, and deletion obligations require concrete
  schema and query validation before implementation is treated as settled.

## Validation evidence

The specification and technology plan describe a bitemporal ledger; no `pgx` write-path spike,
schema, or performance result is included here. Acceptance requires the planned ledger spike,
review of temporal edge cases, and a retention/privacy assessment.

## Revisit when

The spike demonstrates unacceptable write/query costs, or retention and deletion requirements
cannot be met while preserving the required audit history.
