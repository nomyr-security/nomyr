#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
dev_workflow="$root/.github/workflows/dev-compose.yml"
product_workflow="$root/.github/workflows/security-supply-chain.yml"

grep -Fq '"$scheme:$image" --fail-on high --output json --file "$report"' "$dev_workflow"
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

echo 'development-only Medium backlog and product Medium gate boundaries passed'
