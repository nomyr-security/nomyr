# Phase 0 Security Harnesses

The isolated Go tooling in `tools/securityharness/` checks two Phase 0 security
boundaries without adding product handlers or authentication behavior.

Run the checks from the repository root:

```sh
go test ./tools/securityharness/...
go run ./tools/securityharness/cmd verify
```

The verifier reads `api/openapi/v1.yaml` without resolving external references
and sorts declared paths and methods. Its reviewed Phase 0 inventory contains
exactly `GET /healthz` (`getHealth`, implemented and public) and
`GET /capabilities` (`listCapabilities`, known unimplemented and authenticated
by contract). It checks the operation ID and effective OpenAPI security against
that inventory, including tenant-scope declarations. Any new operation, changed
ID, security or tenant scope, invalid scope declaration, or missing reviewed
operation is `not_verified`; `verify` fails with a fixed diagnostic and emits no
report until inventory review is updated.

Operations are classified as list, get, aggregate, or other. Tenant scope comes
from an explicit inventory entry for reviewed operations, or from tenant path
segments, tenant path parameters, or `x-nomyr-tenant-scoped: true` for unknown
operations. Unknown operations remain `not_verified` even if no tenant marker is
found; they are never counted as tenant-isolation passes or N/A. The current
known inventory explicitly reports health and capabilities isolation as
`not_applicable`, with different reasons for public health and an unimplemented
handler.

The current contract declares `GET /healthz` and `GET /capabilities`. Health is
public and implemented. Capabilities is declared but has no handler. Both have
tenant isolation reported as `not_applicable`; neither is reported as an
isolation pass. Configurable synthetic list, get, and aggregate operation
targets use a requester principal in `synthetic-requester` to request
`synthetic-foreign`. Their expected denial contract includes an exact status
and media type; RFC 9457 problem denials must include `type`, `title`, matching
`status`, and a contract-shaped `code`. Malformed problems, wrong statuses or
media types, and any denial body containing the foreign tenant identifier fail.
An unsafe 200 response also fails. These fixtures do not exercise product
authorization.

The canary is generated from cryptographic random bytes at run time. `scan`
accepts a canary file with mode `0600` and repeatable `--file`, `--artifact`,
`--log`, `--trace`, and `--database-dump` input paths. A real file-reader adapter
scans inputs in bounded chunks. Callers use `NewCanary`, write it only to a
private mode-`0600` file, and remove the temporary directory after the scan.
Findings contain only a source kind, fixed finding code, and count. Open/read
failures use fixed diagnostics and omit underlying error text. The CLI exits
nonzero on a leak, leaves stdout empty, and prints only a fixed sanitized error
to stderr.

CI tests the actual CLI process using a random canary file with mode `0600` and
temporary safe and leaking files for all five input categories. The test
removes the canary and fixtures at completion; the workflow uploads no scan
artifacts. No live database dump, production log, trace backend, build output,
or persisted artifact is scanned. Database coverage is explicitly reported as
`not_exercised`.

The GitHub Actions workflow is read-only, uses SHA-pinned checkout and Go setup
actions, and runs Go tests plus the synthetic verifier. It emits only sanitized
JSON summary data to the job log and does not upload artifacts or access product
credentials.
