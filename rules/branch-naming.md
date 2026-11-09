---
description: >-
  Git branch names for Jira-backed work — feat/fix prefix plus Jira key.
---

# Branch naming (Jira-backed work)

When **creating** a git branch for work tracked in Jira — in any repo
listed in `atlas.yaml` `repos`, or before TRD planning — use this naming:

| Purpose                         | Required prefix   | Example          |
| ------------------------------- | ----------------- | ---------------- |
| New feature (story, task, epic) | `feat/<JIRA-KEY>` | `feat/TEAM-123`  |
| Bug fix                         | `fix/<JIRA-KEY>`  | `fix/TEAM-99`    |

## Rules

1. **`<JIRA-KEY>`** — exact Atlassian issue key (project from `atlas.yaml`
   `jira.project` unless the user names another). Use the **implementation**
   ticket for the branch when it differs from the TRD folder key; record
   both in TRD frontmatter.
2. **feat vs fix** — from Jira issue type (Bug → `fix/`; Story, Task, Epic,
   sub-task for new behavior → `feat/`) or explicit user intent when type
   is ambiguous.
3. Branch name must **begin with** the chosen prefix and key. An optional
   lowercase slug after the key is allowed (`feat/TEAM-123-checkout`).
4. Do **not** use `feature/`, bare keys without prefix, or ad-hoc prefixes
   for **new** branches.

Do not rename existing remote branches unless the user asks.

Commits: [`commit-message.md`](commit-message.md).
