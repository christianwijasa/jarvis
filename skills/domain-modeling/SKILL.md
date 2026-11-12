---
name: domain-modeling
description: >-
  Build and sharpen the project domain model. Use when resolving terminology
  during grill-with-docs or TRD planning — write the glossary and docs/adr/
  as terms and decisions crystallise.
---

<!--
Adapted from mattpocock/skills skills/engineering/domain-modeling/SKILL.md
at 49dd158d1076134a641b33efb035946536778336, MIT, Copyright (c) 2026 Matt Pocock.
Hub UPSTREAM.md skipped the upstream CONTEXT → GLOSSARY rename on purpose.
-->

# Domain Modeling

Actively build and sharpen the project's domain model as you design. Challenge terms, invent edge-case scenarios, and write the glossary and decisions down the moment they crystallise.

Glossary and ADRs live in the history repo (`jarvis.yaml` `history.path`), not in this CLI repo.

## File structure

```
<history.path>/
├── docs/
│   ├── CONTEXT.md
│   ├── adr/
│   └── plans/<JIRA-KEY>/
```

Create files lazily: only when you have something to write.

## During the session

### Challenge against the glossary

When the user uses a term that conflicts with the glossary, call it out immediately.

### Sharpen fuzzy language

When the user uses vague or overloaded terms, propose a precise canonical term.

### Discuss concrete scenarios

Stress-test domain relationships with specific scenarios. Invent edge cases that force precise boundaries.

### Cross-reference with code

When the user states how something works, check graphify/code. Surface contradictions.

### Update the glossary inline

When a term is resolved, update the glossary right there. Don't batch. Use [CONTEXT-FORMAT.md](./CONTEXT-FORMAT.md).

The glossary is not implementation detail and not spec prose.

### Offer ADRs sparingly

Only offer `docs/adr/NNNN-slug.md` when all three are true:

1. **Hard to reverse**
2. **Surprising without context**
3. **Real trade-off** with genuine alternatives

Use [ADR-FORMAT.md](./ADR-FORMAT.md). Reference the ADR from the TRD ADR block; don't duplicate full text in the TRD.
