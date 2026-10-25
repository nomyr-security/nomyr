# Phase 0 architecture spikes

These isolated experiments are evidence, not product implementations. They are outside the root
Go module and are not imported by any Nomyr binary. No provider credentials, production accounts,
or billable resources are used.

## Commands

From this directory:

```sh
go test ./temporal ./opa
NOMYR_SPIKE_POSTGRES_URL="$DISPOSABLE_POSTGRES_URL" go test ./pgx
NOMYR_SPIKE_NATS_URL='nats://127.0.0.1:4222' go test ./nats
```

The PostgreSQL and NATS tests require the loopback-only services. They skip when their explicit URLs
are absent. The architecture-spikes workflow starts digest-pinned PostgreSQL and NATS test services;
the NATS container enables JetStream explicitly. The tests never connect to a cloud provider or
create product data. Use only a disposable local test database; the PostgreSQL test creates and
drops a uniquely named table.

## Evidence boundaries

- Temporal uses the official Go SDK test environment. It exercises retry ordering and confirms a
  runtime-generated canary is rejected if serialized into the workflow payload. It does not prove
  server persistence, worker failover, or production replay compatibility.
- OPA uses the embedded Go bundle reader, signing, and verification APIs. The signing key is
  generated in memory for each test; acceptance, tampering, and an untrusted verification key are
  exercised. No policy is installed in the Nomyr runtime.
- `pgx` tests run actual SQL against PostgreSQL when the explicit local URL is provided. They check
  valid-time and record-time interval boundaries. No database run is claimed when the URL is absent.
- NATS tests use the official client against the existing JetStream development service. No local
  result is claimed when that service is unavailable.
- Connector enumeration is blocked by the current public protobuf contract, which defines only
  `Describe`. Provider-specific list operations, fixture shapes, sanitization, and coverage cannot
  be represented without changing that contract. This ticket does not invent a public API.
- The web tree already has one static identity route. The build can demonstrate export and Go
  embedding, but it does not provide arbitrary identity routes, retrieval, or authorization.
- Embedded PostgreSQL is not validated here. The repository has no selected maintained embedded
  PostgreSQL distribution with the requested six-platform binary support and provenance. No
  database binary is downloaded or executed. Cross-compilation is not platform runtime evidence.
- The existing NATS service is a container service, not an embedded NATS server. The client test
  validates the local service boundary only.

Exact observed versions, command results, measurements, and remaining gaps are recorded in
[`../../docs/spikes/phase-0-architecture.md`](../../docs/spikes/phase-0-architecture.md).
