# Jarvis review — standards checklist (subagents)

One-page Standards axis. Full workflow: [development/SKILL.md](../../development/SKILL.md).
Karpathy/Ponytail are always-on rules.

## Process

- Surgical diff; every changed line traces to the request
- Match target repo style and patterns
- YAGNI; reuse before new abstractions

## Coverage gates

Percents come from `jarvis.yaml` `gates`. Blocker if changed logic lacks
tests or misses the gate on non-excluded production files.

## Client apps (including Flutter)

- Format and analyze clean on touched files
- Meet `gates.client_coverage_percent` on changed production logic
- See development skill and [`flutter-conventions.md`](../../development/references/flutter-conventions.md)

## Test-case / test-data disclosure

When `gates.test_case` is on and implementation did not use readable
test cases, changed production files need **`test-case-gate:`** comments.
When `gates.test_data` is on and data setup was waived, fixture/persona
files need **`test-data-gate:`**.

## Author evidence (PR mode, every round)

- Docs/rules-only PRs are exempt: green CI is enough.
- Otherwise: fresh evidence for the reviewed head. Missing, stale or failing evidence is a blocker.

## Failing test first (Tier A)

- Required for Tier A state, race and ordering logic, and bug fixes there ([test-first](../../development/references/test-first.md))
- No accepted proof is a blocker
- Elsewhere test-first is author guidance only

## Regression signals (always report)

- Deleted, skipped, or weakened tests — blocker unless justified
- Removed error handling — blocker unless justified
- Cite the removed lines (`file:line` on the base side)

## Blocker wording

Each blocker gives **Fix** and **Preserve** in 2–3 sentences, `file:line`,
using the codebase/TRD terms. Nits go in a separate list.

## Merge safety

- **Flags at default:** merge must not change live behaviour unless the spec says so
- **Companion PRs:** dependencies in other repos are merged and deployed first, or merge is safe without them
- **Checks:** applicable CI must be green. Pending or red: stop, do not post, do not poll

## Commits (when present in diff scope)

- Branch: `feat/<KEY>` or `fix/<KEY>`
- Subject: `feat({topic}): [{key}] …` or `fix({topic}): …`

## Smells (judgement)

Flag only clear issues; repo conventions win.
