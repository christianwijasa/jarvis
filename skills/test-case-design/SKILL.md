---
name: test-case-design
description: >-
  Write docs/plans/<KEY>/TRD-Testing.md as a local-testing guide: cases
  proving the TRD/PR delivers the story, plus regression cases for existing
  flows, each with how to run it locally. Use when creating a Tier A TRD,
  when the user asks for TRD-Testing on Tier B, or when reviewing a Tier A
  TRD/PR with no TRD-Testing.
---

# Test case design (TRD-Testing local-testing guide)

**Output:** one file, `<history.path>/docs/plans/<KEY>/TRD-Testing.md`
(history repo from `jarvis.yaml`). It answers two questions for whoever
tests locally: **(a) does the TRD/PR deliver what the story asks?** and
**(b) do existing flows still work?**

A merged TRD-Testing is a valid source for the
[test case gate](../../rules/plan-execution-test-case-guard.md). While
its PR is open, implementers may already use it as guidance.

**Who writes it:** the person creating the TRD, or any teammate following
this skill. Fill **Author** with whoever writes it.

## Triggers

| Trigger | Deliver as |
| ------- | ---------- |
| A **Tier A** TRD is created | The TRD author writes it on the TRD PR |
| User asks for TRD-Testing on a **Tier B** TRD | Same |
| Review of a **Tier A** TRD/PR with no TRD-Testing | Ask the author |
| **Tier B** with no TRD-Testing (Jira is the test source) | Do not write or request one unless the user asks |

**Tiers:** only **Tier A** requires a TRD-Testing. Keep historical
`TRD-Testing.md` files; do not delete them.

## Workflow

1. Read the Jira story AC (`jarvis.yaml` `jira.site`), the implementation TRD, and the PR diff if one exists.
2. Find existing coverage (read-only): linked cases in the team’s test tool, tagged automated specs for the touched endpoints/screens.
3. Build the AC → case map, then the regression list.
4. Fill traceability, env/data/flags, functional cases, cross-service cases, regression. Every case gets **How to run**.
5. List open questions at the end for the TRD author and reviewer.

## AC → case mapping

- Each AC gets at least one happy-path case (P0). Each AC clause about a fallback, error, empty state, permission, or flag state gets its **own** case. One case = one behaviour, verified by several assertions.
- Tie cases to ACs in the Objective (`Verify AC 1: …`) or a label. IDs in the traceability table and the case table match one-to-one.
- Size: expect 4–12 feature cases plus regression rows; split by client when the story spans more than one app.
- If an AC is not delivered by the TRD/PR, keep the case and flag it as a gap.

## IDs and names

| Item | Convention |
| ---- | ---------- |
| TC ID | `TC-<AREA>-NN`; `E2E-NN` cross-service; `REG-NN` regression |
| Case name | `[Platform][Module][Sub-area] <Actor> should <observable outcome> when <condition>` |
| API name | `[API][<Service>] <METHOD> <path> should <outcome>` |
| Flag variants | Put the flag state in the name |
| Regression | Same pattern, about the existing behaviour: `… should still …` |

## Fields per case

Write Objective, Precondition, Steps and Expected in plain English:
[plain-english](../../rules/plain-english.md).

| Field | Write it like |
| ----- | ------------- |
| Objective | One sentence: `Verify …`, naming the AC |
| Precondition | Numbered list of concrete state: persona, flag, data, app version |
| Steps | Numbered user or API actions |
| Test data | Shared-env or seed/fixture data only, never customer or production data. Concrete placeholders, never "valid data" |
| Expected | Checkable bullets: copy, HTTP code, DB column, "request not sent" |
| Priority | P0–P2 (map to High / Normal / Low if the team’s tool uses those words) |
| How to run | Enough to reproduce without asking |

## How to run locally

Put shared setup once; per case, write only what differs.

- **API:** an HTTP client collection path, or a `curl` with method, path, headers (`Authorization: Bearer $TOKEN`), and body.
- **UI:** app, login persona, numbered taps/clicks, and what to look at.
- **Events:** exchange/queue and a sample payload.
- **DB check:** table and `WHERE` clause, expected column values.
- **Env and flags:** env name; flag name and value per case (on vs off are separate cases).
- **Seed data:** record IDs, SQL/fixture, and how to reset. Mutable shared personas go in `test-data-setup.md` ([gate](../../rules/plan-execution-test-data-guard.md)).
- Never put tokens or passwords in the doc; name the env var instead.

## Negative and edge checklist

Pick what applies. Each pick is its own case.

- Missing/null field shows a placeholder. Empty state. Error view with retry.
- Auth and scope: 401; data isolation between tenants or roles.
- Validation: self-target or duplicate, blank mandatory fields ignored, non-terminal status triggers nothing.
- Boundaries and config: value just outside a min/max, filter combinations, chained rules.
- State transitions: edit flow resets dependent fields, re-select keeps data, re-open refreshes counts.
- Source variants: flag on vs off, web vs app, sibling product types.

## Regression: existing flows that must not break

1. **Impact list:** every touched endpoint, event, screen, shared table, and feature flag, from the TRD and PR diff.
2. **Reuse first:** cite existing case keys or spec paths, with the command to re-run them. Write new `REG-NN` cases only where nothing fits.
3. **Always include:** old client / additive contract; flag off = legacy; the untouched sibling variant; unchanged copy, navigation, and recipients for related notifications.
4. Priority P1 by default; P0 when the sibling is a core money, identity, or journal flow.

## Learning from the team’s test tool (read-only)

If the team keeps cases in Jira or another test tool, read them for the
same module or linked story; borrow scenarios and cite real keys. Never
create, edit, or delete cases there from this skill.

## Output checklist

- [ ] Every AC mapped; undelivered AC flagged
- [ ] Each case has Objective, Precondition, numbered Steps, concrete Test data, checkable Expected
- [ ] Negative/edge cases cover isolation, validation, empty/error, variants, flag states
- [ ] Every case has How to run; no secrets
- [ ] Regression lists impacted flows, reuses existing cases, covers old client, flag off, sibling variants
