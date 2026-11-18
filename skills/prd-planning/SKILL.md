---
name: prd-planning
description: >-
  Write or update a PRD (Product Requirements Document) for a work-item key in
  the history repo. Use when a product owner or PM asks to plan a feature,
  write a PRD, or turn a ticket into requirements. Grill first, then fill.
---

# PRD planning

A PRD says **what** to build and **why**, from the user's point of view. It
does not say how the system builds it. An engineer writes the TRD from it
later.

The PRD lives in the **history repo**, at `docs/plans/<KEY>/PRD.md`. The
folder name is the exact work-item key. TRDs for the same key go in the same
folder later.

## When to use

- A product owner or PM wants a new PRD, or wants to update one.
- A ticket must become clear requirements before engineering starts.

**Hard gate:** do not write PRD content before grilling round 1 is posted and
answered. Scaffolding the empty file is fine.

## 1. Resolve context

1. Run `jarvis project current --json`. Read `role`, `role_skills`, and
   `plan_type`. If `role_source` is `default` and a project is set, ask the
   user for their role in that project and run
   `jarvis project update <name> --role product` (or `engineer`). Never guess.
   If there is no `role` field, the binary is too old; ask the user to run
   `jarvis update`.
2. Run `jarvis plan <KEY> --json` (the product role defaults to `--type prd`).
   Read `plan_dir`, `prd_file`, `source`, and `tracker_mcp`.
3. If `action` is `found`, update the existing PRD. Do not start a second one.

## 2. Fetch facts

- If `source` is `tracker`, fetch the work item through the host MCP named in
  `tracker_mcp`. If `source` is `user`, use the user's text or the
  "Source notes" section.
- Read the history repo glossary, earlier PRDs and TRDs in `docs/plans/`, and
  any decision records the repo keeps.
- Never read secrets files or tokens.

## 3. Grill

Load [grilling](../grilling/SKILL.md) with the product question set in
[grilling/references/product-questions.md](../grilling/references/product-questions.md).
Work the design tree in rounds: problem, users, goals, scope, flows, business
rules, acceptance criteria, edge cases, dependencies, rollout.

Small story: if the owner says the story is small, ask only scope, rules,
acceptance criteria, and risks.

## 4. Engineering checks

Some questions are about how the current system behaves. Examples: "Does the
order screen already show a discount?" or "Is there a daily limit today?"

- If you cannot confirm the answer from documents, add a row to the PRD
  "Engineering checks" table with status `open`.
- Never guess the answer. Never ask the product owner to answer it.
- Keep grilling the other questions. An open check does not block the PRD.

An engineer answers each check later. The answer goes into the TRD and back
into this table (status `answered`, the answer, the commit read, and a link to
the TRD). Ask the product owner to review those answers when they arrive.

## 5. Model terms

Load [domain-modeling](../domain-modeling/SKILL.md). Under the product role,
write glossary terms only. Do not write architecture decision records.
Product decisions go in the PRD "Decisions" section.

## 6. Fill the template

Fill the scaffold from `jarvis plan`. Rules:

- Each business rule (`BR-1`, `BR-2`, ...) is one testable sentence.
- Each acceptance criterion (`AC-1`, ...) uses Given / When / Then and names
  the rule or flow step it proves. See
  [references/acceptance-criteria.md](references/acceptance-criteria.md).
- Use numbers, not adjectives ("within 2 seconds", not "fast").
- Remove empty optional sections and the scaffold note.
- Do not put architecture, service names, or database design in the PRD.

## 7. Review loop

1. Check the draft with
   [references/prd-review-checklist.md](references/prd-review-checklist.md).
2. Run [plan-review](../plan-review/SKILL.md) on the PRD: fresh-context
   reviewers explain it back and apply the PRD lenses, then resolve the
   findings with the owner's input until its stop rule is met. If the host
   has no sub-agents, say so and record `plan-review: not run`.
3. Show the PRD and the plan-review result to the owner. Apply their edits.
4. Set `status: in-review` while the owner reviews.
5. Set `status: signed-off` only when the owner says so. A signed-off PRD is
   settled: engineers do not ask those questions again.

## 8. Hand off

- Commit `PRD.md` on a branch in the history repo and open a pull request.
  Follow the team's normal process from config or the history repo's own docs.
- Keep the PRD and the commit message about the product. Do not name the
  planning tool in the PRD, the commit, or the pull request.

## Done when

- The grilling frontier is empty.
- No placeholder text is left.
- Open engineering checks are listed in the table.
- plan-review met its stop rule, or the owner accepted the open findings.
- The owner confirmed the PRD.
