---
title: JSON Schema interoperability
description: Normative supported JSON Schema import and export subset for Mergeway specification v1.
type: docs
mode: reference
weight: 100
bookToC: true
---

# JSON Schema interoperability

> **Normative v1.** This page applies to repositories declaring `mergeway.version: 1` and is normative.

## Import

An entity MAY replace native `fields` with `json_schema`, whose path is resolved relative to the declaring configuration document. The file MUST contain a JSON object schema with `properties`. `fields` and `json_schema` are mutually exclusive, and a JSON-Schema-backed entity cannot participate in inheritance.

Version 1 supports this conversion subset:

| JSON Schema construct | Mergeway field result |
| --- | --- |
| `type: string`, `integer`, `number`, `boolean` | Corresponding primitive field. |
| `type: object` plus `properties` | Nested `object` and properties. |
| `type: array` plus one `items` schema | `repeated: true` element field. |
| `required` | Required fields at that object level. |
| string `enum`, string `const`, or `oneOf` string constants | `enum` field. |
| `description`, `format`, `pattern`, `default` | Corresponding field metadata. |
| `x-reference-type` | One entity reference. |
| in-document `$ref` | Resolved JSON Pointer. |

A `type` array MAY contain `null` plus one supported non-null type. Multiple non-null types are unsupported. External `$ref` values, reference unions in `x-reference-type`, non-string enum values, and unsupported JSON Schema types are configuration errors. A `$schema` declaration MAY identify draft 2020-12, but version 1 implements the subset above; it does not perform arbitrary draft 2020-12 keyword validation.

The checked import document is:

{{< spec-example path="testdata/specification-v1/schemas/customer.json" language="json" >}}

Keywords outside the conversion subset do not add Mergeway validation constraints. Authors MUST NOT rely on such keywords for repository conformance.

## Export

Configuration export produces a draft 2020-12 object schema with `properties` and `required`. Primitive, enum, nested object, array, description, and single-entity reference fields are represented; references use `x-reference-type`. Path-derived fields are marked `readOnly`.

Export is intentionally lossy in v0.10.0: identifiers, `unique`, `pattern`, `format`, and `default` are not all represented. A native reference union cannot be exported and produces an error. Consumers MUST treat exported JSON Schema as an interoperability view, not a round-trip serialization of all Mergeway configuration.
