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
printf '%s\n' 'test-version'
EOF
	chmod +x "$temp_dir/$tool"
done

if PATH="$temp_dir:/usr/bin:/bin" /bin/sh tools/bootstrap/check >"$temp_dir/missing-tool.out" 2>&1; then
	printf '%s\n' 'expected check to fail when govulncheck is missing' >&2
	exit 1
fi
/usr/bin/grep -q 'available: go (test-version)' "$temp_dir/missing-tool.out"
/usr/bin/grep -q 'missing: govulncheck' "$temp_dir/missing-tool.out"
