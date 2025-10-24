# Schema Specification

## Purpose

Define the configuration language that describes database structure, so the CLI and future tooling can understand object types, constraints, and cross-type references.

## Configuration Files

- Primary entry point: `mergeway.yaml` (path configurable when invoking the CLI).
- YAML is the authoritative format for all configuration, keeping schema metadata decoupled from object files.
- Config files may `include` other files via glob patterns, letting teams split large schemas into focused modules per domain or type.

```yaml
version: 1
include:
  - types/**/*.yaml
```

The CLI resolves `include` relative to the parent file, expanding globs before merging the referenced documents.

## Top-Level Structure

Every configuration document must conform to the following shape after include expansion:

```yaml
version: <integer>
entities:
  <TypeName>:
    identifier: <IdentifierDefinition|string>
    file_patterns:
      - <glob>
    fields:
      <FieldName>: <FieldDefinition>
```

- `version`: configuration schema version (start with `1`).
- `entities`: map keyed by type identifiers (must start with an uppercase letter and otherwise follow identifier constraints outlined in `database-requirements.md`).

## Entity Definition

Each `<EntityDefinition>` entry provides the authoritative schema for a single object type.

```yaml
entities:
  User:
    identifier:
      field: id
    file_patterns:
      - data/users/*.yaml
    fields:
      id:
        type: string
        required: true
      name:
        type: string
        required: true
      email:
        type: string
        format: email
        required: true
      profile:
        type: object
        properties:
          bio:
            type: string
          website:
            type: string
            format: uri
      teams:
        type: Team
        repeated: true
```

`identifier` accepts either a plain string (e.g., `identifier: id`) or a mapping with `field`, optional `generated`, and `pattern` keys when you need additional behavior.

### Field Specification

- `type`: primitive (`string`, `integer`, `number`, `boolean`), `object`, `enum`, or another defined type name to indicate a reference.
- `required`: defaults to `false`; set to `true` for mandatory fields.
- `repeated`: when `true`, the field stores an array of values described by the rest of the field definition.
- `format`: optional semantic hint (URI, email, date-time, etc.); aligns with JSON Schema semantics.
- `enum`: array of allowed values when `type: enum`.
- `default`: optional default value filled in by tooling.
- `properties`: nested field definitions used when `type: object`. When combined with `repeated: true`, each array element must respect the nested definition.

### Validation Extensions

Extra validation knobs let future tooling derive JSON Schema while preserving richer semantics:

- `unique`: ensure a field’s value is unique across all objects of this type. Only applicable to non-repeated fields.
- When `type` equals another defined type name, the field stores the referenced type’s identifier (similar to a foreign key). For repeated fields this models one-to-many links implicitly.
- `computed`: mark fields derived during build/publish; the CLI can warn if supplied manually.

## File Association

`file_patterns` bind a type to one or more data files. Patterns may select both YAML and JSON documents. When matching files, the CLI infers the type from the configuration; a top-level `type` field is optional and only used as an override/sanity check.

```yaml
# multi-object file example
items:
  - id: post-001
    title: Hello
  - id: post-002
    title: World
```

## Future JSON Schema Generation

- The configuration should contain enough detail to emit JSON Schema for downstream tooling.
- Repeated fields map to JSON Schema arrays with `items` derived from the base field definition.
- Fields whose `type` is another defined type map to foreign-key style validations and can translate to JSON Schema annotations or custom keywords.
- Keep the configuration expressive but deterministic so transformation tooling can remain stable.

## Validation Hooks

- Format validation uses `type`, `format`, and `enum` metadata.
- Schema validation leverages `required`, `repeated`, `properties`, and `unique`.
- Referential integrity ensures that fields whose `type` matches another defined type reference existing identifiers.

This specification provides the foundation for building linting, code generation, and future schema evolution features while remaining human-editable.
