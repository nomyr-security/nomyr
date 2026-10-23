# Dependency security review — 2026-10-23

This is a point-in-time review, not a certification or a claim that all CVEs
are exploitable in Nomyr. Review window: 2026-10-16 through 2026-10-23 UTC,
including earlier upstream fixes referenced by advisories published that week.
Changes are proposed in PRs #58, #59, and #60 until merged and validated together.

## Runtime and web dependencies

| Surface | Reviewed version | Evidence and disposition |
| --- | --- | --- |
| Go toolchain | 1.26.10 | Already pinned. [GO-2026-6809](https://pkg.go.dev/vuln/GO-2026-6809), CVE-2026-111943, is fixed in this supported release line. |
| Node.js toolchain | 22.23.4 | Already pinned. The [October 21 security release](https://nodejs.org/en/blog/vulnerability/october-2026-security-releases) fixes the affected 22.x DNS, TLS, Brotli, socket, entropy-cache, and permission-model issues. CI must not retain older independent Node pins. |
| Next.js | 16.4.2 | Already newer than the 16.4.1 fix for [the recent upstream RCE and DoS advisories](https://nextjs.org/blog/nextjs-security-update-october-14-2026). Static export is not a reason to keep vulnerable build or development dependencies. |
| React / React DOM | 19.3.1 | Updated together to the current stable patch. This is dependency maintenance, not an assertion that the previous patch was exploitable. |
| Vite | 8.3.5 | Explicit current stable pin, above the 8.3.3 fixes including [CVE-2026-112946](https://github.com/vitejs/vite/security/advisories/GHSA-rq7h-c2jc-7f22). |
| Vitest | 5.0.5 | Updated from 4.1.12, including a safe version of the mocker for [CVE-2026-112949](https://github.com/vitest-dev/vitest/security/advisories/GHSA-qxg6-89mg-rqv8). The previously pinned 4.1.12 was also a patched release; this update does not establish previous exposure. |
| Babel / jsdom / test utilities | Babel 8, jsdom 30, current compatible utilities | Updated and pinned direct dependencies; migrated explicit JSX syntax configuration following [Babel's migration guide](https://babeljs.io/docs/v8-migration). |
| ESLint | 9.39.5 | Compatibility exception, not an all-clear: [ESLint 9 is end-of-life](https://eslint.org/version-support/), while eslint-plugin-jsx-a11y 6.10.2 declares support only through ESLint 9. Upstream [ESLint 10 support remains unresolved](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/issues/1075). Retained without forced peer bypass or removal of accessibility rules; [#61](https://github.com/nomyr-security/nomyr/issues/61) tracks migration with a separately validated equivalent gate. |
| Node.js types | 22.20.5 | Pinned to the runtime's major line rather than floating to types for a different Node runtime. |

All direct web dependencies use explicit versions and the npm lockfile records
the resolved transitive dependency graph. Version freshness and vulnerability
status are separate checks; a newer unsupported major is not automatically safer.

## Verified web checks

Run from the repository root under the pinned mise environment:

```sh
cd web
npm ci --ignore-scripts
npm test
npm run build
npm audit --json
```

After the refresh, eight Vitest tests, TypeScript checking, ESLint, and the
Next.js static export passed. npm audit reported zero known vulnerabilities
across 426 resolved dependencies, including development dependencies. This
result is time-bound and limited to npm's advisory coverage; it does not cover
Go modules, Python tools, image OS packages, or unknown vulnerabilities.

## Verified Go checks

Go modules were refreshed, including kin-openapi 0.152.0, oapi-codegen/runtime
1.8.0, jsonpointer 1.0.2, pflag 1.0.10, and their resolved transitive graph.
govulncheck was updated to 1.9.0. Module-level scans across all 17 Go package
directories, plus `govulncheck -scan=package ./...` and
`govulncheck -scan=symbol ./...`, reported no vulnerabilities. Combined
`make test`, `make build`, and `git diff --check` passed after the Go/web refresh.
These results do not certify unscanned provider integrations or image OS packages.

## Developer-tool binary finding

The latest upstream golangci-lint 2.15.0 prebuilt binary used Go 1.27.0.
An independent `govulncheck -mode=binary` scan found 13 standard-library
findings, including GO-2026-6809 / CVE-2026-111943. A current tool version
therefore did not mean its embedded compiler runtime was patched.

The mise pin now selects the Go-source backend for the same linter release,
with automatic toolchain switching disabled, so it builds using pinned Go
1.26.10. The bootstrap checker verifies the embedded compiler version;
tests reject the old vulnerable build and mismatched/unverifiable tools.
The CLI baseline also checks development tool provenance. Independent
post-rebuild `go version -m` confirmed Go 1.26.10 and binary-mode govulncheck
reported no vulnerabilities. Buf 1.75.0 was built with Go 1.27.3 and the
source-built govulncheck 1.9.0 used Go 1.26.10 when inspected.

## Separate acceptance boundaries

- Go direct and transitive modules require both reachability-aware and
  module-level vulnerability review; an unreachable finding must not be hidden
  or described as fixed solely because symbol-level checks pass.
- PR #59 owns development images. Its known-vulnerable archived MinIO image is
  not acceptable merely because its limitation is documented. Any removed or
  unavailable replacement must remain an explicit Phase 0 acceptance gap.
- PR #60 owns scanner/action/container and Python contract-tool dependency
  updates, and ongoing security gates. Image SBOM generation alone is not a
  vulnerability scan, and an unrun scanner is not a clean result.
- GitHub Dependency Review initially failed because the repository dependency
  graph was unavailable. After the repository owner enabled it, the SBOM API
  returned SPDX 2.3 with 456 packages and the previously failing review job
  passed on rerun. This resolves the configuration blocker without skipping or
  downgrading the gate. An empty alert list alone still does not establish a
  complete inventory or replace independent ecosystem/image scans.
- Trusted-main artifact signing, complete Compose health/teardown, and the
  final integrated security workflows need their own execution evidence.

Keep dependency updates reviewable, preserve SHA-pinned actions and hash-locked
Python installation, and never expose reusable credentials in scan fixtures,
logs, SBOM examples, or remediation reports.
