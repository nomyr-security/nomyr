# Contract tooling

The files in `api/`, `proto/`, and `schemas/` remain declarative Phase 0
contracts. Their generated outputs provide Go server interfaces and TypeScript
types; they do not register handlers or make declared operations available.
Update and review the contract before any future handler or client behavior.

Run `tools/contracts/generate.sh` after an intentional contract change. The
generator versions are pinned in `versions.env`; CI regenerates both outputs
and fails if the checked-in results drift. Go generation is strict-server
only, while the TypeScript output is produced from the same OpenAPI document.

The contract workflow runs Buf lint and file-level breaking checks, OpenAPI
lint and breaking checks against `main`, and JSON Schema draft 2020-12 checks.
JSON Schema fixtures include both accepted examples and intentionally malformed
documents. Fixtures are synthetic and contain no customer data or reusable
credentials.

CI uses Go 1.27.3 only for contract tooling because Vacuum v0.35.0 requires
that version. The product toolchain in `mise.toml` remains Go 1.26.10. The
contract workflow checks generated outputs with its pinned toolchain.
