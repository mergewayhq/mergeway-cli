---
title: Terminology
description: Normative vocabulary for Mergeway specification version 1.
type: docs
mode: reference
weight: 20
bookToC: true
---

# Terminology

> **Normative v1.** These definitions apply to repositories declaring `mergeway.version: 1` and are normative.

## Core terms

**Repository**
: The root used to resolve entity data-source paths and to express normalized record paths. A repository is also called a workspace in older CLI output.

**Entry configuration**
: The configuration document passed to the implementation, normally `mergeway.yaml` beneath the repository root.

**Configuration document**
: A YAML file with a required `mergeway.version` block and optional configuration includes or entity definitions.

**Entity**
: A named logical type with an identifier definition, a field schema, and zero or more records. Entity names begin with an uppercase ASCII letter.

**Record**
: One map of field names to values associated with exactly one entity. Older internal help may call a record an “object”; JSON object retains its format-specific meaning.

**Data source**
: Inline `data`, a file matched by an entity `include`, an `items` array, or values selected from a document.

**Field**
: A named value in a record governed by a field definition. A property is a nested field inside an `object` field.

**Identifier**
: The non-empty string representation that uniquely addresses a record within its entity and assignable inheritance hierarchy.

**Reference**
: A scalar identifier value, or repeated array of identifier values, whose field type names one or more target entities.

**Assignable hierarchy**
: An entity together with all descendants that extend it. Parent-typed queries and references include descendants.

## Path terms

**Configuration-relative path**
: A path resolved from the directory containing the configuration document. Top-level configuration `include` entries and `json_schema` use this base.

**Repository-relative path**
: A slash-normalized path relative to the repository root. Entity data `include` entries use the repository root as their base, including when the entity was declared in an included configuration document.

Normative pages use these terms consistently. Advice and conceptual background remain in [the repository-model explanation](../../../explanation/repository-model/).
