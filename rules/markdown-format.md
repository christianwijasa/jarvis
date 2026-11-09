---
description: >-
  Keep Markdown formatted consistently. Teams may use Prettier or another
  formatter; match the repo you are editing.
---

# Markdown formatting

Format Markdown the same way the target repo already does.

## After you edit Markdown

1. Format changed files with that repo’s documented command (for example
   `npx prettier --write -- '<path>'` when Prettier is configured).
2. Do not hand-tweak spacing the formatter normalizes (trailing spaces,
   final newline, list wrapping).

## Before `git push`

Run the repo’s format check and commit any formatting changes. If git
hooks format on push, do not fight them.

## Scope

Do not force-format generated output, vendored trees, or local env files.
