# Security scanner toolchain

`scanners.env` pins the scanner source module versions, Go module checksums,
upstream VCS commits, patched Go runtime, and the govulncheck version. Run
`mise exec -- sh tools/security/build-scanners.sh OUTPUT_DIRECTORY` to build the
three tools into `OUTPUT_DIRECTORY/bin/`; the directory also contains source
provenance and binary SHA-256 checksums. Builds verify upstream module checksums
and commit identities before compiling. Syft and Grype use upstream version
metadata and platform-compatible build tags. Syft intentionally omits the
optional `containers_image_openpgp` tag: this CI scanner consumes only local
`docker-archive:` and `sbom:` inputs, not the Linux containers-storage provider
that requires that tag. This avoids linking the legacy OpenPGP parser without
changing supported scan modes. Grype retains the upstream release tags.
Gitleaks is built from its pinned
module source with the pinned transitive dependency updates in
`patches/gitleaks-transitive-security-updates.patch` applied.

Run `mise exec -- sh tools/security/audit-scanner-binaries.sh OUTPUT_DIRECTORY`
to retain raw govulncheck JSON, the complete set of finding IDs (including
module-wide and trace-less findings), and the exact source dependency graph
captured during each build with the matching CGO mode and build tags. No new
finding is discarded based on trace shape. Each scanner must match its
explicitly reviewed ID set, and the excluded source packages must be absent
from both the recorded graph and the unstripped binary symbol table. The
resulting OpenVEX and text evidence bind dispositions to exact source commits,
module checksums, and binary hashes; any changed ID, source pin, graph, or
binary namespace fails for review. The affected Docker daemon IDs reported by
govulncheck for Grype remain in the raw report and are considered not
applicable only while the exact Grype build graph and symbol checks prove that
the advisory-affected daemon packages are not linked. `@nomyr-ai` is the
owner-designated maintainer/decision-owner account as of 2026-10-23; the human
or legal identity behind that account is not independently verified here.

The workflow uploads the build directory as `nomyr-security-scanner-tools-<sha>`.
Consumers use `bin/gitleaks`, `bin/syft`, and `bin/grype`; the reports directory
contains govulncheck JSON, IDs, source package graphs, symbol listings, and
scanner-specific VEX records. This artifact
is a CI tool input, not a product release. It is produced with read-only
permissions and is not signed or published as a runtime artifact.

## Cosign blob-command boundary

The workflow pins Cosign v3.1.4 and audits the official Linux/amd64 release
asset before use, without executing it. Its SHA-256 is pinned to the GitHub
release asset digest (`424e2c97d0b45fd7a6b608a87f090f35633f4c23ece5f092d48bba11aaf4bbfa`),
and the report checks Go build metadata and retains the raw govulncheck JSON,
all finding IDs, and affected symbol evidence. The verified binary reports
Go 1.27.3. The audit retains GO-2026-5932 and GO-2026-6225; it does not claim
the binary is vulnerability-free.

The GO-2026-5932 report identifies only `golang.org/x/crypto/openpgp` packages
as affected, with no known fixed version. The exact binary's symbol listing has
no `golang.org/x/crypto/openpgp` symbols. GO-2026-6225 is different: the ACR
credential helper and `ACRCredHelper.Get` are linked into the binary, so their
absence is not asserted. The upstream v3.1.4 source calls
`NewACRCredentialsHelper` only from `RegistryOptions.GetRegistryClientOpts`
when the Kubernetes keychain is selected
([registry.go](https://github.com/sigstore/cosign/blob/v3.1.4/cmd/cosign/cli/options/registry.go)).
This workflow invokes only `cosign sign-blob` on local archive/SBOM/checksum
files and `cosign verify-blob` on the local image archive. The corresponding
upstream command implementations call blob signing and verification paths
without registry options
([signblob.go](https://github.com/sigstore/cosign/blob/v3.1.4/cmd/cosign/cli/signblob.go),
[sign_blob.go](https://github.com/sigstore/cosign/blob/v3.1.4/cmd/cosign/cli/sign/sign_blob.go),
[verify.go](https://github.com/sigstore/cosign/blob/v3.1.4/cmd/cosign/cli/verify.go),
[verify_blob.go](https://github.com/sigstore/cosign/blob/v3.1.4/cmd/cosign/cli/verify/verify_blob.go)).
The not-applicable disposition is limited to these exact commands and reviewed
v3.1.4 source; it is not a general exemption for Cosign or registry use. The
boundary test fails if the workflow adds registry/image `sign` or `verify`
commands or changes the pinned Cosign release, requiring a fresh review.
