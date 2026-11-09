---
description: >-
  Load when a change depends on seeded test data. Trigger: development
  skill. Honours atlas.yaml gates.test_data.
---

# Plan execution — test data setup gate

## Scope

Run when `atlas.yaml` has `gates.test_data: true`. Skip entirely when
that flag is false.

**Trigger:** you are about to edit **data-dependent** artifacts for a
Jira-backed feature: fixtures, env persona files, or flows that assume
pre-seeded **mutable** shared-environment records.

**Allowed without gate:** TRD and grill-with-docs, docs-only work,
backend code with no new fixture/env coupling, read-only investigation.

Skill: **Test data setup gate** in [development](../skills/development/SKILL.md).

Respect **story tier** and plan frontmatter. Lower tiers: skip unless
the story is data-dependent or the user flags env/persona work.

## Order relative to test case gate

When both apply: pass **test case gate** first, then **test data setup
gate**, then implement.

## What counts as ready

A `test-data-setup.md` (or filled TRD-Testing data section) in the
history repo under `docs/plans/<JIRA-KEY>/`, with enough IDs and reset
steps to run the cases. No extra GitHub approval is required for that
file.

## Shared-environment reads

If the team has a read-only data tool configured, use it to confirm
row state. Do **not** run DML/DDL during gate verification unless the
user explicitly requests a seed or fix.

## Waivers

User must waive explicitly. Add **`test-data-gate:`** on changed
fixture or persona files.

## After the gate

One-line summary: data source path, env keys, verify result, sign-off type.
