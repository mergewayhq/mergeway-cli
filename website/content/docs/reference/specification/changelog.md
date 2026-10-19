---
title: Changelog and migration notes
description: Publication history and change policy for the normative Mergeway specification.
type: docs
mode: reference
weight: 130
bookToC: true
---

# Changelog and migration notes

> **Normative v1.** This page governs documentation of changes affecting repositories declaring `mergeway.version: 1`.

## 2026-10-19 — Version 1 baseline

Published the first consolidated normative specification, aligned with Mergeway release v0.10.0 and `mergeway.version: 1`. It covers the native configuration model, inheritance, reference unions, JSONPath selectors, JSON Schema interoperability, path-derived data, `$path` identifiers, OKF records, and phased validation.

Existing conforming version 1 repositories require no migration for this documentation publication.

## Change policy

This page records changes in publication order:

- A **compatible addition** extends documented version 1 behavior already accepted by released version 1 implementations, or adds an optional capability without invalidating conforming repositories. The entry names the first supporting release.
- A **clarification** resolves ambiguity without changing conforming behavior and links the affected page.
- A **deprecation** identifies behavior that remains valid, the release that introduced the notice, the preferred replacement, and any announced removal boundary.
- A **breaking change** requires a new integer `mergeway.version`, migration instructions, and an explicit compatibility matrix. It is not silently added to version 1.

Unreleased proposals MUST NOT appear as normative requirements. They may be discussed outside this section only when clearly labeled non-normative and unreleased.

## Migration-note format

Each future migration entry will state the source and target configuration versions, minimum tool release, affected constructs, mechanical steps, validation command, and rollback considerations. Version-specific routes may be introduced only when more than one normative version must be maintained; version 1 remains at this stable unversioned route.
