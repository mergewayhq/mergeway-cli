# Storage Layout

## Purpose

Describe conventional on-disk structures for the file-backed database so that teams can organize configuration and object data consistently while retaining flexibility for simpler setups.

## High-Level Structure

The following is a conventional layout for a Mergeway database repository. Treat it as a starting point rather than a mandate—lightweight projects can stay within a single `mergeway.yaml`, while larger teams often break files out by responsibility.

```
repo-root/
  mergeway.yaml
  entities/
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

- `mergeway.yaml`: root configuration entry file that wires together entity definitions via includes. It can also embed the entire schema and inline data for small datasets.
- `entities/`: YAML files that define types and field metadata (optional; use when the schema benefits from modularization).
- `data/`: Object files stored in JSON or YAML; folders are a convenience, not a requirement, and can be introduced gradually.
- `docs/`: Specification documents (this folder).
- `examples/`: Sample datasets referenced in documentation or tests.

## Configuration Layout

- `mergeway.yaml` is the entry point referenced by the CLI via `--config` (defaults to this path when present).
- Split complex schemas into `entities/<TypeName>.yaml` files (or another folder of your choosing) and include them from the entry point using glob patterns when the project size warrants it.

```yaml
# mergeway.yaml
mergeway:
  version: 1

include:
  - entities/*.yaml
```

## Data Layout Guidelines

- File naming should reflect object identifiers (`post-hello-world.yaml`).
- Use `.yaml` for human-edited files by default; `.json` remains supported for automation.
- Files may hold a single object or multiple objects of a single type.
- Multi-object files typically wrap records under `items:` with a shared `type:` header to avoid ambiguity.

## Choosing a Layout

Pick an organization strategy that mirrors how teams maintain the data:

- **Single-file setups**: Keep everything inline inside `mergeway.yaml` when the dataset is tiny, the schema rarely changes, or a single team owns both definitions and data. Inline records stay easy to review and avoid file churn.
- **Split schemas**: Move entities into `entities/` (or another folder name) when the schema grows complex or when you want focused ownership per domain. This lets you gate changes via tooling such as `CODEOWNERS` or directory-specific reviews.
- **Sharded data**: Break `data/` into multiple folders when several teams contribute records or when automation writes to specific subsets. Directory boundaries map cleanly to ownership rules, continuous integration filters, or deployment pipelines.
- **Hybrid**: Combine inline records for lookup tables with file-backed data for high-churn entities. The CLI treats both sources uniformly, so you can pick whatever mix keeps editing friction low.

Revisit the layout as the repository evolves; reorganizing includes and folders is safe as long as references stay consistent.

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

- Object identifiers use alphanumeric characters, hyphens, or underscores.
- Identifiers are unique per type, so different types may reuse the same string.
- Preserve case when naming files to match the identifier.

## Cross-Type References

- To link records, define fields with `reference: <TypeName>` in the configuration. These fields store the identifier of the target type.
- Referenced type names start with an uppercase letter and match a defined type.
- When a field’s `type` equals another defined type, it stores that type’s identifier. Combined with `repeated: true` it models one-to-many relationships implicitly. File contents do not need to repeat the type name; the configuration determines the target type per file pattern.

## Validation Interaction

- The CLI crawls files determined by each type’s `include`.
- Try to keep unrelated assets outside the listed directories or use explicit `include` to avoid accidental ingestion.
- Aggregated validation errors include file paths, so consistent structure improves debugging.

## Future-Proofing

- Consider a dedicated `includes/` folder if you plan to share fragments across multiple databases.
- Keep `data/` sharded by stable heuristics (alphabetical prefix, domain grouping) once datasets grow, updating `include` accordingly.
- Binary assets are best kept outside the database scope until support is introduced.

## Worked Example

See `examples/` in this repository for a complete dataset. `examples/mergeway.yaml` wires together multiple types, while `examples/data/` contains YAML and JSON objects (including repeated fields and cross-type references). The example is exercised by automated tests and can be inspected with commands such as:

```bash
mw --root examples list --type User
mw --root examples get --type Post post-001 --format yaml
mw --root examples validate
```
