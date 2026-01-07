---
title: "mergeway-cli create"
linkTitle: "create"
---

> **Synopsis:** Create a new object file that conforms to an entity definition.

## Usage

```bash
mw [global flags] create --type <type> [--file path] [--id value]
```

| Flag     | Description                                                                 |
| -------- | --------------------------------------------------------------------------- |
| `--type` | Required. Type identifier to create.                                        |
| `--file` | Optional path to a YAML/JSON payload. If omitted, data is read from STDIN.  |
| `--id`   | Optional identifier override. Useful when the payload omits the `id` field. |

## Example

Run the command from the workspace root (or pass `--root` if you are elsewhere). Create a user by piping a YAML document and letting Mergeway write the file under `data/users/`:

```bash
cat <<'PAYLOAD' > user.yaml
name: Bob Example
PAYLOAD

mw create --type User --file user.yaml --id user-bob
```

Output:

```
User user-bob created
```

The command writes `data/users/user-bob.yaml` with the provided fields. Remove the temporary `user.yaml` file afterward and run `mw validate` to confirm the new object passes checks.

## Related Commands

- [`mw update`](update.md) — modify an existing object.
- [`mw delete`](delete.md) — remove an object.
