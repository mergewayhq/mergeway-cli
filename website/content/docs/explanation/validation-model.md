---
title: The validation model
description: Understand Mergeway's format, schema, and relationship validation phases and how tools share their results.
type: docs
mode: explanation
weight: 30
bookToC: true
---

# The validation model

Validation answers one repository-level question: can Mergeway load this file-based database as a consistent set of entities, records, and relationships? It proceeds in layers so a useful local error is reported before a dependent check produces noise.

## Format

The format phase reads configured YAML and JSON sources. Syntax errors, unsupported document shapes, or data that cannot be decoded stop the affected source from becoming records. Formatting and parsing are related but distinct: canonical formatting makes reviews predictable, while format validation determines whether data can be read.

## Schema

The schema phase checks each loaded record against its entity definition. It covers required fields, scalar types, enum membership, repeated values, and other declared constraints. Identifier rules also allow Mergeway to address records consistently and detect collisions.

Entity definitions can use Mergeway's native field model or supported JSON Schema sources. Both feed the repository model, but their authoring syntax and feature boundaries differ.

## References

The reference phase checks relationships after records and identifiers are known. When a `Post.author` field has type `User`, its value must identify a loaded `User` record. Repeated and union relationships apply the same principle to each permitted target.

This phase turns separately valid files into a connected database. Deleting or renaming a target record can therefore invalidate a record in another file even when neither file has a syntax or field-type error.

## Consistent feedback

`mergeway-cli validate` is the explicit whole-repository check. The LSP reuses the same core to produce editor diagnostics, and other components load the same configuration and data model for their operations. Consumers should still validate before publishing or exporting: successful parsing alone does not prove schema or relationship integrity.

By default the CLI reports all discovered problems. `--fail-fast` stops earlier, and repeatable `--phase` flags can narrow a diagnostic run. Those controls change how checks are executed, not the repository's validity rules.

Run all three layers in [Build your first Mergeway repository](../../tutorials/first-repository/), review [the repository vocabulary](../repository-model/), or see how the [CLI](../../components/cli/) and [LSP](../../components/lsp/) expose validation.
