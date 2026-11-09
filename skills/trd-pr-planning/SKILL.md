---
name: trd-pr-planning
description: >-
  Create or find TRD plan documents in the history repo from atlas.yaml by
  Jira key before opening PRs. Every new plan must run grill-with-docs first
  (interview rounds, wait for user) before any docs/plans file. Use when the
  user mentions a Jira key, creates a PR, analyzes a task, or plans
  implementation or architecture work.
---

# TRD plan before PR

Plan documents live in the **history repo** (`atlas.yaml` `history.path`),
under `docs/plans/`. PRs for product code open in the target repo from
`atlas.yaml` `repos`. Folder name = **exact Jira issue key**.

This CLI repo is not a TRD store. `atlas plan <JIRA-KEY>` will later
scaffold from `templates/` into that history path.

TRD enhancement: [grill-with-docs](../grill-with-docs/SKILL.md) — based on
[grill-with-docs](https://www.aihero.dev/grill-with-docs).

## Resolve history repo

| Situation | Action |
| --------- | ------ |
| `atlas.yaml` found | Read/write plans under `history.path` |
| No config | Stop and run `atlas init`, or point at the history checkout |

Confirm target repo(s) via `atlas.yaml` `repos` (short name → local path).

## Find existing plan

1. Given Jira key `TEAM-123`, open `docs/plans/TEAM-123/` in the history repo.
2. Read YAML **frontmatter** on `TRD-Architecture.md` or primary `TRD-Implementation.md` (`tier`, `test_source`, `repos`).
3. Load that primary TRD body; pull `TRD-Testing.md`, per-repo impl, or `test-data-setup.md` only when tier or task requires them.
4. If the folder is missing, search `docs/plans/` for the key.
5. Use found docs for task analysis before writing code or opening PRs.

## Templates

Start from this CLI’s `templates/trd.md`, or from templates the history
repo keeps (architecture, backend/frontend implementation, testing). Copy
structure into the output path; do not invent a second template tree here.

## Output path

**Single-repo:**

```
docs/plans/<JIRA-KEY>/TRD-Implementation.md
docs/plans/<JIRA-KEY>/TRD-Architecture.md
docs/plans/<JIRA-KEY>/TRD-Testing.md
```

**Multi-repo (same ticket):**

```
docs/plans/<JIRA-KEY>/TRD-Architecture.md
docs/plans/<JIRA-KEY>/<repo-name>/TRD-Implementation.md
```

`<repo-name>` is a key from `atlas.yaml` `repos`.

## Frontmatter (required)

```yaml
---
jira: TEAM-123
tier: A | B | C
test_source: atlas:TRD-Testing | jira:TEAM-123 | user-prompt | waived
repos: [api]
type: implementation
status: draft
---
```

Story tiers: `atlas.yaml` `tiers` (default when unsure: `tiers.default`).

## Workflow

0. **Source baseline** — record local revision. Planning does not fetch,
   checkout, pull, rebase or extract implicitly. Graph navigation is
   advisory ([graph-first-analysis](../../rules/graph-first-analysis.md)).
1. Resolve Jira key; confirm target repo(s) via `atlas.yaml` `repos`.
2. Fetch the Jira issue from `atlas.yaml` `jira.site` (when that
   integration exists). This skeleton does not call the Jira API yet.
3. Check `docs/plans/<JIRA-KEY>/` — update existing doc if present, do not duplicate.
4. Confirm doc type: architecture, backend implementation, frontend implementation, or testing plan.
5. **Classify story tier** with the user. Default is `atlas.yaml` `tiers.default`. Record `tier` and `test_source`. **Tier A:** write `TRD-Testing.md` with [test-case-design](../test-case-design/SKILL.md). Never leave it empty. **Tier B:** skip `TRD-Testing.md` unless the user asks; Jira remains a valid test source. **Tier C:** minimal plan; user may waive full grill.
6. Classify work: **New plan** / **Update** / **Major rewrite** (treat like new for grilling).
7. **Grill with docs** (mandatory before steps 8+ for every new plan and major rewrite; abbreviated or skipped for Tier C when the user confirms) — follow [grill-with-docs](../grill-with-docs/SKILL.md) in full:
   - Load the team glossary and ADR tree (paths the history repo uses; often `docs/CONTEXT.md` and `docs/adr/`).
   - **Hard stop:** post grilling **round 1**, then wait. Do not open TRD templates or write under `docs/plans/` until grilling is done and the user confirms shared understanding.
8. Read the matching template; copy structure into the output path.
9. Fill document information from the history repo’s team file if it has one (squad, author, reviewers). **Related Story (Jira)** must match the folder key. Link issues using `jira.site` from `atlas.yaml`.
10. Complete required sections using grilled decisions. Link the glossary; link `docs/adr/` for persistent decisions. Skip sections that add no value. Do not duplicate Jira acceptance criteria into a sync file.
11. If implementation follows an architecture doc, link it. **TRD-Testing (Tier A):** write it with [test-case-design](../test-case-design/SKILL.md).
12. Create the PR in the **target product repo**; add the plan doc link to the PR body. When the round is final, add the team’s review label if they use one (do not invent a second process).
13. Update the plan doc **Related PR** field with the PR URL.

## TRD-Testing

`TRD-Testing.md` is the team’s own guide for checking the feature locally.
QA, if the team has a separate QA function, tests outside this flow and
does not approve the Atlas doc.

- **On TRD creation:** for a new **Tier A** TRD, write `TRD-Testing.md` with [test-case-design](../test-case-design/SKILL.md). Never leave it empty. **Tier B:** skip unless asked. **Tier C:** only when someone asks.
- **On review:** the reviewer asks for it on **Tier A** when it is missing. Request only; it never blocks the review.
- **Style:** the writer may read existing test-management cases (read-only) to match naming. Do not create cases in that system from this skill.

## PR body minimum

```markdown
## Summary

<1-3 bullets>

## Plan document

<link to history-repo TRD>

## Test plan

<checklist>
```

Do not run `gh pr create` until the plan doc exists and the user has reviewed it (or explicitly waived).
