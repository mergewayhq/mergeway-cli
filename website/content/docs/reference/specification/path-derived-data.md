---
title: Path-derived data
description: Normative path identifiers and synthetic field sources for Mergeway specification v1.
type: docs
mode: reference
weight: 110
bookToC: true
---

# Path-derived data

> **Normative v1.** This page applies to repositories declaring `mergeway.version: 1` and is normative.

## Normalized path

Path-derived behavior starts with the record source path relative to the repository root. Implementations clean the platform path and expose it with `/` separators. A source outside the root therefore retains leading `..` segments.

## `$path` identifiers

`identifier: $path` uses that normalized path as the record identifier. Each matched file MUST yield exactly one record. Inline data, top-level `items` arrays, and selectors that return multiple records are invalid with `$path`. The record does not need a literal `$path` field.

## Synthetic field sources

A string-like, non-repeated field MAY define `source` with exactly one selector:

| Selector | Value |
| --- | --- |
| `path: true` | Entire normalized repository-relative path. |
| `path_segment: n` | Zero-based segment counted from the beginning. |
| `path_segment_rev: n` | Zero-based segment counted backward from the filename. |

Indexes MUST be non-negative and in range for every record. Source fields cannot be `integer`, `number`, `boolean`, or `object`; cannot be repeated; and cannot define a default. An entity containing a source field MUST NOT define inline data.

The checked example covers all three selectors and `$path` identity:

{{< spec-example path="testdata/specification-v1/entities/files.yaml" >}}

Derived values are inserted for reads, filtering, export, and validation. They are not written back to source files. A source field MAY be `required`, because derivation supplies its value before schema validation.
