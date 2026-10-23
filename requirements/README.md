# Requirement tracker

`catalog.json` contains requirement identifiers and status metadata. It does not copy specification descriptions or private planning material. Expected category counts and identifier sets are checked during validation.

Run `go run ./tools/requirements` to validate the catalog and print its deterministic Markdown status summary. In a checkout containing the local traceability matrix, validation also checks that catalog IDs still match it. To create or refresh ID/source records from that matrix, run `go run ./tools/requirements -seed`; existing statuses and evidence are preserved for matching IDs. The dedicated GitHub Actions workflow publishes the summary in its run summary.

## Statuses and evidence

- `specified`: recorded in the traceability matrix; no implementation claim.
- `designed`: a reviewed design exists.
- `implemented`: repository behavior exists and has implementation evidence.
- `tested`: focused verification evidence exists for the requirement.
- `certified`: the applicable certification/release evidence exists.
- `unavailable`: the capability is explicitly unavailable, unsupported, or not applicable, with evidence and rationale recorded by the linked evidence source.

Normal transitions are `specified → designed → implemented → tested → certified`. A requirement may move from any pre-certification state to `unavailable`. `certified` and `unavailable` are terminal in this initial tracker. Every transition requires a valid repository `path#Lline` evidence link; non-`specified` records must retain at least one evidence link. The validator checks that both the source file and referenced non-empty line exist. A status is not inferred from plans, names, or implementation proximity.

Use `go run ./tools/requirements -transition ID=STATUS -evidence path#Lline` to record one validated transition. Evidence should identify the implementation, test, or reviewed decision that supports the new state. Do not place secret values, customer data, or private infrastructure details in records or evidence.

## Scope and limits

The catalog currently records 132 functional requirements, 48 detections, 32 acceptance tests, 14 workflows, and 28 connector-family rows, all as `specified`. This is a traceability index, not proof that these capabilities are implemented. Connector records represent the matrix's grouped connector families, not individual certified provider adapters. Detailed requirement descriptions remain in specification material and are not copied here.

## Source register

The source field on each record links to the matching entry here. The IDs were checked against these local specification sections when seeded; this public tracker intentionally contains no copied requirement text.

- Functional requirements (`FR-*`): product specification §5.
- Detections (`D*`): product specification §8.2.
- Acceptance tests (`AT*`): product specification §14.2.
- Workflows (`6.1`–`6.14`): product specification §6.
- Connector-family IDs use the exact row labels from product specification §7.2.
