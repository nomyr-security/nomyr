# Phase 0 architecture spikes

These isolated experiments are evidence, not product implementations. They are outside the root
Go module and are not imported by any Nomyr binary. No provider credentials, production accounts,
or billable resources are used.

## Commands

From this directory:

```sh
go test ./temporal ./opa
NOMYR_SPIKE_POSTGRES_URL="$DISPOSABLE_POSTGRES_URL" go test ./pgx
go test ./nats
govulncheck -test ./...
```

The PostgreSQL test requires an explicit URL for a disposable local database and skips if it is
absent. The architecture-spikes workflow starts digest-pinned PostgreSQL bound to loopback only.
The NATS test starts the pinned server library in-process on `127.0.0.1` with an ephemeral port
and temporary JetStream store. The tests never connect to a cloud provider or create product data.
The PostgreSQL test creates and drops a uniquely named table. Use `govulncheck -test ./...` to
include SDK and server dependencies referenced only by tests; the default scan excludes test code.

## Evidence boundaries

- Temporal uses the official Go SDK test environment. It exercises retry ordering and confirms a
  runtime-generated canary is rejected if serialized into the workflow payload. It does not prove
  server persistence, worker failover, or production replay compatibility.
- OPA uses the embedded Go bundle reader, signing, and verification APIs. An ephemeral RSA private
  key signs each test bundle; verification receives only its public key. Acceptance, tampering, and
  a different untrusted public key are exercised. No policy is installed in the Nomyr runtime.
- `pgx` tests run actual SQL against PostgreSQL when the explicit local URL is provided. They check
  valid-time and record-time interval boundaries. No database run is claimed when the URL is absent.
- NATS uses the official server library in-process with loopback binding, ephemeral port, and
  temporary storage; the official client publishes a synthetic event, receives the JetStream
  acknowledgement, and reads it back. This does not demonstrate production persistence or HA.
- Connector enumeration is blocked by the current public protobuf contract, which defines only
  `Describe`. Provider-specific list operations, fixture shapes, sanitization, and coverage cannot
  be represented without changing that contract. This ticket does not invent a public API.
- The web tree already has one static identity route. The build can demonstrate export and Go
  embedding, but it does not provide arbitrary identity routes, retrieval, or authorization.
- Embedded PostgreSQL is not validated here. The repository has no selected maintained embedded
  PostgreSQL distribution with the requested six-platform binary support and provenance. No
  database binary is downloaded or executed. Cross-compilation is not platform runtime evidence.

Exact observed versions, command results, measurements, and remaining gaps are recorded in
[`../../docs/spikes/phase-0-architecture.md`](../../docs/spikes/phase-0-architecture.md).
