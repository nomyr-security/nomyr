# Project non-human identity inventory

This inventory records only identities and trust relationships supported by
repository evidence. It contains no credential values. A blank or `unknown`
field means it has not been verified; it does not mean that no identity exists.
Review this inventory when automation, applications, or cloud trust is added.

| Identity or trust | Evidence in this checkout | Owner | Credential or trust material | Status |
|---|---|---|---|---|
| GitHub Actions workloads | Repository workflows; repository default workflow token permission is read-only in live settings | Unknown; human owner roster not published | No cloud OIDC trust is declared in the reviewed public workflow configuration. Workflow-specific credentials are not inventoried here. | Partial; live org/app ownership and workflow-specific grants not verified |
| GitHub Apps and installed applications | No complete installation inventory is available from this checkout | Unknown | Not recorded | Unknown; organization-level inventory requires administrator review |
| Cloud OIDC trusts | No cloud trust configuration is present in the reviewed repository paths | Unknown | Not recorded | Unknown; account-side trusts have not been inspected |
| Release, package, or deployment bots | No complete public project bot roster is documented | Unknown | Not recorded | Unknown; verify organization and repository installations before release automation |

The GitHub organization and repository settings are externally administered.
Do not treat this partial repository inventory as a complete account-wide
identity audit. Update it only with a verified identity, purpose, owner, scope,
rotation or revocation path, and evidence date; never include secret values.
