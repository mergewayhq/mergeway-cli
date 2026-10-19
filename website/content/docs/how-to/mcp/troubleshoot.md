---
title: Troubleshoot MCP connections and stdout
description: Diagnose startup, stdio framing, missing tools, stale data, allow-list, and HTTP endpoint failures.
type: docs
weight: 40
bookToC: true
---

# Troubleshoot MCP connections and stdout

## The process exits immediately

Run the configured command directly with the same absolute paths. Diagnostics are on stderr:

```sh
/absolute/path/to/mergeway-mcp --root /absolute/path/to/repository --help
/absolute/path/to/mergeway-mcp --root /absolute/path/to/repository < /dev/null
```

Check that the root exists, contains `mergeway.yaml` or `mergeway.yml` directly, and every `--entity` uses exact case. Exit `2` indicates flags; exit `1` indicates repository or server startup/runtime.

## The client reports invalid JSON or loses the stdio connection

Remove wrappers that print banners, shell tracing, progress, or logging to stdout. Stdio reserves stdout for newline-delimited JSON-RPC. Send wrapper diagnostics to stderr. Do not run the server through a formatter or pipe that rewrites stdout.

Use an absolute executable and root path because GUI clients can have a smaller `PATH` and an unexpected working directory. Restart the client after changing configuration.

## Tools are missing or access is denied

The released catalog is exactly `entity_list`, `entity_show`, `object_list`, `object_get`, `repository_export`, and `files_list`. Check client-side tool filters as well as repeated server `--entity` flags. An unknown entity did not exist in the reloaded config; an entity-not-allowed error means it exists but was excluded at startup.

## Saved changes do not appear

Each request reloads files from disk; unsaved editor buffers are invisible. Save, run `mergeway-cli validate`, and call the tool again. Restart if you changed the entity set supplied through `--entity`.

## HTTP returns 404 or cannot connect

Match the exact endpoint: `/` by default or the cleaned `--http-base-path`. Confirm the host/port and that `--transport=http` was supplied. A refused connection usually means the process failed to bind or exited; inspect stderr. A path outside the configured mount intentionally returns `404`.
