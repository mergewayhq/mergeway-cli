# Diff Merge Readiness

This note captures the invariants that make `mergeway-cli diff` reusable for a
future semantic merge command.

## Core invariants

1. `diff` ignores configuration by design.
   Config files are only used to discover Mergeway-managed data. Configuration
   edits must never become diff entries.

2. Logical identity is path-independent.
   Objects are keyed by Mergeway identity, not by file location. Moving an
   unchanged object between files must not create a false value modification.

3. Normalization is deterministic.
   The logical database builder and diff engine both emit stable ordering so the
   same repository state produces the same semantic facts and the same rendered
   output.

4. Same normalized value across different paths is not a modification.
   When the logical value is unchanged and only location metadata differs, the
   diff engine reports relocation facts rather than modified field values.

5. Working-tree snapshot modes are explicit and tested.
   `diff` uses `HEAD` vs unstaged-only working tree, `diff <left>` uses the
   full current working tree, and `diff <left> <right>` is revision-only. Those
   modes stay separate because merge work will need the same distinctions.

6. Semantic diff output is stable enough for merge planning.
   `DiffResult` carries field-level changes plus old/new source metadata so a
   future merge command can answer:
   what changed, where it changed semantically, and whether the object moved.

## Reuse for merge

The current layering is intentionally reusable:

- snapshot resolution chooses repository states without leaking Git details into
  the diff engine
- data-only snapshot loading excludes configuration changes up front
- logical database building removes path identity while keeping source metadata
- semantic diffing produces machine-readable facts that are already serialized
  by `mergeway-cli --format json diff`

Future merge work should reuse these layers rather than re-deriving object
identity or file movement from path-based Git diffs.
