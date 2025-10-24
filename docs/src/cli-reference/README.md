# CLI Reference

Last updated: 2025-10-22

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

- [`mw init`](cli/init.md)
- [`mw validate`](cli/validate.md)
- [`mw version`](cli/version.md)

## Schema utilities

- [`mw type list`](cli/type-list.md)
- [`mw type show`](cli/type-show.md)
- [`mw config lint`](cli/config-lint.md)
- [`mw config export`](cli/config-export.md)
- [Schema Format](schema-spec.md)

## Object operations

- [`mw list`](cli/list.md)
- [`mw get`](cli/get.md)
- [`mw create`](cli/create.md)
- [`mw update`](cli/update.md)
- [`mw delete`](cli/delete.md)
- [`mw export`](cli/export.md)

Need a refresher on terminology? See the [Concepts](../concepts/README.md) chapter.
