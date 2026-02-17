---
title: "mergeway-cli init"
linkTitle: "init"
description: "Scaffold the directory layout and default configuration for a Mergeway workspace."
---

> **Synopsis:** Scaffold the directory layout and default configuration for a Mergeway workspace.

## Usage

```bash
mergeway-cli [global flags] init
```

`mergeway-cli init` targets the directory referenced by `--root` (default `.`) and does not accept positional arguments. Use `mkdir`/`cd` before running the command if you want to initialize a new folder.

Need a walkthrough after initialization? Continue with the [Getting Started guide](../getting-started/README.md).

## Example

```bash
mkdir blog-metadata
cd blog-metadata
mergeway-cli init
```

Output resembles:

```
Initialized repository at .
```

`mergeway-cli init` ensures a starter `mergeway.yaml` exists in the target directory. Add folders such as `entities/` or `data/` yourself once the project grows; keeping everything in a single file is perfectly valid. Re-run the command safely—it won't overwrite existing files.

The default configuration contains:

```yaml
# mergeway-cli configuration
mergeway:
  version: 1

entities: {}
```

## Related Commands

- [`mergeway-cli validate`](validate.md) — run after adding schema and data files.
- [`mergeway-cli config lint`](config-lint.md) — verify configuration changes once you edit `mergeway.yaml`.
