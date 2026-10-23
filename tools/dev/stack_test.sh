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
if grep -q 'minio\|MINIO' "$repo_root/deploy/compose/compose.yaml" "$repo_root/tools/dev/stack"; then
	echo "archived MinIO service unexpectedly remains in the dev stack" >&2
	exit 1
fi
if grep -E '^    image:' "$repo_root/deploy/compose/compose.yaml" | grep -vE '@sha256:[0-9a-f]{64}$'; then
	echo "development images must be pinned by registry digest" >&2
	exit 1
fi
if grep -E '^    ports:' "$repo_root/deploy/compose/compose.yaml" | grep -v '127\.0\.0\.1:'; then
	echo "development stack published a non-loopback port" >&2
	exit 1
fi
grep -q 'contents: read' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'persist-credentials: false' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'anchore/grype/releases/download/v0.121.0/grype_0.121.0_linux_amd64.tar.gz' "$repo_root/.github/workflows/dev-compose.yml"
grep -q '64e07133d3486115de647a66581461dccf5c732748a0a3119077395565f15519' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'registry:$image' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'actions/upload-artifact@cf430e030ddbb5b0abf93d22962f4752f3646cd9' "$repo_root/.github/workflows/dev-compose.yml"
grep -q 'mcr.microsoft.com/devcontainers/base:ubuntu-24.04@sha256:' "$repo_root/.devcontainer/devcontainer.json"
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
