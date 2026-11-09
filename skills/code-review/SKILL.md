---
name: code-review
description: >-
  Review a branch diff or GitHub PR along two axes: Standards (team
  conventions) and Spec (TRD / Jira issue). Optional over-engineering
  pass: references/ponytail-review.md. Trigger: "code-review", a PR URL,
  or after implementation before opening a PR. Author side:
  references/receiving-code-review.md. Based on
  https://www.aihero.dev/skills-code-review — Atlas uses TRDs in the
  history repo as the spec source.
---

<!--
§3 target-repo standards search adapted from mattpocock/skills
skills/engineering/code-review/SKILL.md at 49dd158, MIT, Copyright (c) 2026
Matt Pocock. Ponytail-review from DietrichGebert/ponytail. Receiving review
from obra/superpowers. See upstream.yaml.
-->

Two-axis review of the diff between `HEAD` and a fixed point the user
supplies (or a GitHub PR's base…head):

- **Standards**: does the code conform to documented standards for that repo?
- **Spec**: does the code faithfully implement the TRD / Jira issue?

Both axes run as **parallel sub-agents** when the host supports them;
this skill aggregates findings without merging or reranking axes.

History repo path and Jira site come from `atlas.yaml`.

## Process

### 0. Detect PR mode (when applicable)

If the user pastes a GitHub PR URL or names an open PR:

1. Parse `owner/repo` + PR number from the URL.
2. Fetch metadata + diff via `gh` (do **not** require a local checkout of the PR branch).
3. Fixed point = PR base branch. Pin the **reviewed head SHA**; every verdict applies to that SHA only.
4. **Round** = number of distinct reviewed head SHAs. Rebuild the findings ledger from prior Atlas reviews (`## Atlas review` prefix). If the latest prior verdict is `APPROVE` and its `commit_id` ≠ head, that approval is **stale**; review at least `<approved-sha>..<head>`.
5. After aggregate, apply the **review-readiness gate**, then submit.

Local/WIP review (no PR): skip submit.

On any `gh`/`git` auth failure (`HTTP 401`, `Bad credentials`), retry
once with `env -u GH_TOKEN -u GITHUB_TOKEN`. Never print or edit tokens.

### 1. Pin the fixed point (once)

User supplies commit, branch, tag, `main`/`master`, `HEAD~N`, etc.
Default for feature work: merge-base with remote default branch. In PR
mode, use the PR base…head.

```bash
git diff <fixed-point>...HEAD
git diff <fixed-point>...HEAD --stat
git log <fixed-point>..HEAD --oneline
```

Pass sub-agents **file paths** and **file:line** hunks — do not paste
the full diff twice. Include **untracked files** when reviewing WIP.

### 2. Refresh and identify the spec source (every review, every round)

Jira key from branch/title; parent story when the branch is a sub-task.
Per-repo implementation TRD:
`<history.path>/docs/plans/<KEY>/<repo-name>/TRD-Implementation.md`;
single-repo plans keep it at `docs/plans/<KEY>/TRD-Implementation.md`.

Never trust the local history checkout alone. Fetch (read-only; no
pull, checkout or reset) and read plans from remote refs of the
history repo (`atlas.yaml` `history.path`, and `history.github` when
set for `gh pr list`).

**Refresh gate (no verdict on a stale history repo):** if fetch fails
after the auth retry, **post no verdict**. If a merged history-repo PR
for the key is not on the fetched default branch, **post no verdict**.
No merged plan PR, or no key: the gate **passes**; note `no spec` in
the Spec section.

Check **all three** locations (and the parent key) before deciding:

1. Default branch: `docs/plans/<KEY>/`.
2. The ticket's plan branch on origin.
3. Links in the PR body (history-repo `blob`/`tree` URLs, plan paths, Jira).

Then: user-provided spec path or Jira AC. If sources disagree, prefer
the PR-linked revision, else the newest commit touching the plan, and
note the divergence.

**TRD-Testing request (PR mode, Tier A only):** when the plan is
**Tier A** and there is no `TRD-Testing.md` on any ref checked, ask
the author to write it with [test-case-design](../test-case-design/SKILL.md).
Once per key. **Tier B:** do not ask unless the user asked. Request
only: not a blocker or nit.

Record the spec revision used (path, ref, short SHA) in the review's
collapsed `<details>` block.

### 3. Standards checklist (subagents)

Pass sub-agents [references/standards-checklist.md](references/standards-checklist.md).
Karpathy/Ponytail already apply globally.

**Target repo standards:** search the pinned tree of the repo under
review for every file that documents how code should be written. When
`CONTRIBUTING.md` or `CODING_STANDARDS.md` exists, it must be on the
list. Those docs add to the Atlas checklist and never relax a hub
rule: a target repo cannot waive coverage gates from `atlas.yaml`,
the evidence rule, or `test-case-gate:` / `test-data-gate:`.

**Optional — graph hint:** when the diff touches an integration
client or its config, look up matching nodes in the team graph
([graphify](../graphify/SKILL.md)) and verify every caller against
source before citing it.

### 4. Spawn both sub-agents in parallel

**Standards** — pinned diff stat, checklist, target-repo standards docs.
**Spec** — changed files relevant to TRD, TRD excerpts from the recorded
spec revision, Jira summary if needed. Skip Spec when no TRD exists.

### 5. Aggregate

Classify each finding as **blocker** or **nit**:

- **Blocker**: broken behavior, missing required TRD/Jira behavior, security issue, coverage gate fail, or a clear standards violation that must land before merge.
- **Nit**: style, naming, optional cleanup. Author may defer.

**Every blocker states its required fix and what to preserve:** name
the adjacent behaviour the fix must keep.

**Missing evidence is a blocker unless the PR is docs/rules-only.**
The author must have posted fresh evidence for the reviewed head
([development § Evidence before done](../development/SKILL.md#evidence-before-done)).
Your own run of a check does not replace the author's evidence; a red
result from your run is its own blocker.

**Always flag** deleted, skipped or weakened tests and removed error
handling. Each is a blocker unless the PR or spec justifies it.

**Round 1 does the full pass:** report every nit found.

**Plan / TRD PRs** in the history repo: check that Tier A repos have
implementation TRDs; run the history repo's markdown/contract checks
when those scripts exist; ask for TRD-Testing on Tier A when missing
(request only).

**Merge safety** (PR mode):

- Behaviour on merge with **every new or changed flag at its default**.
- **Companion PRs** the change depends on are merged and deployed first, or merge is provably safe without them.
- **Review-readiness:** applicable CI on the reviewed head must be green. If any applicable check is **pending** or **red**, stop: do not post a review, and do not poll.

#### 5a. Findings ledger (re-review only)

Carry **every** finding from prior Atlas reviews forward: Fixed, Open,
or Withdrawn (reason and evidence required). An `Open` blocker keeps
the verdict at `REQUEST_CHANGES`.

| Verdict | When |
| ------- | ---- |
| `REQUEST_CHANGES` | ≥1 blocker on Standards, Spec or Merge safety |
| `APPROVE` | zero blockers (nits OK) |

If Spec was skipped after the step 2 checks and Standards and Merge
safety have no blockers → `APPROVE`; add one line under the verdict.

### 6. Submit review to the PR (PR mode only)

**Review-readiness gate** first. If checks are pending or red, **stop**.

Write the posted text in plain English: [plain-english](../../rules/plain-english.md).

1. Opening line = verdict + blocker count: `## Atlas review: Request changes, 2 blockers.` or `## Atlas review: Approve.`
2. One numbered entry per blocker: `file:line`, what is wrong, the fix, and `Keep …`.
3. Nits in a short separate list.
4. Re-reviews show only what changed since the last round.
5. Mandatory collapsed `<details>`: reviewed head, round, spec revision, full findings ledger, merge-safety, evidence.

Submit with `gh pr review` (`--approve` or `--request-changes`). If
GitHub rejects that (own PR, missing permissions), fall back to
`--comment` with `Intended verdict:` in the opening line.

An `APPROVE` covers only the reviewed head. Later commits make it stale
until a re-review.

### Same-head corrections

When a prior Atlas review exists on the **same `commit_id`** and only
the review needs correcting, edit or dismiss-and-resubmit rather than
posting another `REQUEST_CHANGES` that reads as an extra round.

## Multi-repo work

When the branch spans repos in `atlas.yaml` `repos`, review each
touched repo against the same TRD set. One aggregate report with
per-repo subsections. For a single-repo PR URL, review that repo only.

## Optional: ponytail review (over-engineering)

[references/ponytail-review.md](references/ponytail-review.md) when the
user asks for `/ponytail-review` or "what can we delete". Does not
replace this review or change GitHub verdict logic.

## Chain position

`grill-with-docs → trd-pr-planning → implement (development) → **code-review** → PR`

Author acting on a review: [`references/receiving-code-review.md`](references/receiving-code-review.md).
