#!/bin/sh
set -eu

config_home=$(mktemp -d)
trap 'rm -rf "$config_home"' EXIT
printf '{"enabled":false}\n' > "$config_home/telemetry.json"
export NOMYR_CONFIG_HOME="$config_home"

output=$(go run ./cmd/nomyr version)
printf '%s' "$output" | grep -q '^nomyr '
output=$(go run ./cmd/nomyr doctor)
printf '%s' "$output" | grep -q 'product services.*not implemented'
if go run ./cmd/nomyr demo --listen 0.0.0.0:8080 >"$config_home/stdout" 2>"$config_home/stderr"; then
  echo 'unsafe non-loopback demo bind unexpectedly succeeded' >&2
  exit 1
fi
grep -q 'non-loopback' "$config_home/stderr"
