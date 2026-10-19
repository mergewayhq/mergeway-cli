---
title: Conformance and requirement language
description: Requirement keywords and conformance classes for Mergeway specification version 1.
type: docs
mode: reference
weight: 30
bookToC: true
---

# Conformance and requirement language

> **Normative v1.** This page applies to repositories declaring `mergeway.version: 1` and is normative.

## Requirement keywords

The keywords **MUST**, **MUST NOT**, **SHOULD**, **SHOULD NOT**, and **MAY** express requirement strength as described by RFC 2119 and RFC 8174 when, and only when, shown in capitals.

- **MUST** and **MUST NOT** identify unconditional requirements.
- **SHOULD** and **SHOULD NOT** permit a departure only when its consequences are understood.
- **MAY** identifies optional behavior.

Descriptive present-tense statements on normative pages are requirements of the version 1 data model even when a keyword would add no clarity. Examples are normative for structure and behavior but their names and values are non-normative.

## Repository conformance

A version 1 repository conforms when:

1. every loaded configuration document parses and normalizes under [repository and configuration](../repository-configuration/);
2. every record passes the [format and schema rules](../validation/); and
3. every reference passes referential validation.

An empty entity data glob does not by itself make a repository non-conforming. A configuration include glob that matches no documents does.

## Implementation conformance

A conforming version 1 reader MUST accept every repository required by this specification, MUST apply the specified identity and validation rules, and MUST NOT silently reinterpret an unsupported future `mergeway.version` as version 1. An implementation MAY expose extra commands or read-only views when they do not change repository meaning.

Release v0.10.0 is the implementation baseline for this publication. Behavior introduced only after that release is outside this version of the normative text.
