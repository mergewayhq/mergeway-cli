---
title: Entity definitions
description: Normative entity members, inline data, and inheritance rules for Mergeway specification v1.
type: docs
mode: reference
weight: 50
bookToC: true
---

# Entity definitions

> **Normative v1.** This page applies to repositories declaring `mergeway.version: 1` and is normative.

## Names and members

An entity name MUST begin with an uppercase ASCII letter. Remaining characters MAY be ASCII letters, digits, hyphens, or underscores.

| Member | Requirement and meaning |
| --- | --- |
| `description` | Optional human-readable string. |
| `identifier` | Required on a root entity; inherited by descendants. |
| `extends` | Optional name of one parent entity. |
| `include` | Optional array of [record sources](../records/). |
| `fields` | Native field map; mutually exclusive with `json_schema`. |
| `json_schema` | Configuration-relative JSON Schema path; mutually exclusive with `fields` and inheritance. |
| `data` | Optional array of inline record maps. |

An entity MUST define native `fields` or `json_schema`. A descendant using native fields MAY inherit all its fields and add none. A concrete entity with no descendants MUST provide at least one `include` or inline record. A schema-only entity MAY omit both only when another entity extends it.

The checked example defines a schema-only base, one descendant, two inline entities, all identifier forms, and native field forms:

{{< spec-example path="testdata/specification-v1/entities/core.yaml" >}}

## Inline records

Each `data` item is a record governed by the same schema and reference rules as a file-backed record. Inline records participate in reads and validation. They have no backing record file and are read-only to file mutation operations.

If inline and file-backed records have the same entity and identifier, loading for ordinary data operations gives precedence to the file-backed record. Whole-repository validation still detects duplicate identifiers among collected records. An entity using `identifier: $path` or any path-derived field MUST NOT define inline data.

## Inheritance

`extends` names exactly one configured parent. Cycles and unknown parents are errors. A descendant inherits the parent's identifier, fields, and all ancestor fields. It MAY add fields but MUST NOT redefine an inherited field. If it repeats `identifier`, the complete mapping MUST equal the inherited definition.

Identifier values MUST be unique across each assignable hierarchy. Querying a parent includes descendant records. A reference to a parent accepts a record of the parent or any descendant.

An entity that uses `json_schema` MUST NOT extend another entity and MUST NOT have descendants in version 1.
