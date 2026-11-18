<!--
Adapted from google/elephant-goldfish-model at
210977c1f76f07961d11d42b34bfd768c9267327 (agents/comprehension.md,
agents/zoom_out.md, agents/design_critic.md, agents/codebase_details.md,
resources/document_standards/feature-spec-standard.md), Apache-2.0.
Modified: rewritten in Jarvis's own words and split by PRD/TRD role.
See LICENSES/egm-Apache-2.0.txt.
-->

# Reviewer prompts

Every reviewer gets the same opening, then one lens.

## Opening (all reviewers)

> You are reviewing a plan document with no prior context. You have not seen
> any conversation about it. Read only the file at `<path>` and the files it
> links. Do not edit anything. Do not be polite: look for what is wrong,
> missing, or risky. Report each finding with a severity (Blocker, Major,
> Nit), the section or lines, what is wrong, and why it matters. If you find
> nothing at a severity, say so.

## Stage 1: explain it back (PRD and TRD)

> First, in your own words, write what this document proposes and how the
> existing system works, as the document describes it. Then list every
> section you could not restate (Blocker) and every term that is neither
> defined in the document, nor in a linked glossary, nor linked to its own
> doc (Major). Do not judge whether the plan is good. Judge only whether a
> new team member could follow it.

## Stage 2 lenses: PRD

Pick two or three. Default: **scope**, **testability**, **rollout**.

- **Scope and YAGNI.** Does every in-scope item serve the stated problem and
  goal? Is there a smaller version that still moves the metric? Is
  "Out of scope" real? Are rejected options and their reasons written down?
- **Testability.** Is each business rule one sentence a tester could check?
  Does each rule have at least one acceptance criterion, and does each
  criterion name its rule or flow step? Are numbers used instead of
  adjectives? Are empty, invalid, duplicate, late, and permission cases
  covered where they apply?
- **Rollout and risk.** Who gets it first? What number or event triggers a
  rollback? What existing data or users are affected? Which dependencies
  (teams, vendors, legal) could block release?
- **Boundary.** Does the PRD drift into architecture, service names, or
  database design? Are questions about current system behaviour in the
  engineering-checks table instead of being answered by guess?

## Stage 2 lenses: TRD

Pick two or three. Default: **zoom out**, **design critic**, **codebase
details**.

- **Zoom out.** Does the design fix the root problem in the simplest way?
  Does it reuse what the codebase already has? Flag speculative abstractions,
  early optimisation, and logic placed in the wrong layer.
- **Design critic.** Were real alternatives considered and rejected for
  stated reasons? What fails under load, retries, partial failure,
  concurrency, version skew, or rollback? Are security and data-migration
  risks named with a mitigation?
- **Codebase details.** Do the files, types, functions, endpoints, and config
  keys the TRD names exist where it says (or is it clear they are new)? Do
  existing behaviours it relies on actually work that way? Cite the file and
  symbol you read.
- **PRD trace (when a PRD exists).** Does every PRD acceptance criterion
  (`AC-n`) map to a TRD section or test case? Does the TRD silently change a
  signed-off PRD rule?
