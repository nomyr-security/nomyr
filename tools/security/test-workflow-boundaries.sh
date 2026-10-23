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
	if grep -Eq 'github/codeql-action/(init|autobuild|analyze)@' "$file"; then
		echo "advanced CodeQL workflows conflict with the configured GitHub default setup: $file" >&2
		exit 1
	fi
done
if [ -e "$root/.github/workflows/codeql-schedule.yml" ]; then
	echo 'remove the duplicate advanced CodeQL schedule while GitHub default setup is configured' >&2
	exit 1
fi

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
grep -Fq 'tools/security/audit-cosign-binary.sh' "$workflow"
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
checksum_steps=$(grep -Fc 'sha256sum --check scanner-build.SHA256SUMS' "$workflow")
chmod_steps=$(grep -Fc 'chmod +x bin/*' "$workflow")
if [ "$checksum_steps" -ne 2 ] || [ "$chmod_steps" -ne 2 ]; then
	echo 'artifact consumers must verify scanner hashes before restoring executable modes' >&2
	exit 1
fi
grep -Fq 'export DOCKER_HOST="unix://$RUNNER_TEMP/no-daemon.sock"' "$workflow"
grep -Fq 'docker-archive:$RUNNER_TEMP/nomyr-${{ matrix.plane }}.image.tar' "$workflow"
grep -Fq 'sbom:$RUNNER_TEMP/nomyr-${{ matrix.plane }}.cdx.json' "$workflow"
grep -Fq 'cosign-release: v3.1.4' "$workflow"
grep -Fq 'COSIGN_VERSION=3.1.4' "$root/tools/security/scanners.env"
grep -Fq 'COSIGN_LINUX_AMD64_SHA256=424e2c97d0b45fd7a6b608a87f090f35633f4c23ece5f092d48bba11aaf4bbfa' "$root/tools/security/scanners.env"
grep -Fq "printf '%s  %s\\n' '424e2c97d0b45fd7a6b608a87f090f35633f4c23ece5f092d48bba11aaf4bbfa' \"\$(command -v cosign)\" | sha256sum --check -" "$workflow"
cosign_install_line=$(grep -nF 'uses: sigstore/cosign-installer@' "$workflow" | cut -d: -f1)
cosign_checksum_line=$(grep -nF "printf '%s  %s\\n' '424e2c97d0b45fd7a6b608a87f090f35633f4c23ece5f092d48bba11aaf4bbfa'" "$workflow" | cut -d: -f1)
cosign_sign_line=$(grep -nF 'cosign sign-blob --bundle' "$workflow" | head -n1 | cut -d: -f1)
if [ -z "$cosign_install_line" ] || [ -z "$cosign_checksum_line" ] || [ -z "$cosign_sign_line" ] || [ "$cosign_install_line" -ge "$cosign_checksum_line" ] || [ "$cosign_checksum_line" -ge "$cosign_sign_line" ]; then
	echo 'installed Cosign bytes must match the audited release digest before signing' >&2
	exit 1
fi
grep -Fq 'GO-2026-5932' "$root/tools/security/audit-cosign-binary.sh"
grep -Fq 'GO-2026-6225' "$root/tools/security/audit-cosign-binary.sh"
sign_blob_count=$(grep -Fc 'cosign sign-blob --bundle' "$workflow")
verify_blob_count=$(grep -Fc 'cosign verify-blob --bundle' "$workflow")
if [ "$sign_blob_count" -ne 3 ] || [ "$verify_blob_count" -ne 1 ]; then
	echo 'Cosign signing workflow must stay confined to the reviewed blob commands' >&2
	exit 1
fi
grep -Fq 'cosign sign-blob --bundle "artifacts/nomyr-${{ matrix.plane }}.image.sigstore.json" "artifacts/nomyr-${{ matrix.plane }}.image.tar"' "$workflow"
grep -Fq 'cosign sign-blob --bundle "artifacts/nomyr-${{ matrix.plane }}.cdx.sigstore.json" "artifacts/nomyr-${{ matrix.plane }}.cdx.json"' "$workflow"
grep -Fq 'cosign sign-blob --bundle "artifacts/nomyr-${{ matrix.plane }}.SHA256SUMS.sigstore.json" "artifacts/nomyr-${{ matrix.plane }}.SHA256SUMS"' "$workflow"
grep -Fq -- '--certificate-identity https://github.com/nomyr-security/nomyr/.github/workflows/security-supply-chain.yml@refs/heads/main' "$workflow"
if grep -Eq 'cosign (sign|verify)([[:space:]]|$)' "$workflow"; then
	echo 'Cosign workflow must not add registry/image commands without a new applicability review' >&2
	exit 1
fi
grep -Fq -- '--certificate-identity https://github.com/nomyr-security/nomyr/.github/workflows/security-supply-chain.yml@refs/heads/main' "$workflow"
if grep -Fq -- '--certificate-identity-regexp' "$workflow"; then
	echo 'keyless signature verification must require the exact trusted workflow identity' >&2
	exit 1
fi
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
sh "$root/tools/security/test-dev-image-policy.sh"
