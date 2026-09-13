# ADR-0001: Origin as Source of Truth for peradijakartabarat

**Status:** Accepted  
**Date:** 2026-09-13  
**Context:** First ADR for jarvis hub; establishes SoT pattern for manifest repos

---

## Decision

**peradijakartabarat** uses **Origin** (`fchr/peradijakartabarat`) as the **canonical source of truth** for Cloud Agent work and hub documentation references.

The GitHub sibling (`christianwijasa/peradijakartabarat`) exists but is **not** the primary remote for agent-driven development in this hub.

---

## Context

Two remotes exist for the same codebase:
- **Origin:** https://cursor.com/codebase/fchr/peradijakartabarat (Cursor-hosted)
- **GitHub:** https://github.com/christianwijasa/peradijakartabarat

Cloud Agents default to Origin repositories when available. The hub's `workstation/repos.manifest` must point to a single canonical remote per repo to avoid drift.

---

## Rationale

### Why Origin
1. **Cloud Agent native:** Origin repos integrate directly with Cursor Cloud Agents; authentication is automatic
2. **Collaboration context:** fchr namespace suggests this is the owner's personal/work org on Origin
3. **Single SoT simplifies sync:** Agents and local clones track one remote by default

### Why not GitHub
- GitHub sibling may be a mirror, backup, or legacy location
- No indication GitHub is the active development remote for this project
- Mixing remotes without explicit sync strategy causes divergence

### Trade-offs
- **Accepted:** If GitHub is the "real" public repo, manual sync from Origin → GitHub is required
- **Accepted:** Contributors without Origin access cannot clone via manifest (must use GitHub manually)
- **Mitigated by:** Manifest supports multiple remotes if sync strategy is later documented

---

## Consequences

### Immediate
- `workstation/repos.manifest` lists Origin URL as primary remote
- Future hub docs and scratch issues assume Origin is the working copy
- Agent-generated PRs target Origin by default

### Future
- If dual-remote sync is needed, document the strategy (e.g., GitHub Actions mirror, manual push)
- If GitHub becomes primary, revisit this ADR and update manifest

---

## Alternatives Considered

1. **GitHub as SoT** — Rejected: no evidence GitHub is more canonical; Origin is agent-native
2. **Dual remotes in manifest** — Deferred: adds complexity without clear sync process; revisit if needed
3. **No manifest entry** — Rejected: defeats hub's "second brain" purpose

---

**References:**
- [workstation/repos.manifest](../../workstation/repos.manifest)
- [docs/plans/peradijakartabarat/design.md](../plans/peradijakartabarat/design.md)
