# CLI Reference

Every command shares a set of global flags:

| Flag | Description |
| --- | --- |
| `--root` | Path to the workspace (defaults to `.`). |
| `--config` | Explicit path to `mergeway.yaml` (defaults to `<root>/mergeway.yaml`). |
| `--format` | Output format (`yaml` or `json`, default `yaml`). |
| `--fail-fast` | Stop after the first validation error (where supported). |
| `--yes` | Auto-confirm prompts (useful for `delete`). |
| `--verbose` | Emit additional logging. |

## Repository setup

- [`mw init`](init.md)
- [`mw validate`](validate.md)
- [`mw version`](version.md)

## Schema utilities

- [`mw entity list`](entity-list.md)
- [`mw entity show`](entity-show.md)
- [`mw fmt`](fmt.md)
- [`mw config lint`](config-lint.md)
- [`mw config export`](config-export.md)
- [Schema Format](schema-spec.md)

## Object operations

- [`mw list`](list.md)
- [`mw get`](get.md)
- [`mw create`](create.md)
- [`mw update`](update.md)
- [`mw delete`](delete.md)
- [`mw export`](export.md)

Need a refresher on terminology? See the [Concepts](../concepts/README.md) chapter.
