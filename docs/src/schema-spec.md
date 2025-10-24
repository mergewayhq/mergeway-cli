# Schema Format

Last updated: 2025-10-22

Schemas live under `types/` and describe one entity per file. Mergeway also expects object data under `data/`; both pieces work together to give you referential integrity.

## Configuration Entry (`mergeway.yaml`)

The workspace entry file declares the schema version and the files to load:

```yaml
version: 1
includes:
  - types/*.yaml
```

- `version` tracks breaking changes in the configuration format (keep it at `1`).
- `includes` is a list of glob patterns. Each matching file is merged into the configuration. Patterns must resolve to at least one file; otherwise Mergeway reports an error.

## Schema Files (`types/*.yaml`)

A schema file declares one or more entity definitions. The example below defines a `Post` entity:

```yaml
entities:
  Post:
    identifier: id
    file_patterns:
      - data/posts/*.yaml
    fields:
      id:
        type: string
        required: true
      title:
        type: string
        required: true
      body:
        type: string
      author:
        type: User
        required: true
```

For advanced scenarios you can expand `identifier` into a mapping:

```yaml
entities:
  Post:
    identifier:
      field: id
      generated: true
    file_patterns:
      - data/posts/*.yaml
    fields:
      # ...
```

### Required Sections

| Key | Description |
| --- | --- |
| `identifier` | Name of the identifier field inside each record (must be unique per entity). Provide either a string (the field name) or a mapping with `field`, optional `generated`, and `pattern`. The identifier value itself can be a string, integer, or number. |
| `file_patterns` | Glob patterns pointing at the data files that belong to this entity. |
| `fields` | Map of field definitions. |

### Field Attributes

| Attribute | Example | Notes |
| --- | --- | --- |
| `type` | `string`, `number`, `boolean`, `list[string]`, `User` | Lists are written as `list[type]`. A plain string (e.g., `User`) references another type. |
| `required` | `true` / `false` | Required fields must appear in every record. |
| `repeated` | `true` / `false` | Indicates an array field. |
| `description` | `Service owner team` | Optional but recommended. |
| `enum` | `[draft, active, retired]` | Allowed values. |
| `default` | Any scalar | Value injected when the field is missing. |

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

You can store one object per file (as above) or provide an `items:` array to keep several objects together. Adding the `type` key is optional when the file already matches the schema’s `file_patterns`, but keeping it makes each file self-describing.

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
