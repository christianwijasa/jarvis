# Team graphs

If `jarvis.yaml` has `project: <name>`, prefer the **project graph**
(`jarvis project graph-path <name>` / `$JARVIS_HOME/projects/<name>/graphify-out/`)
which spans every member repo with repo-prefixed node ids. Otherwise use
this working tree's `graphify-out/`.

Treat `graphify query` results as hints, then verify in current source.

Before query, run `jarvis project graph <name> --if-stale` (may fetch and
extract). `graphify query` does not rebuild. `jarvis project status`
without `--fetch` uses local origin refs only. A full rebuild is
`jarvis project graph <name>`.

See [query.md](query.md) for the graphify query interface.
