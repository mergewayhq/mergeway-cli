# Schema Format

Last updated: 2025-10-22

Schemas live under `types/` and describe one entity per file. Mergeway also expects object data under `data/`; both pieces work together to give you referential integrity.

## Configuration Entry (`mergeway.yaml`)

The workspace entry file declares the schema version and the files to load:

```yaml
version: 1
include:
  - types/*.yaml
```

- `version` tracks breaking changes in the configuration format (keep it at `1`).
- `include` is a list of glob patterns. Each matching file is merged into the configuration. Patterns must resolve to at least one file; otherwise Mergeway reports an error.

## Schema Files (`types/*.yaml`)

A schema file declares one or more entity definitions. The example below defines a `Post` entity:

```yaml
entities:
  Post:
    identifier: id
    include:
      - data/posts/*.yaml
    fields:
      id: string
      title:
        type: string
        required: true
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
entities:
  Post:
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
entities:
  User:
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
| `fields`     | Map of field definitions. Use either the shorthand `field: type` (defaults to optional) or the expanded mapping for advanced options.                                                                                                                     |
| `data`       | Optional array of inline records. Each entry must contain the identifier field and follows the same schema rules as external data files.                                                                                                                  |

### Inline Data

Inline data is helpful for tiny lookup tables or bootstrapping a demo without creating additional files. Define records directly inside the entity specification:

```yaml
entities:
  Person:
    identifier: id
    include:
      - data/people/*.yaml
    fields:
      id: string
      name:
        type: string
        required: true
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

Keep schema files small and focused—one entity per file is the easiest to maintain.

## Data Files (`data/...`)

Each data file must declare a `type` and provide the fields required by its entity definition:

```yaml
type: Post
id: post-001
title: Launch Day
author: user-alice
body: |
  We are excited to announce the product launch.
```

You can store one object per file (as above) or provide an `items:` array to keep several objects together. Adding the `type` key is optional when the file already matches the schema’s `include`, but keeping it makes each file self-describing.

JSONPath selectors let you extract objects from nested structures—handy when you need to read a subset of a larger document. For example, `selector: "$.users[*]"` walks through the `users` array in a JSON file and emits one record per element. Mergeway validates that the selector returns objects; any other shape triggers a format error.

Identifier fields accept numeric payloads as well. For example, the following record is valid when the schema marks `id` as an `integer`:

```yaml
type: Person
id: 42
name: Numeric Identifier
```

## Good Practices

- Prefer references (`type: User`) over duplicating identifiers.
- Group files in predictable folders (`data/posts/`, `data/users/`, etc.).
- Run `mw validate` after every change to catch problems immediately.

Need more context? Return to the [Concepts](../concepts/README.md) page for the bigger picture.
