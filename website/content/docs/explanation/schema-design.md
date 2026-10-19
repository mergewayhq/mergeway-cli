---
title: Schema design choices
description: Choose native or JSON Schema fields, record granularity, inheritance, unions, and identifier strategy.
type: docs
mode: explanation
weight: 70
bookToC: true
---

# Schema design choices

Mergeway does not prescribe one repository layout. The durable choice is the one that makes identity, ownership, review, and automation clear for the team maintaining the records.

## Native fields or JSON Schema

Native `fields` integrate directly with entity references, inheritance, defaults, enums, and path-derived fields. They are usually the clearest choice when Mergeway owns the schema.

Use `json_schema` when an external system owns an object contract or when its supported draft 2020-12 constraints are valuable. Mergeway converts only its documented subset, and JSON-Schema-backed entities cannot use inheritance. The [JSON Schema guide](../../how-to/repositories/reuse-json-schema/) demonstrates the boundary.

## One record per file or grouped records

One record per file gives Git history, ownership, and merge conflicts a narrow unit. It also enables `$path` identity. Grouped YAML `items` or JSONPath-selected arrays reduce file count and can match upstream exports, but a change to one record shares a backing file with its neighbors.

Directory structure has no meaning unless an include or path-derived source gives it meaning. Start with the smallest layout that reflects actual ownership, then split configuration and records using the [repository-model guide](../../how-to/cli/split-files/).

## Inheritance or a reference union

Inheritance says one entity is assignable to its parent and inherits its native fields. It is appropriate for a real common schema and parent-level queries. A union says one value may refer to several otherwise independent entities. It does not merge their fields or make them interchangeable elsewhere.

Both approaches tighten identifier rules. Parent/descendant identifiers share an assignable namespace, while a union value must resolve in exactly one target. Work through [Model inheritance and reference unions](../../how-to/repositories/inheritance-reference-unions/) before adopting either.

## Field or path identity

A field identifier survives file moves and is explicit in exported data. `$path` avoids duplicating an identity already enforced by layout, but moving a file changes the identifier and can break references. Path-derived fields provide a middle ground: keep a stable field identifier while exposing location metadata as read-only values.

The [normative specification](../../reference/specification/) remains authoritative for syntax and validation. These choices are maintenance trade-offs rather than additional conformance rules.
