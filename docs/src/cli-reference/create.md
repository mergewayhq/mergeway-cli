---
title: "mergeway-cli create"
linkTitle: "create"
description: "Create a new object file that conforms to an entity definition."
---

> **Synopsis:** Create a new object file that conforms to an entity definition.

## Usage

```bash
mergeway-cli [global flags] create --type <type> [--file path] [--id value]
```

| Flag     | Description                                                                 |
| -------- | --------------------------------------------------------------------------- |
| `--type` | Required. Type identifier to create.                                        |
| `--file` | Optional path to a YAML/JSON payload. If omitted, data is read from STDIN.  |
| `--id`   | Optional identifier override for field-based identifiers. Required when the entity uses `identifier: $path`, in which case the value must be the workspace-relative file path to create. |

## Example

Run the command from the workspace root (or pass `--root` if you are elsewhere). Create a user by piping a YAML document and letting Mergeway write the file under `data/users/`:

```bash
cat <<'PAYLOAD' > user.yaml
name: Bob Example
PAYLOAD

mergeway-cli create --type User --file user.yaml --id user-bob
```

Output:

```
User user-bob created
```

The command writes `data/users/user-bob.yaml` with the provided fields. Remove the temporary `user.yaml` file afterward and run `mergeway-cli validate` to confirm the new object passes checks.

When an entity uses `identifier: $path`, pass the target file path with `--id`, for example `--id data/notes/alpha.yaml`. Mergeway uses that workspace-relative path as the object ID and does not persist a `$path` field into the file.

## Related Commands

- [`mergeway-cli update`](update.md) — modify an existing object.
- [`mergeway-cli delete`](delete.md) — remove an object.
