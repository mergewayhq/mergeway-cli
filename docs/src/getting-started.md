# Getting Started with Mergeway

Last updated: 2025-10-22

Goal: scaffold a workspace, define one schema, add an object, and watch Mergeway enforce referential integrity as you expand it.

> All commands assume the `mw` binary is on your `PATH`.

## 1. Create a workspace

```bash
mkdir blog-metadata
cd blog-metadata
mw init
```

`mw init` creates the basic layout:

```
.
├── mergeway.yaml   # configuration entry point
├── data/           # object files live here
└── types/          # schema definitions live here
```

Open `mergeway.yaml` and replace its contents with:

```yaml
version: 1
include:
  - types/*.yaml
```

This tells Mergeway to load every schema stored under `types/`.

## 2. Describe the first schema

Create `types/Post.yaml`:

```yaml
entities:
  Post:
    identifier: id
    include:
      - data/posts/*.yaml
    fields:
      id:
        type: string
        required: true
      title:
        type: string
        required: true
      body: string
```

This schema maps every YAML file in `data/posts/` to a `Post`. The `id` field acts as the primary key. The `body` field uses the shorthand `body: string`, which is equivalent to the longer mapping with `required: false`.

## 3. Add the first record

Create `data/posts/launch.yaml`:

```yaml
type: Post
id: post-001
title: Launch Day
body: |
  We are excited to announce the product launch.
```

### Optional: seed inline data

If you only need a couple of seed rows, you can embed them alongside the schema by adding a `data` section:

```yaml
entities:
  Post:
    identifier: id
    include:
      - data/posts/*.yaml
    fields:
      id:
        type: string
        required: true
      title:
        type: string
        required: true
      body:
        type: string
    data:
      - id: post-inline
        title: Inline Example
        body: Inline data lives in the schema file.
```

Inline records load together with file-based data. They are intentionally read-only—commands such as `mw data update` and `mw data delete` only modify files on disk.

## 4. Inspect what Mergeway sees

List the known entities:

```bash
mw type list
```

Output:

```
Post
```

Inspect the normalized schema (use `--format json` if you prefer JSON):

```bash
mw --format yaml type show Post
```

Output (abridged):

```yaml
name: Post
filepatterns:
  - data/posts/*.yaml
fields:
  id:
    type: string
    required: true
  title:
    type: string
    required: true
  body:
    type: string
```

Validate the workspace:

```bash
mw validate
```

Output:

```
validation succeeded
```

## 5. Extend the model with references

Add a user schema (`types/User.yaml`):

```yaml
entities:
  User:
    identifier: id
    fields:
      id:
        type: string
        required: true
      name:
        type: string
        required: true
    include:
      - data/users/*.yaml
```

Update `types/Post.yaml` so each post points to a user:

```yaml
entities:
  Post:
    identifier: id
    fields:
      id:
        type: string
        required: true
      title:
        type: string
        required: true
      body:
        type: string
      author:
        type: User
        required: true
    include:
      - data/posts/*.yaml
```

Run validation again:

```bash
mw validate
```

Output:

```
- phase: schema
  type: Post
  id: post-001
  file: data/posts/launch.yaml
  message: missing required field "author"
```

Mergeway highlights that the record is missing the new `author` field.
Errors are emitted in the same format you request via `--format` (YAML by default).

## 6. Provide the referenced data

Create `data/users/alice.yaml` and update the post:

```yaml
# data/users/alice.yaml
type: User
id: user-alice
name: Alice Example
```

```yaml
# data/posts/launch.yaml
type: Post
id: post-001
title: Launch Day
author: user-alice
body: |
  We are excited to announce the product launch.
```

Validate again:

```bash
mw validate
```

Output:

```
validation succeeded
```

## 7. Explore the data

List posts and fetch a record:

```bash
mw list --type Post
mw get --type Post post-001 --format yaml
```

Example output:

```
post-001
```

```yaml
author: user-alice
body: |
  We are excited to announce the product launch.
id: post-001
title: Launch Day
```

Run a configuration check when you change schemas:

```bash
mw config lint
# configuration valid
```

Next steps:

- Use `mw create`, `mw update`, and `mw delete` to manage records from the command line.
- Review the [CLI Reference](cli-reference/README.md) for every command and flag.
- Keep the [Schema Format](schema-spec.md) page handy while evolving your types.
