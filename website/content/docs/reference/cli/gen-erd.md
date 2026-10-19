---
title: mergeway-cli gen-erd
description: Render an entity-relationship diagram with Graphviz.
weight: 90
aliases: [/cli-reference/gen-erd.html, /cli-reference/gen-erd/]
---
# `mergeway-cli gen-erd`

## Synopsis
```sh
mergeway-cli [global flags] gen-erd --path <output-file>
```

`--path` is required and has no default. The extension selects Graphviz output (for example `.svg` or `.png`); with no extension, PNG is used. `dot` must be on `PATH`.

## Output, effects, and exits

Loads entity definitions, sends a generated graph to `dot`, and creates or replaces the requested file. It displays fields, data paths, references, and inheritance. Success exits `0`; missing paths, configuration errors, or Graphviz failures exit `1`.

```sh
mergeway-cli --root examples/full gen-erd --path schema.svg
```

See [generate an ERD](../../../how-to/cli/generate-erd/) and [`entity show`](../entity-show/).
