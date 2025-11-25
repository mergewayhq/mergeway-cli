# `mw export`

> **Synopsis:** Export repository objects into a single JSON or YAML document.

## Usage

```bash
mw [global flags] export [--output <path>] [entity...]
```

| Flag | Description |
| --- | --- |
| `--output` | Optional path to write the exported document. Defaults to STDOUT. |
| `entity...` | Optional list of type names to include. Omitting the list exports every entity defined in the workspace. |

The export format matches the global `--format` flag (`yaml` by default).

## Examples

Export every entity in the repository as YAML to the terminal:

```bash
mw export
```

Export only the `User` and `Post` entities as JSON into a file:

```bash
mw --format json export --output snapshot.json User Post
```

Each top-level key in the output map is the entity name; the value is an array of records sorted by ID.

## Related Commands

- [`mw list`](list.md) — inspect available identifiers before exporting.
- [`mw get`](get.md) — fetch a single object instead of the full dataset.
