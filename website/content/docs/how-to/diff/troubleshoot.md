---
title: Troubleshoot semantic diff
description: Diagnose revision resolution, invalid repositories, empty changes, and automation failures.
type: docs
mode: how-to
weight: 50
bookToC: true
---

# Troubleshoot semantic diff

## A revision cannot be resolved

Run `git rev-parse <revision>` in the same repository. `mergeway-diff` accepts Git revisions and the working-tree sentinel documented in the [reference](../../../reference/diff/); it does not interpret arbitrary filesystem paths as revisions.

## A snapshot cannot be loaded

Validate the relevant checked-out state with `mergeway-cli validate`. Diff must load configuration and records at both endpoints before comparing them. Missing configuration, malformed sources, or an invalid identifier can prevent snapshot construction.

## The result is empty

Confirm that the changed file is matched by an entity include and that its selected record data changed. Formatting-only edits, key order, and a backing-file move that preserves logical identity need not become a semantic record change. Read [Semantic diff versus file diff](../../../explanation/semantic-diff/) for the model.

## JSON automation fails

Use the JSON mode and schema described in [Consume JSON diff output](../json-automation/). Treat a non-zero process status as failure before parsing output, and do not parse the human-readable renderer.

For known exit behavior and snapshot rules, use the complete [`mergeway-diff` reference](../../../reference/diff/).
