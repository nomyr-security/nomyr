#!/bin/sh
# SPDX-FileCopyrightText: 2026 Nomyr technologies
# SPDX-License-Identifier: AGPL-3.0-only
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
temporary=$(mktemp -d)
trap 'rm -rf "$temporary"' EXIT HUP INT TERM
context="$temporary/context"
output="$temporary/output"
mkdir -p "$context"
cp "$root/.dockerignore" "$context/.dockerignore"

write_marker() {
	path=$1
	mkdir -p "$context/$(dirname "$path")"
	printf 'synthetic context-boundary marker\n' > "$context/$path"
}

for path in .dev.env .env .env.local .env.production .git/config docs/README.md docs/spec/private.md docs/plan/private.md docs/phases/private.md docs/adr/private.md .cache/bootstrap/private.bin .cache/security/nomyr-unexpected deploy/private.env deploy/compose/private.env deploy/compose/images/keycloak/README.md deploy/compose/images/temporal/README.md web/node_modules/private.js web/.next/private.js web/out/private.js node_modules/private.js; do
	write_marker "$path"
done

for binary in nomyr-experience nomyr-control-api nomyr-control-worker nomyr-collection nomyr-action; do
	write_marker ".cache/security/$binary"
done
mkdir -p "$context/deploy/compose/images/temporal"
cp "$root/deploy/compose/images/temporal/dependencies.patch" \
	"$context/deploy/compose/images/temporal/dependencies.patch"

cat > "$temporary/Dockerfile" <<'EOF'
FROM scratch
COPY . /
EOF

docker buildx build --file "$temporary/Dockerfile" --output "type=local,dest=$output" "$context"

for path in .dev.env .env .env.local .env.production .git/config docs/README.md docs/spec/private.md docs/plan/private.md docs/phases/private.md docs/adr/private.md .cache/bootstrap/private.bin .cache/security/nomyr-unexpected deploy/private.env deploy/compose/private.env deploy/compose/images/keycloak/README.md deploy/compose/images/temporal/README.md web/node_modules/private.js web/.next/private.js web/out/private.js node_modules/private.js; do
	if [ -e "$output/$path" ]; then
		echo "ignored Docker context path was exported: $path" >&2
		exit 1
	fi
done

for binary in nomyr-experience nomyr-control-api nomyr-control-worker nomyr-collection nomyr-action; do
	if [ "$(cat "$output/.cache/security/$binary")" != 'synthetic context-boundary marker' ]; then
		echo "required product binary path was not retained: $binary" >&2
		exit 1
	fi
done
cmp "$root/deploy/compose/images/temporal/dependencies.patch" \
	"$output/deploy/compose/images/temporal/dependencies.patch"

echo 'root Docker context excludes secret/private/cache paths and retains explicit image inputs'
