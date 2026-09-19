<picture>
  <source media="(prefers-color-scheme: dark)" srcset="brand/svg/nomyr-lockup-on-dark.svg">
  <img src="brand/svg/nomyr-lockup.svg" alt="Nomyr" width="240">
</picture>

# Nomyr — Open-Source Non-Human Identity Security

**Non-human identity security. From discovery to retirement.**

Nomyr is building an open-source, self-hostable **non-human identity (NHI) security platform** for the identities that power software: service accounts, cloud roles, API keys, certificates, OAuth applications, workload identities, and AI agents.

Our goal is to give security and platform teams one place to understand machine access, establish ownership, reduce risk, and govern the complete identity lifecycle across cloud, SaaS, Kubernetes, and on-premises environments.

[Explore the project](https://github.com/nomyr-security/nomyr) · [Contribute](CONTRIBUTING.md) · [Ask a question](https://github.com/nomyr-security/nomyr/issues)

## Why non-human identity security?

Applications, services, and AI agents need access to infrastructure and data. Their identities and credentials can outlive the workloads that created them, accumulate permissions, or lose an accountable owner. Securing them requires understanding both the identity and how its access is used.

Nomyr brings that problem into one product scope: machine identity discovery, service account governance, credential lifecycle management, workload access, and AI agent security. It is designed for security engineers, IAM teams, and platform teams responsible for machine access.

## Non-human identity lifecycle: discovery to retirement

Nomyr’s product scope connects six essential parts of non-human identity security:

- **Discover and map.** Build an inventory of identities, credentials, permissions, and activity. Connect each identity to its consumers, resources, and accountable owners.
- **Understand and prioritize risk.** Identify excessive access, exposed credentials, stale accounts, ownership gaps, and suspicious behavior with the context needed to act.
- **Govern access.** Bring ownership, access reviews, policy decisions, approvals, and third-party application oversight into a shared workflow.
- **Manage credentials and identities.** Coordinate provisioning, vault and federation bindings, credential rotation, certificate renewal, and verified retirement.
- **Secure workloads and AI agents.** Govern short-lived access, delegated permissions, and agent sessions with scoped policies and attributable activity.
- **Respond and verify.** Connect investigation to approved containment and remediation, verify outcomes, and retain evidence for audit and reporting.

## Built around control and evidence

The platform is designed to work with existing identity providers, cloud IAM, vaults, certificate authorities, and runtime gateways. Teams should be able to start with visibility, introduce approved actions, and adopt policy-driven automation within explicitly authorized scopes.

Nomyr’s design puts self-hosting, clear ownership, separation of duties, and verifiable outcomes at the center. Security decisions should explain what is known, what remains uncertain, and whether an action actually achieved its intended result.

## Local development

Requirements: Go (see `go.mod`) and Node.js 22 with npm.

```sh
git clone https://github.com/nomyr-security/nomyr.git
cd nomyr
make web-install
make test-fast
make build
./bin/nomyr demo
```

Open http://127.0.0.1:8080. The local demo uses synthetic data and does not
connect to providers or execute security actions. Keep it bound to loopback;
it does not provide authentication or TLS.

## Contribute to Nomyr

Help shape open-source machine identity security through code, documentation, security research, and real-world use cases. Read the [contribution guidelines](CONTRIBUTING.md) for setup, checks, and pull request expectations. Use [GitHub issues](https://github.com/nomyr-security/nomyr/issues) for bugs, feature proposals, and questions.

Report security vulnerabilities privately to [oss@nomyr.io](mailto:oss@nomyr.io). Please keep credentials and sensitive environment details out of public issues.

Follow the [Nomyr GitHub organization](https://github.com/nomyr-security) for project activity.

## Open-source licensing

Nomyr core is licensed under [AGPL-3.0-only](LICENSE). SDKs and public
API/schema contracts are licensed under Apache-2.0. See [LICENSING.md](LICENSING.md)
for directory boundaries and third-party license notices.

## Brand assets

The complete [Nomyr brand kit](brand/README.md) includes SVG masters, PNG exports, web icons, and social artwork. Runtime assets live in `web/public`; its manifest uses relative URLs to support deployment paths. Use the supplied lockups without retyping the wordmark.
