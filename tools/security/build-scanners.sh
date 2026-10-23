#!/bin/sh
# SPDX-FileCopyrightText: 2026 Nomyr technologies
# SPDX-License-Identifier: AGPL-3.0-only
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
. "$root/tools/security/scanners.env"

if [ "$(go version | sed -E 's/^go version go([^ ]+).*/\1/')" != "$GO_VERSION" ]; then
	echo "scanner builds require Go $GO_VERSION" >&2
	exit 1
fi
if ! govulncheck -version | grep -Fq "govulncheck@v$GOVULNCHECK_VERSION"; then
	echo "scanner audits require govulncheck v$GOVULNCHECK_VERSION" >&2
	exit 1
fi

out=${1:?usage: build-scanners.sh OUTPUT_DIRECTORY}
mkdir -p "$out/bin" "$out/reports"
out=$(CDPATH= cd -- "$out" && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM

verify_module() {
	module=$1
	version=$2
	expected_sum=$3
	expected_commit=$4
	metadata=$(go mod download -json "$module@$version")
	actual_sum=$(printf '%s\n' "$metadata" | jq -r '.Sum')
	actual_commit=$(printf '%s\n' "$metadata" | jq -r '.Origin.Hash')
	if [ "$actual_sum" != "$expected_sum" ] || [ "$actual_commit" != "$expected_commit" ]; then
		echo "source pin mismatch for $module@$version (sum=$actual_sum commit=$actual_commit)" >&2
		exit 1
	fi
	printf '%s\n' "$metadata" | jq -r '.Dir'
}

syft_dir=$(verify_module github.com/anchore/syft "v$SYFT_VERSION" "$SYFT_SUM" "$SYFT_COMMIT")
grype_dir=$(verify_module github.com/anchore/grype "v$GRYPE_VERSION" "$GRYPE_SUM" "$GRYPE_COMMIT")
gitleaks_dir=$(verify_module github.com/zricethezav/gitleaks/v8 "v$GITLEAKS_VERSION" "$GITLEAKS_SUM" "$GITLEAKS_COMMIT")
verify_module github.com/ulikunitz/xz "v$XZ_VERSION" "$XZ_SUM" "$XZ_COMMIT" >/dev/null
verify_module github.com/klauspost/compress "v$COMPRESS_VERSION" "$COMPRESS_SUM" "$COMPRESS_COMMIT" >/dev/null
verify_module github.com/nwaples/rardecode/v2 "v$RARDECODE_VERSION" "$RARDECODE_SUM" "$RARDECODE_COMMIT" >/dev/null
verify_module golang.org/x/text "v$XTEXT_VERSION" "$XTEXT_SUM" "$XTEXT_COMMIT" >/dev/null
verify_module github.com/docker/docker "$DOCKER_MODULE_VERSION" "$DOCKER_MODULE_SUM" "$DOCKER_MODULE_COMMIT" >/dev/null
verify_module golang.org/x/crypto "v$XCRYPTO_VERSION" "$XCRYPTO_SUM" "$XCRYPTO_COMMIT" >/dev/null

(
	cd "$syft_dir/cmd/syft"
	CGO_ENABLED=0 go build -tags 'exclude_graphdriver_btrfs' \
		-trimpath -buildvcs=false \
		-ldflags "-w -X main.version=$SYFT_VERSION -X main.gitCommit=$SYFT_COMMIT -X main.buildDate=source-build -X main.gitDescription=verified-source" \
		-o "$out/bin/syft" .
	CGO_ENABLED=0 go list -tags 'exclude_graphdriver_btrfs' -deps -f '{{.ImportPath}}' . > "$out/reports/syft.packages.txt"
)
(
	cd "$grype_dir/cmd/grype"
	CGO_ENABLED=0 go build -tags 'containers_image_openpgp exclude_graphdriver_btrfs' \
		-trimpath -buildvcs=false \
		-ldflags "-w -X main.version=$GRYPE_VERSION -X main.gitCommit=$GRYPE_COMMIT -X main.buildDate=source-build -X main.gitDescription=verified-source" \
		-o "$out/bin/grype" .
	CGO_ENABLED=0 go list -tags 'containers_image_openpgp exclude_graphdriver_btrfs' -deps -f '{{.ImportPath}}' . > "$out/reports/grype.packages.txt"
)

cp -R "$gitleaks_dir" "$tmp/gitleaks"
chmod -R u+w "$tmp/gitleaks"
(
	cd "$tmp/gitleaks"
	patch --batch --forward -p1 < "$root/tools/security/patches/gitleaks-transitive-security-updates.patch"
	go mod download "github.com/ulikunitz/xz@v$XZ_VERSION" "github.com/klauspost/compress@v$COMPRESS_VERSION" \
		"github.com/nwaples/rardecode/v2@v$RARDECODE_VERSION" "golang.org/x/crypto@v$XCRYPTO_VERSION" \
		"golang.org/x/text@v$XTEXT_VERSION"
	CGO_ENABLED=0 go build -mod=mod -trimpath -buildvcs=false \
		-ldflags "-X=github.com/zricethezav/gitleaks/v8/version.Version=v$GITLEAKS_VERSION" \
		-o "$out/bin/gitleaks" .
	CGO_ENABLED=0 go list -mod=mod -deps -f '{{.ImportPath}}' . > "$out/reports/gitleaks.packages.txt"
)

cat > "$out/scanner-build-provenance.txt" <<EOF
go=$GO_VERSION
goos=$(go env GOOS)
goarch=$(go env GOARCH)
syft_build_tags=exclude_graphdriver_btrfs (no containers-storage provider)
grype_build_tags=containers_image_openpgp,exclude_graphdriver_btrfs
govulncheck=$GOVULNCHECK_VERSION
syft=v$SYFT_VERSION commit=$SYFT_COMMIT sum=$SYFT_SUM
grype=v$GRYPE_VERSION commit=$GRYPE_COMMIT sum=$GRYPE_SUM
gitleaks=v$GITLEAKS_VERSION commit=$GITLEAKS_COMMIT sum=$GITLEAKS_SUM
gitleaks-patch=tools/security/patches/gitleaks-transitive-security-updates.patch
xz=v$XZ_VERSION commit=$XZ_COMMIT sum=$XZ_SUM
compress=v$COMPRESS_VERSION commit=$COMPRESS_COMMIT sum=$COMPRESS_SUM
rardecode=v$RARDECODE_VERSION commit=$RARDECODE_COMMIT sum=$RARDECODE_SUM
docker-module=$DOCKER_MODULE_VERSION commit=$DOCKER_MODULE_COMMIT sum=$DOCKER_MODULE_SUM
xcrypto=v$XCRYPTO_VERSION commit=$XCRYPTO_COMMIT sum=$XCRYPTO_SUM
xtext=v$XTEXT_VERSION commit=$XTEXT_COMMIT sum=$XTEXT_SUM
EOF
(cd "$out" && sha256sum bin/gitleaks bin/grype bin/syft > scanner-build.SHA256SUMS)
