---
title: Field definitions and types
description: Normative native field syntax, types, defaults, and constraints for Mergeway specification v1.
type: docs
mode: reference
weight: 60
bookToC: true
---

# Field definitions and types

> **Normative v1.** This page applies to repositories declaring `mergeway.version: 1` and is normative.

## Names and forms

A field or property name MUST be non-empty and contain only ASCII letters, digits, hyphens, or underscores. A scalar string is shorthand for an optional, non-repeated field with that `type`:

```yaml
fields:
  title: string
```

The expanded mapping accepts:

| Member | Meaning |
| --- | --- |
| `type` | Required primitive, `object`, entity reference, or reference union. |
| `required` | When true, a value or non-null default MUST exist. Defaults to false. |
| `repeated` | When true, the record value MUST be an array and each item is validated. Defaults to false. |
| `format` | String annotation; `email`, `uri`, and `url` are enforced case-insensitively. |
| `enum` | Array of allowed strings for `type: enum`. |
| `default` | Value used during validation when the record field is absent. |
| `properties` | Nested field map required by `type: object`. |
| `unique` | When true, present values MUST be unique within the entity. |
| `pattern` | Go regular expression enforced on `string` and `enum` values. |
| `description` | Optional human-readable string. |
| `source` | One [path-derived selector](../path-derived-data/). |

`default` satisfies validation but does not mutate the record or promise that every read/export surface materializes the value. Unrecognized `format` strings are retained as annotations and are not validated in version 1. Extra record fields are not rejected by schema validation.

## Types

- `string` accepts a string.
- `integer` accepts an integer numeric value.
- `number` accepts an integer or floating-point numeric value.
- `boolean` accepts a Boolean.
- `enum` accepts a string and, when `enum` is non-empty, requires membership.
- `object` accepts a map and MUST define `properties`; nested properties use these same rules.
- A configured entity name defines a [reference](../identifiers-references/).
- Two or more entity names separated by `|` define a reference-only union. Members are trimmed, unique, configured entity names; primitive union members are invalid.

`repeated: true` changes the outer value to an array; it MUST NOT be combined with `unique: true`. `properties` is valid only for `type: object`; the legacy nested key `fields` is not accepted.

## Constraint behavior

`required` treats an absent or null value as missing. An empty array is present. `unique` compares normalized scalar values and serialized arrays or maps among records of the same concrete entity; absent values are not compared.

The full checked native-field example is part of [the entity example](../entities/#names-and-members). Path-derived fields add further restrictions defined on their own page.
