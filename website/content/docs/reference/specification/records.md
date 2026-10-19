---
title: Records and data sources
description: Normative inline, file, multi-record, and selector source rules for Mergeway specification v1.
type: docs
mode: reference
weight: 70
bookToC: true
---

# Records and data sources

> **Normative v1.** This page applies to repositories declaring `mergeway.version: 1` and is normative.

## Entity includes

An entity `include` is an array. Each entry is either a non-empty glob string or a mapping with non-empty `path` and optional `selector`:

```yaml
include:
  - data/users/*.yaml
  - path: data/catalog.json
    selector: "$.items[*]"
```

Paths are resolved from the [repository root](../repository-configuration/#path-bases), not the configuration document. Matches are sorted. Duplicate entries with the same path and selector are collapsed. A data glob MAY match no files.

Without a selector, `.json` is decoded as JSON-compatible YAML, `.md` and `.okf` use the [OKF mapping](../open-knowledge-format/), and other suffixes are decoded as YAML. A regular document represents one record unless its top-level `items` member is an array of record maps.

## Entity association

A non-selector document MAY have a top-level `type` string. If present, it MUST equal the including entity and is removed from record fields. If absent, the including entity supplies the type.

A selector is a JSONPath expression applied to the decoded YAML or JSON value. It MUST parse, match at least one value, and each selected value MUST be a map. A selected map MAY declare matching `type` or `Type`; those keys are removed. One match is a single record and multiple matches form a multi-record source.

The checked fixture exercises an `items` source and a selector source:

{{< spec-example path="testdata/specification-v1/data/events.yaml" >}}

{{< spec-example path="testdata/specification-v1/data/catalog.json" language="json" >}}

## Record values

Each record is a map. Defined fields are validated under [field definitions and types](../fields/). A record identifier is required by [identifiers and references](../identifiers-references/). Unknown record members are preserved as data but are not rejected by version 1 schema validation.

Multi-record files and selector sources are valid read sources. The v0.10.0 mutation API refuses create, update, and delete for selector-backed records because one selected value cannot be safely rewritten in isolation. A `$path` identifier MUST have exactly one record per file.

Inline records are specified with their [entity](../entities/#inline-records). Path-derived values are injected into the in-memory record before schema validation and do not alter the source file.
