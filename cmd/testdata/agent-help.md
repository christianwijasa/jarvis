# Jarvis agent guide

Stable compact help for AI agents (Cursor, Claude Code, Codex, Copilot, and others).
Humans can use `jarvis --help` and `jarvis <cmd> --help`.

Do not prompt the user through Jarvis. Never wait on stdin.
A human TTY may see a zsh-style update offer (`Would you like to update? [Y/n]`). That check is skipped without a TTY, with `--json`, for `help --agent` / `agent-help` / `update`, when the binary version is `dev`, and when `JARVIS_DISABLE_AUTO_UPDATE` is set.

## What Jarvis is

Jarvis is a small, tool-agnostic CLI for software-team planning.
It is not an editor plugin. It does not call a model or a tracker API.

- Project settings: `jarvis.yaml` (tracker, history repo, repo map, gates, skill choice).
- TRDs (design docs): a **separate history repo** named in `jarvis.yaml` (`history.path`), not this CLI repo.
- Bundled skill text (original/adapted) ships in the binary.
- Vendored skills (archify) are fetched at a **pinned commit** from `upstream.yaml`. Never pull `HEAD`/`main`.
- Rules live in the Jarvis repo (`rules/`) and stay in-tree.

## Recommended workflow

1. `jarvis help --agent` (this document) or `jarvis help --agent --json`
2. `jarvis init --yes` (or `--recommended` / `--skills` plus `--scope`) in the working repo
3. `jarvis doctor` — fail if not ok
4. `jarvis skills` — see what is installed
5. `jarvis skills show <name>` — load a skill body before following it
6. `jarvis plan <KEY>` — find or scaffold a TRD into `history.path` (`source: tracker` or `source: user`)
7. `jarvis project current` / `show --json` / `graph` — combined graph across sibling repos

## Where skills live

Canonical files: `$JARVIS_HOME/skills/<name>/SKILL.md`

- `JARVIS_HOME` env overrides all else.
- Else `skills.home` in `jarvis.yaml`.
- Else `~/.jarvis`.

Project-local (agents in a repo often look here): `skills.local_dir` (default `.agents/skills`).
`skills.scope` is `global`, `local`, or `both`.
Local copies are **symlinks** from `$JARVIS_HOME` unless `skills.copy` is true or `--copy` was passed.

Rules: shipped under `rules/` in the Jarvis CLI repo, not installed per project.

## Commands

| Command | When to use | Non-interactive notes |
| --- | --- | --- |
| `jarvis help --agent` | First thing an agent should run | No prompts |
| `jarvis agent-help` | Alias of `help --agent` | No prompts |
| `jarvis version` | Identify the binary and skill pack | `--json` |
| `jarvis doctor` | Check git, gh, graphify, jarvis.yaml, skill links | `--json`; exit 1 if any check fails |
| `jarvis init` | Create `jarvis.yaml` and install skills | Requires `--yes` or `--recommended` or `--skills` when stdin is not a TTY. Optional `--scope`, `--copy`, `--local-dir`, `--agents-md` |
| `jarvis skills` | List installed vs available | `--json` |
| `jarvis skills show <name>` | **Load a skill** (print SKILL.md) | `--json` |
| `jarvis skills install <name...>` | Install catalog entries at the pin | Needs `jarvis.yaml`; `--copy` |
| `jarvis skills remove <name>` | Uninstall | Needs `jarvis.yaml` |
| `jarvis skills upstream` | Read-only pin vs latest | `--json`; never applies updates |
| `jarvis plan <KEY>` | Find or scaffold TRD via `trd-pr-planning` | `--json`; `--from`; no model/tracker API. `source: tracker` uses host MCP; `source: user` uses markdown. Never read project secrets. |
| `jarvis update` | Latest **tag**, not main | needs `gh`; `--check` / `--json`; never follows `main` |
| `jarvis project list` / `show` / `current` / `graph-path` / `status` | Named repo groups and graph location | `--json` on these reads; status SHA freshness vs origin; `--fetch` optional |
| `jarvis project graph <name>` | Build combined graphify graph | needs `graphify` on PATH; `--if-stale` fetches origin and extracts only stale SHAs; empty extract falls back to in-repo `graphify-out` or skips |
| `jarvis project create` / `update` / `add-repo` / `clone` / `use` / `delete` | Mutate project yaml | create/update take optional `--tracker-*`; token in secrets not yaml; delete needs `--yes` off TTY |

## How to load a skill

```
jarvis skills show grilling
jarvis skills show grilling --json
```

Use the installed copy when present, else the bundled text. Vendored skills that are not installed error; run `jarvis skills install <name>`.

## Projects (combined graph)

A project is a named group of repositories (`shop` with `cart-service`, `payment-service`, …). Storage: `$JARVIS_HOME/projects/<name>/project.yaml`. The **point** is one graphify knowledge graph spanning every member repo so agents can trace A→B calls.

Find the active project and its graph:

```
jarvis project current --json
jarvis project show shop --json          # includes graph_json and repos[].path
jarvis project graph-path shop --json
```

When an active project has a graph, run `jarvis project graph shop --if-stale` then **query that graph first** (for example `graphify query "..." ` with the project's graph path/cwd) before reading files across repos. `--if-stale` fetches origin, compares `stamps.json` to `origin/main` or `origin/master`, and re-extracts only stale members. Full rebuild: `jarvis project graph shop`. It prints `extract i/n <repo>` and `merge n graphs` on stderr and streams graphify logs (no GUI). Output is `$JARVIS_HOME/projects/<name>/graphify-out/` (not inside member checkouts). Jarvis uses graphify v0.9.15 `extract --out` per repo then `merge-graphs` (node ids prefixed with the repo name). An empty or failed `--out` extract copies that repo's in-repo `graphify-out/graph.json` when present, otherwise skips the repo and continues.

Per-repo `graphify-out/` is only for a repo that is **not** in an active project.

## Non-interactive flags (never hit a prompt)

If stdin is not a TTY and Jarvis would ask a question, it **exits 1** and names the flag. It does not hang.

Init:

```
jarvis init --yes
jarvis init --recommended --scope both
jarvis init --skills grilling,graphify --scope local --copy
jarvis init --yes --agents-md
jarvis init --yes --project shop
```

`--yes` = recommended skill set (10 bundled + archify) and scope `both`, unless `--skills` / `--scope` override.
`--agents-md` appends a marked block to `AGENTS.md` telling agents to run `jarvis help --agent` first. It never overwrites a whole existing file; re-running refreshes the marked block only.

## Exit codes

| Code | Meaning |
| --- | --- |
| 0 | Success |
| 1 | Error (unknown skill, doctor failed, missing flags, missing jarvis.yaml, fetch failed) |

Agents never wait on stdin. Usage mistakes also exit 1.

## JSON (`--json`)

Every JSON document has `"schema_version": "jarvis.cli/v1"` and a `"kind"` string.

| kind | Command |
| --- | --- |
| `agent_help` | `jarvis help --agent --json` |
| `version` | `jarvis version --json` |
| `skills` | `jarvis skills --json` |
| `skill` | `jarvis skills show --json <name>` |
| `doctor` | `jarvis doctor --json` |
| `upstream` | `jarvis skills upstream --json` |
| `project_list` | `jarvis project list --json` |
| `project` | `jarvis project show --json <name>` |
| `project_current` | `jarvis project current --json` |
| `project_status` | `jarvis project status --json <name>` |
| `project_graph_path` | `jarvis project graph-path --json <name>` |
| `update` | `jarvis update --json` |
| `plan` | `jarvis plan --json <KEY>` |

Unknown `schema_version` means the agent should not guess the rest of the shape.

## Catalog (recommended set)

Bundled (in the binary): `trd-pr-planning`, `grilling`, `grill-with-docs`, `domain-modeling`, `test-case-design`, `development`, `debugging`, `code-review`, `generate-pr-description`, `graphify`.

Vendored (fetch at pin): `archify` (tt-a1i/archify v3.2.0).
