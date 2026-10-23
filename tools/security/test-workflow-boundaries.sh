#!/bin/sh
# SPDX-FileCopyrightText: 2026 Nomyr technologies
# SPDX-License-Identifier: AGPL-3.0-only
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
workflow="$root/.github/workflows/security-supply-chain.yml"
scorecard="$root/.github/workflows/scorecard.yml"

for file in "$root"/.github/workflows/*.yml; do
	uses_count=$(grep -Ec '^[[:space:]]+- uses: [^[:space:]@]+@[0-9a-f]{40}([[:space:]]|$)' "$file" || true)
	all_uses_count=$(grep -Ec '^[[:space:]]+- uses:' "$file" || true)
	if [ "$uses_count" -ne "$all_uses_count" ]; then
		echo "every action reference must use a full commit SHA: $file" >&2
		exit 1
	fi
done

grep -Fq "if: github.event_name == 'push' && github.ref == 'refs/heads/main'" "$workflow"
grep -Fq 'id-token: write' "$workflow"
if grep -Eq 'contents: write|pull-requests: write|packages: write' "$workflow"; then
	echo 'untrusted build workflow must not grant repository write permissions' >&2
	exit 1
fi

grep -Fq "if: github.event_name == 'pull_request'" "$workflow"
grep -Fq "if: github.ref == 'refs/heads/main'" "$scorecard"
grep -Fq 'publish_results: false' "$scorecard"
grep -Fq 'publish_results: true' "$scorecard"
grep -Fq 'npm audit --audit-level=moderate' "$workflow"
grep -Fq 'pip-audit --strict --requirement tools/contracts/requirements.txt' "$workflow"
grep -Fq 'pip install --disable-pip-version-check --require-hashes --only-binary=:all:' "$workflow"
grep -Fq 'docker-archive:$RUNNER_TEMP/nomyr-${{ matrix.plane }}.image.tar' "$workflow"
grep -Fq 'sbom:$RUNNER_TEMP/nomyr-${{ matrix.plane }}.cdx.json' "$workflow"
if grep -Eq -- '--ignore-vuln|--exclude|--only-fixed|fail-on high|fail-on critical' "$workflow"; then
	echo 'security scans must not suppress findings or weaken the medium threshold' >&2
	exit 1
fi

echo 'workflow actions are SHA-pinned and write-capable jobs are trust-gated'
