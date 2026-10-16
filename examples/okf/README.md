# Open Knowledge Format example

This workspace loads [Open Knowledge Format (OKF)](https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md) concepts from both Markdown (`.md`) and OKF (`.okf`) files. Their frontmatter `type` maps to the `Knowledge` entity; `id` is derived from the file path through the entity's `id.source.path` definition. The Markdown content after the closing frontmatter delimiter is exposed as the `body` field.

Run:

```sh
mergeway-cli --root examples/okf validate
mergeway-cli --root examples/okf --format json get --type Knowledge customer-orders.md
mergeway-cli --root examples/okf --format json get --type Knowledge order-returns.okf
```
