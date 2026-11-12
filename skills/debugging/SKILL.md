---
name: debugging
description: >-
  Root-cause debugging for repos listed in jarvis.yaml. Use on demand when a
  test fails, CI is red, a bug is reported, or behaviour is unexpected and
  the cause is not yet known, before proposing a fix. Steps: reproduce,
  trace the root cause, then fix with a failing test first. Not for planned
  feature work (use development).
---

<!--
Adapted from obra/superpowers skills/systematic-debugging/ at commit
bb92a77741419a4ab5f06e711a283343f1ada0c3 (v7.0.0), MIT,
Copyright (c) 2025 Jesse Vincent. See LICENSES/superpowers-MIT.txt.
-->

# Debugging

**No fix before you know the root cause.** A patch that makes the
symptom go away without an explanation is a guess, and guesses come
back as the next ticket.

Process and solution rules already live in always-on rules:

- [`karpathy-guidelines`](../../rules/karpathy-guidelines.md): state assumptions, surgical diff, verifiable goal.
- [`ponytail`](../../rules/ponytail.md): fix the shared function once at the root.

Implementation, coverage and PR gates stay in [`development`](../development/SKILL.md).
Never paste payloads or personal data into PRs, Jira or chat.

## 1. Reproduce

1. **Read the whole error.** Full stack trace, the first `Caused by:`, the failing assertion's expected vs actual, the CI step that failed. The first error in the log is usually the real one; later ones are fallout.
2. **Make it happen on demand** with the smallest command you can. Note the exact input, environment and commit. If it does not reproduce, gather more evidence; do not guess.
3. **Check what changed.** `git log -p <last-green>..<first-red> -- <area>`, dependency bumps, config or flag changes per env, and companion PRs that deployed in between.

| Stack | Narrowest reproduction |
| ----- | ---------------------- |
| Java / Spring | `mvn -Dtest='ClassTest#method' test` (or the repo’s equivalent). If private libraries do not resolve locally, quote the failing test from CI. |
| Go | `go test ./path/to/pkg -run 'TestName' -count=1 -race -v`. For workflow bugs, rebuild event order in an in-process test. |
| Flutter / Dart | `flutter test test/path/x_test.dart --plain-name '<test name>'`. Web vs mobile traps: [`flutter-conventions.md`](../development/references/flutter-conventions.md). |
| Data / SQL | Read-only `SELECT` against the shared env the team uses. Never mutate shared rows to "make it reproduce". |

## 2. Trace to the root cause

Start where the bad value or wrong state shows up and walk **backwards**
until you find where it was first produced. Fix there, not where it
surfaced.

- **One component at a time.** When the path crosses services, log or inspect what enters and leaves each hop once, then work only on the hop where good input turns into bad output.
- **Compare with a working case.** Find the closest path that works and list every difference. Read the reference implementation in full before copying its pattern.

Stack hints:

| Stack | Where to look |
| ----- | ------------- |
| Java / Spring | Request/trace id across logs; HTTP client boundary; conditional `UPDATE ... WHERE` row count; `@Transactional` boundaries; which `application-<env>.yaml` actually loaded. |
| Go / workflows | Event order, signal and timer timestamps, retries. Non-determinism means the workflow code changed under running histories: compare against history and versioning. Side effects and time belong in activities or the workflow clock. |
| Flutter | Controller/state at the screen boundary, the API response actually parsed, platform branch (web vs mobile), localization keys. |

Write down one hypothesis: "X is the root cause because Y". Test it with
the smallest possible change or probe, one variable at a time. If it is
wrong, form a new hypothesis; do not stack fixes.

Say so when you do not understand something. A confident wrong fix is not useful.

## 3. Fix

1. **Failing test first.** Reproduce the bug in an automated test and watch it fail for the right reason ([test-first](../development/references/test-first.md)). Required for Tier A state, race and ordering bugs.
2. **One fix, at the root.** No bundled refactors or "while I'm here" changes.
3. **Verify.** The new test passes, the rest of the suite for the changed area passes, and the original symptom is gone. Paste the evidence ([development § Evidence before done](../development/SKILL.md#evidence-before-done)).

**Three failed fixes = stop.** If each fix uncovers a new problem somewhere else, the design is likely wrong. Stop and raise it with the user or the TRD owner before attempting a fourth.

**Truly environmental?** (timing on a shared env, an external outage, a provider limit). Record what you checked, then add the handling the TRD allows. Most "no root cause" findings are an unfinished trace.

## Flaky tests

Replace sleeps with waiting on the actual condition. Keep a fixed delay
only when the delay itself is under test, and say why in a comment.
