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
the advisory-affected daemon packages are not linked. The individual
accountable human reviewer remains unassigned.

The workflow uploads the build directory as `nomyr-security-scanner-tools-<sha>`.
Consumers use `bin/gitleaks`, `bin/syft`, and `bin/grype`; the reports directory
contains govulncheck JSON, IDs, source package graphs, symbol listings, and
scanner-specific VEX records. This artifact
is a CI tool input, not a product release. It is produced with read-only
permissions and is not signed or published as a runtime artifact.
