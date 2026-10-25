# Phase 0 Security Harnesses

The isolated Go tooling in `tools/securityharness/` checks two Phase 0 security
boundaries without adding product handlers or authentication behavior.

Run the checks from the repository root:

```sh
go test ./tools/securityharness/...
go run ./tools/securityharness/cmd verify
```

The verifier reads `api/openapi/v1.yaml` without resolving external references,
sorts declared paths and methods, and reports each operation's kind, whether a
handler is implemented in the current demo server, and whether tenant isolation
applies. Classification recognizes list operation IDs, item paths for get,
aggregate/summary/statistics names, tenant path segments or parameters, and the
`x-nomyr-tenant-scoped: true` operation extension. Implemented routes are
explicitly mapped from the current server. This mapping must be updated with
source changes; contract declaration alone never makes an operation implemented.

The current contract declares `GET /healthz` and `GET /capabilities`. Health is
public and implemented. Capabilities is declared but has no handler. Both have
tenant isolation reported as `not_applicable`; neither is reported as an
isolation pass. A future declared tenant-scoped operation without an implemented
handler is also not applicable. An implemented tenant-scoped operation is
reported as `not_tested` until an authorization assertion is connected to that
actual handler. The synthetic HTTP fixture uses a requester principal in
`synthetic-requester` to request `synthetic-foreign`; only a 401, 403, or 404
denial without the foreign tenant identifier in its body passes. Its unsafe
200 response makes the assertion fail. This does not exercise product
authorization.

The canary is generated from cryptographic random bytes at run time. Scanners
accept caller-owned readers tagged as files, artifacts, logs, traces, or database
dumps. Inputs are scanned in bounded chunks. Findings contain only a source
kind, fixed finding code, and count. Open/read failures use fixed diagnostics
and omit underlying error text. CI exercises synthetic readers for all five
input kinds and a deliberate leak.
No live database dump, production log, trace backend, build output, or persisted
artifact is scanned by this harness yet. Database coverage is explicitly
reported as `not_exercised`.

The GitHub Actions workflow is read-only, uses SHA-pinned checkout and Go setup
actions, and runs Go tests plus the synthetic verifier. It emits only sanitized
JSON summary data to the job log and does not upload artifacts or access product
credentials.
