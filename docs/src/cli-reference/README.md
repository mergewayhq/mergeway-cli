---
title: "mergeway-cli Reference"
linkTitle: "mergeway-cli Reference"
description: "Reference for the mergeway-cli binary."
weight: 30
---

`mergeway-cli` is the primary workspace-management binary. It handles repository setup, schema inspection, validation, formatting, and object operations.

Use `--long-name`; single-dash long flags like `-root` are not supported. Global flags can appear before or after the command name.

| Flag          | Description                                                            |
| ------------- | ---------------------------------------------------------------------- |
| `--root`      | Path to the workspace (defaults to `.`).                               |
| `--config`    | Explicit path to `mergeway.yaml` (defaults to `<root>/mergeway.yaml`). |
| `--format`    | Output format (`yaml` or `json`, default `yaml`).                      |
| `--fail-fast` | Stop after the first validation error (where supported).               |
| `--yes`       | Auto-confirm prompts (useful for `delete`).                            |
| `--verbose`   | Emit additional logging.                                               |

## `mergeway-cli` Repository Setup

- [`init`](init.md)
- [`validate`](validate.md)
- [`version`](version.md)

## `mergeway-cli` Schema Utilities

- [`entity list`](entity-list.md)
- [`entity show`](entity-show.md)
- [`fmt`](fmt.md)
- [`config lint`](config-lint.md)
- [`config export`](config-export.md)
- [`gen-erd`](gen-erd.md)

For more information on the schema, please consult the [Schema Format](../getting-started/schema-spec.md)

## `mergeway-cli` Object Operations

- [`list`](list.md)
- [`files`](files.md)
- [`get`](get.md)
- [`create`](create.md)
- [`update`](update.md)
- [`delete`](delete.md)
- [`export`](export.md)

For the other binaries, see [mergeway-diff Reference](diff.md) and [mergeway-lsp Reference](lsp.md). Need a refresher on terminology? See the [Basic Concepts](../getting-started/README.md) page.
