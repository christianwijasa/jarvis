# Domain docs

How agent skills use domain documentation in **personal-hub** and the repos listed in the manifest.

## Before exploring, read these

- **`CONTEXT.md`** at the hub root — shared glossary (vocabulary only, not implementation detail). Created lazily by `/grill-with-docs` or domain-modeling.
- **`docs/adr/`** — durable, hard-to-reverse decisions that span efforts.
- **`docs/plans/<slug>/`** — design notes for one effort (e.g. `design.md` after grilling).

If `CONTEXT.md` or `docs/adr/` do not exist yet, proceed silently. The domain-modeling skill creates them when terms or decisions resolve during `/grill-with-docs`.

## File structure

Single-context layout (this hub):

```
personal-hub/
├── CONTEXT.md              ← glossary
├── docs/
│   ├── adr/                ← 0001-slug.md, …
│   ├── agents/             ← agent config (this file, issue-tracker.md)
│   └── plans/<slug>/       ← per-effort design notes
└── .agents/skills/
```

Code repos under `~/projects/<name>/` hold implementation; the hub holds shared language and plans. Do not duplicate the full glossary inside plan bodies — link to `CONTEXT.md` and cite `docs/adr/` when a decision is recorded there.

## Glossary vs ADR vs plan

| Artifact               | Purpose                                      | Where                            |
|------------------------|----------------------------------------------|----------------------------------|
| **CONTEXT.md term**    | Canonical word for a concept                 | Root glossary                    |
| **docs/adr/NNNN-*.md** | Hard-to-reverse decision with trade-offs     | Persistent ADR file              |
| **docs/plans/.../design.md** | Decisions for *this effort* for implementers | Plan folder                      |

During `/grill-with-docs`: resolve vocabulary in `CONTEXT.md` first; if a decision passes all three ADR gates (see domain-modeling skill), write `docs/adr/` and reference it from the plan instead of duplicating prose.

## Vocabulary

When naming concepts in plans, scratch issues, or APIs, use terms as defined in `CONTEXT.md`. If the glossary defines _Avoid_ synonyms, do not drift to them.

If a concept is missing from the glossary, resolve it in a grill round before finalizing the plan.

## ADR conflicts

If plan content contradicts an existing `docs/adr/` entry, surface it explicitly:

> _Contradicts ADR-0007 (…), but worth reopening because…_

## Hub-local grill skill

When working in this repo, use [`.agents/skills/grill-with-docs/SKILL.md`](../../.agents/skills/grill-with-docs/SKILL.md) (not TRD/Jira). Global primitives: **grilling** and **domain-modeling** under `~/.agents/skills/`.
