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

## Set up development

Install Go at the version specified in `go.mod` and Node.js 22 with npm. Fork
the repository, clone your fork, and create a branch for your change.

```sh
make web-install
make test-fast
make build
./bin/nomyr demo
```

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

## Validate and submit

Run these checks before opening a pull request:

```sh
make test
make build
git diff --check
```

In the pull request, describe what changed, why, and how you tested it. Link the
relevant issue and include screenshots for visible interface changes. Mention
breaking changes, dependency additions, and operational implications.

Use descriptive commit messages. Keep generated build output, local editor or
agent configuration, and internal planning documents out of your changes.
Respond to review feedback and rerun affected checks after revisions.

## Licensing and contribution rights

Contributions use the license of the directory being changed: **AGPL-3.0-only**
for Nomyr core and **Apache-2.0** for SDKs and designated public contracts. See
[LICENSING.md](LICENSING.md) for the exact boundaries. Preserve third-party
copyright and license notices, and identify the source of reused material.

Submit only work you have the right to contribute under the applicable license.
Sign off your commits to certify the [Developer Certificate of Origin](https://developercertificate.org/):

```sh
git commit -s
```

Use your own contributor identity for the sign-off. A DCO sign-off does not
transfer copyright or grant a separate right to relicense your contribution.

## Security reports

Do not disclose vulnerabilities in public issues or pull requests. Email
[oss@nomyr.io](mailto:oss@nomyr.io) with a brief description and a safe way to
contact you. Do not include live secrets or customer data. Coordinate detailed
reproduction information privately with the maintainers.

## Community

Be respectful, specific, and constructive. Review ideas and code on their
merits, make room for different experience levels, and help others contribute.
