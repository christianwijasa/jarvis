# Team graphs

If `atlas.yaml` has `project: <name>`, prefer the **project graph**
(`atlas project graph-path <name>` / `$ATLAS_HOME/projects/<name>/graphify-out/`)
which spans every member repo with repo-prefixed node ids. Otherwise use
this working tree's `graphify-out/`.

Treat `graphify query` results as hints, then verify in current source.

Refresh or extract is a separate explicit action (`atlas project graph`
for a project graph). Query and status must not fetch, pull, or rebuild.

See [query.md](query.md) for the graphify query interface.
