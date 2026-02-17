---
title: "mergeway-cli entity list"
linkTitle: "entity list"
description: "Show every entity Mergeway discovered from your configuration."
---

> **Synopsis:** Show every entity Mergeway discovered from your configuration.

## Usage

```bash
mergeway-cli [global flags] entity list
```

No command-specific flags. Add the global `--root` flag if you need to inspect another workspace.

## Example

List entities for the `examples/` workspace bundled with the repository:

```bash
mergeway-cli --root examples/full entity list
```

Output:

```
Comment
Post
Tag
User
```

Entities are listed alphabetically.

## Related Commands

- [`mergeway-cli entity show`](entity-show.md) — inspect an individual schema definition.
- [`mergeway-cli config lint`](config-lint.md) — verify the configuration if an entity is missing.
