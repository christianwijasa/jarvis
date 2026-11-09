<!--
Adapted from DietrichGebert/ponytail skills/ponytail-review/SKILL.md
at 1b1a0c51629436cf7831017648f7f3f0827196e4 (unchanged at 09c434a for the
taken lines), MIT, Copyright (c) 2026 DietrichGebert.
Hub UPSTREAM.md skipped the Ponytail 5 full-quality-review rewrite.
-->

# Ponytail review (over-engineering only)

Optional pass after or beside full [code-review](../SKILL.md). Same
repo scope as [development](../../development/SKILL.md).

Review diffs for unnecessary complexity. One line per finding: location,
what to cut, what replaces it. The diff's best outcome is getting shorter.

**Triggers:** "review for over-engineering", "what can we delete", "is
this over-engineered", "simplify review", `/ponytail-review`.

## Format

`<N>. L<line>: <tag> <what>. <replacement>.`, or `<N>. <file>:L<line>: ...`
for multi-file diffs. Number findings `1.`, `2.`, ... across the whole
report, so the user can say "fix 2 and 5".

Tags:

- `delete:` dead code, unused flexibility, speculative feature. Replacement: nothing.
- `stdlib:` hand-rolled thing the standard library ships. Name the function.
- `native:` dependency or code doing what the platform already does. Name the feature.
- `reuse:` equivalent helper, util, or pattern already in this repo. Name the path.
- `yagni:` abstraction with one implementation, config nobody sets, layer with one caller.
- `shrink:` same logic, fewer lines. Show the shorter form.

## Examples

❌ "This EmailValidator class might be more complex than necessary, have you
considered whether all these validation rules are needed at this stage?"

✅ `1. L12-38: stdlib: 27-line validator class. "@" in email, 1 line, real validation is the confirmation mail.`

✅ `2. L4: native: moment.js imported for one format call. Intl.DateTimeFormat, 0 deps.`

✅ `3. L18-29: reuse: slugify helper duplicates src/lib/slug.ts slugify. Delete it, import the existing one.`

✅ `4. repo.py:L88: yagni: AbstractRepository with one implementation. Inline it until a second one exists.`

✅ `5. L52-71: delete: retry wrapper around an idempotent local call. Nothing replaces it.`

✅ `6. L30-44: shrink: manual loop builds dict. dict(zip(keys, values)), 1 line.`

## Scoring

End with the only metric that matters: `net: -<N> lines possible.`

If there is nothing to cut, say `Lean already. Ship.` and stop.

## Boundaries

Scope: over-engineering and complexity only. Correctness bugs, security
holes, and performance are out of scope. Route them to Standards/Spec in
[code-review](../SKILL.md). A single smoke test or `assert`-based
self-check is the ponytail minimum, not bloat; never flag it for deletion.
Does not apply the fixes, only lists them.
