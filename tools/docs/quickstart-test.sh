#!/bin/sh
set -eu

config_home=$(mktemp -d)
demo_pid=
trap 'if [ -n "$demo_pid" ]; then kill "$demo_pid" 2>/dev/null || true; wait "$demo_pid" 2>/dev/null || true; fi; rm -rf "$config_home"' EXIT
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
go build -o "$config_home/nomyr" ./cmd/nomyr
"$config_home/nomyr" demo --listen 127.0.0.1:18082 >"$config_home/demo.log" 2>&1 &
demo_pid=$!
ready=false
attempt=0
while [ "$attempt" -lt 60 ]; do
  attempt=$((attempt + 1))
  if curl --noproxy '*' --fail --silent http://127.0.0.1:18082/api/v1/healthz >"$config_home/health.json"; then
    ready=true
    break
  fi
  if ! kill -0 "$demo_pid" 2>/dev/null; then
    cat "$config_home/demo.log" >&2
    exit 1
  fi
  sleep 1
done
if [ "$ready" != true ]; then
  cat "$config_home/demo.log" >&2
  echo 'loopback demo did not become ready' >&2
  exit 1
fi
grep -q '"status":"ok"' "$config_home/health.json"
curl --noproxy '*' --fail --silent http://127.0.0.1:18082/api/v1/demo-estate >"$config_home/estate.json"
grep -q '"watermark":"demo data"' "$config_home/estate.json"
grep -q '"provider":"synthetic"' "$config_home/estate.json"
