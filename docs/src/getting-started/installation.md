---
title: "Install Mergeway CLI"
linkTitle: "Installation"
description: "Install the Mergeway CLI using a release download, Docker, Go install, or Nix."
weight: 10
---

Pick the method that fits your setup. You can install a local `mergeway-cli` binary or run the public container image directly.

## Option 1 – Download a Release (macOS, Linux)

Use the pre‑built archives published on GitHub releases. The example below downloads version `v0.3.0` for your platform and moves the binary into `/usr/local/bin`:

```bash
curl -L https://github.com/mergewayhq/mergeway-cli/releases/download/v0.3.0/mergeway-cli-\
  $(uname | tr '[:upper:]' '[:lower:]')-amd64.tar.gz | tar -xz
sudo mv mergeway-cli /usr/local/bin/
````

Check the published SHA‑256 checksum before moving the binary if you operate in a locked‑down environment.

## Option 2 – Docker

Use the public GitHub Container Registry image to run the CLI without installing the binary locally:

```bash
docker run ghcr.io/mergewayhq/mergeway-cli version
```

## Option 3 – Go Install (for contributors)

If you have Go installed you can build the CLI directly from the repository using `go install`:

```bash
go install github.com/mergewayhq/mergeway-cli@latest
```

This drops the binary in `$GOPATH/bin` (often `~/go/bin`). Prefer tagged versions in production.

## Option 4 – Nix / Flake Install

The repository defines a [Nix flake](https://nixos.wiki/wiki/Flakes) that packages the CLI. Using the Nix package manager you can install, run or develop the CLI without managing Go toolchains manually:

### Install into your Nix profile

```bash
nix profile install github:mergewayhq/mergeway-cli
```

This command builds the flake’s default package (`mergeway‑cli`) and adds it to your user profile. The binary is symlinked into `$HOME/.nix-profile/bin`.

### Run without installing

```bash
nix run github:mergewayhq/mergeway-cli -- help
```

Use `nix run` to execute the CLI directly from the flake without permanently installing it. Append `--` and your subcommand (e.g. `nix run github:mergewayhq/mergeway-cli -- version`) to pass arguments to the CLI.

### Build locally

If you clone the repository you can build the binary via Nix:

```bash
# Clone and enter the repository
git clone https://github.com/mergewayhq/mergeway-cli.git
cd mergeway-cli

# Build the CLI from the flake
nix build

# The binary appears in ./result/bin/mergeway-cli
./result/bin/mergeway-cli version
```

This method uses the `flake.nix` file to produce reproducible builds.

### Development shell

For contributors, the flake exposes a development shell that provides Go 1.24.x, linters and documentation tooling. Run `nix develop` (or the provided `devenv shell`) from the project root to enter a shell with all dependencies and pre‑commit hooks installed.

## Option 5 – Build from Source

You can also build the CLI manually from source. Clone the repository and build using the provided `Makefile`:

```bash
git clone https://github.com/mergewayhq/mergeway-cli.git
cd mergeway-cli
make build      # compiles the CLI into bin/mergeway-cli
./bin/mergeway-cli version
```

This approach requires Go 1.24.x and is recommended for people packaging the CLI themselves or contributing to the project.

## Verify the installation

After installation, confirm that the `mergeway‑cli` binary is on your `PATH` and prints version information:

```bash
mergeway-cli --version
```

You should see output similar to:

```
Mergeway CLI v0.3.0 (commit abc1234)
```

If the command is missing, confirm that the installation path is on your `PATH`.

Move on to the [Getting Started](README.md) guide once the binary is available.
