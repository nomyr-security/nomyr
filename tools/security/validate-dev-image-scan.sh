#!/bin/sh
# SPDX-FileCopyrightText: 2026 Nomyr technologies
# SPDX-License-Identifier: AGPL-3.0-only
set -u

if [ "$#" -ne 8 ]; then
	echo "usage: $0 IMAGE GATE GRYPE_EXIT SBOM_EXIT GRYPE_JSON SBOM_JSON SUMMARY FINDINGS" >&2
	exit 2
fi

image=$1
gate=$2
grype_status=$3
syft_status=$4
report=$5
sbom=$6
summary=$7
findings=$8

fail() {
	printf '%s: %s\n' "$image" "$1" >&2
	exit 1
}

case "$gate" in
	high|report-only) ;;
	*) fail "unsupported scan gate: $gate" ;;
esac
case "$grype_status:$syft_status" in
	*[!0-9:]*) fail 'scanner exit status is not numeric' ;;
esac
[ "$grype_status" -ge 0 ] && [ "$syft_status" -ge 0 ] || fail 'scanner exit status is invalid'
[ -s "$report" ] || fail 'Grype JSON report is missing or empty'
[ -s "$sbom" ] || fail 'Syft SBOM is missing or empty'

if ! jq -e '
	type == "object"
	and (.descriptor | type == "object")
	and (.descriptor.version | type == "string" and length > 0)
	and (.matches | type == "array")
	and all(.matches[];
		type == "object"
		and (.vulnerability | type == "object")
		and (.vulnerability.severity | type == "string")
		and (.vulnerability.id | type == "string")
		and (.artifact | type == "object")
		and (.artifact.name | type == "string")
		and (.artifact.version | type == "string")
	)
' "$report" >/dev/null 2>&1; then
	fail 'Grype report is invalid JSON or does not match the expected report schema'
fi
if ! jq -e '
	type == "object"
	and .bomFormat == "CycloneDX"
	and (.specVersion | type == "string" and length > 0)
	and (.components | type == "array")
' "$sbom" >/dev/null 2>&1; then
	fail 'Syft output is invalid JSON or does not match the CycloneDX SBOM schema'
fi
if [ "$syft_status" -ne 0 ]; then
	fail "Syft failed with exit $syft_status"
fi
if [ "$grype_status" -ne 0 ] && [ "$grype_status" -ne 2 ]; then
	fail "Grype failed with exit $grype_status"
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
if ! jq -r --arg image "$image" '
	[.matches[] | select(.vulnerability.severity == "Critical" or .vulnerability.severity == "High" or .vulnerability.severity == "Medium")]
	| "\($image): \(length) Medium-or-higher matches; complete findings in JSON report"
' "$report" >"$tmp/summary"; then
	fail 'could not summarize Grype findings'
fi
if ! jq -r '
	.matches[]
	| [.vulnerability.severity, (.vulnerability.fix.state // "unknown"), (.vulnerability.fix.versions // [] | join(",")), .vulnerability.id, .artifact.name, .artifact.version]
	| @tsv
' "$report" >"$tmp/findings"; then
	fail 'could not extract Grype findings'
fi
if ! sort -u "$tmp/findings" >>"$findings"; then
	fail 'could not persist Grype findings'
fi
if ! cat "$tmp/summary" >>"$summary"; then
	fail 'could not persist Grype summary'
fi

has_blocking_finding=0
if [ "$gate" = high ]; then
	if jq -e 'any(.matches[]; .vulnerability.severity == "Critical" or .vulnerability.severity == "High")' "$report" >/dev/null; then
		has_blocking_finding=1
	else
		jq_status=$?
		if [ "$jq_status" -ne 1 ]; then
			fail 'could not evaluate High/Critical findings'
		fi
	fi
fi

if { [ "$gate" = high ] && [ "$grype_status" -eq 2 ]; } || [ "$has_blocking_finding" -eq 1 ]; then
	printf '%s: High-or-higher vulnerability threshold exceeded\n' "$image" >>"$summary"
	exit 2
fi
if [ "$grype_status" -eq 2 ]; then
	fail 'Grype returned threshold-exceeded for a report-only scan'
fi
