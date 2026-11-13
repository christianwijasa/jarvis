# Cloud Logging MCP (optional)

Read-only Google Cloud Logging for a Jarvis **project that opts in**. Other
projects omit the key and agents must not call this server.

Catalog (editor-agnostic): [`mcp.json`](mcp.json). Cursor workspace copy:
[`.cursor/mcp.json`](../../.cursor/mcp.json).

## When to use it

Call `cloud-logging` only if `$JARVIS_HOME/projects/<name>/project.yaml`
lists `cloud-logging` under `mcp`. Check with `jarvis project show <name> --json`.

If `mcp` is missing or does not include `cloud-logging`, skip GCP logs. Do
not ask the user to connect Google Cloud.

`bfi-bravo-agency` and `bfi-lora-agency` opt in. A later project leaves `mcp`
unset unless that team needs logs.

## Server

Local package [`cloud-logging-mcp`](https://www.npmjs.com/package/cloud-logging-mcp)
pinned to **1.1.1**. Start: `npx -y cloud-logging-mcp@1.1.1`.

Tools: `queryLogs`, `getLogDetail`, `aggregateLogs`, `queryLogMetrics`,
`listProjects`.

Do not use the remote endpoint `https://logging.googleapis.com/mcp`.

Before changing the pin, review the npm package. It uses the user's ADC
token.

## Auth and quota

Auth is per user. Run `gcloud auth application-default login`. The server
has no credentials in git. Access follows IAM on the signed-in account.

`GOOGLE_CLOUD_PROJECT` is the **quota** project (`x-goog-user-project`).
Set it in the environment. Do not put a product GCP id in the hub catalog.
Do not use a production project as the quota project unless the account
can bill that project.

Need `roles/logging.viewer` on each log project and
`serviceusage.services.use` on the quota project.

## `projectId` vs `resourceNames`

| Field                                | Meaning                                |
| ------------------------------------ | -------------------------------------- |
| `projectId` / `GOOGLE_CLOUD_PROJECT` | Auth and quota anchor                  |
| `resourceNames`                      | Projects that **hold** the log entries |

Read `logging.quota_project` and `logging.resources` from the **named**
project yaml. Those values are per project, not hub defaults.

Always pass a time window. Prefer `jsonPayload.correlationId` (or the
field the service actually logs) over a broad text search.

`aggregateLogs` counts one page only. Do not report that number as a
total.

If MCP is not connected, query Logging REST `entries:list` with the ADC
token and the same quota header.

## Cursor

Merge [`mcp.json`](mcp.json) into the Jarvis workspace MCP (already mirrored
in `.cursor/mcp.json`), then reload. The process may sit unused when the
active project has no `mcp` key. Agents still must not call it then.
