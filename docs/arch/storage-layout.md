# Storage Layout

## Purpose

Describe the recommended on-disk structure for the file-backed database so that teams can organize configuration and object data consistently.

## High-Level Structure

```
repo-root/
  mergeway.yaml
  types/
    User.yaml
    Post.yaml
  data/
    users/
      user-alice.yaml
    posts/
      post-*.yaml
  docs/
  examples/
```

- `mergeway.yaml`: root configuration entry file that wires together entity definitions via includes.
- `types/`: YAML files that define types and field metadata.
- `data/`: Object files stored in JSON or YAML; folders are a convenience, not a requirement.
- `docs/`: Specification documents (this folder).
- `examples/`: Sample datasets referenced in documentation or tests.

## Configuration Layout

- `mergeway.yaml` is the entry point referenced by the CLI via `--config` (defaults to this path when present).
- Split complex schemas into `types/<TypeName>.yaml` files and include them from the entry point using glob patterns.

```yaml
# mergeway.yaml
version: 1
include:
  - types/*.yaml
```

## Data Layout Guidelines

- File naming should reflect object identifiers (`post-hello-world.yaml`).
- Use `.yaml` for human-edited files by default; `.json` remains supported for automation.
- Files may hold a single object or multiple objects of a single type.
- Multi-object files should wrap records under `items:` with a shared `type:` header to avoid ambiguity.

### Single Object Example

```yaml
# data/users/user-alice.yaml
id: User-Alice
name: Alice Example
email: alice@example.com
```

### Multi-Object Example

```yaml
# data/posts/posts-batch-001.yaml
items:
  - id: Post-001
    title: First Post
    author: User-Alice
    tags:
      - Tag-Writing
      - Tag-Product
  - id: Post-002
    title: Second Post
    author: User-Alice
    tags:
      - Tag-Writing
```

## Identifier Rules

- Object identifiers must use alphanumeric characters, hyphens, or underscores.
- Identifiers are unique per type, so different types may reuse the same string.
- Preserve case when naming files to match the identifier.

## Cross-Type References

- To link records, define fields with `reference: <TypeName>` in the configuration. These fields store the identifier of the target type.
- Referenced type names must start with an uppercase letter and match a defined type.
- When a field’s `type` equals another defined type, it stores that type’s identifier. Combined with `repeated: true` it models one-to-many relationships implicitly. File contents do not need to repeat the type name; the configuration determines the target type per file pattern.

## Validation Interaction

- The CLI crawls files determined by each type’s `file_patterns`.
- Keep unrelated assets outside the listed directories or use explicit `file_patterns` to avoid accidental ingestion.
- Aggregated validation errors include file paths, so consistent structure improves debugging.

## Future-Proofing

- Consider a dedicated `includes/` folder if you plan to share fragments across multiple databases.
- Keep `data/` sharded by stable heuristics (alphabetical prefix, domain grouping) once datasets grow, updating `file_patterns` accordingly.
- Binary assets should stay outside the database scope until support is introduced.

## Worked Example

See `examples/` in this repository for a complete dataset. `examples/mergeway.yaml` wires together multiple types, while `examples/data/` contains YAML and JSON objects (including repeated fields and cross-type references). The example is exercised by automated tests and can be inspected with commands such as:

```bash
mw --root examples list --type User
mw --root examples get --type Post post-001 --format yaml
mw --root examples validate
```
