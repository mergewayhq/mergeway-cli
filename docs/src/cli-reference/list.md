---
title: "mergeway-cli list"
linkTitle: "list"
description: "List object identifiers for a given type, optionally filtered by a field."
---

> **Synopsis:** List object identifiers for a given type, optionally filtered by a field.

## Usage

```bash
mergeway-cli [global flags] list --type <type> [--filter key=value]
```

| Flag       | Description                                                                                                                    |
| ---------- | ------------------------------------------------------------------------------------------------------------------------------ |
| `--type`   | Required. Type identifier to query.                                                                                            |
| `--filter` | Optional `key=value` string used to filter objects before listing their IDs. The comparison is a simple string equality check. |

## Example

Run the command from the workspace root. If you need to operate on another directory, add the global `--root` flag.

List all posts in the quickstart workspace:

```bash
mergeway-cli list --type Post
```

Output:

```
post-001
```

Filter by author:

```bash
mergeway-cli list --type Post --filter author=user-alice
```

Output:

```
post-001
```

## Related Commands

- [`mergeway-cli get`](get.md) — inspect a specific object.
- [`mergeway-cli create`](create.md) — add a new object when an ID is missing.
