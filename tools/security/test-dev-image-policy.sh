#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
dev_workflow="$root/.github/workflows/dev-compose.yml"
product_workflow="$root/.github/workflows/security-supply-chain.yml"

grep -Fq '"$scheme:$image" --fail-on high --output json --file "$report"' "$dev_workflow"
grep -Fq 'validate-dev-image-scan.sh' "$dev_workflow"
grep -Fq -- '--fail-on medium --output table' "$product_workflow"
grep -Fq -- '--fail-on medium --output sarif' "$product_workflow"
if grep -Fq -- '--fail-on high' "$product_workflow"; then
	echo 'product-plane image threshold must remain Medium' >&2
	exit 1
fi
grep -Fq 'docs/dev-image-vulnerabilities.md' "$dev_workflow"
grep -Fq 'issues/67' "$root/docs/dev-image-vulnerabilities.md"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
mkdir -p "$tmp/images"
printf '%s\n' 'docker:nomyr-dev/service:scan' >"$tmp/images/service.source"
printf '%s\n' 'sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef' >"$tmp/images/service.image-id"
cat >"$tmp/images/service.json" <<'JSON'
{
  "matches": [
    {
      "artifact": {"name": "package-a", "version": "1.2.3", "type": "apk", "purl": "pkg:apk/package-a@1.2.3"},
      "vulnerability": {"id": "CVE-TEST-MEDIUM", "severity": "Medium", "fix": {"state": "not-fixed", "versions": []}}
    },
    {
      "artifact": {"name": "package-b", "version": "4.5.6", "type": "apk", "purl": "pkg:apk/package-b@4.5.6"},
      "vulnerability": {"id": "CVE-TEST-MEDIUM", "severity": "Medium", "fix": {"state": "fixed", "versions": ["4.5.7"]}}
    },
    {
      "artifact": {"name": "package-c", "version": "7.8.9", "type": "apk", "purl": "pkg:apk/package-c@7.8.9"},
      "vulnerability": {"id": "CVE-TEST-LOW", "severity": "Low", "fix": {"state": "not-fixed", "versions": []}}
    }
  ]
}
JSON
sh "$root/tools/security/summarize-dev-medium-backlog.sh" "$tmp/images" "$tmp/inventory.json" 'https://example.invalid/run/1'
jq -e '
    .scope == "development-only"
    and .owner == "@nomyr-ai"
    and .evidence_url == "https://example.invalid/run/1"
    and .scanner.name == "Grype"
    and .blocking_threshold == "High and Critical"
    and .medium_package_match_count == 2
    and .unique_medium_advisory_count == 1
    and .scanned_images[0].platform == "linux/amd64"
    and .scanned_images[0].image_id == "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
    and .scanned_images[0].severity_matches.Medium == 2
    and .findings[0].id == "CVE-TEST-MEDIUM"
    and (.findings[0].packages | length) == 2
    and any(.findings[0].packages[]; .fix_state == "fixed" and .fix_versions == ["4.5.7"])
    and any(.findings[0].packages[]; .fix_state == "not-fixed" and .fix_versions == [])
    and (all(.findings[]; .id != "CVE-TEST-LOW"))
' "$tmp/inventory.json" >/dev/null
if sh "$root/tools/security/summarize-dev-medium-backlog.sh" "$tmp/missing" "$tmp/empty.json" 'https://example.invalid/run/1' >/dev/null 2>&1; then
	echo 'inventory generation unexpectedly accepted missing reports' >&2
	exit 1
fi

validator="$root/tools/security/validate-dev-image-scan.sh"
report="$tmp/grype.json"
sbom="$tmp/sbom.json"
summary="$tmp/scan-summary.txt"
findings="$tmp/findings.tsv"
cat >"$report" <<'JSON'
{"descriptor":{"version":"0.121.0"},"matches":[]}
JSON
cat >"$sbom" <<'JSON'
{"bomFormat":"CycloneDX","specVersion":"1.6","components":[]}
JSON
run_validator() {
	sh "$validator" image:tag "$1" "$2" "$3" "$report" "$sbom" "$summary" "$findings"
}
run_validator high 0 0
if sh "$validator" image:tag high 0 0 "$tmp/missing.json" "$sbom" "$summary" "$findings" >/dev/null 2>&1; then
	echo 'validator accepted a missing Grype report' >&2
	exit 1
fi
: >"$report"
if run_validator high 0 0 >/dev/null 2>&1; then
	echo 'validator accepted an empty Grype report' >&2
	exit 1
fi
printf '%s\n' '{not-json' >"$report"
if run_validator high 0 0 >/dev/null 2>&1; then
	echo 'validator accepted malformed Grype JSON' >&2
	exit 1
fi
printf '%s\n' '{"descriptor":{"version":"0.121.0"},"matches":{}}' >"$report"
if run_validator high 0 0 >/dev/null 2>&1; then
	echo 'validator accepted a Grype report with an invalid matches schema' >&2
	exit 1
fi
cat >"$report" <<'JSON'
{"descriptor":{"version":"0.121.0"},"matches":[{"artifact":{"name":"pkg","version":"1"},"vulnerability":{"id":"CVE-HIGH","severity":"High"}}]}
JSON
if run_validator high 0 0 >/dev/null 2>&1; then
	echo 'High finding passed the development gate when Grype returned success' >&2
	exit 1
else
	test "$?" = 2
fi
cat >"$report" <<'JSON'
{"descriptor":{"version":"0.121.0"},"matches":[{"artifact":{"name":"pkg","version":"1"},"vulnerability":{"id":"CVE-MEDIUM","severity":"Medium"}}]}
JSON
run_validator high 0 0
if run_validator high 1 0 >/dev/null 2>&1; then
	echo 'validator accepted a Grype operational error' >&2
	exit 1
fi
if run_validator high 0 1 >/dev/null 2>&1; then
	echo 'validator accepted a Syft operational error' >&2
	exit 1
fi
printf '%s\n' '{"bomFormat":"SPDX","components":[]}' >"$sbom"
if run_validator high 0 0 >/dev/null 2>&1; then
	echo 'validator accepted an invalid CycloneDX SBOM' >&2
	exit 1
fi

mkdir -p "$tmp/oci-images" "$tmp/arm64"
printf '%s\n' 'oci-archive:/tmp/service.oci.tar' >"$tmp/oci-images/service.source"
cp "$report" "$tmp/oci-images/service.json"
printf '%s  %s\n' 'abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789' 'service.oci.tar' >"$tmp/arm64/OCI-SHA256SUMS"
sh "$root/tools/security/summarize-dev-medium-backlog.sh" "$tmp/oci-images" "$tmp/oci-inventory.json" 'https://example.invalid/run/2'
jq -e '.scanned_images[0].image_ref == "oci-archive:linux/arm64/service@sha256:abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"' "$tmp/oci-inventory.json" >/dev/null
	summarize_line=$(grep -n 'summarize-dev-medium-backlog.sh' "$dev_workflow" | head -n 1 | cut -d: -f1)
	checksum_line=$(grep -n 'sha256sum.*basename.*archive' "$dev_workflow" | cut -d: -f1)
	delete_line=$(grep -n 'find.*-delete' "$dev_workflow" | cut -d: -f1)
if [ "$checksum_line" -lt "$summarize_line" ] && [ "$delete_line" -gt "$summarize_line" ]; then :; else
	echo 'OCI checksum generation/deletion must bracket durable inventory summarization' >&2
	exit 1
fi

echo 'development-only Medium backlog and product Medium gate boundaries passed'
