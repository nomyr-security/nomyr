# ADR-0011 — Signing-key purposes

- **Status:** Proposed (cryptographic review and key-lifecycle design pending)
- **Date:** 2026-10-23
- **Decision owner:** Project owner (acceptance pending)
- **Scope:** Planned signing and verification keys; exact algorithms and providers are undecided

## Context

The specification assigns signing roles to policy bundles, execution approvals, and release or
artifact provenance. Reusing one key across trust domains would couple compromise and rotation
procedures. Verification must identify what was signed, by whom, and under which key purpose.

## Proposed decision

Separate keys and trust policies by purpose, environment, and (where appropriate) tenant. Define
distinct signing identities for policy bundles, approval or execution authorization, and release
provenance. Include key identifier, purpose, algorithm, and validity metadata in signed
envelopes; verifiers must reject an unexpected purpose or algorithm. Specify generation,
protected storage, rotation, revocation, recovery, and audit before production use. Do not place
private key material in control-plane records or reusable-secret payloads.

## Consequences

- Purpose separation limits cross-domain use after a key compromise but increases key inventory
  and lifecycle operations.
- Trust roots, algorithm selection, hardware protection, and tenant boundaries remain decisions
  for a cryptographic threat review.
- This ADR does not establish that signing, key storage, or verification is implemented.

## Validation evidence

The specification identifies signing-key purposes, but no cryptographic review, key-management
spike, rotation test, or verifier implementation is recorded here. Acceptance requires those
artifacts and negative tests for wrong-purpose, revoked, expired, and unknown keys.

## Revisit when

Threat analysis, deployment constraints, or key-provider capabilities require changes to purpose
separation, trust roots, or lifecycle controls.
