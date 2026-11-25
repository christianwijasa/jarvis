# Temporal MCP (optional)

Read and write Temporal workflows for a Jarvis **project that opts in**.
Other projects omit the key and agents must not call this server.

Catalog (editor-agnostic): [`mcp.json`](mcp.json). Cursor workspace copy:
[`.cursor/mcp.json`](../../.cursor/mcp.json).

## When to use it

Call `temporal` only if `$JARVIS_HOME/projects/<name>/project.yaml`
lists `temporal` under `mcp`. Check with `jarvis project show <name> --json`.

If `mcp` is missing or does not include `temporal`, skip Temporal MCP.
Do not ask the user to connect a cluster.

A project that needs workflow tools lists `temporal`. Other projects leave
`mcp` without that id.

## Server

PyPI package [`temporal-mcp-server`](https://pypi.org/project/temporal-mcp-server/)
pinned to **1.12.0**. Source:
[GethosTheWalrus/temporal-mcp](https://github.com/GethosTheWalrus/temporal-mcp).

Start: `uvx temporal-mcp-server==1.12.0`.

Need `uvx` on PATH. Connection values come from a local env file.
Do not put host, namespace, or API key in the hub catalog.

| Env | Meaning | Default in the server |
| --- | --- | --- |
| `TEMPORAL_HOST` | Host and port (`host:port`) | `localhost:7233` |
| `TEMPORAL_NAMESPACE` | Default namespace | `default` |
| `TEMPORAL_API_KEY` | Temporal Cloud API key | unset |
| `TEMPORAL_TLS_ENABLED` | Force TLS | auto-detect |
| `TEMPORAL_ALLOWED_NAMESPACES` | Comma list, or `*` | default namespace only |

Some apps split host and port (`TEMPORAL_HOST` plus `TEMPORAL_PORT`).
This MCP wants one `TEMPORAL_HOST` value with the port included.

When `TEMPORAL_API_KEY` is set, the server enables TLS.

Payload encryption is a codec concern. Without
`TEMPORAL_CODEC_ENDPOINT`, history payloads stay encoded.

## Tools

Prefer read tools:

- `list_workflows`
- `describe_workflow`
- `get_workflow_history`
- `get_workflow_event`
- `get_workflow_result`
- `query_workflow`
- `list_schedules`
- `describe_schedule`

Do not call write or batch tools unless the user names the workflow and
asks to change it:

- `start_workflow`
- `signal_workflow`
- `cancel_workflow`
- `terminate_workflow`
- `continue_as_new`
- `batch_signal`
- `batch_cancel`
- `batch_terminate`
- `create_schedule`
- `pause_schedule`
- `unpause_schedule`
- `delete_schedule`
- `trigger_schedule`

## Cursor

Merge [`mcp.json`](mcp.json) into the Jarvis workspace MCP (already mirrored
in `.cursor/mcp.json`), then reload. The process may sit unused when the
active project has no `temporal` id. Agents still must not call it then.

Create `~/.cursor/temporal.env` on your laptop (not in git). Cursor loads
it through `envFile`. Then reload MCP.

Example:

```
TEMPORAL_HOST=region.aws.api.temporal.io:7233
TEMPORAL_NAMESPACE=default
TEMPORAL_API_KEY=
TEMPORAL_TLS_ENABLED=true
TEMPORAL_ALLOWED_NAMESPACES=default
```
