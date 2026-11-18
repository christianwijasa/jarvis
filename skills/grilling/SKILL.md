---
name: grilling
description: >-
  Grill the user relentlessly about a plan, decision, or idea, in rounds.
  Required before a new TRD (via grill-with-docs) or a new PRD (via
  prd-planning): post round 1 and wait before any docs/plans file. Use when
  stress-testing TRD or PRD plans, architecture or product choices, or domain
  language.
---

<!--
Adapted from mattpocock/skills skills/productivity/grilling/SKILL.md
at 49dd158d1076134a641b33efb035946536778336, MIT, Copyright (c) 2026 Matt Pocock.
See LICENSES/mattpocock-MIT.txt and https://www.aihero.dev/skills
-->

Interview the user relentlessly until you reach a shared understanding.

When trd-planning, grill-with-docs, or prd-planning is active for a **new plan**, the first user-visible output after fact-fetch must be **round 1** (frontier questions + recommended answers), not a TRD or PRD draft. Map this as a **design tree**: every decision branches into the decisions that hang off it.

Work the tree in **rounds**. The **frontier** is every decision whose prerequisites are already settled: the questions you can ask _now_ without guessing at answers you haven't heard yet. Ask the whole frontier in one round: number each question and give your recommended answer. Then wait for the user's answers before the next round.

Format a round like so:

```
❓ **Q1** - **<question title>**: <question body, might be multiple paragraphs, including multiple choices>

➡️ <your recommended answer>

---

❓ **Q2** - **<question title>**: <question body, might be multiple paragraphs, including multiple choices>

➡️ <your recommended answer>
```

Word each question so "yes" accepts your recommended answer.

**Diagram the gaps (optional, at most one per round).** When the story has 3 or more states or services and a diagram beats a paragraph, add one picture of _my reading of the story_ above the questions. Mark each unclear part as an `OPEN Qn` box tied to the matching numbered question. Prefer Mermaid unless the user asks for the optional `archify` pack.

**Question set.** The rounds method, the frontier, the recommended answers, and "yes accepts" are the same for every plan. Only the topics change:

- Engineer plans (TRD): use the design topics in `grill-with-docs`.
- Product plans (PRD): use [references/product-questions.md](references/product-questions.md). Do not ask about architecture, services, tiers, or test gates.

Each round the user answers reshapes the tree: settled decisions push the frontier outward and unblock questions that depended on them. Recompute the frontier and ask the next round. A question whose answer depends on another question still open in this round belongs to a _later_ round, not this one.

Finding _facts_ is your job, never the user's. When a frontier question needs a fact from the environment (documents, the tracker, the history repo, existing PRDs and TRDs, and code or the project graph when available), look it up yourself; don't ask the user for anything you could look up. Don't block on it: a running exploration is an unsettled prerequisite, so only the questions downstream of it wait; ask the rest of the frontier now. The _decisions_ are the user's: put each to them and wait.

**Needs engineering check.** When a fact needs code or systems you cannot read, do not guess and do not ask the user. Record it as a `needs engineering check` item (in a PRD, the "Engineering checks" table) and move on.

The session is done when the frontier is empty: every branch of the design tree visited, nothing left silently assumed. Do not act on it until the user confirms you have reached a shared understanding.
