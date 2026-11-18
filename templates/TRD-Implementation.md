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
- Related Story: {{.Issue}}
- Related PR: (add the PR URL when it opens)
- Architecture: (link the architecture TRD, if any)
- Repos: {{.ReposCSV}}

## Goals

-

## Non-goals

-

## Files to change

Every file to create, modify, or delete, repo-relative. Include tests,
config, schema, and docs. No design choices are left to the implementer:
data shapes, API paths, error handling, and defaults are decided here.

| File | Create / modify / delete | What changes and why |
| --- | --- | --- |
| | | |

## Phases

Each phase leaves the code building and its tests passing.

1.

## Verification

Test cases (happy path and edge cases) and how to check the change locally.

-

## Open questions

Mark each one blocking or not blocking, with an owner.

-
