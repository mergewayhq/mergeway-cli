# Schema Format

Schemas can live entirely inside `mergeway.yaml` or be split across additional include files (for example under an `entities/` folder) for readability. Likewise, object data may be defined inline or stored under `data/`. Pick the mix that matches your editing workflow—comments below highlight conventions for modular repositories without requiring them. See [Storage Layout](../arch/storage-layout.md) for heuristics on choosing a structure.

## Configuration Entry (`mergeway.yaml`)

The workspace entry file declares the schema version and the files to load:

```yaml
mergeway:
  version: 1

include:
  - entities/*.yaml
```

- `mergeway.version` tracks breaking changes in the configuration format (keep it at `1`).
- `include` is a list of glob patterns. Each matching file is merged into the configuration. Patterns must resolve to at least one file; otherwise Mergeway reports an error.

## Schema Files (optional includes)

A schema file declares one or more entity definitions. Store them in whichever folder makes sense for your workflow (many teams use `entities/`); the location has no semantic impact. The example below defines a `Post` entity:

```yaml
mergeway:
  version: 1

entities:
  Post:
    description: Blog posts surfaced on the marketing site
    identifier: id
    include:
      - data/posts/*.yaml
    fields:
      id: string
      title:
        type: string
        required: true
        description: Human readable title
      body: string
      author:
        type: User
        required: true
    data:
      - id: post-inline
        title: Inline Example
        author: user-alice
        body: Inline data lives in the schema file.
```

For advanced scenarios you can expand `identifier` into a mapping:

```yaml
mergeway:
  version: 1

entities:
  Post:
    description: Blog posts surfaced on the marketing site
    identifier:
      field: id
      generated: true
    include:
      - data/posts/*.yaml
    fields:
      # ...
```

When several objects live in one file, provide a JSONPath selector to extract them:

```yaml
mergeway:
  version: 1

entities:
  User:
    description: Directory of account holders sourced from JSON
    identifier: id
    include:
      - path: data/users.json
        selector: "$.users[*]"
    fields:
      # ...
```

Strings remain a shorthand for `path` with no `selector`; Mergeway then reads the entire file as a single object (or uses the `items:` array if present).

### Required Sections

| Key          | Description                                                                                                                                                                                                                                               |
| ------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `identifier` | Name of the identifier field inside each record (must be unique per entity). Provide either a string (the field name) or a mapping with `field`, optional `generated`, and `pattern`. The identifier value itself can be a string, integer, or number.    |
| `include`    | List of data sources. Each entry can be a glob string (shorthand) or a mapping with `path` and optional `selector` property. Omit only when you rely exclusively on inline `data`. Without a selector, Mergeway treats the whole file as a single object. |
| `fields`     | Map of field definitions. Use either the shorthand `field: type` (defaults to optional) or the expanded mapping for advanced options. Provide either `fields` or `json_schema` for each entity.                                                        |
| `json_schema`| Path to a JSON Schema (draft 2020-12) file relative to the schema that declares the entity. When present, Mergeway derives field definitions from the JSON Schema and the `fields` block must be omitted.                                             |
| `data`       | Optional array of inline records. Each entry must contain the identifier field and follows the same schema rules as external data files.                                                                                                                  |

Add `description` anywhere you need extra context. Entities accept it alongside `identifier`, and each field definition supports its own `description` value.

### Inline Data

Inline data is helpful for tiny lookup tables or bootstrapping a demo without creating additional files. Define records directly inside the entity specification:

```yaml
mergeway:
  version: 1

entities:
  Person:
    description: Lightweight profile objects
    identifier: id
    include:
      - data/people/*.yaml
    fields:
      id: string
      name:
        type: string
        required: true
        description: Preferred display name
      age: integer
    data:
      - id: person-1
        name: Alice
        age: 30
      - id: person-2
        name: Bob
        age: 42
```

Inline records are loaded alongside file-based data. If a record with the same identifier exists both inline and on disk, the file wins. Inline records are read-only at runtime—`mw data update` and `mw data delete` target files only.

### Field Shorthand

When a field only needs a type, map entries can use the compact `field: type` syntax. These fields default to `required: false` and behave identically to the expanded form otherwise. Switch to the full mapping whenever you need attributes like `required`, `repeated`, or `format`.

### Field Attributes

| Attribute     | Example                                               | Notes                                                                                     |
| ------------- | ----------------------------------------------------- | ----------------------------------------------------------------------------------------- |
| `type`        | `string`, `number`, `boolean`, `list[string]`, `User` | Lists are written as `list[type]`. A plain string (e.g., `User`) references another type. |
| `required`    | `true` / `false`                                      | Required fields must appear in every record.                                              |
| `repeated`    | `true` / `false`                                      | Indicates an array field.                                                                 |
| `description` | `Service owner team`                                  | Optional but recommended.                                                                 |
| `enum`        | `[draft, active, retired]`                            | Allowed values.                                                                           |
| `default`     | Any scalar                                            | Value injected when the field is missing.                                                 |

### JSON Schema Entities

For larger teams it can be convenient to author schemas once and consume them in multiple places. Entities now support a `json_schema` property that points to an on-disk JSON Schema document (draft 2020-12). The path is resolved relative to the file that declares the entity and must live inside the repository—external `$ref` documents and network lookups are rejected.

When `json_schema` is present, omit the `fields` map. Mergeway parses the JSON Schema and converts to its native field definitions:

- `type: object` becomes nested field groups, preserving `required` entries for each level.
- `type: array` sets `repeated: true` and uses the `items` schema to determine the element type.
- `enum`, `const`, or `oneOf` blocks translate into Mergeway enums (string values only).
- `$ref` segments are resolved within the same JSON Schema file (e.g., `#/$defs/...`).
- Custom references to other entities use the same `x-reference-type` property emitted by `mw config export`.

See `examples/json-schema` for a runnable workspace that demonstrates this flow end-to-end.

Keep schema files small and focused—one entity per file is the easiest to maintain.

## Data Files (`data/...`)

Each data file provides the fields required by its entity definition. Declaring a `type` at the top is optional—the CLI infers it from the entity that referenced the file (through `include`/`selector`) and only errors when a conflicting `type` value is present. Keeping it in the file can still be helpful for humans who open an arbitrary YAML document.

```yaml
type: Post             # optional; falls back to the entity that included this file
id: post-001
title: Launch Day
author: user-alice
body: |
  We are excited to announce the product launch.
```

You can store one object per file (as above) or provide an `items:` array to keep several objects together. Mergeway removes any top-level `type` key before validating the record, so referencing the same file from multiple entities requires the selector approach described below.

JSONPath selectors let you extract objects from nested structures—handy when you need to read a subset of a larger document. For example, `selector: "$.users[*]"` walks through the `users` array in a JSON file and emits one record per element. Mergeway validates that the selector returns objects; any other shape triggers a format error.

Identifier fields accept numeric payloads as well. For example, the following record is valid when the schema marks `id` as an `integer`:

```yaml
id: 42
name: Numeric Identifier
```

## Good Practices

- Prefer references (`type: User`) over duplicating identifiers.
- Group files in predictable folders (`data/posts/`, `data/users/`, etc.).
- Run `mw validate` after every change to catch problems immediately.

Need more context? Return to the [Concepts](../concepts/README.md) page for the bigger picture.
