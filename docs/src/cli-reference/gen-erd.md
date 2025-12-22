# `mw gen-erd`

Generates an Entity Relationship Diagram (ERD) of your data model.

```bash
mw gen-erd --path <output-file>
```

This command inspects your configuration and generates a visual representation of your entities and their relationships. It relies on [graphviz](https://graphviz.org/) (specifically the `dot` command) to produce the output image.

## Arguments

| Argument | Description |
| --- | --- |
| `--path` | **Required.** The path where the generated image will be saved. The file extension determines the output format (e.g., .png, .svg). |

## Examples

Generate a PNG image of your schema:

```bash
mw gen-erd --path schema.png
```

Generate an SVG:

```bash
mw gen-erd --path schema.svg
```

## Requirements

The `dot` executable from [Graphviz](https://graphviz.org/) must be installed and available in your system's PATH.
