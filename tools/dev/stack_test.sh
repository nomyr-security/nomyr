#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
temp_root=$(mktemp -d "${TMPDIR:-/tmp}/nomyr-stack-test.XXXXXX")
cleanup() {
	rm -rf "$temp_root"
}
trap cleanup EXIT HUP INT TERM

export NOMYR_DEV_ENV_FILE="$temp_root/.dev.env"

cat >"$temp_root/docker" <<'MOCK'
#!/bin/sh
printf '%s\n' "$*" >>"$NOMYR_DOCKER_CALLS"
case "$*" in
	*' up -d --wait'*)
		if [ "${NOMYR_TEST_COMPOSE_FAIL:-}" = 1 ]; then
			echo 'simulated compose startup failure' >&2
			exit 17
		fi
		;;
	*' ps --all'*) echo 'container exited (1)' ;;
	*' logs --no-color --timestamps'*) printf 'keycloak startup password=%s\n' "$NOMYR_TEST_LOG_SECRET" ;;
esac
MOCK
cat >"$temp_root/go" <<'MOCK'
#!/bin/sh
exec sleep 30
MOCK
cat >"$temp_root/npm" <<'MOCK'
#!/bin/sh
exit 7
MOCK
chmod +x "$temp_root/docker"
chmod +x "$temp_root/go" "$temp_root/npm"
export NOMYR_DOCKER_CALLS="$temp_root/docker-calls"
PATH="$temp_root:$PATH"
export PATH

config_output=$("$repo_root/tools/dev/stack" config)
test -z "$config_output"
test -f "$NOMYR_DEV_ENV_FILE"
if [ "$(uname -s)" = Darwin ]; then
	file_mode=$(stat -f '%Lp' "$NOMYR_DEV_ENV_FILE")
else
	file_mode=$(stat -c '%a' "$NOMYR_DEV_ENV_FILE")
fi
test "$file_mode" = 600
test "$(wc -l <"$NOMYR_DEV_ENV_FILE" | tr -d ' ')" = 4
replacement_env_file="$temp_root/replacement.env"
while IFS='=' read -r key value; do
	if [ "$key" = NOMYR_DEV_POSTGRES_PASSWORD ]; then
		value='literal[0-9]&$\\value'
	fi
	printf '%s=%s\n' "$key" "$value"
done <"$NOMYR_DEV_ENV_FILE" >"$replacement_env_file"
mv "$replacement_env_file" "$NOMYR_DEV_ENV_FILE"
chmod 600 "$NOMYR_DEV_ENV_FILE"
export NOMYR_TEST_LOG_SECRET
NOMYR_TEST_LOG_SECRET=$(sed -n 's/^NOMYR_DEV_POSTGRES_PASSWORD=//p' "$NOMYR_DEV_ENV_FILE")
export NOMYR_TEST_COMPOSE_FAIL=1
if "$repo_root/tools/dev/stack" smoke >"$temp_root/compose-diagnostics.log" 2>&1; then
	echo 'development smoke unexpectedly succeeded after Compose failure' >&2
	exit 1
else
	compose_status=$?
fi
test "$compose_status" = 17
grep -q 'container exited (1)' "$temp_root/compose-diagnostics.log"
grep -Fq 'keycloak startup password=[REDACTED]' "$temp_root/compose-diagnostics.log"
if ! grep -Fq 'literal[0-9]&$\\value' "$NOMYR_DEV_ENV_FILE"; then
	echo 'Compose test did not exercise regex-sensitive secret characters' >&2
	exit 1
fi
if grep -Fq "$NOMYR_TEST_LOG_SECRET" "$temp_root/compose-diagnostics.log"; then
	echo 'Compose diagnostics leaked a generated secret' >&2
	exit 1
fi
unset NOMYR_TEST_COMPOSE_FAIL
if grep -q 'minio\|MINIO' "$repo_root/deploy/compose/compose.yaml" "$repo_root/tools/dev/stack"; then
	echo "archived MinIO service unexpectedly remains in the dev stack" >&2
	exit 1
fi
if grep -E '^    image:' "$repo_root/deploy/compose/compose.yaml" | grep -vE '(@sha256:[0-9a-f]{64}|nomyr-dev/(temporal|openbao|keycloak|postgres):scan)$'; then
	echo "development images must be digest-pinned upstream or built from a reviewed local recipe" >&2
	exit 1
fi
postgres_config=$(sed -n '/^  postgres:/,/^  clickhouse:/p' "$repo_root/deploy/compose/compose.yaml")
printf '%s\n' "$postgres_config" | grep -Fq 'image: nomyr-dev/postgres:scan'
printf '%s\n' "$postgres_config" | grep -Fq 'context: ./images/postgres'
printf '%s\n' "$postgres_config" | grep -Fq 'dockerfile: Dockerfile'
grep -Fq 'postgres:18.6-alpine3.24@sha256:898c60cb390d1085bc97bdd6a1ae968757421df394b071848b52b1c09c89d6a5' "$repo_root/deploy/compose/images/postgres/Dockerfile"
grep -Fq 'brotli-libs=1.2.1-r0' "$repo_root/deploy/compose/images/postgres/Dockerfile"
grep -Fq 'brotli-libs-1.2.1-r0' "$repo_root/deploy/compose/images/postgres/Dockerfile"
grep -Fq 'apk list --installed -q brotli-libs' "$repo_root/deploy/compose/images/postgres/Dockerfile"
grep -A4 '^  temporal:' "$repo_root/deploy/compose/compose.yaml" | grep -q 'dockerfile: deploy/compose/images/temporal/Dockerfile'
grep -q -- '--db-filename' "$repo_root/deploy/compose/compose.yaml"
grep -q 'temporal-data:/home/temporal/data' "$repo_root/deploy/compose/compose.yaml"
grep -q '^  temporal-data:' "$repo_root/deploy/compose/compose.yaml"
grep -A20 '^  keycloak:' "$repo_root/deploy/compose/compose.yaml" | grep -q 'GET /health/ready'
grep -A20 '^  keycloak:' "$repo_root/deploy/compose/compose.yaml" | grep -q 'grep -q.*HTTP/1.1 200'
grep -q 'GET /health/ready' "$repo_root/tools/dev/stack"
grep -q 'test -w /opt/keycloak/lib/quarkus' "$repo_root/tools/dev/stack"
grep -q 'test ! -w /opt/keycloak/lib/lib/deployment' "$repo_root/tools/dev/stack"
grep -q 'test ! -w /opt/keycloak/lib/lib/deployment/appmodel.dat' "$repo_root/tools/dev/stack"
grep -q 'test ! -w /opt/keycloak/lib/lib/deployment/deployment-class-path.dat' "$repo_root/tools/dev/stack"
grep -q 'test ! -w /opt/keycloak/bin/kc.sh' "$repo_root/tools/dev/stack"
grep -q 'test -w /opt/keycloak/data' "$repo_root/tools/dev/stack"
grep -q 'chown -R 1000:1000 /opt/keycloak/lib/quarkus' "$repo_root/deploy/compose/images/keycloak/Dockerfile"
if grep -Eq 'chown -R .* /opt/keycloak(/lib)?([[:space:]]|$)' "$repo_root/deploy/compose/images/keycloak/Dockerfile"; then
	echo 'Keycloak runtime must not make the full installation or lib tree writable' >&2
	exit 1
fi
grep -A4 '^  openbao:' "$repo_root/deploy/compose/compose.yaml" | grep -q 'dockerfile: deploy/compose/images/openbao/Dockerfile'
grep -A4 '^  keycloak:' "$repo_root/deploy/compose/compose.yaml" | grep -q 'dockerfile: deploy/compose/images/keycloak/Dockerfile'
grep -Fq 'for service in temporal openbao keycloak postgres; do' "$repo_root/.github/workflows/dev-compose.yml"
grep -Fq 'context=deploy/compose/images/postgres' "$repo_root/.github/workflows/dev-compose.yml"
grep -Fq 'github.com/labstack/echo/v4 v4.16.1 // indirect' "$repo_root/deploy/compose/images/temporal/dependencies.patch"
grep -Fq 'github.com/labstack/echo/v4 v4.16.1 h1:h05NuF1GvNu6jSAZxOv++AdjD/NJK+mZdGy0thiiuns=' "$repo_root/deploy/compose/images/temporal/dependencies.patch"
grep -Fq 'github.com/labstack/echo/v4 v4.16.1/go.mod h1:V/EQaTVHYJH92QhbQe7kFQ4mjgcjxLQQUQtAN87zZdw=' "$repo_root/deploy/compose/images/temporal/dependencies.patch"
grep -Fq '${NOMYR_DEV_OPENBAO_TOKEN:?run make dev-stack-up}' "$repo_root/deploy/compose/compose.yaml"
clickhouse_config=$(sed -n '/^  clickhouse:/,/^  temporal:/p' "$repo_root/deploy/compose/compose.yaml")
clickhouse_healthcheck=$(printf '%s\n' "$clickhouse_config" | sed -n '/^    healthcheck:/,$p')
printf '%s\n' "$clickhouse_healthcheck" | grep -Fq 'test: ["CMD", "wget", "--spider", "--no-check-certificate", "-T", "1", "-q", "http://127.0.0.1:8123/ping"]'
if printf '%s\n' "$clickhouse_healthcheck" | grep -q -- '--password'; then
	echo 'ClickHouse healthcheck must not expose credentials in argv' >&2
	exit 1
fi
grep -Fq "dpkg-query -W -f='ca-certificates version \${Version}\\n' ca-certificates;" "$repo_root/deploy/compose/images/openbao/Dockerfile"
if grep -E '^    ports:' "$repo_root/deploy/compose/compose.yaml" | grep -v '127\.0\.0\.1:'; then
	echo "development stack published a non-loopback port" >&2
	exit 1
fi
grep -q 'contents: read' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'persist-credentials: false' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'tools/security/build-scanners.sh' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'tools/security/audit-scanner-binaries.sh' "$repo_root/.github/workflows/dev-compose.yml"
if grep -q 'anchore/grype/releases/download' "$repo_root/.github/workflows/dev-compose.yml"; then
	echo "workflow must build scanner tools from verified source" >&2
	exit 1
fi
grep -q '"$scheme:$image"' "$repo_root/.github/workflows/dev-compose.yml"
grep -q -- '--fail-on high' "$repo_root/.github/workflows/dev-compose.yml"
grep -q -- '--fail-on medium' "$repo_root/.github/workflows/security-supply-chain.yml"
grep -q 'issues/67' "$repo_root/docs/dev-image-vulnerabilities.md"
grep -q 'High/Critical' "$repo_root/docs/dev-image-vulnerabilities.md"
grep -q 'CVE-2025-60876' "$repo_root/docs/dev-image-vulnerabilities.md"
grep -q 'GHSA-mjcv-p78q-w5fw' "$repo_root/docs/dev-image-vulnerabilities.md"
grep -q 'summarize-dev-medium-backlog.sh' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'actions/upload-artifact@cf430e030ddbb5b0abf93d22962f4752f3646cd9' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'mcr.microsoft.com/devcontainers/base:ubuntu-24.04@sha256:' "$repo_root/.devcontainer/Dockerfile"
grep -q 'apt-get upgrade --yes' "$repo_root/.devcontainer/Dockerfile"
if grep -q '"features"' "$repo_root/.devcontainer/devcontainer.json"; then
	echo "devcontainer must not reference an unverified Feature" >&2
	exit 1
fi
grep -q 'MISE_VERSION=2026.10.17' "$repo_root/.devcontainer/Dockerfile"
grep -q '17565e268c5542d8ce944321e17af81e0b741b65af4aa7e8e010ee77376a4aa3' "$repo_root/.devcontainer/Dockerfile"
grep -q 'cac510e51793ad45e65fe92305bdbe4309c9c60b5907badbc7eba111738ab08f' "$repo_root/.devcontainer/Dockerfile"
grep -q '/usr/local/share/mise/provenance.txt' "$repo_root/.devcontainer/Dockerfile"
grep -q 'install -D -m 0755 /tmp/mise/bin/mise /usr/local/bin/mise' "$repo_root/.devcontainer/Dockerfile"
grep -q 'mise version | grep -F' "$repo_root/.devcontainer/Dockerfile"
grep -q '^USER vscode$' "$repo_root/.devcontainer/Dockerfile"
grep -Fq 'RUN test -w "$HOME" \' "$repo_root/.devcontainer/Dockerfile"
grep -Fq '    && command -v mise \' "$repo_root/.devcontainer/Dockerfile"
grep -q 'nomyr-image-scan' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'nomyr-devcontainer:scan' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'go build -mod=readonly.*golang.org/x/vuln/cmd/govulncheck' "$repo_root/.github/workflows/dev-compose.yml"
grep -q '^tool golang.org/x/vuln/cmd/govulncheck$' "$repo_root/tools/security/govulncheck/go.mod"
test -s "$repo_root/tools/security/govulncheck/go.sum"
grep -q 'for service in temporal openbao keycloak postgres; do' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'docker compose.*build --pull --no-cache "\$service"' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'compose-build-errors.txt' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'docker image inspect "\$image"' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'local image is unavailable; check the Compose build failure logs; Syft and Grype skipped' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'deploy/compose/images/\*\*' "$repo_root/.github/workflows/dev-compose.yml"
if grep -q 'pnpm audit\|openbao-ui-pnpm-audit' "$repo_root/.github/workflows/dev-compose.yml"; then
	echo "headless OpenBao workflow must not build or audit the optional UI" >&2
	exit 1
fi
grep -q 'docker buildx build --platform linux/arm64' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'oci-archive high' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'sha256sum.*basename.*archive' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'find.*-name.*\.oci\.tar.*-delete' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'docker/setup-qemu-action@29109295f81e9208d7d86ff1c6c12d2833863392' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'docker/setup-buildx-action@f87e5991a6d7451dcb8d9637bfbc97413f497069' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'moby/buildkit@sha256:cec9f139f45e93c5c69c60f8b07cfad9f43f4ef6b6a6cd917527fea5ff2e3dea' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'if: always()' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'runner.temp' "$repo_root/.github/workflows/dev-compose.yml"
if "$repo_root/tools/dev/stack" invalid >/dev/null 2>&1; then
	echo "unknown stack command unexpectedly succeeded" >&2
	exit 1
fi
"$repo_root/tools/dev/stack" down
test -f "$NOMYR_DEV_ENV_FILE"
"$repo_root/tools/dev/stack" reset
test ! -e "$NOMYR_DEV_ENV_FILE"
grep -q 'config --quiet' "$NOMYR_DOCKER_CALLS"
grep -q 'down --volumes --remove-orphans' "$NOMYR_DOCKER_CALLS"
(
	if "$repo_root/tools/dev/dev" >/dev/null 2>&1; then
		dev_status=0
	else
		dev_status=$?
	fi
	printf '%s\n' "$dev_status" >"$temp_root/dev.status"
) &
launcher_pid=$!
sleep 2
if [ ! -f "$temp_root/dev.status" ]; then
	kill "$launcher_pid" 2>/dev/null || true
	wait "$launcher_pid" 2>/dev/null || true
	echo "dev launcher did not stop after its web process failed" >&2
	exit 1
fi
wait "$launcher_pid"
test "$(cat "$temp_root/dev.status")" = 7
"$repo_root/tools/dev/stack" reset
test ! -e "$NOMYR_DEV_ENV_FILE"
printf '%s\n' 'development stack lifecycle checks passed'
