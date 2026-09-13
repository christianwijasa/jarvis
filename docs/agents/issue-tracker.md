# Issue tracker (local scratch)

Work is tracked as **markdown files** in this repo, not GitHub Issues on the hub.

## Location

- Directory: [`.scratch/issues/`](../../.scratch/issues/)
- Filename pattern: `<id>-<slug>.md` (e.g. `001-auth-refactor.md`)
- `id`: zero-padded counter or date prefix you choose consistently
- `slug`: short kebab-case title

Issue **bodies** (`*.md`) are gitignored; the directory layout is committed (`.gitkeep`).

## File format

```yaml
---
status: open | in-progress | done | wontfix
created: YYYY-MM-DD
plan: docs/plans/auth-refactor/   # optional
repos: my-side-project            # optional manifest repo name(s)
---

# Title

## Problem

## Notes

## Links
```

Agents may create and update these files when the user asks to track work. Do not call `gh issue create` for hub tracking.

## Pairing with plans

When working on scratch issue `001-auth-refactor`:

1. Read the scratch file (create if missing and user requested tracking).
2. Check `docs/plans/auth-refactor/` for existing `design.md`.
3. Run `/grill-with-docs` before a large implementation or plan rewrite.
4. After grilling, update `docs/plans/<slug>/design.md` and set `plan:` in the scratch frontmatter.

## Branches and commits (personal repos)

- Branches: `feat/<slug>` or `fix/<slug>` unless the repo already has another convention.
- Commits: Conventional Commits — see [`.agents/rules/commit-message.mdc`](../../.agents/rules/commit-message.mdc).

## Wayfinding

No Jira, no TRD folders. Slug in scratch filename should match `docs/plans/<slug>/` when a plan exists.
