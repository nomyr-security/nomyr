# OpenBao development image

This image rebuilds OpenBao from the upstream `v2.7.1` release at commit
`a5db72cef75c24b920ade02065b18dd8eb666bac`. GitHub reports the commit signature
as verified ([release](https://github.com/openbao/openbao/releases/tag/v2.7.1),
[commit verification](https://github.com/openbao/openbao/commit/a5db72cef75c24b920ade02065b18dd8eb666bac)).
The upstream tag's published Go module graph contains
`golang.org/x/net v0.58.0`; the build uses the patched `v0.60.0` release to
address GO-2026-6612, GO-2026-6603, GO-2026-6611, GO-2026-6617, and related
fixed advisories in the 2026-10-23 Grype baseline
([Go advisory](https://pkg.go.dev/vuln/GO-2026-6612)). The change is applied
through an alternate Go module file and leaves upstream source files untouched.
The upstream `go.sum` and Go checksum database verify the selected module
content. This module version requires `x/crypto v0.57.0`, `x/sys v0.48.0`,
`x/term v0.46.0`, and `x/text v0.42.0`; the same selected graph pins
`x/sync v0.23.0`, `x/mod v0.41.0`, and `x/tools v0.49.0`. The Go compiler
image supplies Go 1.27.3, whose official release
includes the intervening standard-library security fixes
([Go 1.27 release history](https://go.dev/doc/devel/release#go1.27.3)).

The compiler stage uses the official `golang:1.27.3-bookworm` multi-platform
image at index digest
`sha256:972b4110ba0f47539a975c31e52c64765e1ff19f2d74469ae4a595c6676e80a5`
(linux/amd64 manifest `sha256:fea12ba16096adec0f699ce1c4da2154d320450e56c1f63671f95d2a3551c77b`,
linux/arm64/v8 manifest `sha256:d3f61539a5743b1f1815ba8dd3627202f55262fb09380feda1d585ccf5a9850e`).
Before compiling, it reinstalls Debian Bookworm `ca-certificates`
`20250419~deb12u2` through signed APT metadata (published package SHA-256
`b9f76b5271a7ac373fd5e7e1b5b826a71ddd702b42a0d104955ecb1acfffc94b`),
disables the emSign C1 and ECC C3 roots, regenerates the bundle, and fails the
build if C1, ECC C3, or ECC G3 is present. Debian tracks CVE-2026-114406 as
`NOT-FOR-US`; these roots are removed to avoid trusting the compromised chain
described in the [upstream CVE report](https://bugzilla.redhat.com/show_bug.cgi?id=2556046).
The build records the package version and regenerated bundle SHA-256; the
matching Debian copyright notice is retained in the image.

The Go application is built without the `ui` build tag. Upstream provides
`internal/http/assets_stub.go` under the `!ui` build constraint, and its
Makefile's `dev` target omits `ui` while `dev-ui` explicitly adds it. This
headless mode preserves the server/API, CLI, and `bao status` health behavior
required by this Phase 0 stack; it intentionally does not bundle or serve the
OpenBao browser management UI. No architecture or Compose contract changes.

The final runtime is
`scratch`; it contains the statically linked OpenBao binary, CA certificates,
the upstream MPL-2.0 license, and only the directories needed for writable
runtime state. Both image digests were resolved from Docker Official Images
registries on 2026-10-23; per-architecture child manifests were verified from
the Docker Registry manifest index.

No CVE count or clean-image result is asserted here. The baseline matches are
package/advisory evidence, not proof of reachable vulnerable code. A local
`govulncheck` binary scan of the Go 1.27.3 build reported no vulnerable symbols
or imported vulnerable packages. It did report `GO-2026-5932` for the
required `golang.org/x/crypto` module (`openpgp` is marked unmaintained, with no
fixed version); the scanner found no calls into the affected package. Keep this
finding visible in the hosted image report and reassess it when the upstream
module graph changes.

## Build and run

Run from the repository root so Compose uses the repository-root build context:

```sh
make dev-stack-config
docker compose --env-file .dev.env -f deploy/compose/compose.yaml build --pull --no-cache openbao
docker compose --env-file .dev.env -f deploy/compose/compose.yaml up -d openbao
docker compose --env-file .dev.env -f deploy/compose/compose.yaml exec -T openbao bao version
docker compose --env-file .dev.env -f deploy/compose/compose.yaml exec -T openbao bao status -address=http://127.0.0.1:8200
docker compose --env-file .dev.env -f deploy/compose/compose.yaml stop openbao
docker compose --env-file .dev.env -f deploy/compose/compose.yaml rm --force openbao
```

## Rejected UI-builder candidate

The UI build was evaluated and then removed from this recipe because Phase 0
requires only the API/CLI/health contract and the UI graph had unresolved
critical advisories. The UI's `ui/package.json` declares Node `20`; its
upstream lockfile was tested with frozen install and production build using
Node `22.23.4` and pnpm `10.34.4`, with engine/Ember compatibility warnings.
Those UI outputs are not included in this image. On 2026-10-23,
`pnpm audit --json` on the frozen UI graph exited 1 with 16 low, 52 moderate,
77 high, and 6 critical findings. `pnpm audit --prod --json` exited 1 with 11
low, 27 moderate, 40 high, and 3 critical findings. Critical production-graph
findings include `babel-traverse` with no patched version listed and
`handlebars@4.7.9` (two advisories; patched version `>=4.7.10`). These are
rejected-candidate findings, not components in the headless image. Complete
local JSON reports were retained outside the repository for integration review.
No UI lockfile or broad dependency upgrade is included.

The running executable is `/usr/bin/bao`; `/usr/bin/vault` remains an alias.
The image runs as UID/GID `10001`; `/openbao/{config,file,logs}`, `/home/openbao`,
and `/tmp` are writable by that user. It accepts the existing Compose
`server -dev -dev-listen-address=0.0.0.0:8200 -dev-root-token-id=…` command,
exposes container port `8200`, and supports the Compose health command
`bao status -address=http://127.0.0.1:8200`. Compose binds the host port to
`127.0.0.1:8200` and disables container logging for this service. The generated
development token remains supplied by the existing local stack script; it is
not an image default or build argument.

## Hosted scan and runtime checks

Build the service through Compose, then retain the full report and scanner
metadata as CI artifacts:

```sh
make dev-stack-config
docker compose --env-file .dev.env -f deploy/compose/compose.yaml build --pull --no-cache openbao
grype nomyr-dev/openbao:scan --output json > /tmp/openbao-grype.json
docker image inspect nomyr-dev/openbao:scan --format '{{.Id}} {{.Os}}/{{.Architecture}}'
docker compose --env-file .dev.env -f deploy/compose/compose.yaml up -d openbao
docker compose --env-file .dev.env -f deploy/compose/compose.yaml ps --format json
docker compose --env-file .dev.env -f deploy/compose/compose.yaml exec -T openbao bao status -address=http://127.0.0.1:8200
docker compose --env-file .dev.env -f deploy/compose/compose.yaml stop openbao
docker compose --env-file .dev.env -f deploy/compose/compose.yaml rm --force openbao
```

Record the Grype version and database build time alongside the JSON report.
Review all remaining matches without blanket ignores; package matches alone do
not establish runtime reachability. Docker image build, scratch runtime
permissions, Compose health, teardown, and the image-level Grype scan require
the hosted Docker runner and are not verified locally because its daemon is
unavailable. A native headless build of the same OpenBao source with Go 1.27.3
and `x/net v0.60.0` succeeded; both `bao status` and `/v1/sys/health` returned
healthy dev-server responses with a generated temporary token. These checks do
not replace the hosted image build and scan.
