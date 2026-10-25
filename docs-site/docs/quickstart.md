---
title: Local quickstart
---

The supported local demo uses synthetic in-memory data and binds to loopback. It does not authenticate, connect to providers, or execute security actions. Do not expose it to an untrusted network.

With Go 1.26.10 and Node.js 22.23.4 available, run from the repository root:

```sh
make web-install
make build
NOMYR_CONFIG_HOME="$(mktemp -d)" ./bin/nomyr demo --listen 127.0.0.1:8080
```

Open `http://127.0.0.1:8080`. The demo is intentionally loopback-only. To exercise CLI output without starting a server:

```sh
NOMYR_CONFIG_HOME="$(mktemp -d)" ./bin/nomyr --output json doctor
```

The docs-as-tests workflow runs the CLI smoke checks with telemetry disabled and tests the synthetic demo using Playwright.

Run `make docs-capture` to record the version and doctor commands with VHS. The capture also disables telemetry and uses no provider credentials or real identity data.
