---
title: Implementation status
---

This page describes the checked-in foundation behavior. Requirements and contracts can describe intended interfaces without being wired into a running service.

| Evidence level | Meaning |
| --- | --- |
| Design intent | Described by project specifications or plans; not evidence of executable behavior. |
| Scaffold behavior | Present in source, with no claim of complete product capability. |
| Tested behavior | Covered by a named automated test or executed docs check. |
| Certified behavior | Supported by an explicit certification or release record. No product capability is certified by this scaffold. |

The friendly `nomyr` command provides version, doctor, context, local demo, usage-metrics controls, and non-functional connector/rule scaffolds. The demo uses synthetic data, has no provider connection, and rejects non-loopback listen addresses. Login and product service processes remain not implemented.

The repository contains API, protobuf, and JSON Schema contracts. A declaration in a contract does not mean the corresponding endpoint, provider integration, or event pipeline exists. Requirement records currently describe status metadata and do not independently prove functionality.

The current unversioned docs build is the source for a future first version snapshot. Pull requests build the site and retain it as an Actions artifact for review; no public preview host or Pages deployment is configured by this change.

No provider, inventory, authentication/RLS, product custody, action execution, or certification capability is included in this Phase 0 site.
