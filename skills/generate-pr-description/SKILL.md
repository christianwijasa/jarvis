---
name: generate-pr-description
description: >-
  Generate brief, clear, and high-signal pull request descriptions from
  git diff and commit history. Produces copy-paste-ready markdown focused
  on brevity, with extra detail reserved only for critical parts. Use
  before opening PRs in the history repo or any repo listed in atlas.yaml.
---

# Generate PR Description

Generate brief, clear, and high-signal pull request descriptions
formatted in clean bullet points, ready to copy and paste.

## Core Principles

- **Brief & Clear**: Keep the description short, crisp, and readable in under 30 seconds. Avoid fluff or narrating mechanical edits.
- **High Signal-to-Noise**: Focus on _what_ changed and _why_ at a functional level.
- **Selective Depth**: Only attach extra details to **critical parts** (breaking changes, data/schema migrations, security/auth impacts, subtle edge-case behavior).

## When to Use

- When the user asks to generate, draft, or write a PR description.
- Before opening a PR in a repo listed in `atlas.yaml` `repos`, or in the history repo.
- After finishing an implementation branch.

## Workflow

### 1. Identify the Base Reference (Fixed Point)

- Determine the target/base branch (typically `origin/main` or `origin/master`).
- If on a feature branch (e.g. `feat/TEAM-123`), find the merge base:
  ```bash
  git merge-base HEAD origin/main || git merge-base HEAD origin/master
  ```
- Inspect status and diff:
  ```bash
  git status
  git diff --stat <base>...HEAD
  git log <base>..HEAD --oneline
  ```

### 2. Check for Linked Spec / TRD

- If the branch name or commits reference a Jira key:
  - Check `<history.path>/docs/plans/<JIRA-KEY>/` (from `atlas.yaml`).
  - Read YAML frontmatter for **`tier`** and **`test_source`**.
- PR body should include a link to that TRD using the history repo host.

### 3. Extract & Condense Core Changes

- Filter out noise.
- Group into 2–5 punchy bullets.
- Isolate critical items (breaking API contract, schema change, config prerequisite).

### 4. Output Format Guidelines

Always provide the output in a single copy-paste-ready `markdown` code block.

```markdown
## Summary

<1–2 concise sentences explaining what changed and why.>

## Key Changes

- **<Component/Area>**: <Concise 1-line summary of change.>

<!-- Only add extra detail / sub-bullets if critical or breaking -->

- **⚠️ <Critical / Breaking Change>**: <Concise summary>
  - _Impact/Migration_: <Brief explanation or required action.>

## Plan document

- [TRD …](<history repo URL>/blob/<ref>/docs/plans/<JIRA-KEY>/…)

## Story tier

- **Tier:** <A|B|C from plan frontmatter or user>
- **Test source:** <atlas:TRD-Testing | jira:KEY | user-prompt | waived>

## Test plan

- [ ] …

## Verification

- <Concise bullet on test or validation performed>

## Related Issue

- Closes [<JIRA-KEY>](<jira.site>/browse/<JIRA-KEY>)
```

Build Jira links from `atlas.yaml` `jira.site`. Do not hard-code a
company Jira host.

### 5. Open PR

After the user approves the description and the branch is ready:

1. Create the PR in the **product repo** (not this CLI repo), unless the change is Atlas-CLI-only.
2. Run local checks, push **once** for this review round, then add the team’s review label on that **final** head if they use one.
