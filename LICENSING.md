# Nomyr licensing

Nomyr's core is licensed under the **GNU Affero General Public License,
version 3 only (`AGPL-3.0-only`)**. The complete text is in [LICENSE](LICENSE).
This is the default license for original Nomyr files unless a more specific
license notice applies.

The following original Nomyr SDK and interface files are licensed under the
**Apache License, version 2.0 (`Apache-2.0`)**, as indicated by the license file
in each directory:

| Directory | Scope |
|---|---|
| [`sdk/`](sdk/LICENSE) | All Nomyr SDKs placed in this directory |
| [`api/`](api/LICENSE) | Public API contract definitions |
| [`proto/`](proto/LICENSE) | Protocol definitions and their supporting configuration |
| [`schemas/`](schemas/LICENSE) | JSON Schema definitions |

The `sdk/` directory currently contains only its license; it does not provide an
implemented SDK. New SDK packages must carry Apache-2.0 licensing when added.
Generated SDK distributions must preserve the applicable license and attribution
notices, including any notices required by their generators or dependencies.

The Apache-2.0 exceptions do not relicense the server, command-line application,
web application, internal implementation, collectors, or action implementations.
Those remain under AGPL-3.0-only. A directory license covers original Nomyr files
in that directory and its descendants unless a file carries a different notice.

Third-party software retains its own licenses and copyright notices. Nomyr's
license choices do not replace dependency licenses, including licenses for
bundled code, generated code, assets, or build tooling. Preserve the relevant
license and attribution notices when distributing those materials.

Canonical license texts were obtained from the [GNU project](https://www.gnu.org/licenses/agpl-3.0.txt)
and the [Apache Software Foundation](https://www.apache.org/licenses/LICENSE-2.0.txt).
Unmodified copies are also stored in [`LICENSES/`](LICENSES/).
