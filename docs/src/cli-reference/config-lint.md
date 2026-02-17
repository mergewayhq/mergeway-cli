---
title: "mergeway-cli config lint"
linkTitle: "config lint"
description: "Validate configuration files (including includes) without touching data."
---

> **Synopsis:** Validate configuration files (including includes) without touching data.

## Usage

```bash
mergeway-cli [global flags] config lint
```

No additional flags.

## Example

Run the command from the workspace root (or pass `--root`):

```bash
mergeway-cli config lint
```

Output:

```
configuration valid
```

If the command encounters a problem (for example, an include pattern that matches no files), it prints the error and exits with status `1`.

Run this command whenever you edit `mergeway.yaml` or add new entity definitions to catch syntax mistakes early.

## Related Commands

- [`mergeway-cli config export`](config-export.md) — derive a JSON Schema for a type.
- [`mergeway-cli validate`](validate.md) — validate both schemas and data.
