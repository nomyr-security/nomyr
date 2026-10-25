# ADR-0012 — Single-binary embedded UI and lite mode

- **Status:** Proposed; owner acceptance pending
- **Date:** 2026-10-25
- **Decision owner:** `@nomyr-ai` (owner-designated maintainer account)
- **Scope:** Static UI embedding and possible local-mode packaging

## Context

The Phase 0 repository already contains a static-export Next.js application embedded in the Go
binary. This is a build and packaging boundary; it does not mean that local identity retrieval,
authorization, or a complete lite product mode exists. The proposed architecture also calls for
embedded PostgreSQL behavior across macOS, Linux, and Windows on amd64 and arm64.

## Proposed decision

Retain the existing static-export and Go embedding path as a proposed packaging direction. Keep
“lite mode” unaccepted until platform, persistence, route, and operational requirements have
reproducible evidence and the decision owner records a disposition.

## Evidence and limitations

The repository source has a static `/identities/demo-service-account/` export route and a dynamic
`as_of` query display. The demo server provides static fallback behavior, not arbitrary identity
retrieval. The assigned spike report records build and cold-start evidence when measured. No
embedded PostgreSQL distribution has been selected or run across the requested platform matrix;
cross-compilation is not treated as runtime evidence. The isolated NATS spike exercises the actual
NATS server library in-process with JetStream enabled; it does not establish production persistence,
availability, or platform behavior.

## Consequences

- Static assets can be packaged in Go without a Node.js runtime requirement, subject to the
  existing build pipeline.
- The single-binary distribution and lite-mode decision remains open pending persistence and
  platform evidence.
- The spike report does not establish identity storage, retrieval, authentication, or
  authorization as implemented behavior.

## Revisit when

A maintained, provenance-reviewed embedded PostgreSQL option is identified and tested on each
supported OS/architecture, or the distribution and local-mode requirements change.
