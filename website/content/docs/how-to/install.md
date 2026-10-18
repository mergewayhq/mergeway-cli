---
title: Install Mergeway tools
description: Install the released Mergeway CLI, Diff, MCP, and LSP binaries from an archive, or choose a clearly scoped alternative.
type: docs
mode: how-to
weight: 10
bookToC: true
aliases:
  - /getting-started/installation.html
  - /getting-started/installation/
---

# Install Mergeway tools

GitHub release archives are the default installation route. One platform archive contains `mergeway-cli`, `mergeway-diff`, `mergeway-mcp`, and `mergeway-lsp` built from the same release.

## Install a release archive

1. Open the [latest Mergeway release](https://github.com/mergewayhq/mergeway-cli/releases/latest).
2. Download the archive matching your operating system and architecture:

   - Linux: `mergeway-cli_Linux_x86_64.tar.gz` (`amd64`) or `mergeway-cli_Linux_arm64.tar.gz` (`arm64`).
   - macOS: `mergeway-cli_Darwin_x86_64.tar.gz` (`amd64`) or `mergeway-cli_Darwin_arm64.tar.gz` (`arm64`).
   - Windows: `mergeway-cli_Windows_x86_64.zip` (`amd64`) or `mergeway-cli_Windows_arm64.zip` (`arm64`).

3. Extract the archive and put the binaries you need on `PATH`. On Linux or macOS, from the extracted directory:

   ```sh
   install -d "$HOME/.local/bin"
   install -m 0755 mergeway-cli mergeway-diff mergeway-mcp mergeway-lsp "$HOME/.local/bin/"
   ```

   Add `$HOME/.local/bin` to `PATH` if your shell does not already include it. On Windows, move the `.exe` files to a directory listed in `PATH`.

4. Verify the command-line tools:

   ```sh
   mergeway-cli version
   mergeway-diff --help
   mergeway-mcp --help
   mergeway-lsp --help
   ```

`mergeway-lsp` normally runs as an editor subprocess over standard input and output. `mergeway-mcp` normally runs as an MCP client subprocess or an HTTP server; neither needs to remain running for this installation check.

## Choose an alternative

| Method | Provides | When to use it |
| --- | --- | --- |
| Release archive | CLI, Diff, MCP, and LSP | Recommended for tutorials and matched suite versions. |
| Docker | CLI only | Run validation or other CLI operations in a Linux container. |
| Go install | One requested binary at a time | Install from Go module source with a local Go toolchain. |
| Nix flake | CLI only | Install or run the CLI from Nix. |
| Source build | CLI, Diff, MCP, and LSP | Develop or test this repository. |

### Docker

The published container contains `mergeway-cli` only. Mount a repository at `/work`, which is the image working directory:

```sh
docker run --rm -v "$PWD:/work" ghcr.io/mergewayhq/mergeway-cli validate
```

### Go

Install the CLI, then request any other binaries you need separately:

```sh
go install github.com/mergewayhq/mergeway-cli@latest
go install github.com/mergewayhq/mergeway-cli/cmd/mergeway-diff@latest
go install github.com/mergewayhq/mergeway-cli/cmd/mergeway-mcp@latest
go install github.com/mergewayhq/mergeway-cli/cmd/mergeway-lsp@latest
```

For reproducible automation, replace `@latest` with a release tag.

### Nix

The flake exposes the CLI, not the other three binaries:

```sh
nix profile install github:mergewayhq/mergeway-cli
nix run github:mergewayhq/mergeway-cli -- version
```

### Source

A source build requires Git, Go, and Make:

```sh
git clone https://github.com/mergewayhq/mergeway-cli.git
cd mergeway-cli
make build
```

The four binaries are written to `bin/`. This is a contributor workflow, not the tutorial installation path.

## Next steps

- Follow [Build your first Mergeway repository](../../tutorials/first-repository/).
- Learn [how the components fit together](../../components/).
- Read [the repository model](../../explanation/repository-model/) before adapting an existing dataset.
