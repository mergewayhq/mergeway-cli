---
title: Repository and configuration
description: Normative repository roots, configuration documents, versions, and include resolution for specification v1.
type: docs
mode: reference
weight: 40
bookToC: true
---

# Repository and configuration

> **Normative v1.** This page applies to repositories declaring `mergeway.version: 1` and is normative.

## Entry and version

An implementation loads one YAML entry configuration. The conventional path is `mergeway.yaml`, but callers MAY select another path. The repository root is supplied separately and is the base for entity data paths.

Every configuration document MUST contain:

```yaml
mergeway:
  version: 1
```

Version `1` is the only version defined here. A missing `mergeway` block, missing version, or other version is a configuration error.

## Top-level members

A configuration document has these version 1 members:

| Member | Value | Meaning |
| --- | --- | --- |
| `mergeway.version` | integer `1` | Selects this specification. |
| `include` | array of glob strings | Loads more configuration documents. |
| `entities` | map | Defines entities by name. |

The checked entry document is:

{{< spec-example path="testdata/specification-v1/mergeway.yaml" >}}

## Configuration includes

Each top-level `include` entry is a glob resolved relative to the document containing it. Matches are sorted before loading. Every pattern MUST match at least one file. Includes are recursive; cycles are errors.

All included documents MUST declare version `1`. An entity name MUST be defined exactly once across the resolved configuration graph. Repeating a name in two documents is an error rather than an override or merge.

## Path bases

Configuration includes and an entity's `json_schema` path are configuration-relative. Entity record `include` paths are repository-relative. These bases are intentionally different.

An entity data path MAY contain `..` and resolve outside the repository root for reading. Such paths preserve `..` in repository-relative identity. Tool-specific write commands may refuse to modify records outside the root; that safety policy does not change their validity as readable records.

See [entity definitions](../entities/) for the entity map and [records and data sources](../records/) for entity include semantics.
