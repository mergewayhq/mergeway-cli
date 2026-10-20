---
title: Mergeway
description: Explore the Mergeway specification and tools for working with a file-based database.
layout: landing
params:
  bookBreadcrumbs: false
---

<img class="landing-logo" src="/mergeway-logo.svg" alt="Mergeway logo">

# Mergeway {anchor=false}

Mergeway is a specification and suite of tools for reading and modifying a file-based database.  
Schemas and records stay in ordinary YAML and JSON files, so teams can review data changes in Git while Mergeway checks structure and relationships.

{{< button href="/docs/how-to/install/" class="btn-primary">}}Install mergeway-cli{{< /button >}}
{{< button href="/docs/tutorials/first-repository/" >}}Build your first repository{{< /button >}}

## One repository model, several ways to work

Use `mergeway-cli` to validate, inspect, modify, format, and export records. Add `mergeway-diff` for semantic reviews, `mergeway-python` for typed Python automation, `mergeway-mcp` for read-only AI-tool access, or `mergeway-lsp` for editor diagnostics and navigation. Every component works with the same repository model defined by the Mergeway specification.

{{% columns %}}

- ### Define and validate {anchor=false}
  Use the [version 1 specification](/docs/reference/specification/) and `mergeway-cli` to model, inspect, validate, and export repositories.

- ### Review changes {anchor=false}
  Use [`mergeway-diff`](/docs/components/diff/) to compare repository states as records and relationships instead of raw text.

- ### Connect tools {anchor=false}
  Give AI clients read-only access through [`mergeway-mcp`](/docs/components/mcp/) or editor clients diagnostics and navigation through [`mergeway-lsp`](/docs/components/lsp/).

- ### Automate with Python {anchor=false}
  Use the [`mergeway` Python package](/docs/components/python/) to generate typed models and automate repository operations.

{{% /columns %}}

## Start here
- [Build your first Mergeway repository](/docs/tutorials/first-repository/) with a fixed, tested example.
- [Install the Mergeway tools](/docs/how-to/install/) for your platform and workflow.
- [Understand the repository model](/docs/explanation/repository-model/) before adapting Mergeway to existing data.
- [Choose a component](/docs/components/) when you already know the kind of integration you need.
