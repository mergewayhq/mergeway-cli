---
title: "mergeway-cli export"
linkTitle: "export"
description: "Export repository objects into a single JSON or YAML document."
---

> **Synopsis:** Export repository objects into a single JSON or YAML document.

## Usage

```bash
mergeway-cli [global flags] export [--output <path>] [entity...]
```

| Flag        | Description                                                                                              |
| ----------- | -------------------------------------------------------------------------------------------------------- |
| `--output`  | Optional path to write the exported document. Defaults to STDOUT.                                        |
| `entity...` | Optional list of type names to include. Omitting the list exports every entity defined in the workspace. |

The export format matches the global `--format` flag (`yaml` by default).

## Examples

Export every entity in the repository as YAML to the terminal:

```bash
mergeway-cli export
```

Export only the `User` and `Post` entities as JSON into a file:

```bash
mergeway-cli --format json export --output snapshot.json User Post
```

Each top-level key in the output map is the entity name; the value is an array of records sorted by ID.

## Related Commands

- [`mergeway-cli list`](list.md) — inspect available identifiers before exporting.
- [`mergeway-cli get`](get.md) — fetch a single object instead of the full dataset.
