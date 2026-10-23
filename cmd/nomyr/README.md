# CLI behavior

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
