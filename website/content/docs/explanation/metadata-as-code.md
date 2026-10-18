---
title: Why metadata as code?
description: Understand why Mergeway stores structured records in files and applies software review practices to data changes.
type: docs
mode: explanation
weight: 10
bookToC: true
---

# Why metadata as code?

Metadata often begins as a small configuration file and grows into shared operational data: catalogs, ownership mappings, content, product definitions, or service inventories. A conventional database makes that data queryable, but its changes can become detached from the code, review, and deployment process that depends on it.

Mergeway takes the opposite starting point. YAML and JSON records live in a Git repository beside explicit entity schemas. A proposed data change is therefore an ordinary change set: it has an author, a diff, reviewers, automated checks, and recoverable history.

## What the approach provides

- **Reviewable intent.** A pull request shows which schemas and records changed and why.
- **Reproducible state.** A Git revision identifies the repository data used by a build or deployment.
- **Early feedback.** Validation catches malformed files, schema violations, duplicate identifiers, and broken relationships before merge.
- **Tool choice.** People can edit text directly, use an editor integration, automate with the CLI, or inspect through a read-only MCP client without changing the underlying format.
- **Portable output.** Export turns the file layout into a consolidated YAML or JSON snapshot for downstream tools.

## What it does not try to replace

Mergeway is not a transactional database server, query engine, or high-frequency runtime store. Git-centered review is most useful when changes should be deliberate and auditable. Applications can consume a validated export or transform it into a runtime-specific store after review.

The file layout is also not the logical model. Teams may split records across many files or mix YAML and JSON; Mergeway loads those sources through entity definitions and presents one consistent repository view.

## One model across the suite

The specification defines the model. `mergeway-cli` operates on it, `mergeway-diff` compares revisions semantically, `mergeway-lsp` reports editor feedback, and `mergeway-mcp` exposes read-only inspection. Because these tools share loading and validation behavior, a relationship that fails in the CLI should not silently mean something different in an editor or integration.

Continue with [The Mergeway repository model](../repository-model/), try [the first-repository tutorial](../../tutorials/first-repository/), or compare the [suite components](../../components/).
