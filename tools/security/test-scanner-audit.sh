#!/bin/sh
# SPDX-FileCopyrightText: 2026 Nomyr technologies
# SPDX-License-Identifier: AGPL-3.0-only
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
mkdir -p "$tmp/bin" "$tmp/reports" "$tmp/mock-bin"
go_binary=$(command -v go)
go_binary=$(CDPATH= cd -- "$(dirname -- "$go_binary")" && pwd)/$(basename "$go_binary")
for scanner in syft gitleaks grype; do
	ln -s "$go_binary" "$tmp/bin/$scanner"
done
printf '%s\n' github.com/docker/docker/client > "$tmp/reports/grype.packages.txt"
printf '%s\n' github.com/example/tool > "$tmp/reports/syft.packages.txt"
printf '%s\n' github.com/example/tool > "$tmp/reports/gitleaks.packages.txt"

cat > "$tmp/mock-bin/govulncheck" <<'EOF'
#!/bin/sh
set -eu
binary=
for argument do binary=$argument; done
cat "$MOCK_REPORT_DIR/$(basename "$binary").json"
EOF
chmod +x "$tmp/mock-bin/govulncheck"

write_findings() {
	scanner=$1
	shift
	: > "$tmp/mock-reports/$scanner.json"
	for vulnerability do
		jq -nc --arg id "$vulnerability" '{finding:{osv:$id}}' >> "$tmp/mock-reports/$scanner.json"
	done
}
mkdir -p "$tmp/mock-reports"
write_findings syft GO-2026-5932
write_findings gitleaks GO-2026-5932
write_findings grype GO-2026-4883 GO-2026-4887 GO-2026-5617 GO-2026-5668 GO-2026-5746 GO-2026-5932

PATH="$tmp/mock-bin:$PATH" MOCK_REPORT_DIR="$tmp/mock-reports" sh "$root/tools/security/audit-scanner-binaries.sh" "$tmp"
grep -Fxq GO-2026-5932 "$tmp/reports/syft.finding-ids.txt"
jq -e '.statements | length == 1' "$tmp/reports/syft.openvex.json" >/dev/null

# A new module-wide finding without any trace must fail rather than disappear.
write_findings syft GO-2026-5932 GO-2026-NEW-TRACELESS
if PATH="$tmp/mock-bin:$PATH" MOCK_REPORT_DIR="$tmp/mock-reports" sh "$root/tools/security/audit-scanner-binaries.sh" "$tmp" > "$tmp/unexpected.stdout" 2> "$tmp/unexpected.stderr"; then
	echo 'audit accepted an unreviewed trace-less finding' >&2
	exit 1
fi
grep -Fq GO-2026-NEW-TRACELESS "$tmp/unexpected.stderr"

# A source graph that now includes the excluded package invalidates the VEX.
write_findings syft GO-2026-5932
printf '%s\n' golang.org/x/crypto/openpgp > "$tmp/reports/syft.packages.txt"
if PATH="$tmp/mock-bin:$PATH" MOCK_REPORT_DIR="$tmp/mock-reports" sh "$root/tools/security/audit-scanner-binaries.sh" "$tmp" > "$tmp/unexpected.stdout" 2> "$tmp/unexpected.stderr"; then
	echo 'audit accepted an excluded package added to the source graph' >&2
	exit 1
fi
grep -Fq 'includes excluded package golang.org/x/crypto/openpgp' "$tmp/unexpected.stderr"

echo 'scanner audit retains trace-less findings and rejects changed source graphs'
