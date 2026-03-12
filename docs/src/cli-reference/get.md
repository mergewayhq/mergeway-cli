---
title: "mergeway-cli get"
linkTitle: "get"
description: "Print the fields of one object."
---

> **Synopsis:** Print the fields of one object.

## Usage

```bash
mergeway-cli [global flags] get --type <type> <id>
```

| Flag     | Description                                                      |
| -------- | ---------------------------------------------------------------- |
| `--type` | Required. Type identifier that owns the object.                  |
| `<id>`   | Required positional argument representing the object identifier. For entities that use `identifier: $path`, this is the workspace-relative file path. |

Use `--format json` if you prefer JSON output.

## Example

Run the command from the workspace root. Use `--root` if you need to target another workspace.

Fetch the `post-001` record as YAML:

```bash
mergeway-cli --format yaml get --type Post post-001
```

Output:

```yaml
author: user-alice
body: |
  We are excited to announce the product launch.
id: post-001
title: Launch Day
```

For path-based identifiers, the lookup uses the file path instead:

```bash
mergeway-cli --format yaml get --type Note data/notes/alpha.yaml
```

## Related Commands

- [`mergeway-cli list`](list.md) — discover identifiers before calling `get`.
- [`mergeway-cli update`](update.md) — change object fields.
