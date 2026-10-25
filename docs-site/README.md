# Documentation site

This Docusaurus site is a static build with local search. It uses the repository's exact Node.js pin (`22.23.4`). Install the lockfile with `npm ci --ignore-scripts`; build with `npm run build`. To create a release snapshot, run `npm run version:cut -- 0.1.0` after reviewing the current docs.

`make docs-required` runs generated-reference drift tests, the site build, docs lint and links, a full lockfile audit at the `moderate` threshold, a telemetry-disabled CLI quickstart, and synthetic-demo Playwright screenshots. `make docs-capture` creates the VHS terminal capture with telemetry disabled. VHS and ttyd versions are pinned in `tools/docs/versions.env`; CI verifies the ttyd release checksum. Audit findings remain in the command output. The initial lockfile audit reports four low-severity KaTeX findings in the Markdown-lint tool dependency chain; no advisory is suppressed. Moderate, high, and critical findings fail the check.

The generator reads only the exact file allowlist in `tools/docgen/main.go`. Do not add globs or copy the private planning/specification directories, ignored ADR sources, OpenWiki, or telemetry deployment material into the site. Pull requests retain the static site build and synthetic screenshots as workflow artifacts. This repository change does not configure public Pages hosting or a live preview service.
