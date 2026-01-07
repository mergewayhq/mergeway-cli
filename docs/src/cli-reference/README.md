---
title: "CLI Reference"
linkTitle: "CLI Reference"
weight: 30
---

Every command shares a set of global flags:

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
- [Schema Format](/cli/docs/schema-spec/)

## Object operations

- [`list`](list.md)
- [`get`](get.md)
- [`create`](create.md)
- [`update`](update.md)
- [`delete`](delete.md)
- [`export`](export.md)

Need a refresher on terminology? See the [Concepts](/cli/docs/concepts/) page.
