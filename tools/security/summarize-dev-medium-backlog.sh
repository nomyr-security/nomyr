#!/bin/sh
set -eu

if [ "$#" -ne 3 ]; then
	echo "usage: $0 IMAGE_REPORT_DIRECTORY OUTPUT_JSON EVIDENCE_URL" >&2
	exit 2
fi

report_dir=$1
output=$2
evidence_url=$3
tmp=$(mktemp)
images_tmp=$(mktemp)
trap 'rm -f "$tmp" "$images_tmp"' EXIT HUP INT TERM

found_report=0
scanner_version=
database_schema=
database_built=
database_source=
for report in "$report_dir"/*.json; do
	[ -f "$report" ] || continue
	case "$report" in
		*.cdx.json) continue ;;
	esac
	found_report=1
	if [ -z "$scanner_version" ]; then
		scanner_version=$(jq -r '.descriptor.version // "unknown"' "$report")
		database_schema=$(jq -r '.descriptor.db.status.schemaVersion // "unknown"' "$report")
		database_built=$(jq -r '.descriptor.db.status.built // "unknown"' "$report")
		database_source=$(jq -r '.descriptor.db.status.from // "unknown"' "$report")
	fi
	source_file=${report%.json}.source
	if [ ! -s "$source_file" ]; then
		echo "missing image reference for Grype report: $report" >&2
		exit 1
	fi
	image_ref=$(cat "$source_file")
	case "$image_ref" in
		registry:mcr.microsoft.com/devcontainers/base:*) continue ;;
		oci-archive:*)
			archive=${image_ref##*/}
			service=${archive%.oci.tar}
			checksum_file="$report_dir/../arm64/OCI-SHA256SUMS"
			checksum=$(awk -v archive="$archive" '$2 == archive { print $1 }' "$checksum_file")
			if [ -z "$checksum" ]; then
				echo "missing retained archive checksum for $image_ref" >&2
				exit 1
			fi
			image_ref="oci-archive:linux/arm64/$service@sha256:$checksum"
			;;
	esac
	platform=linux/amd64
	case "$image_ref" in
		oci-archive:linux/arm64/*) platform=linux/arm64 ;;
	esac
	jq -c --arg image_ref "$image_ref" --arg platform "$platform" '
		{
			image_ref: $image_ref,
			platform: $platform,
			severity_matches: {
				Critical: ([.matches[]? | select(.vulnerability.severity == "Critical")] | length),
				High: ([.matches[]? | select(.vulnerability.severity == "High")] | length),
				Medium: ([.matches[]? | select(.vulnerability.severity == "Medium")] | length),
				Low: ([.matches[]? | select(.vulnerability.severity == "Low")] | length),
				Negligible: ([.matches[]? | select(.vulnerability.severity == "Negligible")] | length),
				Unknown: ([.matches[]? | select(.vulnerability.severity == "Unknown")] | length)
			}
		}
	' "$report" >>"$images_tmp"
	jq -c --arg image_ref "$image_ref" '
		.matches[]?
		| select(.vulnerability.severity == "Medium")
		| {
			id: .vulnerability.id,
			image_ref: $image_ref,
			package: {
				name: .artifact.name,
				version: .artifact.version,
				type: .artifact.type,
				purl: .artifact.purl
			},
			fix_state: (.vulnerability.fix.state // "unknown"),
			fix_versions: (.vulnerability.fix.versions // [])
		}' "$report" >>"$tmp"
done

if [ "$found_report" -ne 1 ]; then
	echo "no Grype JSON reports found in $report_dir" >&2
	exit 1
fi

jq -s --arg evidence_url "$evidence_url" \
	--arg generated_date "$(date -u +%F)" \
	--arg scanner_version "$scanner_version" \
	--arg database_schema "$database_schema" \
	--arg database_built "$database_built" \
	--arg database_source "$database_source" \
	--slurpfile image_reports "$images_tmp" '
    sort_by(.id, .image_ref, .package.name, .package.version, .fix_state)
    | . as $matches
    | {
        scope: "development-only",
        owner: "@nomyr-ai",
        evidence_url: $evidence_url,
        generated_utc_date: $generated_date,
        scanner: {name: "Grype", version: $scanner_version},
        vulnerability_database: {
            schema_version: $database_schema,
            built_utc: $database_built,
            source_with_checksum: $database_source
        },
        scanned_images: ($image_reports | sort_by(.image_ref) | unique_by(.image_ref)),
        blocking_threshold: "High and Critical",
        nonblocking_severity: "Medium (retained and reviewed; not automatically accepted in future scans)",
        medium_package_match_count: ($matches | length),
        unique_medium_advisory_count: ([$matches[].id] | unique | length),
        findings: ($matches | group_by(.id) | map({
            id: .[0].id,
            severity: "Medium",
            image_refs: (map(.image_ref) | unique),
            packages: (map({
                name: .package.name,
                version: .package.version,
                type: .package.type,
                purl: .package.purl,
                fix_state,
                fix_versions
            }) | unique_by([.name, .version, .type, .purl, .fix_state, .fix_versions]) | sort_by(.name, .version))
        }))
    }
' "$tmp" >"$output"
