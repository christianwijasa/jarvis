# Jarvis

Jarvis is a small CLI that gives a software team planning guidance. It is
meant to replace an older mix of rules and stored design documents sitting
in one repo.

This repository is the **executable**. It is not the place Technical
Requirements Documents (TRDs) live.

## What lives where

| Piece | Where |
| --- | --- |
| CLI, bundled skills, rules, TRD templates | this repo, released by version tag |
| Project settings (Jira site, history path, repo map, gates, skill choice) | `jarvis.yaml` in the working tree you run Jarvis from |
| Installed skill files | `$JARVIS_HOME/skills` (default `~/.jarvis/skills`) |
| Named projects + combined graphs | `$JARVIS_HOME/projects/<name>/` |
| Project-local skill links (optional) | `.agents/skills` (or `skills.local_dir`) |
| TRDs and other design history | a **separate history repo** named in `jarvis.yaml` |

The core stays **tool-agnostic**. Jarvis does not assume a particular editor,
agent, or chat product. Original and adapted skill text is shipped inside the
binary. Vendored upstream trees (archify) are **not** stored in this repo;
they are fetched at the pinned commit when you install them.

Releases are **git tags** (for example `v0.1.0`). Pushing a `v*` tag runs
GitHub Actions, which attaches `jarvis-<os>-<arch>` binaries and
`SHA256SUMS`. `jarvis update` installs that latest tag (never `main`).
Needs `gh` on PATH. First install can still be `make build` and a copy
onto PATH; after a tagged release exists, `jarvis update` replaces the
running binary. Network is also used for `jarvis skills upstream`
(read-only) and for installing a vendored skill at its pin.

## Skills

Every skill is **optional** and chosen at setup. `upstream.yaml` is the
catalog: name, upstream repo, path, pinned commit/tag, licence, mode, and a
short description.

`jarvis init` default **Recommended** set (11 skills):

| Skill | Role | In this repo |
| --- | --- | --- |
| `trd-pr-planning` | Scaffold / find a TRD in the history repo. Used by `jarvis plan`. | original |
| `grilling` | Interview rounds until shared understanding. | adapted |
| `grill-with-docs` | Grill before a new TRD file. | adapted |
| `domain-modeling` | Glossary and ADRs in the history repo. | adapted |
| `test-case-design` | Write `TRD-Testing.md`. | original |
| `development` | Karpathy + Ponytail implementation loop. | adapted |
| `debugging` | Reproduce, trace, then fix. | adapted |
| `code-review` | Standards + Spec review of a diff or PR. | adapted |
| `generate-pr-description` | Short PR body from the diff. | original |
| `graphify` | Query a code graph instead of dumping trees into context. | adapted |
| `archify` | Architecture diagrams. Fetched from tt-a1i/archify **v3.2.0** (`208acba`). Apache-2.0. | vendored (not in git) |

Rules live in `rules/` (plain English, markdown format, commit/branch
names, Karpathy, Ponytail, test-case and test-data gates, graph-first
analysis). They stay in the repo for now.

## Install flow

Canonical files always go to **`$JARVIS_HOME/skills/<name>`**. Override the
home with the `JARVIS_HOME` environment variable or `skills.home` in
`jarvis.yaml`.

`jarvis init` asks two things (TTY):

1. **Recommended** (default) vs **Choose** (multi-select from the catalog).
2. Where agents should load skills: **global only** (`~/.jarvis/skills`),
   **project-local** (default `.agents/skills`), or **both** (default).

Local placement **symlinks** each skill from `$JARVIS_HOME` so one install
serves many projects. If the OS cannot symlink, Jarvis copies. Pass `--copy`
to copy on purpose.

Non-interactive flags:

```bash
jarvis init --recommended --yes --scope both
jarvis init --skills grilling,graphify --scope local --copy
jarvis init --yes   # recommended skills, scope both
```

`--yes` skips prompts. Without a TTY you must pass `--recommended`,
`--skills`, or `--yes`.

Install later:

```bash
jarvis skills                 # installed vs available, with paths
jarvis skills install archify
jarvis skills install grilling graphify
jarvis skills remove archify
jarvis skills show grilling
```

Install of a `vendored` catalog row fetches **only** the pinned commit
(`git fetch --depth 1 origin <pin>`, or `gh api …/tarball/<pin>`). It never
follows `HEAD`/`main`. Changing the pin is a reviewed edit to
`upstream.yaml`.

`jarvis doctor` checks `git`, `gh`, `graphify`, `jarvis.yaml`, `JARVIS_HOME`,
and that local skill links still resolve.

## Upstream updates

Every skill and rule is listed in `upstream.yaml` with a mode:

- `vendored` — fetched as-is at the pin (not stored in this git tree)
- `adapted` — rewritten from an upstream file (text lives in `skills/`)
- `original` — Jarvis-written, no upstream

Pinned commits come from the old hub’s `UPSTREAM.md`,
`SUPERPOWERS-NOTICE.md`, `skills-lock.json`, or a verified tag. Missing
pins are `unknown`; we do not guess.

`jarvis skills upstream` is **read-only**. It reports pinned vs latest
and a compare URL. It never copies new text. Taking or ignoring an
upstream change is a manual, reviewed pull request, per entry.

Licence notices are in `LICENSES/`.

## Help for other AI agents

Agents on any machine (Cursor, Claude Code, Codex, Copilot, …) should run
this first — it never prompts:

```bash
jarvis help --agent
jarvis help --agent --json   # schema_version jarvis.cli/v1
jarvis agent-help            # same markdown
```

`--json` is also on `jarvis version`, `jarvis skills`, `jarvis skills show`,
`jarvis skills upstream`, and `jarvis doctor`. Every JSON object includes
`schema_version` (`jarvis.cli/v1`) and `kind`.

If stdin is not a TTY, Jarvis exits 1 and names the flag instead of hanging
(`jarvis init` needs `--yes`, `--recommended`, or `--skills`).

`jarvis init --agents-md` appends a marked block to `AGENTS.md` pointing at
`jarvis help --agent`. It will not overwrite an existing file; re-running
only refreshes the marked block.

## Projects

A **project** is a named group of repositories (example: `shop` with
`cart-service` and `payment-service`). The point is **one graphify graph
that spans every member repo**, so agents can trace cross-repo calls.

Files:

- `$JARVIS_HOME/projects/<name>/project.yaml`
- Combined graph: `$JARVIS_HOME/projects/<name>/graphify-out/` (`graph.json`, `GRAPH_REPORT.md`). Optional `graph.out_dir` in the yaml.

```bash
jarvis project create shop --description "Example shop" \
  --repo cart-service=https://github.com/example/cart-service.git \
  --repo payment-service=https://github.com/example/payment-service.git@~/src/payment-service \
  --tracker-kind jira --tracker-site https://example.atlassian.net
jarvis project update shop --tracker-mcp plugin-atlassian-atlassian
jarvis project use shop
jarvis project clone shop --dir ~/src --depth 1
jarvis project graph shop
jarvis project graph-path shop --json
jarvis project status shop --json
```

`jarvis project graph` shells out to graphify **v0.9.15**:
`extract <checkout> --out $JARVIS_HOME/projects/<name>/extract/<repo> --code-only`
then `merge-graphs … --out …/graphify-out/graph.json`. That keeps
`graphify-out/` **out of member checkouts**. `merge-graphs` prefixes node
ids with the repo name (`parent.parent` of each extract `graph.json`).
Need `graphify` on PATH (`jarvis doctor` checks it). Progress
(`extract i/n <repo>`, `merge n graphs`) and live graphify logs go to
stderr; there is no GUI. Success is one stdout line with the combined
graph path.

When `jarvis.yaml` has `project: shop`, agents should `graphify query`
against the **project** graph before reading files across repos.

## Commands (proposal)

The command list is still under discussion. Each command is a thin stub
so it is cheap to rename, split, or drop.

| Command | Today |
| --- | --- |
| `jarvis help` / `jarvis <cmd> --help` | Human help (purpose, usage, flags, examples). |
| `jarvis help --agent` / `jarvis agent-help` | Compact markdown guide for AI agents. |
| `jarvis version` | CLI version, skill-pack version, config path. `--json`. |
| `jarvis update` | Latest GitHub Release tag (not `main`). `--check`, `--json`. |
| `jarvis plan <KEY>` | Find or scaffold a TRD into `history.path`. Tracker-agnostic key. `--json`, `--from`, `--repo`, `--tier`. No model or tracker API. |
| `jarvis skills` | Installed vs available, with global/local paths. `--json`. |
| `jarvis skills show <name>` | Prints an installed or bundled skill (how agents load a skill). `--json`. |
| `jarvis skills install <name...>` | Fetch-at-pin or copy bundled text into `$JARVIS_HOME`. |
| `jarvis skills remove <name>` | Deletes the global install and local link. |
| `jarvis skills upstream` | Read-only pin vs latest for each registry row. `--json`. |
| `jarvis doctor` | `git`, `gh`, **`graphify`**, `jarvis.yaml`, skill links, active project (missing paths warn). `--json`. |
| `jarvis init` | Writes `jarvis.yaml` and installs the chosen skills. `--agents-md`, `--project`. |
| `jarvis project …` | Named repo groups, clone, status, combined graphify graph. |

## `jarvis.yaml`

Copy `jarvis.yaml.example` or run `jarvis init`, then edit:

- `tracker.kind` / `tracker.site` / `tracker.project` / `tracker.mcp` — optional work-item tracker policy (no token in this file)
- Named project tracker (and token in `$JARVIS_HOME/projects/<name>/secrets`) — `jarvis project create` / `update`
- `history.path` / `history.github` — checkout (and GitHub slug) of the team history repo
- `repos` — short names for local checkouts
- `tiers` — default story tier and placeholder rules
- `gates` — test-case / test-data flags and coverage percents
- `skills.scope` — `global`, `local`, or `both`
- `skills.local_dir` — project path agents should use (default `.agents/skills`)
- `skills.copy` — `true` to copy instead of symlink
- `skills.installed` — names chosen at init or via `jarvis skills install`
- `skills.home` — optional override for `JARVIS_HOME`
- `project` — active named project (`jarvis project use`)

Jarvis looks for `jarvis.yaml` in the current directory, then parent
directories.

## Build

Needs Go 1.22 or newer.

```bash
make build
# or
go build -ldflags "-X github.com/christianwijasa/jarvis/internal/buildinfo.Version=v0.0.0" -o jarvis .
```

Run tests:

```bash
make test
# or
go test ./...
```

The CLI version string is `dev` unless you pass `-ldflags` (or
`make build VERSION=v0.1.0`). A `dev` binary is never treated as
already up to date: `jarvis update` still fetches the latest tag.

GitHub Actions on this repo runs `go build ./...` and `go test ./...`
on pull requests. Pushing a `v*` tag publishes GitHub Release assets.

```bash
jarvis update --check   # current vs latest; no download
jarvis update           # replace this binary
```

On a TTY, other commands may offer that same install (the zsh
`Would you like to update? [Y/n]` pattern). Jarvis checks GitHub at
most once every `JARVIS_UPDATE_DAYS` days (default 13), then asks.
Yes (the default) replaces the binary and re-runs the original
command. `n` skips until the next interval.

The offer never appears without a TTY, with `--json`, for
`jarvis help --agent` / `jarvis update`, when the binary is `dev`,
in CI, or when `JARVIS_DISABLE_AUTO_UPDATE=1`.
`JARVIS_DISABLE_UPDATE_PROMPT=1` installs without asking.

## Layout

```
cmd/                 cobra commands (kept thin)
internal/config/     jarvis.yaml load/find/write, JARVIS_HOME
internal/doctor/     path checks (git, gh, graphify, skill links)
internal/install/    fetch-at-pin, global install, local symlink/copy
internal/upstream/   registry + read-only check
internal/buildinfo/  CLI version set at build time
internal/update/     jarvis update from GitHub Releases
skills/              bundled original/adapted skills (embedded)
rules/               core rules
templates/           TRD template stub
upstream.yaml        provenance catalog
LICENSES/            upstream licence texts
```
