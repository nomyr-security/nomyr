#!/bin/sh
# SPDX-FileCopyrightText: 2026 Nomyr technologies
# SPDX-License-Identifier: AGPL-3.0-only
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
check_definition() {
	plane=$1
	name=$2
	user=$3
	file="$root/containers/$plane/Dockerfile"
	grep -Fq "COPY \${BINARY} /$name" "$file"
	grep -Fq 'FROM scratch' "$file"
	grep -Fq "USER $user:$user" "$file"
	grep -Fq "ENTRYPOINT [\"/$name\"]" "$file"
}

check_definition experience nomyr 10001
check_definition control nomyr-api 10002
check_definition control-worker nomyr-worker 10002
check_definition collection nomyr-collector 10003
check_definition action nomyr-runner 10004

dependencies=$(mktemp)
trap 'rm -f "$dependencies"' EXIT HUP INT TERM
go list -deps ./cmd/nomyr > "$dependencies"
if grep -E '(^|/)internal/custody$|/actions/' "$dependencies"; then
	echo 'the friendly CLI must not link custody or action implementations' >&2
	exit 1
fi

if grep -Eiq '(^|[[:space:]])(ENV|ARG)[[:space:]]+[^=]*(TOKEN|PASSWORD|SECRET|CREDENTIAL)' "$root"/containers/*/Dockerfile; then
	echo 'container definitions must not declare secret-bearing build arguments' >&2
	exit 1
fi

echo 'container definitions preserve plane-specific non-root scratch images'
