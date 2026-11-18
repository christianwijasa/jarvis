<!--
Adapted from google/elephant-goldfish-model at
210977c1f76f07961d11d42b34bfd768c9267327 (skills/egm-resolve/SKILL.md,
skills/egm-review-loop/SKILL.md), Apache-2.0.
Modified: rewritten in Jarvis's own words, shortened, and fitted to the
PRD/TRD owner flow. See LICENSES/egm-Apache-2.0.txt.
-->

# Resolve findings and stop rule

The writing session resolves. It wrote the plan, so wanting to disagree with
a reviewer is not evidence that the reviewer is wrong.

## Label every finding

| Label | Use when | Action |
| ----- | -------- | ------ |
| **FIX** | The finding is right. | Propose the exact edit. |
| **CONTEST** | The finding is wrong, and the proof is in something the reviewer could read (a file and symbol, or a decision already in the document). | Cite that proof. No edit. |
| **ESCALATE** | The finding looks wrong, but the proof is outside every file the reviewer could read (a team habit, a spoken decision). | Ask the user. Do not edit on your own. |
| **DECLINE** | A Nit not worth changing. | Record it. No edit. |

Rules:

- "This was intentional" is not proof. Without readable proof, use FIX or
  ESCALATE.
- Never contest a stage 1 (explain it back) gap. If a fresh reader could not
  follow the text, the text is missing something.
- Do not change a severity. If a Blocker looks like a Nit, CONTEST it and
  say why.
- Add context because a human reader needs it, not only to quiet a
  reviewer. If only a zero-context reviewer needs it, ESCALATE.

## Ask the user

List ESCALATE items first, then FIX proposals, then CONTEST and DECLINE.
Offer: accept all FIX, go item by item, or edit by hand. Apply only what the
user accepts. Then re-read the file before the next round.

## Stop rule

Fresh reviewers sample differently, so a round with zero findings is rare
even on a good document.

- **Settled:** a finding that was contested, escalated and confirmed, or
  declined. If a later reviewer raises the same issue at the same place, do
  not reopen it.
- **Fixed is not settled:** if a later reviewer raises a fixed issue again at
  the same place, the fix did not work. Reopen it.
- **Clean round:** no unsettled Blocker or Major, and no reviewer `not run`.
  Decide this from the review results, before applying that round's fixes.
- **Done:** two clean rounds in a row.
- **Cap:** five rounds. At the cap, or when a fix keeps causing new findings
  at the same place, list the open findings by severity and ask the user to
  accept them, continue, or send the plan back to grilling.

Small change (Tier C, or a user-confirmed small update): one round is
enough when it is clean.
