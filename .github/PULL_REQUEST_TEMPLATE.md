## Problem and result

<!-- What concrete problem does this solve, and what changes for users? -->

## Changes

<!-- List the focused changes that help a reviewer assess the implementation. -->

## Contracts and boundaries

<!-- Identify changed OpenAPI, protobuf, JSON Schema, runtime-plane, licensing, or phase boundaries. Write "None" when they do not apply. -->

## Validation

<!-- Include focused tests and the repository checks you ran. -->

- [ ] `make test`
- [ ] `make build`
- [ ] `git diff --check`

## Visual evidence

<!-- Add before/after screenshots for UI changes at relevant desktop and mobile widths. Remove this section when it does not apply. -->

## Review checklist

- [ ] The PR title follows Conventional Commits, for example `fix(cli): reject unsafe address`.
- [ ] The change is focused and links its issue with `Closes #…` when applicable.
- [ ] Public contracts were updated before implementations that change public behavior.
- [ ] Documentation distinguishes implemented behavior from planned capabilities.
- [ ] Tests cover the affected boundary and failure behavior.
- [ ] No reusable secrets, credentials, customer data, or private infrastructure details were added.
- [ ] New dependencies, licensing effects, and operational tradeoffs are explained.
- [ ] Commits include a DCO sign-off (`git commit -s`).
