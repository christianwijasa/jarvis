---
issue: {{.Issue}}
{{- if .TrackerKind}}
tracker: {{.TrackerKind}}
{{- end}}
tier: {{.Tier}}
test_source: {{.TestSource}}
repos: [{{.ReposCSV}}]
type: testing
status: draft
---

# TRD Testing: {{.Issue}}

Status: draft

Scaffold only. Fill with `test-case-design` for tier A.

## Summary

How to check this change locally.

## Cases

-

## Seed data

-
