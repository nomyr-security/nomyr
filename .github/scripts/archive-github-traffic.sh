#!/usr/bin/env bash

set -euo pipefail

source_repository="${SOURCE_REPOSITORY:?SOURCE_REPOSITORY is required}"
archive_root="${ARCHIVE_ROOT:?ARCHIVE_ROOT is required}"
captured_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
capture_date="${captured_at%%T*}"
repository_root="${archive_root}/data/github/${source_repository}"
snapshot_root="${repository_root}/snapshots/${capture_date}"

mkdir -p \
  "${repository_root}/views" \
  "${repository_root}/clones" \
  "${repository_root}/referrers" \
  "${repository_root}/paths" \
  "${snapshot_root}"

gh api "repos/${source_repository}/traffic/views" > "${snapshot_root}/views.json"
gh api "repos/${source_repository}/traffic/clones" > "${snapshot_root}/clones.json"
gh api "repos/${source_repository}/traffic/popular/referrers" > "${snapshot_root}/referrers.json"
gh api "repos/${source_repository}/traffic/popular/paths" > "${snapshot_root}/paths.json"

while IFS= read -r row; do
  day="$(jq -r '.timestamp | split("T")[0]' <<<"${row}")"
  jq -n \
    --arg date "${day}" \
    --argjson count "$(jq -r '.count' <<<"${row}")" \
    --argjson uniques "$(jq -r '.uniques' <<<"${row}")" \
    '{date: $date, count: $count, uniques: $uniques}' \
    > "${repository_root}/views/${day}.json"
done < <(jq -c '.views[]' "${snapshot_root}/views.json")

while IFS= read -r row; do
  day="$(jq -r '.timestamp | split("T")[0]' <<<"${row}")"
  jq -n \
    --arg date "${day}" \
    --argjson count "$(jq -r '.count' <<<"${row}")" \
    --argjson uniques "$(jq -r '.uniques' <<<"${row}")" \
    '{date: $date, count: $count, uniques: $uniques}' \
    > "${repository_root}/clones/${day}.json"
done < <(jq -c '.clones[]' "${snapshot_root}/clones.json")

cp "${snapshot_root}/referrers.json" "${repository_root}/referrers/${capture_date}.json"
cp "${snapshot_root}/paths.json" "${repository_root}/paths/${capture_date}.json"

jq -n \
  --arg repository "${source_repository}" \
  --arg captured_at "${captured_at}" \
  --slurpfile views "${snapshot_root}/views.json" \
  --slurpfile clones "${snapshot_root}/clones.json" \
  --slurpfile referrers "${snapshot_root}/referrers.json" \
  --slurpfile paths "${snapshot_root}/paths.json" \
  '{
    repository: $repository,
    captured_at: $captured_at,
    trailing_14_days: {
      views: {count: $views[0].count, uniques: $views[0].uniques},
      clones: {count: $clones[0].count, uniques: $clones[0].uniques}
    },
    referrers: $referrers[0],
    popular_paths: $paths[0]
  }' > "${repository_root}/latest.json"

{
  printf 'date,views,unique_visitors,clones,unique_cloners\n'
  for view_file in "${repository_root}"/views/*.json; do
    day="$(jq -r '.date' "${view_file}")"
    clone_file="${repository_root}/clones/${day}.json"
    views="$(jq -r '.count' "${view_file}")"
    unique_visitors="$(jq -r '.uniques' "${view_file}")"
    clones=0
    unique_cloners=0
    if [[ -f "${clone_file}" ]]; then
      clones="$(jq -r '.count' "${clone_file}")"
      unique_cloners="$(jq -r '.uniques' "${clone_file}")"
    fi
    printf '%s,%s,%s,%s,%s\n' \
      "${day}" "${views}" "${unique_visitors}" "${clones}" "${unique_cloners}"
  done
} > "${repository_root}/daily.csv"
