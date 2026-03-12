# Path Identifier Example

This example shows a minimal workspace that uses `identifier: $path` while keeping the entity definition inline in `mergeway.yaml`.

## What it demonstrates
- Inline entity definitions without extra `entities/` files
- One object per file under `data/notes/`
- File paths as stable object identifiers

## Try it

```bash
mergeway-cli --root examples/path-identifier list --type Note
mergeway-cli --root examples/path-identifier get --type Note data/notes/alpha.yaml
mergeway-cli --root examples/path-identifier validate
```

The `Note` objects do not store an `id` field. Instead, Mergeway uses each file's workspace-relative path, such as `data/notes/alpha.yaml`, as the identifier.
