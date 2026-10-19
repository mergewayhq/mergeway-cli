---
title: Overview
description: Scope and structure of the normative Mergeway specification version 1.
type: docs
mode: reference
weight: 10
bookToC: true
---

# Overview

> **Normative v1.** This page applies to repositories declaring `mergeway.version: 1` and is normative unless marked otherwise.

Mergeway version 1 defines a file-based database. A [repository](../terminology/) combines configuration documents, entity definitions, and record sources. Configuration gives independently stored YAML, JSON, Markdown, and OKF files one logical data model.

## Scope

Version 1 specifies:

- recursive configuration loading and entity composition;
- native and JSON-Schema-derived field definitions;
- inline, single-record, multi-record, selector, and OKF data sources;
- field and path identifiers, references, and inheritance;
- format, schema, and reference validation.

It does not require a directory convention, Git, a particular editor, or a public service. Write-command behavior and tool-specific interfaces belong to their component references.

## Minimal repository

The entry configuration below delegates entity definitions to other configuration documents. The path is the checked normative example used by this specification.

{{< spec-example path="testdata/specification-v1/mergeway.yaml" >}}

Every loaded configuration document includes the same version declaration. Continue with [repository and configuration](../repository-configuration/) for resolution rules or [conformance](../conformance/) for the meaning of normative statements.
