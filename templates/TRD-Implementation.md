---
issue: {{.Issue}}
{{- if .TrackerKind}}
tracker: {{.TrackerKind}}
{{- end}}
tier: {{.Tier}}
test_source: {{.TestSource}}
repos: [{{.ReposCSV}}]
type: implementation
status: draft
---

# TRD Implementation: {{.Issue}}

Status: draft

Scaffold only. Load `grill-with-docs` and `trd-pr-planning` before filling.

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
