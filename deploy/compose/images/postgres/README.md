# PostgreSQL development image

This image extends the official PostgreSQL `18.6-alpine3.24` multi-architecture
image at index digest
`sha256:898c60cb390d1085bc97bdd6a1ae968757421df394b071848b52b1c09c89d6a5`.
The official base manifest is pinned; the image updates only `brotli-libs` to
the Alpine-fixed `1.2.1-r0` package for `CVE-2026-111497`, then verifies the
installed package version and restores the upstream `postgres` runtime user.
The package is fetched from Alpine's signed repositories; the Docker build
fails if the fixed version is unavailable. See the [Alpine advisory](https://security.alpinelinux.org/vuln/CVE-2026-111497)
and the [official PostgreSQL image manifest](https://github.com/docker-library/official-images/blob/e00e1bd34ec5c8a8e7ad89b273b3d42efaf6d5bc/library/postgres).

The source run's raw report identified `brotli-libs 1.2.0-r1` as Critical and
listed `1.2.1-r0` as fixed. The new image has not yet been built or scanned on
the hosted runner; amd64 Compose health and both-architecture scan evidence
remain required. Local Docker is unavailable in this environment.
