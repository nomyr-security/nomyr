#!/bin/sh
# SPDX-FileCopyrightText: 2026 Nomyr technologies
# SPDX-License-Identifier: AGPL-3.0-only
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
. "$root/tools/security/scanners.env"

out=${1:?usage: audit-cosign-binary.sh OUTPUT_DIRECTORY}
mkdir -p "$out/reports"
out=$(CDPATH= cd -- "$out" && pwd)
reports="$out/reports"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM

if ! govulncheck -version | grep -Fq "govulncheck@v$GOVULNCHECK_VERSION"; then
	echo "Cosign binary audit requires govulncheck v$GOVULNCHECK_VERSION" >&2
	exit 1
fi

binary="$tmp/cosign-linux-amd64"
curl --proto '=https' --proto-redir '=https' -fsSL "https://github.com/sigstore/cosign/releases/download/v$COSIGN_VERSION/cosign-linux-amd64" -o "$binary"
printf '%s  %s\n' "$COSIGN_LINUX_AMD64_SHA256" "$binary" | sha256sum --check -

go version -m "$binary" > "$reports/cosign.buildinfo.txt"
if [ "$(sed -n '1s/.*: //p' "$reports/cosign.buildinfo.txt")" != "go1.27.3" ]; then
	echo 'Cosign release Go runtime changed; review compiler and runtime advisories' >&2
	exit 1
fi
grep -Eq "^[[:space:]]*mod[[:space:]]+github.com/sigstore/cosign/v3[[:space:]]+v$COSIGN_VERSION[[:space:]]" "$reports/cosign.buildinfo.txt"

go tool nm "$binary" > "$tmp/cosign.symbols.txt"
if grep -Fq 'golang.org/x/crypto/openpgp' "$tmp/cosign.symbols.txt"; then
	echo 'Cosign binary now links an advisory-affected OpenPGP package' >&2
	exit 1
fi
grep -F 'docker-credential-acr-env/pkg/credhelper' "$tmp/cosign.symbols.txt" > "$reports/cosign.acr-helper-symbols.txt"
if ! grep -Eq 'ACRCredHelper\)\.Get$|ACRCredHelper\.Get$' "$reports/cosign.acr-helper-symbols.txt"; then
	echo 'Expected linked ACR helper evidence was not found; review GO-2026-6225 disposition' >&2
	exit 1
fi
printf 'golang.org/x/crypto/openpgp symbol count: 0\n' > "$reports/cosign.openpgp-symbol-check.txt"

set +e
govulncheck -json -mode=binary "$binary" > "$reports/cosign.govulncheck.json" 2> "$reports/cosign.govulncheck.stderr"
scan_exit=$?
set -e
if [ "$scan_exit" -ne 0 ]; then
	cat "$reports/cosign.govulncheck.stderr" >&2
	echo "govulncheck failed for Cosign release binary with status $scan_exit" >&2
	exit "$scan_exit"
fi
jq -s -r '.[] | select(.finding != null) | .finding.osv' "$reports/cosign.govulncheck.json" | LC_ALL=C sort -u > "$reports/cosign.finding-ids.txt"
cat > "$tmp/expected-finding-ids.txt" <<'EOF'
GO-2026-5932
GO-2026-6225
EOF
if ! cmp -s "$tmp/expected-finding-ids.txt" "$reports/cosign.finding-ids.txt"; then
	echo 'Cosign release binary finding IDs changed; review every raw finding before updating the disposition' >&2
	diff -u "$tmp/expected-finding-ids.txt" "$reports/cosign.finding-ids.txt" >&2 || true
	exit 1
fi
if ! jq -s -e 'any(.[]; .finding.osv == "GO-2026-6225" and any(.finding.trace[]?; .module == "github.com/chrismellard/docker-credential-acr-env" and .package == "github.com/chrismellard/docker-credential-acr-env/pkg/credhelper" and .function == "Get"))' "$reports/cosign.govulncheck.json" >/dev/null; then
	echo 'GO-2026-6225 trace changed; review the source-bound applicability evidence' >&2
	exit 1
fi

cat > "$reports/cosign.audit.txt" <<EOF
source=https://github.com/sigstore/cosign/tree/v$COSIGN_VERSION
release_asset=https://github.com/sigstore/cosign/releases/download/v$COSIGN_VERSION/cosign-linux-amd64
sha256=$COSIGN_LINUX_AMD64_SHA256
go_runtime=1.27.3
govulncheck=$GOVULNCHECK_VERSION
raw_findings=GO-2026-5932,GO-2026-6225
disposition=command-bound only; see tools/security/README.md and exact workflow boundary tests
EOF

echo 'Cosign binary audit retained all raw finding IDs and exact release evidence'
