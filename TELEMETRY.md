# Anonymous usage metrics

Nomyr usage metrics are disabled by default. You can explicitly enable them with:

```console
nomyr telemetry enable
```

When enabled, Nomyr sends an `installation_created` event once and a
`command_completed` event after each CLI invocation. The event contains a random anonymous
installation ID, Nomyr version, operating system, CPU architecture, command name, success status,
duration, and timestamp. The payload follows
[`schemas/usage-metrics-event.schema.json`](schemas/usage-metrics-event.schema.json).

Nomyr does not send command arguments beyond the command and subcommand name, environment
variables, file paths, repository names, usernames, hostnames, identity-security data,
credentials, or customer data. The collector does not retain the request IP address or user agent.
Metrics delivery never changes a Nomyr command's result and is bounded by an 800 ms timeout.

The collector accepts at most two events from one installation ID in each request, rejects requests
larger than 16 KiB, validates every field, and limits traffic to 10 requests per minute per
installation ID and 30 requests per minute per network source inside the Worker. Cloudflare also
blocks the ingestion path at the zone edge after 10 requests from one source in 10 seconds. A
higher-priority skip rule exempts the same path on every host except `telemetry.nomyr.io`, so the
landing site and other `nomyr.io` subdomains are outside this rate limit. A source address is used
only as a transient rate-limit key and is not written to the metrics database.
Because the endpoint is public and anonymous, these controls reduce automated abuse but cannot verify
that every event represents a human user. Treat the resulting counts as directional opt-in usage
metrics.

The collector is served only at `https://telemetry.nomyr.io`. Its account-specific `workers.dev`
production and preview URLs are disabled.

Use `nomyr telemetry status` to inspect the setting and `nomyr telemetry disable` to stop sending
events. Disabling also deletes the local anonymous installation ID. The configuration is stored in
the operating system's standard per-user configuration directory under `nomyr/telemetry.json`.
