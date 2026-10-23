# ADR-0004 — License model

- **Status:** Proposed (legal and owner review pending)
- **Date:** 2026-10-23
- **Decision owner:** Project owner (acceptance pending)
- **Scope:** Licensing of original Nomyr material and reuse review; no legal conclusion is made here

## Context

The open-source setup plan proposes AGPL-3.0-only for core services and Apache-2.0 for designated
SDKs and interface material. `LICENSING.md` currently describes those boundaries for the root,
`sdk/`, `api/`, `proto/`, and `schemas/`. Third-party dependencies retain their own terms, and
process separation alone is not a licensing determination.

## Proposed decision

Retain the directory-specific license boundaries documented in `LICENSING.md`: AGPL-3.0-only
as the default for original Nomyr material, with Apache-2.0 exceptions only where the applicable
directory license says so. Require a version-pinned technical and license review before adopting
reused components. Record obligations, distribution paths, technical fit, and unresolved risk;
seek qualified legal review where interpretation is material. Do not infer compatibility from
separate-process deployment.

## Consequences

- SDK and contract consumers can rely on the stated Apache-2.0 exceptions, subject to their
  directory notices.
- Bundled, generated, and redistributed third-party material needs its own attribution and
  obligations review.
- Component versions, notices, and compatibility assessments remain unverified until the
  dependency inventory is pinned and reviewed.

## Validation evidence

Current evidence is the repository's `LICENSING.md` and license files; it does not constitute
owner ratification, legal advice, a complete dependency inventory, or a compatibility review.
The pinned-component review template and completed component reviews remain open work.

## Revisit when

The owner changes the public license promise, a directory boundary changes, or a dependency's
license or distribution model changes the project's obligations.
