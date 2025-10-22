# Mergeway CLI

`mw` is a command-line toolkit for keeping metadata in version control. It stores YAML and JSON objects on disk, validates their schemas, and verifies the integrity of relationships between those objects so your automation stays trustworthy.

## What It Does

- **Track metadata like code**: Check structured data into Git and review changes with normal diffs.
- **Enforce contracts**: Schema and reference validation catch format drift and broken links between related records.
- **Automate comfortably**: Scriptable CRUD commands with YAML/JSON output slot into CI pipelines and other tooling.
- **Spin up fast**: Scaffold a new workspace in seconds with sensible defaults that scale with your dataset.

## Install

### Using Go

```bash
go install github.com/mergewayhq/mergeway-cli@latest
```

Ensure your `GOBIN` (or `GOPATH/bin`) is on `PATH`, then confirm with `mw version`.

### Download a Release Binary

Each GitHub release ships macOS and Linux archives. Drop the `mw` binary somewhere on `PATH` and make it executable (`chmod +x`).

### Build from Source

```bash
git clone https://github.com/mergewayhq/mergeway-cli.git
cd mergeway-cli
make build
./bin/mw version
```

## Quick Start

```bash
# 1. Scaffold a workspace
mw init my-dataset
cd my-dataset

# 2. Inspect the generated types and schemas
mw type list
mw type show User --format json

# 3. Add or update metadata
mw create --type User --file payloads/jane.yaml
mw list --type User

# 4. Validate structure and relations before merging
mw validate --summary
```

## Explore the Examples

A sample workspace lives under `examples/full`. Point `mw` at it to see commands and responses in context:

```bash
mw --root examples/full type list
mw --root examples/full list --type User
mw --root examples/full get --type Post post-001 --format json
mw --root examples/full validate
```

For a deeper command reference, check `docs/cli-behavior.md`.

## Hacking on the CLI

Contributor workflows, tooling, and release notes live in [`DEVELOPING.md`](DEVELOPING.md).
