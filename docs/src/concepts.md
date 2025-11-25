# Concepts

A Mergeway workspace is just a folder with a few predictable parts. Knowing the vocabulary makes the CLI output easier to read.

## Building Blocks

- **Workspace:** Folder tracked in Git that contains `mergeway.yaml`, schemas, and optional objects. All commands run from here.
- **Schema:** YAML/JSON that defines fields and references. Each file describes one entity.
- **Object:** Optional data instances stored under `data/`.
- **Reference:** A link from one schema or field to another (`type: ref`). Mergeway validates referential integrity.

## Validation Flow

1. Mergeway loads `mergeway.yaml` to locate schemas and records.
2. Schemas are parsed and checked for required fields, types, and references.
3. Records (if present) are validated against their schemas.

For field syntax and configuration options, see the [Schema Format](schema-spec.md).
