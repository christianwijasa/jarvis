---
description: >-
  Optional Temporal MCP. Use only when the named Jarvis
  project.yaml lists temporal under mcp.
---

# Temporal MCP

See [`docs/agents/mcp/temporal.md`](../docs/agents/mcp/temporal.md).

Call `temporal` only if the named project's `project.yaml` lists it
under `mcp`. Other projects omit the id. Do not treat this server as
required for every Jarvis project.

Prefer read tools. Do not start, signal, cancel, or terminate workflows
unless the user names the workflow and asks to change it.
