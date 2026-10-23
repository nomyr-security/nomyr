# Contributing to Nomyr

Thank you for helping build Nomyr. Contributions to code, integrations, tests,
documentation, and accessibility are welcome.

## Discuss the change

Search existing [issues](https://github.com/nomyr-security/nomyr/issues) and pull
requests first. Open an issue before starting a substantial feature or changing
a public API so we can agree on scope and design. Small fixes can go straight
to a pull request.

For bug reports, include the version or commit, steps to reproduce, expected and
actual behavior, and sanitized logs. Never include credentials, access tokens,
customer data, or private infrastructure details.

Use the repository issue forms and include enough evidence for another person
to act without guessing:

- Bugs: affected version or commit, component, exact reproduction steps,
  expected and actual behavior, and sanitized logs.
- Features: the user problem, expected outcome, relevant specification or phase,
  and alternatives considered. Do not present roadmap behavior as implemented.
- Documentation: the incorrect or missing text, its location, the proposed
  correction, and primary sources for factual claims.

## Set up development

Install [mise](https://mise.jdx.dev/) using its official instructions. The
checked-in `mise.toml` pins Go, Node.js, Buf, `golangci-lint`, and
`govulncheck`; npm is the version bundled with the pinned Node.js release. Fork
the repository, clone your fork, and create a branch for your change.

`golangci-lint` is built from its pinned Go module source using the selected Go
toolchain rather than downloaded as a prebuilt binary. `make check-dev` verifies
that its embedded Go version matches the selected compiler and scans the lint
binary with `govulncheck` before you rely on it.

```sh
make bootstrap
make check-dev
make test-fast
make test
make build
./bin/nomyr demo
```

`make bootstrap` explicitly installs the versions in `mise.toml` and runs
`npm ci` for the web workspace. It does not install mise or modify global tool
settings. `make check-dev` reports missing development tools and points to the
bootstrap command. `make test-fast` runs Go tests and the web typecheck without
containers; `make test` adds Go race tests, lint, and vulnerability checks.
Run a focused Go fuzz target with
`make fuzz PACKAGE=./path/to/package FOCUS=FuzzTarget` and optionally set
`FUZZTIME=60s`. Only checked-in fuzz targets are runnable.

The demo uses synthetic data and binds to loopback. Do not expose it publicly
or connect it to production credentials.

## Make focused changes

- Keep pull requests small and explain the problem and resulting behavior.
- Update API, protobuf, or JSON Schema contracts before changing public behavior.
- Add tests at the affected boundary. Use synthetic fixtures and cover relevant
  failure cases, permissions, and incomplete data.
- Keep credential handling out of logs, control-plane payloads, and test fixtures.
- Preserve runtime separation: the `nomyr` CLI must not link secret-custody or
  action-execution implementations.
- Format Go changes with `gofmt`. Keep web changes consistent with nearby code.
- Document commands and behavior that users can actually run; identify examples
  and demo data clearly.

## Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/) with a short,
imperative subject:

```text
<type>(optional-scope): <description>
```

Common types are `feat`, `fix`, `docs`, `refactor`, `perf`, `test`, `build`,
`ci`, and `chore`. Useful scopes include `api`, `cli`, `contracts`, `web`,
`connectors`, `actions`, and `docs`.

Examples:

```text
fix(cli): reject non-loopback demo addresses
docs(contracts): clarify capability status semantics
ci: validate pull request titles
```

Keep the subject under 72 characters when practical. Explain the reason and
non-obvious tradeoffs in the body. Add a `BREAKING CHANGE:` footer only when
consumers must take action.

Sign off every commit using your own contributor identity:

```sh
git commit -s -m "fix(cli): reject non-loopback demo addresses"
```

## Validate and submit

Run these checks before opening a pull request:

```sh
make test
make build
git diff --check
```

Open pull requests against `main`. Use a Conventional Commit title because
Nomyr squash-merges pull requests and uses the pull request title for the
resulting commit.

Describe the concrete problem, resulting behavior, and how you tested it. Link
the relevant issue with `Closes #123` when applicable. Include before and after
screenshots for visible interface changes at relevant desktop and mobile widths.
Mention contract changes, breaking changes, dependency additions, security or
operational implications, and any affected phase boundary.

Keep generated build output, local editor or agent configuration, and internal
planning documents out of your changes. Resolve review conversations and rerun
affected checks after revisions.

## Licensing and contribution rights

Contributions use the license of the directory being changed: **AGPL-3.0-only**
for Nomyr core and **Apache-2.0** for SDKs and designated public contracts. See
[LICENSING.md](LICENSING.md) for the exact boundaries. Preserve third-party
copyright and license notices, and identify the source of reused material.

Submit only work you have the right to contribute under the applicable license.
The sign-off certifies the
[Developer Certificate of Origin](https://developercertificate.org/). A DCO
sign-off does not transfer copyright or grant a separate right to relicense your
contribution.

## Security reports

Do not disclose vulnerabilities in public issues or pull requests. Email
[oss@nomyr.io](mailto:oss@nomyr.io) with a brief description and a safe way to
contact you. Do not include live secrets or customer data. Coordinate detailed
reproduction information privately with the maintainers.

## Community

Follow the [Code of Conduct](CODE_OF_CONDUCT.md). For governance and decision
processes, see [GOVERNANCE.md](GOVERNANCE.md), [MAINTAINERS.md](MAINTAINERS.md),
and the [ADR](docs/adr/README.md) and [RFC](docs/rfcs/README.md) process. Support
availability and channels are described in [SUPPORT.md](SUPPORT.md).
