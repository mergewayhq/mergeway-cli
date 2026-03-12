# Reference Union Example

This example shows a reference-only union in a native Mergeway `fields:` definition.

## What it demonstrates
- A field that can reference one of multiple entity types with `User | Team`
- Inline data that validates cleanly because the referenced identifiers are unique across both target sets
- The ambiguity rule: if the same identifier existed in both `User` and `Team`, validation would fail

## Try It

```bash
mergeway-cli --root examples/reference-union entity list
mergeway-cli --root examples/reference-union --format yaml entity show Activity
mergeway-cli --root examples/reference-union list --type Activity
mergeway-cli --root examples/reference-union validate
```
