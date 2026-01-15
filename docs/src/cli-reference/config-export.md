---
title: "mergeway-cli config export"
linkTitle: "config export"
description: "Emit a JSON Schema for one of your types."
type: docs
---

> **Synopsis:** Emit a JSON Schema for one of your types.

## Usage

```bash
mw [global flags] config export --type <type>
```

| Flag     | Description                          |
| -------- | ------------------------------------ |
| `--type` | Required. Type identifier to export. |

## Example

Run the command from the workspace root (or pass `--root`). Export the `Post` type as JSON Schema:

```bash
mw --root examples --format json config export --type Post
```

Output (abridged):

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "properties": {
    "author": {
      "type": "string",
      "x-reference-type": "User"
    },
    "title": {
      "type": "string"
    }
  },
  "required": ["id", "title", "author"],
  "type": "object"
}
```

Fields that reference other types include the `x-reference-type` hint.

Validate your workspace (`mw config lint` or `mw validate`) after editing type files to ensure the exported schema stays in sync.

## Related Commands

- [`mw entity show`](entity-show.md) — view the full Mergeway representation of an entity.
- [`mw validate`](validate.md) — ensure data conforms to the schema you just exported.
