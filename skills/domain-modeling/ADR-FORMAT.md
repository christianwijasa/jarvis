# ADR Format

ADRs live in `<history.path>/docs/adr/` and use sequential numbering:
`0001-slug.md`, `0002-slug.md`, etc.

Create `docs/adr/` lazily when the first ADR is needed.

## Template

```md
# {Short title of the decision}

{1-3 sentences: context, decision, and why.}
```

## Optional sections

Only when they add value: Status frontmatter, Considered Options, Consequences.

## Numbering

Scan `docs/adr/` for the highest existing number and increment by one.

## When to offer an ADR

All three must be true:

1. **Hard to reverse**
2. **Surprising without context**
3. **Real trade-off**

TRD ADR blocks are for _this change_. Persistent cross-ticket decisions
belong in `docs/adr/` and are linked from the TRD.
