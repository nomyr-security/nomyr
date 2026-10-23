#!/usr/bin/env bash
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
source "$root/tools/contracts/versions.env"
cd "$root"

check=false
config=tools/contracts/oapi-codegen.yaml
if [[ "${1:-}" == "--check" ]]; then
  check=true
  temp_dir="$(mktemp -d)"
  trap 'rm -rf "$temp_dir"' EXIT
  go_output="$temp_dir/openapi.gen.go"
  ts_output="$temp_dir/openapi.d.ts"
  config="$temp_dir/oapi-codegen.yaml"
  sed "s|^output: .*|output: $go_output|" tools/contracts/oapi-codegen.yaml > "$config"
else
  go_output=internal/generated/openapi.gen.go
  ts_output=web/lib/api/openapi.d.ts
fi

go run "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$OAPI_CODEGEN_VERSION" \
  -config "$config" api/openapi/v1.yaml
npx --yes "openapi-typescript@$OPENAPI_TYPESCRIPT_VERSION" api/openapi/v1.yaml \
  -o "$ts_output"

if "$check"; then
  cmp "$go_output" internal/generated/openapi.gen.go
  cmp "$ts_output" web/lib/api/openapi.d.ts
  echo "Generated OpenAPI outputs match the checked-in files."
fi
