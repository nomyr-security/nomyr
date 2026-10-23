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
grep -Fq 'python -m pip install --disable-pip-version-check --require-hashes' "$workflow"
grep -Fq -- '--only-binary=:all: --requirement tools/security/pip-audit-requirements.txt' "$workflow"
grep -Fq 'name: Build and audit pinned scanner tools' "$workflow"
grep -Fq 'tools/security/build-scanners.sh' "$workflow"
grep -Fq 'tools/security/audit-scanner-binaries.sh' "$workflow"
grep -Fq 'run: sh tools/security/build-scanners.sh' "$workflow"
grep -Fq 'run: sh tools/security/audit-scanner-binaries.sh' "$workflow"
grep -Fq 'tools/security/test-scanner-audit.sh' "$root/tools/security/test-workflow-boundaries.sh"
grep -Fq "go build -tags 'exclude_graphdriver_btrfs'" "$root/tools/security/build-scanners.sh"
grep -Fq "CGO_ENABLED=0 go list -tags 'exclude_graphdriver_btrfs' -deps" "$root/tools/security/build-scanners.sh"
grep -Fq "CGO_ENABLED=0 go list -tags 'containers_image_openpgp exclude_graphdriver_btrfs' -deps" "$root/tools/security/build-scanners.sh"
if grep -Fq -- '-w -s' "$root/tools/security/build-scanners.sh"; then
	echo 'scanner binaries must retain symbols for binary package auditing' >&2
	exit 1
fi
grep -Fq 'SYFT_COMMIT=f15c3343e29dfb20cf6d82c820e7090b49808c96' "$root/tools/security/scanners.env"
grep -Fq 'GRYPE_COMMIT=afe2742527d98d822da06c9f32eed86fe3d193e0' "$root/tools/security/scanners.env"
grep -Fq 'GITLEAKS_COMMIT=8d1f98c7967eb1e79cb44ac6241a124e145d2165' "$root/tools/security/scanners.env"
grep -Fq 'github.com/ulikunitz/xz v0.5.17' "$root/tools/security/patches/gitleaks-transitive-security-updates.patch"
grep -Fq 'github.com/klauspost/compress v1.20.3' "$root/tools/security/patches/gitleaks-transitive-security-updates.patch"
grep -Fq 'github.com/nwaples/rardecode/v2 v2.4.5' "$root/tools/security/patches/gitleaks-transitive-security-updates.patch"
grep -Fq 'golang.org/x/crypto v0.58.0' "$root/tools/security/patches/gitleaks-transitive-security-updates.patch"
grep -Fq 'golang.org/x/text v0.43.0' "$root/tools/security/patches/gitleaks-transitive-security-updates.patch"
grep -Fq 'GO-2026-4883' "$root/tools/security/audit-scanner-binaries.sh"
grep -Fq 'GO-2026-4887' "$root/tools/security/audit-scanner-binaries.sh"
grep -Fq 'GO-2026-5617' "$root/tools/security/audit-scanner-binaries.sh"
grep -Fq 'GO-2026-5668' "$root/tools/security/audit-scanner-binaries.sh"
grep -Fq 'GO-2026-5746' "$root/tools/security/audit-scanner-binaries.sh"
grep -Fq 'GO-2026-5932' "$root/tools/security/audit-scanner-binaries.sh"
grep -Fq 'jq -s -r' "$root/tools/security/audit-scanner-binaries.sh"
grep -Fq 'go version -m "$binary"' "$root/tools/security/audit-scanner-binaries.sh"
grep -Fq 'module=github.com/anchore/syft' "$root/tools/security/audit-scanner-binaries.sh"
grep -Fq 'module=github.com/anchore/grype' "$root/tools/security/audit-scanner-binaries.sh"
grep -Fq 'module=github.com/zricethezav/gitleaks/v8' "$root/tools/security/audit-scanner-binaries.sh"
if grep -Eq 'linked_findings|has\("function"\)|go list -deps' "$root/tools/security/audit-scanner-binaries.sh"; then
	echo 'scanner audit must not filter module-wide findings or recompute a mismatched source graph' >&2
	exit 1
fi
grep -Fq 'scanner-tools/bin/gitleaks' "$workflow"
grep -Fq 'scanner-tools/bin/syft' "$workflow"
grep -Fq 'scanner-tools/bin/grype' "$workflow"
chmod_steps=$(grep -Fc 'run: chmod +x "$RUNNER_TEMP/scanner-tools/bin/"*' "$workflow")
if [ "$chmod_steps" -ne 2 ]; then
	echo 'artifact consumers must restore executable scanner modes after download' >&2
	exit 1
fi
grep -Fq 'export DOCKER_HOST="unix://$RUNNER_TEMP/no-daemon.sock"' "$workflow"
grep -Fq 'docker-archive:$RUNNER_TEMP/nomyr-${{ matrix.plane }}.image.tar' "$workflow"
grep -Fq 'sbom:$RUNNER_TEMP/nomyr-${{ matrix.plane }}.cdx.json' "$workflow"
grep -Fq 'cosign-release: v3.1.4' "$workflow"
grep -Fq 'mise exec -- tools/bootstrap/check' "$workflow"
if grep -Eq 'github\.com/(gitleaks/gitleaks|anchore/(syft|grype))/releases/download' "$workflow"; then
	echo 'security scanners must be built from source, not downloaded as release binaries' >&2
	exit 1
fi
if grep -Eq -- '--ignore-vuln|--exclude|--only-fixed|fail-on high|fail-on critical' "$workflow"; then
	echo 'security scans must not suppress findings or weaken the medium threshold' >&2
	exit 1
fi

echo 'workflow actions are SHA-pinned and write-capable jobs are trust-gated'
sh "$root/tools/security/test-scanner-audit.sh"
