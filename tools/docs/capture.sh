#!/bin/sh
set -eu

config_home=$(mktemp -d)
trap 'rm -rf "$config_home"' EXIT
printf '{"enabled":false}\n' > "$config_home/telemetry.json"
export NOMYR_CONFIG_HOME="$config_home"
mkdir -p docs-site/captures
vhs tools/docs/quickstart.tape
