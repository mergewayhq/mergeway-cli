---
title: "mergeway-lsp"
linkTitle: "mergeway-lsp"
description: "Run the Mergeway language server over stdio and configure its logging."
---

> **Synopsis:** Run the Mergeway Language Server Protocol server over stdio.

## Usage

```bash
mergeway-lsp [flags]
```

`mergeway-lsp` is a standalone binary. It does not read `mergeway-cli` global flags such as `--root` or `--format`.

## Flags

| Flag           | Description                                             |
| -------------- | ------------------------------------------------------- |
| `--log-file`   | Write server logs to the given file path.               |
| `--log-level`  | Set the log level: `debug`, `info`, `warn`, or `error`. |
| `--log-stderr` | Write server logs to stderr instead of discarding them. |

If you omit all logging flags, the server stays quiet and reserves stdout for JSON-RPC traffic.

## Examples

Start the server and wait for framed LSP messages on stdin:

```bash
mergeway-lsp
```

Enable debug logging on stderr while testing a client manually:

```bash
mergeway-lsp --log-stderr --log-level=debug
```

Write logs to a file instead of stderr:

```bash
mergeway-lsp --log-file /tmp/mergeway-lsp.log --log-level=debug
```

## Environment Variables

The server accepts environment-variable equivalents for its logging settings:

- `MERGEWAY_LSP_LOG_FILE=/tmp/mergeway-lsp.log`
- `MERGEWAY_LSP_LOG_LEVEL=debug`
- `MERGEWAY_LSP_LOG_STDERR=1`

Command-line flags take precedence over environment variables when both are set.

## Notes

- Keep stdout protocol-only. Human-readable output on stdout will break editor integration.
- In normal use, your editor launches `mergeway-lsp` for you rather than you starting it by hand.
- The server discovers Mergeway roots from the files the editor opens and uses the same validation core as `mergeway-cli validate`.

## Related Pages

- [Run mergeway-lsp manually](../getting-started/language-server.md) — startup basics and current capabilities.
- [Set up mergeway-lsp in VS Code and Neovim](../guides/setup-mergeway-lsp-editors.md) — editor wiring examples.
- [`mergeway-cli validate`](validate.md) — compare editor diagnostics with CLI validation output.
