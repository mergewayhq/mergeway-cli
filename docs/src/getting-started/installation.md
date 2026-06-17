---
title: "Install Mergeway CLI"
linkTitle: "Installation"
description: "Install mergeway-cli, mergeway-diff, and mergeway-lsp from a release archive, Docker, Go, Nix, or source."
weight: 15
---

Use one of the installation paths below depending on whether you need only `mergeway-cli` or the wider toolset.

## Option 1: Download a Release Archive

GitHub releases publish platform assets for:

- Linux, macOS, and Windows
- `amd64` and `arm64`
- `mergeway-cli`, `mergeway-diff`, and `mergeway-lsp`

Download the asset or assets for your platform from the latest release page, then install the extracted binaries you need:

```bash
install -m 0755 mergeway-cli /usr/local/bin/mergeway-cli
install -m 0755 mergeway-diff /usr/local/bin/mergeway-diff
install -m 0755 mergeway-lsp /usr/local/bin/mergeway-lsp
```

Put `mergeway-cli` on `PATH` for terminal use, `mergeway-diff` on `PATH` for semantic snapshot workflows, and `mergeway-lsp` on `PATH` for editor integration.

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

The container image does **not** include `mergeway-diff` or `mergeway-lsp`.

## Option 3: Go Install

```bash
go install github.com/mergewayhq/mergeway-cli@latest
```

This installs `mergeway-cli`. Install the other binaries separately if needed:

```bash
go install github.com/mergewayhq/mergeway-cli/cmd/mergeway-diff@latest
go install github.com/mergewayhq/mergeway-cli/cmd/mergeway-lsp@latest
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

The flake path is currently oriented around the CLI. Use a release archive or a local source build if you also need `mergeway-diff` or `mergeway-lsp`.

## Option 5: Build from Source

```bash
git clone https://github.com/mergewayhq/mergeway-cli.git
cd mergeway-cli
make build
./bin/mergeway-cli version
./bin/mergeway-diff --help
./bin/mergeway-lsp --log-stderr --log-level=debug
```

`make build` produces all three binaries under `bin/`.

## Supported Platforms

- Release archives support Linux, macOS, and Windows.
- Each release includes `amd64` and `arm64` builds.
- The container image is Linux-only and CLI-only.

## Verify the Installation

```bash
mergeway-cli version
mergeway-diff --help
mergeway-lsp --log-stderr --log-level=debug
```

If the language server starts successfully, it waits for LSP input on stdin. Stop it with `Ctrl+C`.

For manual startup and editor wiring, continue with [Run mergeway-lsp manually](language-server.md).
