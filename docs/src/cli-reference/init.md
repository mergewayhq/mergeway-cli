# `mw init`

Last updated: 2025-10-22

> **Synopsis:** Scaffold the directory layout and default configuration for a Mergeway workspace.

## Usage

```bash
mw [global flags] init
```

No command-specific flags. Add the global `--root` flag if you want to scaffold somewhere other than the current directory.

## Example

```bash
mkdir blog-metadata
cd blog-metadata
mw init
```

Output:

```
Initialized repository at .
```

`mw init` ensures a starter `mergeway.yaml` exists in the target directory. Add folders such as `entities/` or `data/` yourself once the project grows; keeping everything in a single file is perfectly valid. Re-run the command safely—it never overwrites existing files.
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
