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
test "$(wc -l <"$NOMYR_DEV_ENV_FILE" | tr -d ' ')" = 5
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
