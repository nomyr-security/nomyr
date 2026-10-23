# ADR-0010 — Connector transport

- **Status:** Proposed (protocol spike and threat review pending)
- **Date:** 2026-10-23
- **Decision owner:** Project owner (acceptance pending)
- **Scope:** Planned language-neutral connector interface and transport

## Context

Connectors execute near customer-controlled provider credentials and may be implemented outside
the control-plane language. The specification requires a language-neutral worker interface. The
repository plans protobuf/gRPC and an HTTP-friendly worker option but has not validated a complete
connector transport here.

## Proposed decision

Define versioned protobuf contracts as the canonical internal connector interface. Support gRPC
for managed deployments and a narrowly specified HTTP worker transport where network or language
constraints require it. Authenticate and authorize each collector independently, bind requests to
tenant and connector identity, and send only sanitized batches to the control plane. Reusable
credentials remain at the collector boundary and secret values do not enter control-plane
messages, logs, or fixtures.

## Consequences

- A language-neutral contract supports external workers while retaining a first-class Go SDK.
- Supporting two transports adds conformance and compatibility work; they must have equivalent
  security and semantics.
- Authentication, limits, schema compatibility, and sanitization need protocol-level tests.

## Validation evidence

The plans name gRPC/protobuf and an HTTP worker path. No interoperability spike, protocol
conformance result, or transport threat review is recorded. Acceptance requires a minimal worker
prototype, version-compatibility checks, and tests proving credentials and unsanitized payloads do
not cross the boundary.

## Revisit when

Connector interoperability or security tests show that either transport cannot preserve the same
identity, authorization, sanitization, and compatibility guarantees.
