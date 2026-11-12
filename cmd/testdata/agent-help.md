# Atlas agent guide

Stable compact help for AI agents (Cursor, Claude Code, Codex, Copilot, and others).
Humans can use `atlas --help` and `atlas <cmd> --help`.

Do not prompt the user through Atlas. Never wait on stdin.
A human TTY may see a zsh-style update offer (`Would you like to update? [Y/n]`). That check is skipped without a TTY, with `--json`, for `help --agent` / `agent-help` / `update`, when the binary version is `dev`, and when `ATLAS_DISABLE_AUTO_UPDATE` is set.

## What Atlas is

Atlas is a small, tool-agnostic CLI for software-team planning.
It is not an editor plugin. It does not call a model or a tracker API.

- Project settings: `atlas.yaml` (tracker, history repo, repo map, gates, skill choice).
- TRDs (design docs): a **separate history repo** named in `atlas.yaml` (`history.path`), not this CLI repo.
- Bundled skill text (original/adapted) ships in the binary.
- Vendored skills (archify) are fetched at a **pinned commit** from `upstream.yaml`. Never pull `HEAD`/`main`.
- Rules live in the Atlas repo (`rules/`) and stay in-tree.

## Recommended workflow

1. `atlas help --agent` (this document) or `atlas help --agent --json`
2. `atlas init --yes` (or `--recommended` / `--skills` plus `--scope`) in the working repo
3. `atlas doctor` — fail if not ok
4. `atlas skills` — see what is installed
5. `atlas skills show <name>` — load a skill body before following it
6. `atlas plan <KEY>` — find or scaffold a TRD into `history.path` (`source: tracker` or `source: user`)
7. `atlas project current` / `show --json` / `graph` — combined graph across sibling repos

## Where skills live

Canonical files: `$ATLAS_HOME/skills/<name>/SKILL.md`

- `ATLAS_HOME` env overrides all else.
- Else `skills.home` in `atlas.yaml`.
- Else `~/.atlas`.

Project-local (agents in a repo often look here): `skills.local_dir` (default `.agents/skills`).
`skills.scope` is `global`, `local`, or `both`.
Local copies are **symlinks** from `$ATLAS_HOME` unless `skills.copy` is true or `--copy` was passed.

Rules: shipped under `rules/` in the Atlas CLI repo, not installed per project.

## Commands

| Command | When to use | Non-interactive notes |
| --- | --- | --- |
| `atlas help --agent` | First thing an agent should run | No prompts |
| `atlas agent-help` | Alias of `help --agent` | No prompts |
| `atlas version` | Identify the binary and skill pack | `--json` |
| `atlas doctor` | Check git, gh, graphify, atlas.yaml, skill links | `--json`; exit 1 if any check fails |
| `atlas init` | Create `atlas.yaml` and install skills | Requires `--yes` or `--recommended` or `--skills` when stdin is not a TTY. Optional `--scope`, `--copy`, `--local-dir`, `--agents-md` |
| `atlas skills` | List installed vs available | `--json` |
| `atlas skills show <name>` | **Load a skill** (print SKILL.md) | `--json` |
| `atlas skills install <name...>` | Install catalog entries at the pin | Needs `atlas.yaml`; `--copy` |
| `atlas skills remove <name>` | Uninstall | Needs `atlas.yaml` |
| `atlas skills upstream` | Read-only pin vs latest | `--json`; never applies updates |
| `atlas plan <KEY>` | Find or scaffold TRD via `trd-pr-planning` | `--json`; `--from`; no model/tracker API. `source: tracker` uses host MCP; `source: user` uses markdown. Never read project secrets. |
| `atlas update` | Latest **tag**, not main | needs `gh`; `--check` / `--json`; never follows `main` |
| `atlas project list` / `show` / `current` / `graph-path` / `status` | Named repo groups and graph location | `--json` on these reads |
| `atlas project graph <name>` | Build combined graphify graph | needs `graphify` on PATH; empty extract falls back to in-repo `graphify-out` or skips |
| `atlas project create` / `update` / `add-repo` / `clone` / `use` / `delete` | Mutate project yaml | create/update take optional `--tracker-*`; token in secrets not yaml; delete needs `--yes` off TTY |

## How to load a skill

```
atlas skills show grilling
atlas skills show grilling --json
```

Use the installed copy when present, else the bundled text. Vendored skills that are not installed error; run `atlas skills install <name>`.

## Projects (combined graph)

A project is a named group of repositories (`shop` with `cart-service`, `payment-service`, …). Storage: `$ATLAS_HOME/projects/<name>/project.yaml`. The **point** is one graphify knowledge graph spanning every member repo so agents can trace A→B calls.

Find the active project and its graph:

```
atlas project current --json
atlas project show shop --json          # includes graph_json and repos[].path
atlas project graph-path shop --json
```

When an active project has a graph, **query that graph first** (for example `graphify query "..." ` with the project's graph path/cwd) before reading files across repos. Build/refresh with `atlas project graph shop`. It prints `extract i/n <repo>` and `merge n graphs` on stderr and streams graphify logs (no GUI). Output is `$ATLAS_HOME/projects/<name>/graphify-out/` (not inside member checkouts). Atlas uses graphify v0.9.15 `extract --out` per repo then `merge-graphs` (node ids prefixed with the repo name). An empty or failed `--out` extract copies that repo's in-repo `graphify-out/graph.json` when present, otherwise skips the repo and continues.

Per-repo `graphify-out/` is only for a repo that is **not** in an active project.

## Non-interactive flags (never hit a prompt)

If stdin is not a TTY and Atlas would ask a question, it **exits 1** and names the flag. It does not hang.

Init:

```
atlas init --yes
atlas init --recommended --scope both
atlas init --skills grilling,graphify --scope local --copy
atlas init --yes --agents-md
atlas init --yes --project shop
```

`--yes` = recommended skill set (10 bundled + archify) and scope `both`, unless `--skills` / `--scope` override.
`--agents-md` appends a marked block to `AGENTS.md` telling agents to run `atlas help --agent` first. It never overwrites a whole existing file; re-running refreshes the marked block only.

## Exit codes

| Code | Meaning |
| --- | --- |
| 0 | Success |
| 1 | Error (unknown skill, doctor failed, missing flags, missing atlas.yaml, fetch failed) |

Agents never wait on stdin. Usage mistakes also exit 1.

## JSON (`--json`)

Every JSON document has `"schema_version": "atlas.cli/v1"` and a `"kind"` string.

| kind | Command |
| --- | --- |
| `agent_help` | `atlas help --agent --json` |
| `version` | `atlas version --json` |
| `skills` | `atlas skills --json` |
| `skill` | `atlas skills show --json <name>` |
| `doctor` | `atlas doctor --json` |
| `upstream` | `atlas skills upstream --json` |
| `project_list` | `atlas project list --json` |
| `project` | `atlas project show --json <name>` |
| `project_current` | `atlas project current --json` |
| `project_status` | `atlas project status --json <name>` |
| `project_graph_path` | `atlas project graph-path --json <name>` |
| `update` | `atlas update --json` |
| `plan` | `atlas plan --json <KEY>` |

Unknown `schema_version` means the agent should not guess the rest of the shape.

## Catalog (recommended set)

Bundled (in the binary): `trd-pr-planning`, `grilling`, `grill-with-docs`, `domain-modeling`, `test-case-design`, `development`, `debugging`, `code-review`, `generate-pr-description`, `graphify`.

Vendored (fetch at pin): `archify` (tt-a1i/archify v3.2.0).
