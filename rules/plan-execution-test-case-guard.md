---
description: >-
  Load when implementing from a TRD: readable test cases (TRD-Testing, Jira,
  paste or path). Trigger: development skill. Honours atlas.yaml gates.test_case.
---

# Plan execution — test case gate

## Scope

Run this gate when `atlas.yaml` has `gates.test_case: true` (the example
config turns it on). Skip the entire rule when that flag is false.

**Trigger:** the user asks to **implement**, **execute the plan**, **start
coding from TRD**, **apply TRD**, or you are about to edit **production
code** in any repo listed in `atlas.yaml` `repos` for a Jira-backed
feature. Respect **story tier** from the TRD frontmatter (and
`atlas.yaml` `tiers`).

**Hard stop:** do not edit production code for that feature until
**readable test cases** are loaded. Planning-only work (TRD authoring,
grill-with-docs, graphify, review without code changes) is allowed
without this gate.

Skill hook: **Test case gate** in [development](../skills/development/SKILL.md).

## What counts as ready

Test cases are **ready** when the agent can read **concrete scenarios**
(steps, inputs, expected outcomes — for example Given/When/Then or TC
IDs), not empty template placeholders.

Accept any **one** of:

| Source                     | Requirement |
| -------------------------- | ----------- |
| **History-repo `TRD-Testing.md`** | `<history.path>/docs/plans/<JIRA-KEY>/TRD-Testing.md` (parent story key when the branch uses a sub-task) with filled traceability and/or functional scenarios. |
| **Current user message**   | User pasted test cases or a full scenario table in the prompt. |
| **Jira key (user or TRD)** | User names a key, or the TRD links a test-case issue; fetch from the site in `atlas.yaml` `jira.site`. |
| **Explicit path or URL**   | User gives a file path, wiki page, or export the agent can read. |

A `TRD-Testing.md` counts once it is **merged** on the history repo
default branch. It is the team’s own local-testing guide, written with
[test-case-design](../skills/test-case-design/SKILL.md). A copy that
exists only on an open history-repo PR may be used as guidance; name
the PR in your source line.

## When not ready — ask once, then wait

If no source passes, **stop** and ask where test cases live. Include
what you already checked. Offer:

1. Point to a `TRD-Testing.md` (merged or on an open PR), or write it with [test-case-design](../skills/test-case-design/SKILL.md)
2. Jira key with scenarios
3. Paste scenarios in chat
4. File path or wiki URL
5. Explicit waiver (see below)

Do not guess scenarios from the implementation TRD alone unless the
user confirms that those rows **are** the authoritative test cases.

## Waivers (user must say so explicitly)

Skip the gate only when the user clearly waives it, for example
typo-only, docs-only, config-only, or an emergency hotfix with
reproduction steps already in Jira. Record the waiver before coding.

**In-code disclosure (mandatory on waiver):** any production code
written under a test-case gate waiver must state that behaviour is
**not yet implemented against test cases**. Add a `test-case-gate:`
comment on each **changed production file**. Remove those comments once
test cases are loaded and the implementation is aligned.

## After the gate

Summarize which source you used (one line). During implementation and
[code-review](../skills/code-review/SKILL.md), treat those scenarios as
verification intent alongside TRD rules.
