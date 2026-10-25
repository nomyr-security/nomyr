# Temporal development image

This image rebuilds the upstream Temporal CLI and its embedded development server for the Phase 0 synthetic development stack. It is not a production Temporal deployment or a Nomyr execution plane. The upstream `server start-dev` command is explicitly development-only and starts with in-memory persistence unless `--db-filename` is supplied.

## Pinned inputs

- Temporal CLI `v1.9.1`, GitHub-verified signed commit `1de87a9f26991bf4f5c0a5ff96f2cea8d7a3cbde`; its codeload source archive is pinned by SHA-256 `2417fbc1348cc3b3e5b01166ebe4c40938d779505d8e119b2ed77bb3d6ae965c` ([upstream release](https://github.com/temporalio/cli/releases/tag/v1.9.1)).
- Go builder `golang:1.27.3-bookworm`, pinned to manifest digest `sha256:972b4110ba0f47539a975c31e52c64765e1ff19f2d74469ae4a595c6676e80a5`. The base's copied CA bundle was SHA-256 `e9c7b63ee2623386e9b0a3ee32f499d01a8ea23dc83b8383ddff9f8e749543b7`. The builder reinstalls Debian Bookworm security `ca-certificates` `20250419~deb12u2` through APT's signed repository metadata; the authenticated package SHA-256 published by Debian is `b9f76b5271a7ac373fd5e7e1b5b826a71ddd702b42a0d104955ecb1acfffc94b`. That package still includes emSign C1 and ECC C3 roots, although [Debian tracks CVE-2026-114406 as `NOT-FOR-US`](https://security-tracker.debian.org/tracker/CVE-2026-114406). To avoid trusting the compromised chain identified by [Red Hat's CVE report](https://bugzilla.redhat.com/show_bug.cgi?id=2556046), the build disables those two roots from Debian's bundled certificate selection and regenerates the bundle with `update-ca-certificates --fresh`. ECC G3 is absent from the pinned Debian bundle. A build-time OpenSSL check fails if C1, ECC C3, or ECC G3 remains; the builder also prints the installed CA package version and the SHA-256 of the exact refreshed bundle copied into scratch. Go `crypto/x509` fixes are not treated as CA-store remediation.
- Runtime base is `scratch`; the image contains the statically linked CLI, system CA bundle and zoneinfo, minimal passwd/group entries, and license files. It adds no Alpine/OpenSSL/zlib package layer.
- Explicit patched Go modules: `golang.org/x/net` `v0.60.0`; `github.com/gomarkdown/markdown` `v0.0.0-20260724235538-61af6e1e6971`; `go.opentelemetry.io/otel/exporters/otlp/otlptrace` and its `otlptracegrpc` module at `v1.45.0`; `go.opentelemetry.io/otel/sdk` `v1.45.0`; and `github.com/labstack/echo/v4` `v4.16.1`. The pinned Temporal CLI `v1.9.1` module graph used Echo `v4.15.4`, which is affected by [GHSA-cxf3-9h37-7rr8](https://github.com/advisories/GHSA-cxf3-9h37-7rr8), [GHSA-9969-6676-49h5](https://github.com/advisories/GHSA-9969-6676-49h5), [GHSA-f48p-c9v4-h7xv](https://github.com/advisories/GHSA-f48p-c9v4-h7xv), and [GHSA-h7wr-284f-5qvc](https://github.com/advisories/GHSA-h7wr-284f-5qvc); all list `v4.16.1` as patched. The checked-in `dependencies.patch` applies a frozen `go.mod`/`go.sum` delta to the exact upstream module files, after verifying their Git blob IDs. `GOFLAGS=-mod=readonly` makes the build and dependency-license inventory fail rather than rewrite the selected graph; Go verifies module downloads against the patched `go.sum`. Upstream Go source files remain unmodified.
- The upstream CLI MIT license, license/notice files for imported Go modules, and Debian copyright notices for the copied CA and timezone data are included under `/usr/share/licenses`.

The `v1.9.1` release page reports the tag commit as GitHub-verified. Local `git verify-commit` could not run because `gpg` is not installed; source bytes are independently pinned by the archive checksum in the Dockerfile. The CA bundle digest is emitted by the Docker build after refreshing the authenticated Debian package; retain that build output as the final-bundle provenance record.

## Residual advisory

The [Go advisory `GO-2026-5932`](https://pkg.go.dev/vuln/GO-2026-5932) covers the deprecated `golang.org/x/crypto/openpgp` package and has no fixed version. On the recipe's extracted source and built binary, `govulncheck` `v1.9.0` source mode found no vulnerable imported package or reachable call from `cmd/temporal`, while reporting the required `golang.org/x/crypto` module as containing the advisory. Binary mode reports the OpenPGP symbols, but provides no source call path. The actual module graph contains no `github.com/google/go-github/v30` module, and no package in the command's dependency graph imports `golang.org/x/crypto/openpgp`; therefore a go-github major upgrade or import patch is not applicable to this source. The binary finding remains unresolved rather than suppressed or claimed safe. Preserve it in hosted scans and keep the security gate blocked until the integration review resolves the binary evidence.

## Build and scan

Run these commands from the repository root so the Docker build context matches Compose:

```sh
set -o pipefail
docker build --pull --no-cache --progress=plain \
  --file deploy/compose/images/temporal/Dockerfile \
  --tag nomyr-dev/temporal:scan . | tee /tmp/temporal-build.log

grype nomyr-dev/temporal:scan --output json > /tmp/temporal-grype.json
grype nomyr-dev/temporal:scan --fail-on medium
```

The local source and binary reachability checks used `govulncheck` `v1.9.0`; the exact commands from the extracted upstream source directory are `go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 -mode=source ./cmd/temporal` and `go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 -mode=binary /path/to/temporal`. Source mode reports zero reachable vulnerabilities and one vulnerable required module; binary mode exits nonzero for `GO-2026-5932`. Neither result substitutes for the hosted Grype image gate.

The second Grype command fails on Medium, High, or Critical findings, matching the development image gate. Preserve its complete output and JSON report for the integration review; a clean build alone does not demonstrate that the rebuilt image is free of findings.

## Runtime contract

Compose continues to invoke the CLI entrypoint as:

```text
temporal server start-dev --ip 0.0.0.0
```

The image runs as UID/GID `10001:10001`, with `/usr/local/bin/temporal` as its entrypoint and `/home/temporal` as its working directory. Home and Temporal config directories are root-owned and mode `0555`; only `/home/temporal/data` is writable, owned by UID/GID `10001:10001` and mode `0700`. Source inspection shows `server start-dev` uses in-memory SQLite unless `--db-filename` is set, and creates SQLite schema/WAL state under that filename's parent directory. It does not write its default config path; explicit `temporal config` writes require an external writable config path and are outside this image's start/health contract. No default credentials or provider configuration are included.

Compose's persistence interface is `--db-filename /home/temporal/data/temporal.db` with its existing `temporal-data:/home/temporal/data` named-volume mount retained. The image provides the writable target directory but does not change Compose's persistence behavior. Upstream's `TestServer_StartDev_SQLitePragma` verifies file, `-shm`, and `-wal` creation for `--db-filename`. A native local smoke run against the frozen module graph served `operator cluster health` as `SERVING` with a mode-0555 home, created the specified SQLite database under the writable data directory, and restarted against that database without creating files in home. Hosted container validation must still confirm ownership/modes after `COPY` and the named-volume behavior under UID/GID `10001:10001`.

The existing Compose health check remains compatible and must pass inside the container:

```sh
temporal operator cluster health --address 127.0.0.1:7233
```

For a standalone hosted persistence smoke run, publish only loopback ports and retain the named volume after removing the container:

```sh
docker run --detach --rm --name nomyr-temporal \
  --publish 127.0.0.1:7233:7233 \
  --publish 127.0.0.1:8233:8233 \
  --mount type=volume,src=temporal-data,dst=/home/temporal/data \
  nomyr-dev/temporal:scan server start-dev --ip 0.0.0.0 \
  --db-filename /home/temporal/data/temporal.db

docker exec nomyr-temporal temporal operator cluster health --address 127.0.0.1:7233
docker rm --force nomyr-temporal
```

The container exposes no shell or package manager. The ports are metadata only; host publication must remain bound to loopback as in Compose. This development image provides no TLS, authentication, provider access, custody, or security-action capability.
