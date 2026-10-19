---
title: MCP transports
description: Exact stdio and streamable HTTP behavior, endpoints, combinations, and operational boundaries.
type: docs
weight: 10
bookToC: true
---

# MCP transports

## Stdio

Stdio is the default and is intended for a client that launches one server subprocess:

```sh
mergeway-mcp --root /absolute/path/to/repository
```

The transport is newline-delimited JSON-RPC over stdin and stdout. Closing stdin ends the session. Protocol responses are the only stdout content; diagnostics use stderr. `--http-listen` and `--http-base-path` are rejected when their flags are explicitly supplied with stdio, even if their values match the defaults.

## Streamable HTTP

```sh
mergeway-mcp --root /absolute/path/to/repository --transport=http
```

The default endpoint is `http://127.0.0.1:8080/`. The server uses stateless streamable HTTP with JSON responses. Requests must target the exact normalized mount: unrelated paths return HTTP `404`. A trailing slash in the configured path is cleaned, so `/mcp/` mounts at `/mcp`.

Choose both address and path explicitly when another local service uses the default:

```sh
mergeway-mcp --root /absolute/path/to/repository \
  --transport=http \
  --http-listen 127.0.0.1:9090 \
  --http-base-path /mcp
```

The endpoint is `http://127.0.0.1:9090/mcp`. The base path must begin with `/` and cannot be empty. The listen value must resolve as a TCP address. Listen failures are runtime errors and exit `1`.

## Network boundary

The HTTP server provides no Mergeway authentication, authorization, TLS, rate limiting, or multi-tenant isolation. The safe default binds loopback. Do not bind a public or shared interface without a trusted authenticated TLS reverse proxy and network controls. An `--entity` allow-list limits Mergeway tool results; it does not secure the HTTP endpoint.

Use the [HTTP how-to](../../../how-to/mcp/serve-http/) for an operational procedure or the [troubleshooting guide](../../../how-to/mcp/troubleshoot/) for connection failures.
