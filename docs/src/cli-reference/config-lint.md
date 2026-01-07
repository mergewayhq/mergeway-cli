---
title: "mergeway-cli config lint"
linkTitle: "config lint"
---

> **Synopsis:** Validate configuration files (including includes) without touching data.

## Usage

```bash
mw [global flags] config lint
```

No additional flags.

## Example

Run the command from the workspace root (or pass `--root`):

```bash
mw config lint
```

Output:

```
configuration valid
```

If the command encounters a problem (for example, an include pattern that matches no files), it prints the error and exits with status `1`.

Run this command whenever you edit `mergeway.yaml` or add new entity definitions to catch syntax mistakes early.

## Related Commands

- [`mw config export`](config-export.md) — derive a JSON Schema for a type.
- [`mw validate`](validate.md) — validate both schemas and data.
