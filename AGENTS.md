# Agent instructions — jarvis

Personal multi-repo hub: manifest, shared glossary, local scratch issues, Matt Pocock grill skills. Not BFI atlas — no TRD, no Jira, no agency commit format.

Local path: **`~/projects/jarvis`**. Origin: **`fchr/jarvis`** (`https://origin.cursor.com/fchr/jarvis.git`).

## Invoke: `jarvis …`

User messages prefixed with **`jarvis`** (any case) are commands for this hub—not other “Jarvis” products.

Grammar and verbs: [`.agents/rules/jarvis-invoke.mdc`](.agents/rules/jarvis-invoke.mdc). Examples: `jarvis analyze …`, `jarvis plan <slug>`, `jarvis track …`, `jarvis setup`.

## Workstation

- Manifest: [`workstation/repos.manifest`](workstation/repos.manifest)
- Scripts: [`scripts/setup-workstation.sh`](scripts/setup-workstation.sh), [`scripts/lib/repo-paths.sh`](scripts/lib/repo-paths.sh)
- `make setup` / `make setup-status`

Repos clone to `$PERSONAL_PROJECTS_ROOT/<name>` (default `~/projects/<name>`).

## Domain docs

[`docs/agents/domain.md`](docs/agents/domain.md) — `CONTEXT.md`, `docs/adr/`, `docs/plans/<slug>/`.

## Issue tracker

[`docs/agents/issue-tracker.md`](docs/agents/issue-tracker.md) — local scratch under `.scratch/issues/`.

## Grill (Matt Pocock)

| Invoke | Load | Writes |
|--------|------|--------|
| `/grill-me` | **grilling** (`~/.agents/skills/`) | nothing |
| `/grill-with-docs` | **grilling** + **domain-modeling**; in this hub follow [`.agents/skills/grill-with-docs/SKILL.md`](.agents/skills/grill-with-docs/SKILL.md) | `CONTEXT.md`, `docs/adr/`, optional `docs/plans/<slug>/design.md` |

Start `/grill-me` in a fresh conversation for a loose idea. Use `/grill-with-docs` when vocabulary and decisions should land in the repo.

If grill-with-docs dumps all questions at once with no recommended answers, stop and confirm **grilling** and **domain-modeling** both loaded.

## Git commits

[`.agents/rules/commit-message.mdc`](.agents/rules/commit-message.mdc) — Conventional Commits, no ticket brackets.

## Explicit exclusions

- No TRD templates or `trd-pr-planning`
- No Jira / BAE workflow
- No atlas commit-quality tooling
