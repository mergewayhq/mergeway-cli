---
title: Identifiers and references
description: Normative record identity, identifier patterns, inheritance, and reference integrity for specification v1.
type: docs
mode: reference
weight: 80
bookToC: true
---

# Identifiers and references

> **Normative v1.** This page applies to repositories declaring `mergeway.version: 1` and is normative.

## Identifier definitions

A root entity MUST define `identifier` as one of:

- a field name string, such as `identifier: id`;
- the reserved string `identifier: $path`; or
- a mapping with required `field` plus optional `generated` Boolean and `pattern` string.

Identifier field names use the field-name character set. The field's record value MUST be a non-empty string or numeric scalar. Its canonical identifier is the scalar's string representation. The identifier field does not become required merely because it names a field; a usable value is nevertheless required for every record.

`pattern` is a Go regular expression applied to the canonical identifier. An invalid expression or mismatch is a schema error. `generated: true` is metadata for consumers; version 1 validation and CLI creation still require a supplied identifier.

`$path` derives identity from the normalized repository-relative source path and follows the extra rules in [path-derived data](../path-derived-data/).

## Uniqueness and inheritance

Identifiers MUST be unique among records of one entity. They MUST also be unique across an assignable inheritance hierarchy: a parent and any descendant cannot share an identifier. Unrelated entities MAY use the same identifier.

## References

A field whose `type` is a configured entity name stores the identifier of a record assignable to that entity. A repeated reference stores an array and validates each element. Reference values MAY be strings or numeric scalars and use the same canonical string conversion as identifiers.

A reference union such as `User | Team` accepts a value found in exactly one named target set. No match is a missing-reference error; more than one match is an ambiguity error. A single target that names a parent includes all descendants.

The checked example includes both an inherited parent reference (through `Customer.account_owner`) and a native union (`Activity.owner`):

{{< spec-example path="testdata/specification-v1/data/users/alice.yaml" >}}

Reference validation occurs only after format and schema validation have succeeded for the repository. See [validation and errors](../validation/).
