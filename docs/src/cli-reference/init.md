---
title: "mergeway-cli init"
linkTitle: "init"
---

> **Synopsis:** Scaffold the directory layout and default configuration for a Mergeway workspace.

## Usage

```bash
mw [global flags] init
```

`mw init` targets the directory referenced by `--root` (default `.`) and does not accept positional arguments. Use `mkdir`/`cd` before running the command if you want to initialize a new folder.

Need a walkthrough after initialization? Continue with the [Getting Started guide](/cli/docs/getting-started/).

## Example

```bash
mkdir blog-metadata
cd blog-metadata
mw init
```

Output resembles:

```
Initialized repository at .
```

`mw init` ensures a starter `mergeway.yaml` exists in the target directory. Add folders such as `entities/` or `data/` yourself once the project grows; keeping everything in a single file is perfectly valid. Re-run the command safely—it won't overwrite existing files.

The default configuration contains:

```yaml
# mw configuration
mergeway:
  version: 1

entities: {}
```

## Related Commands

- [`mw validate`](validate.md) — run after adding schema and data files.
- [`mw config lint`](config-lint.md) — verify configuration changes once you edit `mergeway.yaml`.
