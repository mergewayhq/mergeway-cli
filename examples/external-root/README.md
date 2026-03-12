# External Root Example

This example shows that a Mergeway workspace can reference files outside its own directory by using `../...` include paths.

## What it demonstrates
- Top-level config includes resolved relative to the file that declares them
- Entity data includes that point outside the workspace root
- A valid workspace whose schema and records both live in a sibling directory

## Layout
- `examples/external-root/mergeway.yaml` is the workspace entry point and the entity definitions
- `examples/external-root-data-dir/data/**/*.yaml` contains the records loaded by those entities

## Key Paths

Top-level config include in `mergeway.yaml`:

Entity include in the external schema files:

```yaml
include:
  - ../external-root-data-dir/data/suppliers/*.yaml
```

The first path is resolved relative to `mergeway.yaml`. The second is resolved relative to the workspace root passed to the CLI (`examples/external-root`).

## Try It

```bash
mergeway-cli --root examples/external-root export
mergeway-cli --root examples/external-root entity list
mergeway-cli --root examples/external-root list --type Supplier
mergeway-cli --root examples/external-root list --type Product
mergeway-cli --root examples/external-root validate
```
