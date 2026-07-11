# Inheritance Example

This example demonstrates Mergeway entity inheritance with a schema-only base entity.

## What it demonstrates
- `Animal` is a schema-only base entity with no `include` or inline `data`
- `Dog` extends `Animal` and inherits the `id` and `name` fields
- `Kennel.resident` is typed as `Animal`, but it can reference the `Dog` object
- Parent queries such as `list --type Animal` and `get --type Animal dog-1` include descendant objects

## Try It

```bash
mergeway-cli --root examples/inheritance validate
mergeway-cli --root examples/inheritance --format yaml entity show Dog
mergeway-cli --root examples/inheritance list --type Animal
mergeway-cli --root examples/inheritance --format yaml get --type Animal dog-1
```
