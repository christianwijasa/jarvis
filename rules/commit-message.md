---
description: >-
  Git commit subject format for Jira-backed work — feat/fix(scope), Jira
  key in brackets, then summary.
---

# Commit message (Jira-backed work)

When **creating a commit** for work tracked in Jira — in any repo listed
in `atlas.yaml` `repos` — use this **subject line** format:

```text
feat({topic}): [{jira-key}] {commit-message}
fix({topic}): [{jira-key}] {commit-message}
```

| Part               | Meaning                                                                                           |
| ------------------ | ------------------------------------------------------------------------------------------------- |
| `feat` / `fix`     | Same as branch: feature work → `feat`; bug fix → `fix` ([`branch-naming.md`](branch-naming.md)) |
| `{topic}`          | Short scope: repo slice or domain (lowercase, hyphenated)                                         |
| `[{jira-key}]`     | Exact Jira key, brackets required, for example `[TEAM-123]`                                       |
| `{commit-message}` | Imperative summary; lowercase start unless proper noun                                            |

## Examples

```text
feat(checkout): [TEAM-123] publish order moved after customer update
fix(api): [TEAM-99] return bulk error code from agent API
```

## Rules

1. **One line subject** unless the user asks for a body; if a body is needed, keep the subject in this format and put detail below a blank line.
2. **`{topic}`** should match the change area, not the full branch slug.
3. **`[{jira-key}]`** uses the same implementation ticket as the branch when they align.
4. Do **not** omit the prefix, topic, or bracketed key for Jira-backed commits.

See [development](../skills/development/SKILL.md).
