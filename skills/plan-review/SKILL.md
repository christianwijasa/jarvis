---
name: plan-review
description: >-
  Fresh-context review of a PRD or TRD before it goes to its owner or before
  a product PR opens. Runs an explain-it-back check, role lenses, and (TRD
  only) a ready-to-build go/no-go, then resolves findings in the writing
  session. Use when prd-planning or trd-pr-planning reaches review, or when
  the user asks to review a plan document.
---

<!--
Adapted from google/elephant-goldfish-model (Elephant-Goldfish Model, EGM)
at 210977c1f76f07961d11d42b34bfd768c9267327, Apache-2.0, Copyright 2026 Google LLC.
Ideas taken from skills/egm-review, skills/egm-review-loop, skills/egm-resolve,
agents/comprehension.md, agents/implementation_readiness.md, and
resources/document_standards/. Modified: rewritten in Jarvis's own words,
cut to three stages, tied to PRD/TRD roles and story tiers.
See LICENSES/egm-Apache-2.0.txt and https://github.com/google/elephant-goldfish-model
-->

# Plan review (fresh-context)

The session that wrote a plan shares its blind spots. It knows things the
document never says. This skill checks the document with **fresh
reviewers**: sub-agents that see only the document and the files it links,
never the chat.

The writing session keeps the context. It does not review its own draft
here. It only resolves what the fresh reviewers find.

This follows the Elephant-Goldfish Model (EGM): the writing session is the
"elephant" that remembers; each reviewer is a "goldfish" that starts empty.

## When to use

| Caller | When | Target |
| ------ | ---- | ------ |
| prd-planning | after the self-check, before the owner sees the PRD | `docs/plans/<KEY>/PRD.md` |
| trd-pr-planning | after the TRD is filled, before `gh pr create` | primary `TRD-*.md` (and per-repo impl files) |
| user | "review this plan", "plan-review" | any plan file |

Tier rules for TRDs come from trd-pr-planning: **Tier A** must pass, **Tier
B** is advisory, **Tier C** skips this skill unless the user asks.

## Fresh reviewers only

- Start each reviewer as a **new sub-agent** with no chat history. Give it:
  the target file path, the files the document links (inside the history
  repo or target repos), and one prompt from
  [references/lenses.md](references/lenses.md). Nothing else. No summary of
  the conversation.
- Reviewers are **read-only**. They never edit the document or the code.
- Run reviewers of one stage **in parallel** when the host supports it.
- If the host has no sub-agent tool, **stop** and tell the user: a review in
  the same session is not a fresh review. Offer the self-check list instead,
  and record `plan-review: not run (no sub-agents)` in the hand-off.
- A reviewer that fails, times out, or returns nothing **did not review**.
  Retry it once. If it fails again, mark it `not run`. Never count it as a
  reviewer with zero findings.

## Stages

Run in order. Each stage writes its findings before the next starts.

1. **Explain it back.** One reviewer restates, in its own words, what the
   document proposes and how the current system works as the document
   describes it. Any section it cannot restate is a **Blocker**. Undefined
   terms that are not in the glossary or linked are **Major**. Keep the
   restatement in the report: the author must read it, because a wrong
   restatement shows exactly where the text misleads.
2. **Role lenses.** Two or three reviewers, each with one lens for the
   document type (see [references/lenses.md](references/lenses.md)). They
   hunt for gaps, risks, missing alternatives, and needless scope.
3. **Ready to build (TRD only).** One reviewer gives a **GO** or **NO-GO**
   against [references/ready-to-build.md](references/ready-to-build.md). PRDs
   skip this stage; their equivalent is "every rule has an acceptance
   criterion and every open engineering check is listed", which stage 2
   covers.

Each finding has: id (`<stage>-<round>-<n>`), severity (**Blocker**,
**Major**, **Nit**), the section or line range, what is wrong, and why it
matters.

## Resolve

The writing session resolves findings with
[references/resolve.md](references/resolve.md): label each finding, show
the user the labels and proposed edits, apply only what the user accepts,
then re-run the stages until the stop rule says done.

## Report

Write one report per round next to the plan, or in a scratch folder if the
team does not keep review notes in the history repo (ask once; default
scratch):

```markdown
# Plan review: <KEY> <file> (round <n>)

Reviewed: <path> at <commit or "working copy">

| Reviewer | Stage | Status |
| --- | --- | --- |
| explain-back | 1 | completed |
| <lens> | 2 | completed / not run |
| ready-to-build | 3 | GO / NO-GO / not run / n/a |

## Restatement
<stage 1 text, unedited>

## Findings
<Blockers, then Major, then Nits; each with its resolve label>
```

Tell the user the report path and the counts. Do not paste the whole report
into chat.

## Done when

- The stop rule in [references/resolve.md](references/resolve.md) is met,
  or the user accepted the open findings by name.
- For a TRD: stage 3 is **GO**, or the user waived it, or the tier makes it
  advisory or skipped (record which).
- The hand-off (owner review or PR body) states the result in one line,
  for example `plan-review: 2 rounds, GO, 0 open Blocker/Major`.
