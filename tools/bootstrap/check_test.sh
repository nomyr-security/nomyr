#!/bin/sh
set -eu

temp_dir=$(mktemp -d)
trap 'rm -rf "$temp_dir"' EXIT HUP INT TERM

if PATH="$temp_dir" /bin/sh tools/bootstrap/check >"$temp_dir/no-mise.out" 2>&1; then
	printf '%s\n' 'expected check to fail when mise is missing' >&2
	exit 1
fi
/usr/bin/grep -q 'missing: mise' "$temp_dir/no-mise.out"

for tool in mise go node npm buf golangci-lint; do
	cat >"$temp_dir/$tool" <<'EOF'
#!/bin/sh
if [ "$1" = version ] && [ "${2-}" = -m ]; then
	printf '%s: %s\n' "$3" "${TEST_LINTER_GO_VERSION:-go1.26.10}"
	exit 0
fi
if [ "$1" = version ]; then
	printf '%s\n' 'go version go1.26.10 darwin/arm64'
	exit 0
fi
printf '%s\n' 'test-version'
EOF
	chmod +x "$temp_dir/$tool"
done

if PATH="$temp_dir:/usr/bin:/bin" /bin/sh tools/bootstrap/check >"$temp_dir/missing-tool.out" 2>&1; then
	printf '%s\n' 'expected check to fail when govulncheck is missing' >&2
	exit 1
fi
/usr/bin/grep -q 'available: go (go version go1.26.10 darwin/arm64)' "$temp_dir/missing-tool.out"
/usr/bin/grep -q 'missing: govulncheck' "$temp_dir/missing-tool.out"

cat >"$temp_dir/govulncheck" <<'EOF'
#!/bin/sh
if [ "$1" = -version ]; then
	printf '%s\n' 'govulncheck v1.9.0'
	exit 0
fi
if [ "${TEST_VULN_FAIL:-0}" = 1 ]; then
	printf '%s\n' 'synthetic binary vulnerability finding'
	exit 3
fi
printf '%s\n' 'No vulnerabilities found.'
EOF
chmod +x "$temp_dir/govulncheck"

if ! PATH="$temp_dir:/usr/bin:/bin" /bin/sh tools/bootstrap/check >"$temp_dir/valid.out" 2>&1; then
	cat "$temp_dir/valid.out" >&2
	printf '%s\n' 'expected check to accept matching compiler metadata and a clean binary scan' >&2
	exit 1
fi
/usr/bin/grep -q 'verified: golangci-lint built with selected Go (go1.26.10)' "$temp_dir/valid.out"
/usr/bin/grep -q 'No vulnerabilities found.' "$temp_dir/valid.out"

if PATH="$temp_dir:/usr/bin:/bin" TEST_LINTER_GO_VERSION=go1.27.0 /bin/sh tools/bootstrap/check >"$temp_dir/mismatch.out" 2>&1; then
	printf '%s\n' 'expected check to reject a linter built with a different Go toolchain' >&2
	exit 1
fi
/usr/bin/grep -q 'mismatch: golangci-lint was built with go1.27.0; selected Go is go1.26.10' "$temp_dir/mismatch.out"

if PATH="$temp_dir:/usr/bin:/bin" TEST_VULN_FAIL=1 /bin/sh tools/bootstrap/check >"$temp_dir/vulnerable.out" 2>&1; then
	printf '%s\n' 'expected check to fail when the binary vulnerability scan finds a vulnerability' >&2
	exit 1
fi
/usr/bin/grep -q 'synthetic binary vulnerability finding' "$temp_dir/vulnerable.out"
