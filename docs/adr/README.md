# Architecture Decision Records

This directory records durable architecture decisions. Existing records may be
proposals; their status and evidence control whether they are accepted. Do not
infer acceptance from a filename, merge, or reference in a plan.

## Process

1. Open an issue describing the context, constraints, options, and decision
  needed. For substantial cross-cutting work, use the RFC process in
  [`../rfcs/README.md`](../rfcs/README.md).
2. Copy the relevant ADR format from an existing record. State `Proposed`,
  `Accepted`, `Superseded`, or `Rejected`, along with consequences, evidence,
  unresolved questions, and revisit triggers.
3. Keep the decision in `Proposed` until authorized project decision-makers
  record their decision in review. Do not invent approvers or claim an
  unperformed spike as evidence.
4. On acceptance, update affected plans, contracts, and implementation issues.
  Keep planned behavior labeled as planned until source and tests demonstrate
  it.

The project has not yet published named human decision owners. This process
does not itself ratify a decision.
