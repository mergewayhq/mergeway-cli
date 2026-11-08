---
title: "mergeway-cli create"
linkTitle: "create"
description: "Create a new object file that conforms to an entity definition."
---

> **Synopsis:** Create a new object file that conforms to an entity definition.

## Usage

```bash
mergeway-cli [global flags] create --type <type> [--file path] [--id value] [--target path]
```

| Flag     | Description                                                                 |
| -------- | --------------------------------------------------------------------------- |
| `--type` | Required. Type identifier to create.                                        |
| `--file` | Optional path to a YAML/JSON payload. If omitted, data is read from STDIN.  |
| `--id`   | Optional identifier override for field-based identifiers. For `identifier: $path`, either this or `--target` is required. |
| `--target` | Optional workspace-relative output path. It must match a writable `include` for the entity. For `identifier: $path`, it may be used instead of `--id`; if both are supplied, they must identify the same path. |

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

To choose among writable includes explicitly, pass `--target`:

```bash
mergeway-cli create --type User --file user.yaml --id user-bob --target data/users/user-bob.json
```

The target must be relative to the workspace root and match a non-selector `include` configured for `User`. Absolute paths, parent traversal, symlink escapes, and paths outside the entity's includes are rejected. If `--target` is omitted, Mergeway retains the normal include-based destination inference.

When an entity uses `identifier: $path`, pass the target file path with `--target`, for example `--target data/notes/alpha.yaml`. The legacy `--id` form remains supported. Mergeway uses that workspace-relative path as the object ID and does not persist a `$path` field into the file.

`create` only writes inside the workspace root. If a type loads records from an external path such as `../secondary/products/*.yaml`, you can still list, get, validate, and export those records, but `create` will reject IDs that point outside the workspace root.

## Related Commands

- [`mergeway-cli update`](update.md) — modify an existing object.
- [`mergeway-cli delete`](delete.md) — remove an object.
