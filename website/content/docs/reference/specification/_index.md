---
title: Specification
description: Normative version 1 contract for Mergeway repositories, configuration, records, and validation.
type: docs
mode: reference
weight: 10
bookToC: true
---

# Mergeway specification

> **Normative v1.** This section specifies repositories that declare `mergeway.version: 1`. It describes behavior implemented by release v0.10.0.

This route is the stable home of the Mergeway specification. The version belongs to the configuration contract, not the URL; no parallel `/v1/` route exists.

## Contents

- [Overview](overview/) defines scope and the document map.
- [Terminology](terminology/) defines the terms used by every normative page.
- [Conformance and requirement language](conformance/) explains **MUST**, **SHOULD**, and implementation conformance.
- [Repository and configuration](repository-configuration/) defines discovery, configuration documents, and includes.
- [Entity definitions](entities/) defines entities, inline records, and inheritance.
- [Field definitions and types](fields/) defines the native field model and constraints.
- [Records and data sources](records/) defines supported files, selectors, and record shapes.
- [Identifiers and references](identifiers-references/) defines record identity and relationships.
- [Validation and error model](validation/) defines the three validation phases.
- [JSON Schema interoperability](json-schema/) defines the supported import and export subset.
- [Path-derived data](path-derived-data/) defines `$path` identifiers and synthetic path fields.
- [Open Knowledge Format mapping](open-knowledge-format/) defines `.md` and `.okf` records.
- [Changelog and migration notes](changelog/) records specification changes.

The examples rendered in this section come from a fixture validated by the website checks. They are normative where the surrounding text says **MUST** or **MAY**; names and example values are non-normative.
