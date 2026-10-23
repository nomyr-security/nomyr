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
| ESLint | 9.39.5 | Latest peer-compatible 9.x release retained because eslint-plugin-jsx-a11y 6.10.2 declares support only through ESLint 9. No forced peer-dependency bypass. |
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
- GitHub Dependency Review currently fails because the repository dependency
  graph is unavailable. The SBOM API also returns 404. An empty Dependabot alert
  list is therefore not evidence that the dependency inventory is complete.
  Do not skip or downgrade the failing gate to claim a pass.
- Trusted-main artifact signing, complete Compose health/teardown, and the
  final integrated security workflows need their own execution evidence.

Keep dependency updates reviewable, preserve SHA-pinned actions and hash-locked
Python installation, and never expose reusable credentials in scan fixtures,
logs, SBOM examples, or remediation reports.
