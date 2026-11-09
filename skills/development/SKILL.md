---
name: development
description: >-
  Coding standards (Karpathy + Ponytail) for repos listed in atlas.yaml.
  Use when implementing features, fixing bugs, refactoring, or reviewing
  diffs. Before coding, the test case gate must pass when enabled. Meet
  coverage gates from atlas.yaml before opening a PR. Failing test first
  for Tier A state, race and ordering logic.
---

<!--
Sections "Evidence before done" and step 6 test-first adapt obra/superpowers
skills/verification-before-completion and skills/test-driven-development at
commit bb92a77741419a4ab5f06e711a283343f1ada0c3 (v7.0.0), MIT,
Copyright (c) 2025 Jesse Vincent. Intentional shortcuts adapt
DietrichGebert/ponytail at 09c434a. See SUPERPOWERS-NOTICE and UPSTREAM in
the hub snapshot; LICENSES/ in this repo.
-->

# Development

Applies to every repo named in `atlas.yaml` `repos`. This CLI is the
skill source; product code lives in those checkouts. TRDs live in
`history.path`.

## Two layers (always both)

| Layer | Rule file | Role |
| ----- | --------- | ---- |
| **Karpathy** | [`karpathy-guidelines`](../../rules/karpathy-guidelines.md) | Process: assumptions, simplicity, surgical diffs, verifiable goals |
| **Ponytail** | [`ponytail`](../../rules/ponytail.md) | Solution: YAGNI ladder, reuse, shortest working diff |

Karpathy runs first (understand, plan, verify). Ponytail ladder runs
**after** you trace the real flow — not instead of reading the code.

Match style and patterns of the repo you are editing. Paths come from
`atlas.yaml` `repos`. Do not assume a home-directory layout.

## Implementation workflow

0. **Record source baseline** — inspect the selected checkout and preserve unrelated changes. Fetch/pull/checkout/rebase only under explicit authorized scope.

1. **Plan** — TRD doc exists or user waived; see [`trd-pr-planning`](../trd-pr-planning/SKILL.md).
2. **Test case gate** — when `atlas.yaml` `gates.test_case` is true, load [`plan-execution-test-case-guard`](../../rules/plan-execution-test-case-guard.md). Read `tier` / `test_source` from plan frontmatter. **Hard stop** when the gate applies and checks fail. If the user waives, add `test-case-gate:` in-code notes.
3. **Test data setup gate** — when `gates.test_data` is true, load [`plan-execution-test-data-guard`](../../rules/plan-execution-test-data-guard.md) before data-dependent edits.
4. **Understand** — TRD + graph query ([graph-first-analysis](../../rules/graph-first-analysis.md)): map the story to code, then read **only** files/lines at cited locations (plus files you will edit). Trace end-to-end along that path. State assumptions or ask.
5. **Climb ponytail ladder** — stop at first rung that holds.
6. **Implement surgically** — every changed line traces to the request; match target repo style. **Failing test first** is required for Tier A state, race and ordering logic (and any bug fix there), proven by a red CI run on a test-only commit or pasted local red output (commit order alone is not accepted) — [`references/test-first.md`](references/test-first.md). Bug with unknown cause: [`debugging`](../debugging/SKILL.md) first.
7. **Verify** — run the repo’s test suite; meet **coverage gates** below on every production file you changed. Then collect **Evidence before done**.
8. **Commit (when user asks)** — [`commit-message`](../../rules/commit-message.md).
9. **Code review** — [code-review](../code-review/SKILL.md) before opening PR.
10. **Review diff** — optional: [ponytail-review](../code-review/references/ponytail-review.md).
11. **Open PR** — use [generate-pr-description](../generate-pr-description/SKILL.md). When the round is final, add the team’s review label if they use one.

## Coverage gates

Read percents from `atlas.yaml`:

- `gates.backend_coverage_percent` (example: 100) on changed non-excluded production logic in backend repos
- `gates.client_coverage_percent` (example: 80) on changed production logic in client apps

**Do not open a PR** until those gates pass for the files you changed
(“changed” = new or materially edited code in the merge-base…HEAD diff).
Config-only PRs: no unit gate; state manual verify in the TRD/PR.
QA or E2E-only repos use their own rules.

**Workflow**

1. List touched production sources from `git diff <merge-base>...HEAD --name-only`.
2. Add or extend tests until coverage on that set meets the gate.
3. Re-run the repo’s test command; fix failures before review.
4. PR body: say which command you ran.

Non-trivial logic still needs at least one runnable check (Karpathy);
this gate is **in addition**.

Flutter/Dart client extras: format and analyze clean; see
[`flutter-conventions.md`](references/flutter-conventions.md) when the
touched repo is Flutter.

## Evidence before done

**No "fixed", "done", "passing" without fresh evidence.** Fresh = run on
the current head after your last change, in this session.

Paste into the PR (body or a comment after the last push, naming the
head SHA) the output of each **required** check that applies.

**Evidence exemption (docs/rules-only).** Green CI on the head is
enough when the PR is only Markdown docs, skills, rules, or images.
**Not exempt:** code, scripts, tests, contracts, generated code, runtime
or deploy config, CI workflows, and files that control gates. A mixed
PR needs evidence for its non-exempt part.

| Applies when | Evidence |
| ------------ | -------- |
| Code changed | Test output for the changed area (command + summary line). If tests only run in CI, a green CI job on the head plus its link is enough. |
| Diff touches an event, status, HTTP code or field that a TRD names | Contract check across sibling plans, or a written justification per hit |
| Repo has generated code | Generator re-run with no diff |
| SQL, migration or data-shape change | `EXPLAIN` from a shared env (read-only) or a QA result. Mocked repositories do not prove SQL |

If a check fails or cannot run, say so plainly. After review fixes,
re-run all of it ([receiving review](../code-review/references/receiving-code-review.md)).

## Bug fixes

Symptom in ticket ≠ fix location. Grep all callers of the function you
touch; fix shared logic once at the root. Cause not known yet:
[`debugging`](../debugging/SKILL.md).

## Intentional shortcuts

Only for a real corner cut with a known ceiling (e.g. global lock,
O(n²), naive heuristic): mark it with a `ponytail:` comment naming the
ceiling and upgrade path. No comment for ordinary simple code.

## Related skills

| Skill | When |
| ----- | ---- |
| [`trd-pr-planning`](../trd-pr-planning/SKILL.md) | Before PR / plan doc |
| [`graphify`](../graphify/SKILL.md) | Optional codebase analysis |
| [`debugging`](../debugging/SKILL.md) | Failing test, red CI or unknown cause |
| [`receiving-code-review`](../code-review/references/receiving-code-review.md) | Acting on review comments |
