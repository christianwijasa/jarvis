<!--
Adapted from obra/superpowers skills/test-driven-development/ (SKILL.md,
writing-good-tests.md) at commit bb92a77741419a4ab5f06e711a283343f1ada0c3
(v7.0.0), MIT, Copyright (c) 2025 Jesse Vincent.
Red-step "diff against a clean copy" line adapted from mattpocock/skills
skills/engineering/diagnosing-bugs/SKILL.md at 49dd158, MIT,
Copyright (c) 2026 Matt Pocock.
-->

# Test first (red, green, refactor)

Loaded from [development](../SKILL.md) step 6 and from
[debugging](../../debugging/SKILL.md) step 3. It does not change the
coverage gates; it changes the **order** you write things in.

## When it is required

| Change | Failing test first |
| ------ | ------------------ |
| **Tier A** state transitions, races, ordering, fencing, deadlines and idempotency | **Required** |
| Any bug fix in those areas, at any tier | **Required** |
| Other logic | Optional |
| Config, generated code, DTO/mapping boilerplate the coverage gate already excludes | No |

"Required" means the PR proves the test failed before the fix: a link
to a red CI run on a test-only commit, or pasted local output from
before the fix naming the command and the failing assertion. Commit
order alone is not accepted, and a failure for the wrong reason
(compile error, missing mock) does not count. The reviewer treats
missing proof as a blocker. For "Optional" rows test-first is author
guidance.

## The loop

1. **Red.** Write one small test for one behaviour, named after that behaviour. Before writing it, name the production change that should make it fail. Run it and confirm it fails **for that reason**. If you forced the red by changing production code or a fixture, `diff` (or `git diff`) against a clean copy to prove the change really took effect before you trust the failure.
2. **Green.** Write the least code that makes it pass. No extra options or branches the test does not need (ponytail).
3. **Refactor** only while green, without adding behaviour.
4. **Suite.** Run the suite for the changed area, not just your file, and report any failure by name.

## Tests that catch real breaks

- **Literal expectations.** Hand-derived values, never the output of the code under test or its helpers.
- **Behaviour, not mocks.** Assert on the state or the outgoing call your code makes. Asserting only that a mock was called is a test of the mock.
- **Fake at the boundary you own.** Repositories, HTTP clients, message publishers and workflow activities are faked; the logic in between is real.
- **No change detectors.** Do not pin constants or exact log text; test the behaviour that depends on them.
- **Test-only code stays in test sources**, not in production classes.

## Stack notes

Use the test helpers the target repo already uses.

### Go

Table tests; `-count=1 -race`. For workflow engines, rebuild event
order in the in-process test environment (signals, timers, activities).
When a change touches a running workflow type, add a replay test or
state why no open executions exist.

### Java / Spring

JUnit + Mockito, matching existing tests. Fence tests: fake a
conditional update returning 0 rows. One test per allowed and forbidden
state transition. Mock HTTP clients for success, 4xx, 5xx and timeout.

### Flutter / Dart

Unit tests for controllers and state, widget tests for screens, with
the API layer faked as the repo already does.

## What tests cannot show

SQL and data correctness are not proven by mocked repositories. Attach
an `EXPLAIN` from a shared env (read-only) or a QA run, and say which
in the PR evidence.
