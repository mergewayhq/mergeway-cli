---
title: Validation and error model
description: Normative configuration failures and format, schema, and reference validation phases for specification v1.
type: docs
mode: reference
weight: 90
bookToC: true
---

# Validation and error model

> **Normative v1.** This page applies to repositories declaring `mergeway.version: 1` and is normative.

## Configuration loading

Configuration parsing, include resolution, normalization, inheritance construction, and JSON Schema conversion happen before repository validation. A failure in this stage is a configuration error and prevents validation. Examples include an unsupported version, duplicate entity definition, invalid field definition, unknown reference type, or inheritance cycle.

## Validation phases

Repository validation proceeds in dependency order:

1. **Format** resolves entity globs and decodes sources, `items`, selectors, and OKF documents. A conflicting document type, malformed source, invalid selector, or non-map record is an error.
2. **Schema** derives identifiers and path fields, validates identifier patterns and uniqueness, enforces fields and nested properties, and constructs the assignable identifier index.
3. **References** resolves every present reference value against that index, including descendants and union ambiguity.

If format errors exist, later phases do not run. If schema errors exist, references do not run. Requesting reference validation therefore requires schema validation as a prerequisite. A tool MAY expose phase filtering or fail-fast reporting, but those controls do not weaken repository conformance.

## Error information

A repository validation error has a phase, entity type, source location, message, and—when available—record identifier. Multi-record locations identify the one-based item. Inline locations identify their order in the declaring configuration.

Implementations SHOULD report enough of this information to locate the failure. Error wording is not a stable machine interface in version 1.

## Schema coverage

Schema validation checks required/default behavior, arrays, primitive types, enum membership, supported string formats and patterns, nested properties, per-field uniqueness, identifiers, and hierarchy collisions. It does not reject unknown record members. Reference validation checks only fields whose normalized type names configured entities.

For a conceptual introduction, see [the validation model](../../../explanation/validation-model/). That explanation is non-normative.
