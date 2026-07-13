---
title: "Install Mergeway CLI"
linkTitle: "Installation"
description: "Install mergeway-cli, mergeway-diff, mergeway-lsp, and mergeway-mcp from a release archive, Docker, Go, Nix, or source."
weight: 15
---

Use one of the installation paths below depending on whether you need only `mergeway-cli` or the wider toolset.

## Option 1: Download a Release Archive

GitHub releases publish platform assets for:

- Linux, macOS, and Windows
- `amd64` and `arm64`
- `mergeway-cli`, `mergeway-diff`, `mergeway-lsp`, and `mergeway-mcp`

Download the asset or assets for your platform from the latest release page, then install the extracted binaries you need:

```bash
install -m 0755 mergeway-cli /usr/local/bin/mergeway-cli
install -m 0755 mergeway-diff /usr/local/bin/mergeway-diff
install -m 0755 mergeway-lsp /usr/local/bin/mergeway-lsp
install -m 0755 mergeway-mcp /usr/local/bin/mergeway-mcp
```

Put `mergeway-cli` on `PATH` for terminal use, `mergeway-diff` on `PATH` for semantic snapshot workflows, `mergeway-lsp` on `PATH` for editor integration, and `mergeway-mcp` on `PATH` for MCP client integrations.

## Option 2: Docker

The published container image is useful for CLI-only workflows:

```bash
docker run --rm ghcr.io/mergewayhq/mergeway-cli version
```

To run against the workspace in your current directory:

```bash
docker run --rm \
  -v "$PWD:/work" \
  ghcr.io/mergewayhq/mergeway-cli validate
```

The container image does **not** include `mergeway-diff`, `mergeway-lsp`, or `mergeway-mcp`.

## Option 3: Go Install

```bash
go install github.com/mergewayhq/mergeway-cli@latest
```

This installs `mergeway-cli`. Install the other binaries separately if needed:

```bash
go install github.com/mergewayhq/mergeway-cli/cmd/mergeway-diff@latest
go install github.com/mergewayhq/mergeway-cli/cmd/mergeway-lsp@latest
go install github.com/mergewayhq/mergeway-cli/cmd/mergeway-mcp@latest
```

## Option 4: Nix

Install the CLI with Nix:

```bash
nix profile install github:mergewayhq/mergeway-cli
```

Or run it directly:

```bash
nix run github:mergewayhq/mergeway-cli -- help
```

The flake path is currently oriented around the CLI. Use a release archive or a local source build if you also need `mergeway-diff`, `mergeway-lsp`, or `mergeway-mcp`.

## Option 5: Build from Source

```bash
git clone https://github.com/mergewayhq/mergeway-cli.git
cd mergeway-cli
make build
./bin/mergeway-cli version
./bin/mergeway-diff --help
./bin/mergeway-lsp --log-stderr --log-level=debug
./bin/mergeway-mcp --help
```

`make build` produces all four binaries under `bin/`.

## Supported Platforms

- Release archives support Linux, macOS, and Windows.
- Each release includes `amd64` and `arm64` builds.
- The container image is Linux-only and CLI-only.

## Verify the Installation

```bash
mergeway-cli version
mergeway-diff --help
mergeway-lsp --log-stderr --log-level=debug
mergeway-mcp --help
```

If the language server starts successfully, it waits for LSP input on stdin. Stop it with `Ctrl+C`.

If the MCP server starts successfully in stdio mode, it waits for MCP traffic on stdin and reserves stdout for protocol messages. Stop it with `Ctrl+C`.

For manual startup and editor wiring, continue with [Run mergeway-lsp manually](language-server.md).
