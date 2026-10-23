# ADR-0009 — Event bus and transactional outbox

- **Status:** Proposed (delivery semantics and failure testing pending)
- **Date:** 2026-10-23
- **Decision owner:** Project owner (acceptance pending)
- **Scope:** Planned asynchronous event delivery from authoritative transactions

## Context

Ingest and governance transactions need to persist authoritative state and notify projections or
consumers without losing work between a database commit and message publication. The plan calls
for at-least-once delivery and idempotent consumers.

## Proposed decision

Write event records to a PostgreSQL transactional outbox in the same transaction as authoritative
state changes. Relay committed records to NATS JetStream using stable event identifiers. Require
idempotent consumers, bounded retries, observable poison-message handling, and tenant-scoped
subjects and dead-letter handling. Treat delivery as at-least-once; do not claim exactly-once
processing.

## Consequences

- The outbox closes the database/message dual-write gap but adds relay operations, backlog
  management, and retention concerns.
- Consumers must tolerate duplicate and out-of-order delivery where the contract permits it.
- NATS configuration, subject isolation, replay, and dead-letter access need security and
  operational validation.

## Validation evidence

The technology and architecture plans specify PostgreSQL outbox plus NATS JetStream. No relay,
consumer, outage test, or tenancy test is recorded here. Acceptance requires crash/retry tests,
duplicate delivery tests, and verification of tenant-scoped routing.

## Revisit when

Measured event throughput, ordering, retention, or operational constraints cannot be met with the
proposed relay and broker model.
