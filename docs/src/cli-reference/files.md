---
title: "mergeway-cli files"
linkTitle: "files"
description: "List included YAML data files together with the entity type each file belongs to."
---

> **Synopsis:** List included YAML data files together with their entity types.

## Usage

```bash
mergeway-cli [global flags] files [--type <type>]
```

| Flag     | Description                                  |
| -------- | -------------------------------------------- |
| `--type` | Optional. Restrict output to one entity type. |

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

## Related Commands

- [`mergeway-cli list`](list.md) — list object identifiers for one type.
- [`mergeway-cli validate`](validate.md) — confirm the listed files still contain valid records.
