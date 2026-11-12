# Atlas

Atlas is a small CLI that gives a software team planning guidance. It is
meant to replace an older mix of rules and stored design documents sitting
in one repo.

This repository is the **executable**. It is not the place Technical
Requirements Documents (TRDs) live.

## What lives where

| Piece | Where |
| --- | --- |
| CLI, bundled skills, rules, TRD templates | this repo, released by version tag |
| Project settings (Jira site, history path, repo map, gates, skill choice) | `atlas.yaml` in the working tree you run Atlas from |
| Installed skill files | `$ATLAS_HOME/skills` (default `~/.atlas/skills`) |
| Named projects + combined graphs | `$ATLAS_HOME/projects/<name>/` |
| Project-local skill links (optional) | `.agents/skills` (or `skills.local_dir`) |
| TRDs and other design history | a **separate history repo** named in `atlas.yaml` |

The core stays **tool-agnostic**. Atlas does not assume a particular editor,
agent, or chat product. Original and adapted skill text is shipped inside the
binary. Vendored upstream trees (archify) are **not** stored in this repo;
they are fetched at the pinned commit when you install them.

Releases are **git tags** (for example `v0.1.0`). Pushing a `v*` tag runs
GitHub Actions, which attaches `atlas-<os>-<arch>` binaries and
`SHA256SUMS`. `atlas update` installs that latest tag (never `main`).
Needs `gh` on PATH. First install can still be `make build` and a copy
onto PATH; after a tagged release exists, `atlas update` replaces the
running binary. Network is also used for `atlas skills upstream`
(read-only) and for installing a vendored skill at its pin.

## Skills

Every skill is **optional** and chosen at setup. `upstream.yaml` is the
catalog: name, upstream repo, path, pinned commit/tag, licence, mode, and a
short description.

`atlas init` default **Recommended** set (11 skills):

| Skill | Role | In this repo |
| --- | --- | --- |
| `trd-pr-planning` | Scaffold / find a TRD in the history repo. Used by `atlas plan`. | original |
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

Canonical files always go to **`$ATLAS_HOME/skills/<name>`**. Override the
home with the `ATLAS_HOME` environment variable or `skills.home` in
`atlas.yaml`.

`atlas init` asks two things (TTY):

1. **Recommended** (default) vs **Choose** (multi-select from the catalog).
2. Where agents should load skills: **global only** (`~/.atlas/skills`),
   **project-local** (default `.agents/skills`), or **both** (default).

Local placement **symlinks** each skill from `$ATLAS_HOME` so one install
serves many projects. If the OS cannot symlink, Atlas copies. Pass `--copy`
to copy on purpose.

Non-interactive flags:

```bash
atlas init --recommended --yes --scope both
atlas init --skills grilling,graphify --scope local --copy
atlas init --yes   # recommended skills, scope both
```

`--yes` skips prompts. Without a TTY you must pass `--recommended`,
`--skills`, or `--yes`.

Install later:

```bash
atlas skills                 # installed vs available, with paths
atlas skills install archify
atlas skills install grilling graphify
atlas skills remove archify
atlas skills show grilling
```

Install of a `vendored` catalog row fetches **only** the pinned commit
(`git fetch --depth 1 origin <pin>`, or `gh api …/tarball/<pin>`). It never
follows `HEAD`/`main`. Changing the pin is a reviewed edit to
`upstream.yaml`.

`atlas doctor` checks `git`, `gh`, `graphify`, `atlas.yaml`, `ATLAS_HOME`,
and that local skill links still resolve.

## Upstream updates

Every skill and rule is listed in `upstream.yaml` with a mode:

- `vendored` — fetched as-is at the pin (not stored in this git tree)
- `adapted` — rewritten from an upstream file (text lives in `skills/`)
- `original` — Atlas-written, no upstream

Pinned commits come from the old hub’s `UPSTREAM.md`,
`SUPERPOWERS-NOTICE.md`, `skills-lock.json`, or a verified tag. Missing
pins are `unknown`; we do not guess.

`atlas skills upstream` is **read-only**. It reports pinned vs latest
and a compare URL. It never copies new text. Taking or ignoring an
upstream change is a manual, reviewed pull request, per entry.

Licence notices are in `LICENSES/`.

## Help for other AI agents

Agents on any machine (Cursor, Claude Code, Codex, Copilot, …) should run
this first — it never prompts:

```bash
atlas help --agent
atlas help --agent --json   # schema_version atlas.cli/v1
atlas agent-help            # same markdown
```

`--json` is also on `atlas version`, `atlas skills`, `atlas skills show`,
`atlas skills upstream`, and `atlas doctor`. Every JSON object includes
`schema_version` (`atlas.cli/v1`) and `kind`.

If stdin is not a TTY, Atlas exits 1 and names the flag instead of hanging
(`atlas init` needs `--yes`, `--recommended`, or `--skills`).

`atlas init --agents-md` appends a marked block to `AGENTS.md` pointing at
`atlas help --agent`. It will not overwrite an existing file; re-running
only refreshes the marked block.

## Projects

A **project** is a named group of repositories (example: `shop` with
`cart-service` and `payment-service`). The point is **one graphify graph
that spans every member repo**, so agents can trace cross-repo calls.

Files:

- `$ATLAS_HOME/projects/<name>/project.yaml`
- Combined graph: `$ATLAS_HOME/projects/<name>/graphify-out/` (`graph.json`, `GRAPH_REPORT.md`). Optional `graph.out_dir` in the yaml.

```bash
atlas project create shop --description "Example shop" \
  --repo cart-service=https://github.com/example/cart-service.git \
  --repo payment-service=https://github.com/example/payment-service.git@~/src/payment-service
atlas project use shop
atlas project clone shop --dir ~/src --depth 1
atlas project graph shop
atlas project graph-path shop --json
atlas project status shop --json
```

`atlas project graph` shells out to graphify **v0.9.15**:
`extract <checkout> --out $ATLAS_HOME/projects/<name>/extract/<repo> --code-only`
then `merge-graphs … --out …/graphify-out/graph.json`. That keeps
`graphify-out/` **out of member checkouts**. `merge-graphs` prefixes node
ids with the repo name (`parent.parent` of each extract `graph.json`).
Need `graphify` on PATH (`atlas doctor` checks it). Progress
(`extract i/n <repo>`, `merge n graphs`) and live graphify logs go to
stderr; there is no GUI. Success is one stdout line with the combined
graph path.

When `atlas.yaml` has `project: shop`, agents should `graphify query`
against the **project** graph before reading files across repos.

## Commands (proposal)

The command list is still under discussion. Each command is a thin stub
so it is cheap to rename, split, or drop.

| Command | Today |
| --- | --- |
| `atlas help` / `atlas <cmd> --help` | Human help (purpose, usage, flags, examples). |
| `atlas help --agent` / `atlas agent-help` | Compact markdown guide for AI agents. |
| `atlas version` | CLI version, skill-pack version, config path. `--json`. |
| `atlas update` | Latest GitHub Release tag (not `main`). `--check`, `--json`. |
| `atlas plan <JIRA-KEY>` | Would scaffold a TRD using `trd-pr-planning` into `history.path`. No model or Jira API. |
| `atlas skills` | Installed vs available, with global/local paths. `--json`. |
| `atlas skills show <name>` | Prints an installed or bundled skill (how agents load a skill). `--json`. |
| `atlas skills install <name...>` | Fetch-at-pin or copy bundled text into `$ATLAS_HOME`. |
| `atlas skills remove <name>` | Deletes the global install and local link. |
| `atlas skills upstream` | Read-only pin vs latest for each registry row. `--json`. |
| `atlas doctor` | `git`, `gh`, **`graphify`**, `atlas.yaml`, skill links, active project (missing paths warn). `--json`. |
| `atlas init` | Writes `atlas.yaml` and installs the chosen skills. `--agents-md`, `--project`. |
| `atlas project …` | Named repo groups, clone, status, combined graphify graph. |

## `atlas.yaml`

Copy `atlas.yaml.example` or run `atlas init`, then edit:

- `jira.site` / `jira.project` — where issues come from
- `history.path` / `history.github` — checkout (and GitHub slug) of the team history repo
- `repos` — short names for local checkouts
- `tiers` — default story tier and placeholder rules
- `gates` — test-case / test-data flags and coverage percents
- `skills.scope` — `global`, `local`, or `both`
- `skills.local_dir` — project path agents should use (default `.agents/skills`)
- `skills.copy` — `true` to copy instead of symlink
- `skills.installed` — names chosen at init or via `atlas skills install`
- `skills.home` — optional override for `ATLAS_HOME`
- `project` — active named project (`atlas project use`)

Atlas looks for `atlas.yaml` in the current directory, then parent
directories.

## Build

Needs Go 1.22 or newer.

```bash
make build
# or
go build -ldflags "-X github.com/christianwijasabfi/atlas/internal/buildinfo.Version=v0.0.0" -o atlas .
```

Run tests:

```bash
make test
# or
go test ./...
```

The CLI version string is `dev` unless you pass `-ldflags` (or
`make build VERSION=v0.1.0`). A `dev` binary is never treated as
already up to date: `atlas update` still fetches the latest tag.

GitHub Actions on this repo runs `go build ./...` and `go test ./...`
on pull requests. Pushing a `v*` tag publishes GitHub Release assets.

```bash
atlas update --check   # current vs latest; no download
atlas update           # replace this binary
```

On a TTY, other commands may offer that same install (the zsh
`Would you like to update? [Y/n]` pattern). Atlas checks GitHub at
most once every `ATLAS_UPDATE_DAYS` days (default 13), then asks.
Yes (the default) replaces the binary and re-runs the original
command. `n` skips until the next interval.

The offer never appears without a TTY, with `--json`, for
`atlas help --agent` / `atlas update`, when the binary is `dev`,
in CI, or when `ATLAS_DISABLE_AUTO_UPDATE=1`.
`ATLAS_DISABLE_UPDATE_PROMPT=1` installs without asking.

## Layout

```
cmd/                 cobra commands (kept thin)
internal/config/     atlas.yaml load/find/write, ATLAS_HOME
internal/doctor/     path checks (git, gh, graphify, skill links)
internal/install/    fetch-at-pin, global install, local symlink/copy
internal/upstream/   registry + read-only check
internal/buildinfo/  CLI version set at build time
internal/update/     atlas update from GitHub Releases
skills/              bundled original/adapted skills (embedded)
rules/               core rules
templates/           TRD template stub
upstream.yaml        provenance catalog
LICENSES/            upstream licence texts
```
