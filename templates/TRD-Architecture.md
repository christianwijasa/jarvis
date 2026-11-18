---
issue: {{.Issue}}
{{- if .TrackerKind}}
tracker: {{.TrackerKind}}
{{- end}}
tier: {{.Tier}}
test_source: {{.TestSource}}
repos: [{{.ReposCSV}}]
type: architecture
status: draft
---

# TRD Architecture: {{.Issue}}

Status: draft

Scaffold only. Load `grill-with-docs` and `trd-pr-planning` before filling.
If `source` is tracker, fetch the work item via the host MCP named in the project.
If `source` is user, use the pasted requirement or `REQUIREMENT.md`.

## Summary

(one paragraph)

## Context

- Issue: {{.Issue}}
- Related Story: {{.Issue}}
- Related PR: (add the PR URL when it opens)
- Repos: {{.ReposCSV}}

## Goals

-

## Non-goals

-

## Design

Components, data flow, and interfaces. Decide every choice here; do not
leave design choices to the implementer.

## Alternatives considered

| Option | Why rejected |
| --- | --- |
| | |

## Risks and mitigations

| Risk | Mitigation |
| --- | --- |
| | |

## Open questions

Mark each one blocking or not blocking, with an owner.

-
