# jarvis

Personal multi-repo hub (Origin: `fchr/jarvis`): manifest-driven clones under `~/projects/`, shared glossary (`CONTEXT.md`), and Matt Pocock grill skills (no TRD/Jira).

## Quick start

1. Open this folder in Cursor (multi-root with manifest repos optional).
2. Add repos to [`workstation/repos.manifest`](workstation/repos.manifest).
3. `make setup-status` then `make setup` to clone or pull siblings under `~/projects/<repo-name>`.
4. Before a big change: `/grill-with-docs` in this hub or in a code repo. Brainstorm only: `/grill-me` (anywhere).

## Environment

| Variable | Default | Purpose |
|----------|---------|---------|
| `PERSONAL_PROJECTS_ROOT` | `$HOME/projects` | Parent directory for manifest repos |
| `GITHUB_USER` | `YOU` | Default `git@github.com:$GITHUB_USER/<repo>.git` when manifest has no URL |
| `GIT_REMOTE_PREFIX` | (unset) | If set, `clone` uses `$GIT_REMOTE_PREFIX<repo>.git` |

## Scratch issues

Local tickets live under [`.scratch/issues/`](.scratch/issues/). Issue file bodies (`*.md`) are gitignored; see [`docs/agents/issue-tracker.md`](docs/agents/issue-tracker.md).

## Agent entry

Read [`AGENTS.md`](AGENTS.md) and [`docs/agents/`](docs/agents/).
