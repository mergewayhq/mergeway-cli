# Mergeway CLI

`mw` is a command-line toolkit for keeping metadata in version control. It stores YAML and JSON objects on disk, validates their schemas, and verifies the integrity of relationships between those objects so your automation stays trustworthy.

## What It Does

- **Track metadata like code**: Check structured data into Git and review changes with normal diffs.
- **Enforce contracts**: Schema and reference validation catch format drift and broken links between related records.
- **Automate comfortably**: Scriptable CRUD commands with YAML/JSON output slot into CI pipelines and other tooling.
- **Spin up fast**: Scaffold a new workspace in seconds with sensible defaults that scale with your dataset.

## Key Features

- **Workspace scaffolding**: `mw init` bootstraps `mergeway.yaml` in the working directory so you can start committing metadata immediately.
- **Flexible schemas**: Define entities inline in YAML or point `json_schema` at a JSON Schema (draft 2020-12) file; mix inline data, globbed includes, or JSONPath selectors to source records.
- **Complete CRUD workflow**: `list`, `get`, `create`, `update`, and `delete` commands operate on the same files Git tracks, supporting STDIN/STDOUT automation and partial updates.
- **Deterministic formatting**: `mw fmt` rewrites YAML/JSON in place (add `--stdout` to preview) so reviews stay focused on substance.
- **Layered validation**: `mw validate` runs format, schema, and reference phases, surfacing missing fields, enum mismatches, and cross-entity linkage issues before merge time.
- **Schema introspection**: `mw entity show` and `mw config export` emit normalized schemas or JSON Schema for tooling, keeping downstream integrations in sync.

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
mkdir my-dataset
cd my-dataset
mw init

# 2. Inspect the generated entities and schemas
mw entity list
mw entity show User --format json

# 3. Add or update metadata
mw create --type User --file payloads/jane.yaml
mw list --type User

# 4. Validate structure and relations before merging
mw validate

# 5. Export your dataset as a single snapshot
mw export --format json --output snapshot.json
```

## Explore the Examples

A sample workspace lives under `examples/full`. Point `mw` at it to see commands and responses in context:

```bash
mw --root examples/full entity list
mw --root examples/full list --type User
mw --root examples/full get --type Post post-001 --format json
mw --root examples/full validate
```

Curious about JSON Schema-backed entities? `examples/json-schema` demonstrates how an entity can derive its field definitions from an external JSON Schema file.

For a deeper command reference, check [`docs/arch/cli-behavior.md`](docs/arch/cli-behavior.md).

## Hacking on the CLI

Contributor workflows, tooling, and release notes live in [`DEVELOPING.md`](DEVELOPING.md).

## License

Mergeway CLI is released under the [MIT License](LICENSE.md).
