---
name: grill-with-docs
description: >-
  Mandatory first step for every new TRD plan. Interview (grilling rounds)
  before any docs/plans file or template fill; writes the glossary and ADRs
  as terms resolve. Also use for major TRD rewrites, trd-pr-planning on a
  new Jira key, or when the user says /grill-with-docs. Based on
  https://www.aihero.dev/grill-with-docs.
---

<!--
Adapted from mattpocock/skills skills/engineering/grill-with-docs/SKILL.md
at 49dd158d1076134a641b33efb035946536778336, MIT, Copyright (c) 2026 Matt Pocock.
Hub UPSTREAM.md skipped later grill-with-docs upstream edits on purpose.
-->

# Grill with docs (TRD enhancement)

Stateful grilling for Jarvis TRD work. Vocabulary lands in the team
glossary as it resolves; durable decisions land in `docs/adr/` in the
history repo (`jarvis.yaml` `history.path`). Everything else feeds the
TRD sections directly.

## Planning gate (non-negotiable for new plans)

**New plan** = first TRD for a Jira key, first doc type under
`docs/plans/<JIRA-KEY>/`, or user asks for a first draft.

Before any new file under `docs/plans/` or copying a TRD template:

1. Read this skill plus [grilling](../grilling/SKILL.md) and [domain-modeling](../domain-modeling/SKILL.md).
2. Fetch facts (Jira, graphify, code) per workflow below.
3. Post **grilling round 1** — full frontier, numbered questions, recommended answers (`➡️`), then **stop and wait** for the user.
4. Continue rounds until the frontier is empty and the user confirms shared understanding.
5. Only then may trd-pr-planning create plan files and fill templates.

If the agent would skip the interview and write a TRD in one shot, **stop** — that violates this gate. Minor typo/link fixes skip only when the user explicitly says so.

**Upstream:** [grill-with-docs on AI Hero](https://www.aihero.dev/grill-with-docs)

## Preconditions

Read before starting:

- Jira fetch from `jarvis.yaml` `jira.site`
- Glossary + ADR layout in the history repo
- Story tier with the user (`jarvis.yaml` `tiers`)
- Existing glossary, `docs/adr/`, and `docs/plans/<JIRA-KEY>/` if present

## Dependencies (must load both)

1. **[grilling](../grilling/SKILL.md)** — interview rounds until shared understanding
2. **[domain-modeling](../domain-modeling/SKILL.md)** — write glossary + ADRs inline as they resolve

If either dependency is missing, stop and report which skill failed to load.

## When to run

| Trigger                      | Action                                                |
| ---------------------------- | ----------------------------------------------------- |
| New TRD (any type)           | Run before first template fill                        |
| Major TRD rewrite            | Re-run from frontier; update glossary/ADR/TRD together |
| Minor typo / link fix        | Skip                                                  |
| User says `/grill-with-docs` | Run standalone on current topic                       |

## Workflow

0. **Story tier** — confirm with the user. Default is `jarvis.yaml` `tiers.default`.
   - **Tier A:** full grilling rounds until frontier empty.
   - **Tier B:** focus on integration vocabulary and Jira AC gaps; skip empty Testing-file topics.
   - **Tier C:** confirm scope, waive, and risks only — **no multi-round interview** when the user explicitly agrees.
1. **Fetch facts** — Jira issue; graphify/code for existing behaviour; read PRD links from Jira when present.
2. **Grill** — follow grilling skill: one round = whole frontier, numbered questions + recommended answers, then wait (unless Tier C abbreviated path above).
   - **Diagram the gaps** (optional): when the story has 3+ states or services, add at most one diagram per round framed as _my reading of the story_, with `OPEN Qn` boxes tied to the numbered questions.
3. **Model** — follow domain-modeling: resolve terms → glossary immediately; qualifying decisions → `docs/adr/`.
4. **Enhance TRD** — use settled understanding to strengthen Background, ADR blocks, integration contracts, edge cases.
5. **Confirm** — user confirms shared understanding before TRD is marked ready for PR.

## TRD content rules

- Do not paste the full glossary into TRD — link it.
- Do not duplicate ADR prose — TRD ADR blocks summarize; point to `docs/adr/` for persistent decisions.
- Questions the codebase or Jira can answer: answer them yourself; never ask the user.

## Done when

- Grilling frontier is empty and user confirmed understanding.
- TRD required sections reflect grilled decisions (not template placeholders).
- New/resolved terms appear in the glossary (if any vocabulary was settled).
