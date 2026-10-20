---
title: The Mergeway repository model
description: Learn how repositories, workspaces, entities, records, identifiers, files, and relationships fit together.
type: docs
mode: explanation
weight: 20
bookToC: true
---

# The Mergeway repository model

Mergeway separates the logical data model from the files used to store it. Four terms keep that distinction clear.

## Repository

A **repository** is the Git or project boundary containing Mergeway configuration, entity schemas, and record files. Documentation uses this term for the versioned body of data people review and share.

The configuration entry point is normally `mergeway.yaml`. A repository can keep definitions inline or include other files. Record sources can use YAML or JSON and can be divided into directories that make sense to maintainers.

## Workspace

A **workspace** is a repository root detected or selected by a running tool. The distinction matters when an editor has several folders open, when `--root` selects a directory, or when a service resolves files to their owning root. In a simple terminal session, the repository and active workspace are often the same directory.

## Entity

An **entity** is a schema-defined kind, such as `User`, `Post`, or `Service`. Its definition names an identifier, declares fields and constraints, and selects the files that contain its records.

An entity is not a storage directory. One entity can load several globs and data formats, while the configuration controls how those physical sources become one logical collection.

## Record

A **record** is one data instance of an entity. A `Post` record might have the identifier `post-001`, while a `User` record might have `user-alice`. Existing CLI command names such as `create`, `get`, and `update` operate on records even where older internal help text uses “object.” In JSON, the word object retains its normal format-specific meaning.

## Identifiers and relationships

Every entity has an identifier rule. Identifiers distinguish its records and provide stable targets for relationships. A field whose type names another entity holds that entity's identifier; for example, a `Post.author` value of `user-alice` relates the post to that `User` record. Repeated relationship fields can hold several identifiers.

This is why Mergeway loads the repository as a connected model rather than validating each file independently. A field can be syntactically valid and match its local schema while still naming a missing related record.

## Physical files and logical data

Includes connect the two layers:

```text
mergeway.yaml
    -> entity definitions
        -> YAML or JSON record sources
            -> one logical set of entities, records, and relationships
```

Commands such as `files` expose physical sources. Commands such as `list`, `get`, validation, export, Diff, MCP, and LSP operate on the logical model appropriate to their purpose. The Python SDK reaches the same model through the CLI's JSON interface.

Try the model in [Build your first Mergeway repository](../../tutorials/first-repository/), learn [how validation traverses it](../validation-model/), or visit the [Specification hub](../../components/specification/).
