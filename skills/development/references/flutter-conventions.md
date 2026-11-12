# Flutter conventions

Scope: Flutter apps listed in `jarvis.yaml` `repos`. Complements the
client coverage gate in [`SKILL.md`](../SKILL.md).

## Code style

1. **Never hardcode UI text.** Every user-facing string goes through the app’s localization files.
2. **Reuse existing enums/utilities instead of re-deriving logic.** Grep before writing a mapping helper.
3. **Don't create small one-off widgets as their own `StatelessWidget` class.** Prefer widget-builder methods matching the enclosing screen.
4. **Don't assume a pattern — read an existing, similar file first.** Mirror an existing controller+screen pair.

## Before commit

Every Flutter commit — not just the final PR — does all three, in order:

1. **Add/extend tests** for the changed production logic.
2. **`dart format`** the changed files (`dart format --set-exit-if-changed .`).
3. **`dart analyze`** (or `flutter analyze`) — zero errors/warnings on touched files.

## Known gotchas

These are engine/platform traps, not product names:

- **Flutter web: `dart:io` `Platform.is*` throws.** Guard with `!kIsWeb` (`import 'package:flutter/foundation.dart'`).
- **Android native `libflutter.so` mutex crashes in Impeller/Vulkan** are often an engine bug (image decode + GPU upload). Get the full stack first; if it shows Impeller/Vulkan frames, disable Impeller in the Android manifest rather than chasing app-level causes.
- **Store reviews reject unused `READ_MEDIA_*` permissions injected by plugins.** Strip unused permissions with `tools:node="remove"` after confirming the plugin only does incidental file access.
