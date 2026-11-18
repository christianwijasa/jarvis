---
name: trd-planning
description: >-
  Create or find TRD plan documents in the history repo from jarvis.yaml by
  work-item key before opening PRs. Every new plan must run grill-with-docs
  first (interview rounds, wait for user) before filling docs/plans files.
  Run plan-review (fresh-context review, Tier A ready-to-build gate) before
  gh pr create. Use when the user mentions a ticket key, creates a PR,
  analyzes a task, or plans implementation or architecture work.
---

# TRD planning (before PR)

Plan documents live in the **history repo** (`jarvis.yaml` `history.path`),
under `docs/plans/`. PRs for product code open in the target repo from
`jarvis.yaml` `repos`. Folder name = **exact work-item key** (Jira, Trello, or
any path-safe id).

This CLI repo is not a TRD store. `jarvis plan <KEY>` scaffolds from
`templates/` into that history path. Do not fill those files until grilling
is done. Never read `$JARVIS_HOME/projects/<name>/secrets`.

TRD enhancement: [grill-with-docs](../grill-with-docs/SKILL.md) — based on
[grill-with-docs](https://www.aihero.dev/grill-with-docs).

## Resolve history repo

| Situation | Action |
| --------- | ------ |
| `jarvis.yaml` found | Read/write plans under `history.path` |
| No config | Stop and run `jarvis init`, or point at the history checkout |

Confirm target repo(s) via `jarvis.yaml` `repos` (short name → local path).

## Find existing plan

0. If `docs/plans/<KEY>/PRD.md` exists (or `jarvis plan --json` returns
   `prd_file`), read it first. See [Starting from a PRD](#starting-from-a-prd).
1. Given key `TEAM-123`, open `docs/plans/TEAM-123/` in the history repo.
2. Read YAML **frontmatter** on `TRD-Architecture.md` or primary `TRD-Implementation.md` (`tier`, `test_source`, `repos`).
3. Load that primary TRD body; pull `TRD-Testing.md`, per-repo impl, or `test-data-setup.md` only when tier or task requires them.
4. If the folder is missing, search `docs/plans/` for the key.
5. Use found docs for task analysis before writing code or opening PRs.

## Starting from a PRD

A PRD is the product owner's document for the same key. It sits next to the
TRDs in `docs/plans/<KEY>/PRD.md`.

- If the PRD has `status: signed-off`, its goals, scope, business rules,
  acceptance criteria, edge cases, rollout, and decisions are **settled**. Do
  not ask about them again. Ask only about gaps, conflicts between the PRD and
  the code, and technical design.
- If the PRD is still `draft` or `in-review`, tell the user and ask whether to
  wait for sign-off or to plan against the draft.
- **Answer every open Engineering check.** Answer each row from code or the
  project graph. Record the answer in the TRD with the commit you read. Then
  write the same answer back into the PRD "Engineering checks" table: set
  `status` to `answered`, fill `answer` and `answered from (commit)`, and add a
  link to the TRD section. The product owner reviews these answers in the PRD.
  Change only those rows. Do not edit other PRD sections.
- If the code conflicts with a signed-off PRD rule, raise it as a grilling
  question. Do not change the rule silently in the TRD.
- TRD acceptance and test cases link to the PRD acceptance criteria IDs
  (`AC-1`, `AC-2`, ...) instead of copying them.

## Templates

Start from this CLI’s `templates/TRD-*.md`, or from templates the history
repo keeps (architecture, implementation, testing). Copy structure into the
output path; do not invent a second template tree here.

`jarvis plan` writes empty skeletons. If JSON `source` is `tracker`, fetch the
work item with the host MCP named in `tracker_mcp`. If `source` is `user`,
use the user’s markdown or `REQUIREMENT.md`. Do not load Jarvis secrets.

## Output path

**Single-repo:**

```
docs/plans/<KEY>/TRD-Implementation.md
docs/plans/<KEY>/TRD-Architecture.md
docs/plans/<KEY>/TRD-Testing.md
```

**Multi-repo (same ticket):**

```
docs/plans/<KEY>/TRD-Architecture.md
docs/plans/<KEY>/<repo-name>/TRD-Implementation.md
```

`<repo-name>` is a key from `jarvis.yaml` `repos`.

## Frontmatter (required)

```yaml
---
issue: TEAM-123
tracker: jira
tier: A | B | C
test_source: TRD-Testing | tracker:TEAM-123 | user-prompt | waived
repos: [api]
type: implementation
status: draft
---
```

Older plans may say `test_source: jarvis:TRD-Testing`. Treat it as
`TRD-Testing`. Always write `TRD-Testing` in new or updated frontmatter.

Story tiers: `jarvis.yaml` `tiers` (default when unsure: `tiers.default`).

## Workflow

0. **Source baseline** — record local revision. Planning does not fetch,
   checkout, pull, rebase or extract implicitly. Graph navigation is
   advisory ([graph-first-analysis](../../rules/graph-first-analysis.md)).
1. Resolve the work-item key; confirm target repo(s) via `jarvis.yaml` `repos`.
2. If `jarvis plan --json` reports `source: tracker`, fetch the issue with the
   host MCP (`tracker_mcp`). If `source: user`, use the user’s requirement.
   Jarvis does not call the tracker API.
3. Check `docs/plans/<KEY>/` — update existing doc if present, do not duplicate. If `PRD.md` is there, read it first ([Starting from a PRD](#starting-from-a-prd)).
4. Confirm doc type: architecture, backend implementation, frontend implementation, or testing plan.
5. **Classify story tier** with the user. Default is `jarvis.yaml` `tiers.default`. Record `tier` and `test_source`. **Tier A:** write `TRD-Testing.md` with [test-case-design](../test-case-design/SKILL.md). Never leave it empty. **Tier B:** skip `TRD-Testing.md` unless the user asks; Jira remains a valid test source. **Tier C:** minimal plan; user may waive full grill.
6. Classify work: **New plan** / **Update** / **Major rewrite** (treat like new for grilling).
7. **Grill with docs** (mandatory before steps 8+ for every new plan and major rewrite; abbreviated or skipped for Tier C when the user confirms) — follow [grill-with-docs](../grill-with-docs/SKILL.md) in full:
   - Load the team glossary and ADR tree (paths the history repo uses; often `docs/CONTEXT.md` and `docs/adr/`).
   - **Hard stop:** post grilling **round 1**, then wait. Do not open TRD templates or write under `docs/plans/` until grilling is done and the user confirms shared understanding.
8. Read the matching template; copy structure into the output path.
9. Fill document information from the history repo’s team file if it has one (squad, author, reviewers). **Related Story** must match the folder key. Link issues using `tracker.site` from the resolved project/`jarvis.yaml`.
10. Complete required sections using grilled decisions. Link the glossary; link `docs/adr/` for persistent decisions. Skip sections that add no value. Do not duplicate tracker acceptance criteria into a sync file.
    - **No design choices left to the implementer.** Data shapes, API paths and payloads, error handling, retries, defaults, config keys, and new dependencies are decided in the TRD. "Update the service" is not enough.
    - Fill **Alternatives considered** (architecture) and **Files to change**, **Phases**, and **Verification** (implementation). Fill **Related Story** and leave **Related PR** for step 14.
11. If implementation follows an architecture doc, link it. **TRD-Testing (Tier A):** write it with [test-case-design](../test-case-design/SKILL.md).
12. **Plan review** — run [plan-review](../plan-review/SKILL.md) on the TRD (fresh-context reviewers, then resolve):
    - **Tier A:** hard gate. The ready-to-build check must be **GO** and the stop rule met, or the user must waive it by name. No `gh pr create` before that.
    - **Tier B:** advisory. Run it and show the result; open findings do not block the PR.
    - **Tier C:** skip unless the user asks.
    - Record the result in one line for the PR body (for example `plan-review: 2 rounds, GO`, `plan-review: advisory, 1 open Major`, `plan-review: skipped (Tier C)`, or `plan-review: waived by <user>`).
13. Create the PR in the **target product repo**; add the plan doc link and the plan-review line to the PR body. When the round is final, add the team’s review label if they use one (do not invent a second process).
14. Update the plan doc **Related PR** field with the PR URL.

## TRD-Testing

`TRD-Testing.md` is the team’s own guide for checking the feature locally.
QA, if the team has a separate QA function, tests outside this flow and
does not approve the Jarvis doc.

- **On TRD creation:** for a new **Tier A** TRD, write `TRD-Testing.md` with [test-case-design](../test-case-design/SKILL.md). Never leave it empty. **Tier B:** skip unless asked. **Tier C:** only when someone asks.
- **On review:** the reviewer asks for it on **Tier A** when it is missing. Request only; it never blocks the review.
- **Style:** the writer may read existing test-management cases (read-only) to match naming. Do not create cases in that system from this skill.

## PR body minimum

```markdown
## Summary

<1-3 bullets>

## Plan document

<link to history-repo TRD>

<plan-review result line>

## Test plan

<checklist>
```

Do not run `gh pr create` until the plan doc exists, the user has reviewed it (or explicitly waived), and step 12 is satisfied for the tier: **Tier A** needs a ready-to-build **GO** or a named waiver; **Tier B** needs plan-review run (advisory); **Tier C** needs nothing more.
