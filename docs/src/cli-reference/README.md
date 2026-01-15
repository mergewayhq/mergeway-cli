---
title: "CLI Reference"
linkTitle: "CLI Reference"
description: "Reference for Mergeway CLI commands, global flags, and usage."
weight: 30
---

Every command shares a set of global flags (use `--long-name`; single-dash long flags like `-root` are not supported). Global flags can appear before or after the command name.

| Flag          | Description                                                            |
| ------------- | ---------------------------------------------------------------------- |
| `--root`      | Path to the workspace (defaults to `.`).                               |
| `--config`    | Explicit path to `mergeway.yaml` (defaults to `<root>/mergeway.yaml`). |
| `--format`    | Output format (`yaml` or `json`, default `yaml`).                      |
| `--fail-fast` | Stop after the first validation error (where supported).               |
| `--yes`       | Auto-confirm prompts (useful for `delete`).                            |
| `--verbose`   | Emit additional logging.                                               |

## Repository setup

- [`init`](init.md)
- [`validate`](validate.md)
- [`version`](version.md)

## Schema utilities

- [`entity list`](entity-list.md)
- [`entity show`](entity-show.md)
- [`fmt`](fmt.md)
- [`config lint`](config-lint.md)
- [`config export`](config-export.md)
- [`gen-erd`](gen-erd.md)

For more information on the schema, please consult the [Schema Format](../getting-started/schema-spec.md)

## Object operations

- [`list`](list.md)
- [`get`](get.md)
- [`create`](create.md)
- [`update`](update.md)
- [`delete`](delete.md)
- [`export`](export.md)

Need a refresher on terminology? See the [Basic Concepts](../getting-started/README.md) page.
