---
name: grill-with-docs
description: >-
  Interview to sharpen a plan and domain language; writes CONTEXT.md and
  docs/adr/<project>/ as terms and decisions resolve. Jarvis hub variant — no TRD/Jira.
  Based on https://www.aihero.dev/grill-with-docs — config via docs/agents/.
disable-model-invocation: true
---

# Grill with docs (jarvis)

Stateful grilling: vocabulary lands in `CONTEXT.md` as it resolves; durable decisions land in `docs/adr/<project>/`. Optional summary in `docs/plans/<slug>/design.md`.

**Upstream:** [grill-with-docs on AI Hero](https://www.aihero.dev/grill-with-docs)

## Preconditions

Read before starting:

- [`docs/agents/issue-tracker.md`](../../../docs/agents/issue-tracker.md) — scratch issue + plan pairing
- [`docs/agents/domain.md`](../../../docs/agents/domain.md) — glossary + ADR layout
- Existing `CONTEXT.md`, `docs/adr/<project>/`, and `docs/plans/<slug>/` if present
- Manifest repos under `~/projects/` when behaviour lives in code

## Dependencies (must load both)

1. **grilling** — from `~/.agents/skills/grilling/SKILL.md` (interview rounds until shared understanding)
2. **domain-modeling** — from `~/.agents/skills/domain-modeling/SKILL.md` (glossary + ADRs inline)

If either dependency is missing, stop and report which skill failed to load.

## When to run

| Trigger | Action |
|---------|--------|
| New effort or fuzzy plan | Run before large implementation |
| Major plan rewrite | Re-run from frontier; update CONTEXT/ADR/plan together |
| Minor typo / link fix | Skip |
| User says `/grill-with-docs` | Run standalone on current topic |

## Workflow

1. **Fetch facts** — read scratch issue if user named one; read code in hub or manifest repos; read existing plans. Do not use Jira.
2. **Grill** — follow grilling skill: one round = whole frontier, numbered questions + recommended answers, then wait.
3. **Model** — follow domain-modeling skill: resolve terms → `CONTEXT.md` immediately; qualifying decisions → `docs/adr/<project>/`.
4. **Plan** — when user confirms slug or scratch issue, update or create `docs/plans/<slug>/design.md` with:
   - Problem and scope
   - Links to `docs/adr/<project>/` for persistent decisions
   - Integration / edge cases from grilling (not a pasted glossary)
5. **Confirm** — user confirms shared understanding before implementation.

## Content rules

- Do not paste the full glossary into plan files — link `CONTEXT.md`.
- Do not duplicate ADR prose in plans — summarize and point to `docs/adr/<project>/NNNN-*.md`.
- Questions the codebase can answer: answer them yourself; never ask the user.

## Done when

- Grilling frontier is empty and user confirmed understanding.
- Plan or scratch issue reflects grilled decisions (not placeholders).
- New/resolved terms appear in `CONTEXT.md` (if any vocabulary was settled).
