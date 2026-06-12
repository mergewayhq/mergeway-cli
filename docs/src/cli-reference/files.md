---
title: "mergeway-cli files"
linkTitle: "files"
description: "List included YAML data files, or grouped storage containers, together with the entity type."
---

> **Synopsis:** List included YAML data files, or grouped storage containers, together with their entity types.

## Usage

```bash
mergeway-cli [global flags] files [--type <type>] [--group]
```

| Flag      | Description                                                                 |
| --------- | --------------------------------------------------------------------------- |
| `--type`  | Optional. Restrict output to one entity type.                               |
| `--group` | Optional. Collapse one-object-per-file directories to their wildcard path. |

The command only lists YAML-backed files matched by entity `include` globs. Inline records and JSON files are excluded.

## Example

List every included YAML data file:

```bash
mergeway-cli files
```

Output:

```yaml
- type: Post
  file: data/posts/posts.yaml
- type: Tag
  file: data/tags/tag-product.yaml
```

Request JSON output for automation:

```bash
mergeway-cli --format json files --type Tag
```

Show storage containers instead of every concrete file:

```bash
mergeway-cli files --group
```

Example output:

```yaml
- type: Post
  file: data/posts/posts.yaml
- type: User
  file: data/users/*.yaml
```

## Related Commands

- [`mergeway-cli list`](list.md) — list object identifiers for one type.
- [`mergeway-cli validate`](validate.md) — confirm the listed files still contain valid records.
