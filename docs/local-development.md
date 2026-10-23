# Local development stack

The Compose environment is for development only. It is unauthenticated on the
host side, has no TLS, makes no high availability or production safety claim,
and does not implement Nomyr product services. All published ports bind to
`127.0.0.1`. PostgreSQL is a local development dependency; the current demo
server does not connect to it. The Go demo serves synthetic, in-memory data.

## Start and stop

Install the pinned toolchain and web dependencies first:

```sh
make bootstrap
make dev
```

`make dev` generates `.dev.env` on first use with random local-only credentials
(mode `0600`), starts the development services and waits for their health
checks, then runs the Go demo at <http://127.0.0.1:8080> and Next.js at
<http://127.0.0.1:3000>. Next.js forwards `/api/v1` requests to the Go demo.
The launcher uses a temporary telemetry configuration with telemetry disabled;
it does not change the user's normal CLI telemetry preference. Stop both app
processes with Ctrl-C. `make dev-stack-down` stops and removes the containers
while preserving local data and generated credentials.

The services are PostgreSQL, ClickHouse, Temporal's local development server,
NATS JetStream, OpenBao dev mode, and Keycloak dev mode. PostgreSQL, ClickHouse,
and NATS use explicit version tags and registry digests. Temporal, OpenBao, and
Keycloak are built from the pinned upstream sources in `deploy/compose/images/`
with repository-root build context. Local state is held in named Docker volumes.
Keycloak readiness checks require HTTP 200 from the management `/health/ready`
endpoint, not merely an open management port.
Temporal's standalone development server stores its local development database
in the named `temporal-data` volume using the explicit `--db-filename` path; it
does not store Nomyr product state. The current demo does not connect to
PostgreSQL or any other Compose service.

Object storage is intentionally absent: the MinIO Community Edition container
repository is archived and its available image is not safe to retain in this
stack. No current Nomyr demo flow depends on object storage. Adding an object
store requires selecting a maintained, compatible implementation and verifying
its provenance and security before use. This is a development-environment gap,
not an implemented Nomyr capability.
The stack typically needs several GB of memory and several GB of free disk;
image pulls and local volumes determine actual use.

| Service | Host address |
|---|---|
| PostgreSQL | `127.0.0.1:5432` |
| ClickHouse HTTP/native | `127.0.0.1:8123`, `127.0.0.1:9000` |
| Temporal gRPC/UI | `127.0.0.1:7233`, `127.0.0.1:8233` |
| NATS / monitoring | `127.0.0.1:4222`, `127.0.0.1:8222` |
| OpenBao | `127.0.0.1:8200` |
| Keycloak | `127.0.0.1:8081` |

Credentials are created locally and stored only in ignored `.dev.env`. Do not
commit or publish that file. Compose output from startup does not print its
contents. The OpenBao dev-mode token is only for this disposable local stack;
OpenBao container logging is disabled so its startup output cannot persist that
token. Dev-mode services and their credentials are not suitable for production.

## Check and reset

```sh
make dev-stack-config
make dev-stack-test
make dev-smoke
make dev-stack-down
make dev-stack-reset
```

`dev-stack-config` validates Compose without displaying resolved credentials.
`dev-smoke` starts the services if needed, waits for healthy state, then checks
their local health endpoints. Reset removes containers, volumes, and the
generated credential file. It irreversibly deletes local stack data. The
smoke check verifies service availability only; it does not exercise provider
integrations or Phase 1 product flows.

The `development Compose smoke` GitHub Actions workflow builds and health-checks
the local service images on a hosted Docker runner, then scans every Compose
output and the patched devcontainer image with Grype and Syft binaries built
from verified source. It also builds and scans the service images for arm64;
OCI archives are checksummed then removed before artifact upload. OpenBao is
intentionally built headless for this Phase 0 stack, without its optional
management browser UI. Medium,
High, and Critical matches fail the job. Complete JSON reports are uploaded from the
runner temporary directory even when scanning fails. It uses read-only
repository permissions, stores generated credentials in the runner temporary
directory, and always attempts teardown. This is an image-scan and availability
check, not an end-to-end product test.
