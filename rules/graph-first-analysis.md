---
description: Scoped advisory graph navigation with live-source verification
---

# Scoped graph navigation

Architecture, unknown dependency/caller and cross-repository analysis
attempts a scoped graph query first (`graphify query`).

If `jarvis.yaml` has an active `project`, query **that project's combined
graph** (`jarvis project graph-path` / `$JARVIS_HOME/projects/<name>/graphify-out/`)
before reading files across repos. Node ids are prefixed with the repo
name. If there is no active project, use this repo's own `graphify-out/`.

Known file/symbol fixes may read current source directly.

Before a project-graph query, run `jarvis project graph <name> --if-stale`.
That command may `git fetch` and re-extract stale member repos. `graphify
query` itself never fetches or rebuilds. `jarvis project status` without
`--fetch` uses local origin refs only.

Valid results are hints with provenance; current source establishes facts.
Missing, corrupt, blocked or unhelpful graphs use bounded source
fallback without silently expanding scope. Runtime claims require
runtime evidence.

A full rebuild is `jarvis project graph <name>` (no `--if-stale`). Preserve
configured paths from the project yaml and from `jarvis.yaml` `repos`.

Planning or writing TRDs: if the team keeps a business-layer graph,
search process/entity nodes first to map each acceptance criterion to
the code nodes it names, then verify in source. Layer nodes are
inferred hints; treat freshness warnings as a reason to verify every
hint against HEAD. Guide: [team-graphs.md](../skills/graphify/references/team-graphs.md).
