# Full Example

A complete Mergeway sample that stitches together multiple entity files, shared data directories, and cross-entity relationships for a small blogging system.

## What it demonstrates
- Modular config: root `mergeway.yaml` includes entity definitions from `entities/*.yaml`
- Mixed data formats: YAML and JSON records pulled from `data/` via glob patterns
- Rich relationships: `Post` references `User` authors and `Tag`s; `Comment` links back to both `Post` and `User`
- Descriptions on fields and entities to document domain intent

## Diagram
![ERD](./erd.png)
