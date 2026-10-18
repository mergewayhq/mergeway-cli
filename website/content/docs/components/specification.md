---
title: Specification
description: The shared repository and data model implemented by the Mergeway tools.
type: docs
weight: 10
bookToC: true
---

# Specification

The Mergeway specification defines how configuration, entity schemas, records, identifiers, references, and validation fit together in a file-based database. It is the contract shared by the CLI, Diff, MCP, LSP, and editor integrations.

## Who it is for

Schema authors, tool integrators, and maintainers use the specification when they need exact repository behavior rather than a workflow. Most first-time users should begin with the tutorial or repository-model explanation.

## Purpose and boundaries

The specification defines the interoperable file and data model. It does not prescribe a particular directory layout beyond configured paths, a Git workflow, an editor, or a runtime database architecture.

## Relationship to the suite

All released binaries implement or consume the model: the CLI manages and validates it, Diff compares it, MCP exposes read-only views, and LSP supplies editor feedback. The VS Code integration connects to LSP rather than defining another format.

## Availability and compatibility

Specification version 1 is implemented by released tools for repositories declaring `mergeway.version: 1`. The complete reviewed normative reference is assigned to the next documentation phase; until it is published here, released behavior and the existing production documentation remain authoritative. Unreleased proposals are not specification requirements.

## Start here

- [Build your first Mergeway repository](../../tutorials/first-repository/) to use a version 1 repository.
- Read [The Mergeway repository model](../../explanation/repository-model/) for the public vocabulary.
- Read [The validation model](../../explanation/validation-model/) to understand conformance checks.
