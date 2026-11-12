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
- Repos: {{.ReposCSV}}

## Goals

-

## Non-goals

-

## Open questions

-
