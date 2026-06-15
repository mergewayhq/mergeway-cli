# Path Segments Example

This example shows how `mergeway.yaml` can define read-only fields whose values are derived from the backing file path, even when the entity uses a regular field identifier instead of `identifier: $path`.

## What it demonstrates
- Field-based identifiers (`slug`) with records stored in nested directories
- Declared derived fields populated from the file path at read time
- Forward path segments such as `section=guides`
- Reverse path segments such as `filename=install.yaml`
- Full relative file paths such as `relative_path=data/library/guides/install.yaml`

## Layout
- `examples/path-segments/mergeway.yaml` defines the `Page` entity
- `examples/path-segments/data/library/guides/*.yaml` stores guide pages
- `examples/path-segments/data/library/reference/*.yaml` stores reference pages

## Try it

List every page:

```bash
mergeway-cli --root examples/path-segments list --type Page
```

Filter by the directory below `data/library/`:

```bash
mergeway-cli --root examples/path-segments list --type Page --filter 'section=guides'
```

Filter by the derived filename:

```bash
mergeway-cli --root examples/path-segments list --type Page --filter 'filename=install.yaml'
```

Inspect one object with the declared derived fields included:

```bash
mergeway-cli --root examples/path-segments get --type Page guide-install
```

Export the dataset as JSON:

```bash
mergeway-cli --root examples/path-segments --format json export Page
```

## Key detail

For `data/library/guides/install.yaml`, Mergeway derives these declared fields at read time:

```yaml
section: guides
filename: install.yaml
relative_path: data/library/guides/install.yaml
```

These fields are read-only metadata. They can be used for lookups and exports, but Mergeway does not write them into the source files.
