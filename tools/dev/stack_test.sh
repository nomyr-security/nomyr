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
chmod +x "$temp_root/docker"
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
printf '%s\n' 'development stack lifecycle checks passed'
