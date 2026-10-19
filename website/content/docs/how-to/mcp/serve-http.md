---
title: Serve MCP over HTTP
description: Start and connect to a loopback streamable HTTP endpoint at the default or a custom mount.
type: docs
weight: 30
bookToC: true
---

# Serve MCP over HTTP

Prefer stdio for one local client. Use HTTP when a compatible client needs a separately managed streamable endpoint.

Start the loopback-only default:

```sh
mergeway-mcp --root /absolute/path/to/repository --transport=http
```

Configure the client endpoint as `http://127.0.0.1:8080/`. To avoid a port conflict or mount behind a local proxy:

```sh
mergeway-mcp --root /absolute/path/to/repository \
  --transport=http \
  --http-listen 127.0.0.1:9090 \
  --http-base-path /mcp
```

The endpoint is then `http://127.0.0.1:9090/mcp`; `/` and `/mcp/anything` return `404`. Configure `--entity` at server startup when the client should see less than the whole repository.

Stop the foreground process with your terminal interrupt. The server attempts a five-second graceful shutdown. It has no built-in authentication or TLS, so keep it on loopback unless a trusted proxy and network policy supply those controls. Do not treat an entity allow-list as network authorization.
