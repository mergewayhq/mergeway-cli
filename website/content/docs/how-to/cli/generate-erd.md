---
title: Generate an entity-relationship diagram
description: Install Graphviz and render a repository model as SVG.
weight: 70
---
# Generate an entity-relationship diagram

Install Graphviz so its `dot` executable is on `PATH`, then validate and render an SVG:

```sh
dot -V
mergeway-cli config lint
mergeway-cli gen-erd --path schema.svg
```

The extension controls Graphviz's output format. SVG works well in documentation and code review; PNG is convenient for presentations. Regenerate the file whenever definitions change.

![Entity-relationship diagram for the full example, showing Comment, Post, Tag, and User with their reference arrows.](/images/cli/erd-example-full.svg)

The diagram shows normalized entities, direct fields, configured data paths, reference edges, and inheritance. It does not show individual records or validate them. Consult [`gen-erd`](../../../reference/cli/gen-erd/) for errors and the [entity specification](../../../reference/specification/entities/) for model semantics.
