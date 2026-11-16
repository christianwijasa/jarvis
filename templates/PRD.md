---
issue: {{.Issue}}
{{- if .TrackerKind}}
tracker: {{.TrackerKind}}
{{- end}}
type: prd
status: draft
owner:
---

# PRD: {{.Issue}}

Status: draft (draft, in-review, or signed-off)

Scaffold only. Load the `prd-planning` and `grilling` skills before filling.
If the work item is in a tracker, fetch it through the host MCP.
Remove this note and any empty optional section before review.

## 1. Problem

(What problem do users have today? What is the evidence?)

## 2. Users and personas

(Who has the problem? Who uses the feature? Who is affected but does not use it?)

## 3. Goals and success metrics

| Metric | Baseline | Target | How measured |
| --- | --- | --- | --- |
|  |  |  |  |

## 4. Scope

### In scope

-

### Out of scope

-

## 5. User flows

1. (numbered steps for the main flow; add alternative flows below)

## 6. Business rules

- BR-1: (one testable rule)

## 7. Acceptance criteria

- AC-1 (BR-1): Given ..., when ..., then ...

## 8. Edge cases and error states

-

## 9. Dependencies

(Teams, systems, data, legal, vendors.)

## 10. Rollout

(Phases, feature flags, target users, communication, rollback trigger.)

## 11. Decisions

| Date | Decision | Decided by |
| --- | --- | --- |
|  |  |  |

## 12. Open questions

| Question | Owner | Due |
| --- | --- | --- |
|  |  |  |

## 13. Engineering checks

Questions about how the current system behaves. The product owner does not answer these.
An engineer answers each one, records the answer in the TRD, and copies it here with a link.

| ID | Question | Why it matters | Status | Answer | Answered from (commit) | TRD link |
| --- | --- | --- | --- | --- | --- | --- |
| EC-1 |  |  | open |  |  |  |

## 14. Glossary

(Link to the shared glossary. List new terms only.)
{{- if .SourceNotes}}

## Source notes

{{.SourceNotes}}
{{- end}}
