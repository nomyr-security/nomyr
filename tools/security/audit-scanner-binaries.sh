#!/bin/sh
# SPDX-FileCopyrightText: 2026 Nomyr technologies
# SPDX-License-Identifier: AGPL-3.0-only
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
. "$root/tools/security/scanners.env"
out=${1:?usage: audit-scanner-binaries.sh OUTPUT_DIRECTORY}
mkdir -p "$out/reports"

for scanner in syft gitleaks grype; do
	binary="$out/bin/$scanner"
	packages="$out/reports/$scanner.packages.txt"
	if [ ! -x "$binary" ] || [ ! -s "$packages" ]; then
		echo "missing scanner binary or build-time package graph for $scanner" >&2
		exit 1
	fi
	set +e
	govulncheck -json -mode=binary "$binary" > "$out/reports/$scanner.govulncheck.json"
	status=$?
	set -e
	if [ "$status" -ne 0 ]; then
		echo "govulncheck failed for $scanner with status $status" >&2
		exit "$status"
	fi
	go tool nm "$binary" > "$out/reports/$scanner.symbols.txt"
	jq -s -r '[.[] | objects | .finding? | select(type == "object") | .osv] | unique | sort | .[]' \
		"$out/reports/$scanner.govulncheck.json" > "$out/reports/$scanner.finding-ids.txt"
done

expect_ids() {
	scanner=$1
	shift
	printf '%s\n' "$@" | sort -u > "$out/reports/$scanner.expected-finding-ids.txt"
	if ! cmp -s "$out/reports/$scanner.expected-finding-ids.txt" "$out/reports/$scanner.finding-ids.txt"; then
		echo "$scanner govulncheck IDs changed; review every raw finding and source graph" >&2
		diff -u "$out/reports/$scanner.expected-finding-ids.txt" "$out/reports/$scanner.finding-ids.txt" >&2 || true
		exit 1
	fi
}

# Every raw finding ID is compared, including module-wide findings and findings
# whose traces omit function/package data. The raw JSON and exact build-time
# source graphs are retained alongside these reviewed, version-bound sets.
expect_ids syft GO-2026-5932
expect_ids gitleaks GO-2026-5932
expect_ids grype GO-2026-4883 GO-2026-4887 GO-2026-5617 GO-2026-5668 GO-2026-5746 GO-2026-5932

assert_package_absent() {
	scanner=$1
	package_prefix=$2
	if awk -v prefix="$package_prefix" 'index($0, prefix) == 1 && (length($0) == length(prefix) || substr($0, length(prefix) + 1, 1) == "/") { found = 1 } END { exit !found }' "$out/reports/$scanner.packages.txt"; then
		echo "$scanner build-time source graph includes excluded package $package_prefix" >&2
		exit 1
	fi
	if grep -Fq "$package_prefix" "$out/reports/$scanner.symbols.txt"; then
		echo "$scanner binary symbols include excluded package $package_prefix" >&2
		exit 1
	fi
}

for scanner in syft gitleaks grype; do
	assert_package_absent "$scanner" golang.org/x/crypto/openpgp
done
for package in github.com/docker/docker/daemon github.com/docker/docker/daemon/pkg/authz github.com/docker/docker/daemon/pkg/plugin; do
	assert_package_absent grype "$package"
done
if ! grep -Fxq github.com/docker/docker/client "$out/reports/grype.packages.txt"; then
	echo 'Grype build-time dependency graph no longer includes the expected Docker client package' >&2
	exit 1
fi

grype_binary_sha=$(sha256sum "$out/bin/grype" | awk '{print $1}')
gitleaks_binary_sha=$(sha256sum "$out/bin/gitleaks" | awk '{print $1}')
syft_binary_sha=$(sha256sum "$out/bin/syft" | awk '{print $1}')
cat > "$out/reports/scanner-vex-not-applicable.txt" <<EOF
Disposition: not_affected, bounded to exact source and binary identities (not a blanket ignore).
Raw govulncheck JSON, every finding ID, source graph captured during the matching CGO/build-tag build, and binary symbol listing are retained in this directory.
Syft: v$SYFT_VERSION commit $SYFT_COMMIT module checksum $SYFT_SUM binary SHA-256 $syft_binary_sha.
Gitleaks: v$GITLEAKS_VERSION commit $GITLEAKS_COMMIT module checksum $GITLEAKS_SUM binary SHA-256 $gitleaks_binary_sha; transitive dependency patch is tools/security/patches/gitleaks-transitive-security-updates.patch.
Grype: v$GRYPE_VERSION commit $GRYPE_COMMIT module checksum $GRYPE_SUM binary SHA-256 $grype_binary_sha; Docker module $DOCKER_MODULE_VERSION commit $DOCKER_MODULE_COMMIT checksum $DOCKER_MODULE_SUM.
GO-2026-5932 (golang.org/x/crypto/openpgp) is retained in Syft, Gitleaks, and Grype raw reports. Its affected OpenPGP package is absent from each exact build-time package graph and binary symbol listing.
GO-2026-4883 (CVE-2026-33997 / GHSA-pxq6-2prw-chj9), GO-2026-4887 (CVE-2026-34040 / GHSA-x744-4wpc-v9h2), GO-2026-5617 (CVE-2026-42306 / GHSA-rg2x-37c3-w2rh), GO-2026-5668 (CVE-2026-41568 / GHSA-vp62-88p7-qqf5), and GO-2026-5746 (CVE-2026-41567 / GHSA-x86f-5xw2-fm2r) are retained in Grype raw reports. Their advisory-affected Docker daemon packages are absent from Grype's exact build-time package graph and binary symbols; the Docker client package remains present.
Official Go records: https://pkg.go.dev/vuln/GO-2026-5932 https://pkg.go.dev/vuln/GO-2026-4883 https://pkg.go.dev/vuln/GO-2026-4887 https://pkg.go.dev/vuln/GO-2026-5617 https://pkg.go.dev/vuln/GO-2026-5668 https://pkg.go.dev/vuln/GO-2026-5746
The Go records are unreviewed/module-wide and govulncheck traces alone do not establish reachability. These source-bound dispositions are valid only for the listed scanner source identities and while all exact-ID, package-graph, and symbol checks pass. Accountable human owner remains unassigned.
EOF

for scanner in syft gitleaks grype; do
	case "$scanner" in
		syft) version=$SYFT_VERSION; commit=$SYFT_COMMIT; checksum=$SYFT_SUM; binary_sha=$syft_binary_sha ;;
		gitleaks) version=$GITLEAKS_VERSION; commit=$GITLEAKS_COMMIT; checksum=$GITLEAKS_SUM; binary_sha=$gitleaks_binary_sha ;;
		grype) version=$GRYPE_VERSION; commit=$GRYPE_COMMIT; checksum=$GRYPE_SUM; binary_sha=$grype_binary_sha ;;
	esac
	product="pkg:golang/$scanner@$version"
	statements=$(jq -Rsc \
		--arg product "$product" \
		--arg commit "$commit" \
		--arg checksum "$checksum" \
		--arg binary_sha "$binary_sha" \
		--arg scanner "$scanner" \
		'split("\n") | map(select(length > 0) | {"vulnerability":{"name":.},"products":[{"@id":$product}],"status":"not_affected","justification":"vulnerable_code_not_in_execute_path","impact_statement":("Bound to " + $scanner + " source commit " + $commit + " (module checksum " + $checksum + ") and binary SHA-256 " + $binary_sha + ". Raw finding is retained. The matching build-time source graph and unstripped binary symbols prove the advisory-affected package is absent." )})' \
		"$out/reports/$scanner.finding-ids.txt")
	jq -n \
		--arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
		--arg scanner "$scanner" \
		--arg commit "$commit" \
		--argjson statements "$statements" \
		'{"@context":"https://openvex.dev/ns/v0.2.0","@id":("https://github.com/nomyr-security/nomyr/security/vex/" + $scanner + "-" + $commit),"author":"Nomyr project security maintainers (role; individual owner unassigned)","timestamp":$timestamp,"version":1,"statements":$statements}' \
		> "$out/reports/$scanner.openvex.json"
done

echo 'scanner binary audit completed; all raw finding IDs retained and compared against exact reviewed sets'
