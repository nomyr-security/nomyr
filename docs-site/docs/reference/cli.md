---
title: CLI reference
---

Generated from `nomyr --help`; command presence does not imply later-phase product behavior.

```text
Nomyr non-human identity security platform
Usage:
  nomyr [command]
Available Commands:
  completion  Generate shell completion scripts
  context     Show the active context
  demo        Serve a local demo with synthetic data
  doctor      Check the local installation
  help        Help about any command
  login       Authenticate with a Nomyr server
  scaffold    Create non-functional starter files
  telemetry   Manage anonymous usage metrics
  version     Print build version
Flags:
  -h, --help            help for nomyr
      --no-input        disable interactive input
      --output string   output format: text, json, or yaml (default "text")
Use "nomyr [command] --help" for more information about a command.
```

See the [CLI behavior guide](cli-behavior.md) for supported output and scaffold details.

## CLI behavior notes

The `nomyr` CLI is a Phase 0 local scaffold. It does not authenticate, connect
to identity providers, or execute security actions. `login` explicitly returns
the structured not-implemented error.

## Output and input

Use the persistent `--output text|json|yaml` option before or after a command
to select output for implemented command results. Structured failures use the
same format and include `code`, `what`, `cause`, `fix`, and `docs`. The legacy
`version --json` and `doctor --json` options remain supported. `--no-input`
disables interactive input; current commands do not prompt.

`nomyr completion bash|zsh|fish|powershell` writes the shell completion script
to standard output. Scaffold commands create YAML starter files only:

```sh
nomyr scaffold connector new sample-connector --directory ./connectors
nomyr scaffold rule new sample-rule --directory ./rules
```

Connector scaffolds follow the existing capability-manifest contract and
declare no capabilities; rule scaffolds have `scaffold-only` status. Neither
implements provider support or rule execution. Existing files are never
overwritten, and names are restricted to lowercase letters, digits, and
hyphens.

## Exit status

- `0`: command completed successfully.
- `1`: command, argument, or runtime error; structured command errors are
  written to standard error.

The local demo remains loopback-only. Usage metrics retain their documented
default and opt-out behavior; set `NOMYR_CONFIG_HOME` to an isolated directory
with `telemetry.json` containing `{"enabled":false}` for local smoke tests.
The developer-only `go run ./tools/errlint` check exits `1` when it finds a
bare user-facing error constructor and `2` for an input or parse failure.

## `version`

```text
Print build version
Usage:
  nomyr version [flags]
Flags:
  -h, --help   help for version
      --json   emit JSON
Global Flags:
      --no-input        disable interactive input
      --output string   output format: text, json, or yaml (default "text")
```

## `doctor`

```text
Check the local installation
Usage:
  nomyr doctor [flags]
Flags:
  -h, --help   help for doctor
      --json   emit JSON
Global Flags:
      --no-input        disable interactive input
      --output string   output format: text, json, or yaml (default "text")
```

## `login`

```text
Authenticate with a Nomyr server
Usage:
  nomyr login [flags]
Flags:
  -h, --help   help for login
Global Flags:
      --no-input        disable interactive input
      --output string   output format: text, json, or yaml (default "text")
```

## `context`

```text
Show the active context
Usage:
  nomyr context [flags]
Flags:
  -h, --help   help for context
Global Flags:
      --no-input        disable interactive input
      --output string   output format: text, json, or yaml (default "text")
```

## `demo`

```text
Serve a local demo with synthetic data
Usage:
  nomyr demo [flags]
Flags:
  -h, --help            help for demo
      --listen string   loopback listen address (default "127.0.0.1:8080")
Global Flags:
      --no-input        disable interactive input
      --output string   output format: text, json, or yaml (default "text")
```

## `telemetry`

```text
Manage anonymous usage metrics
Usage:
  nomyr telemetry [command]
Available Commands:
  disable     Disable anonymous usage metrics
  enable      Enable anonymous usage metrics
  status      Show anonymous usage metrics status
Flags:
  -h, --help   help for telemetry
Global Flags:
      --no-input        disable interactive input
      --output string   output format: text, json, or yaml (default "text")
Use "nomyr telemetry [command] --help" for more information about a command.
```

## `scaffold`

```text
Create non-functional starter files
Usage:
  nomyr scaffold [command]
Available Commands:
  connector   Manage connector scaffolds
  rule        Manage rule scaffolds
Flags:
  -h, --help   help for scaffold
Global Flags:
      --no-input        disable interactive input
      --output string   output format: text, json, or yaml (default "text")
Use "nomyr scaffold [command] --help" for more information about a command.
```
