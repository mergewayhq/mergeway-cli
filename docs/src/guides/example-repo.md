---
title: "Getting Started with Mergeway CLI Using the Example Repository"
linkTitle: "Example Repository Tutorial"
description: "Use the example repository to get started with Mergeway CLI"
weight: 30
---

Mergeway is a command‑line tool that helps teams describe and validate structured data alongside their code. The [mergeway‑example‑repo](https://github.com/mergewayhq/mergeway-example-repo) demonstrates how the CLI can enforce schemas, format YAML/JSON files and generate relationship diagrams using a minimal blog dataset. This guide shows how to set up Mergeway, explore the example repository and add automation so contributors always follow the rules.

## 1. Cloning the example repository and installing the CLI

Start by cloning the example repository and moving into it:

```bash
git clone git@github.com:mergewayhq/mergeway-example-repo.git
cd mergeway-example-repo
````

The repository’s **Getting started** section explains two ways to access the CLI:

* **Use your own install.** Follow the official installation instructions to install `mergeway‑cli`. Once installed you can run commands directly in the repo.
* **Use the provided dev shell.** If you use [Nix](https://nixos.org/) and optionally [direnv](https://direnv.net/), you can run `nix develop` from the root. This builds `mergeway‑cli` from source and sets up pre‑commit hooks and validation for you.

With the CLI available, try printing the version and usage information:

```bash
mergeway-cli --help    # show usage information
mergeway-cli version   # print the CLI version
```

## 2. Understanding the repository structure

The example repository models a simple blog with four entities: `User`, `Tag`, `Post` and `Comment`. Each entity stores its schema in `entities/` and the corresponding records in `data/`. The top‑level `mergeway.yaml` file ties everything together and tells Mergeway where to find schemas and data.

**Directory overview**:

| Folder          | Purpose                                                                                                                |
| --------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `entities/`     | Contains YAML schemas for each entity (`User.yaml`, `Tag.yaml`, `Post.yaml`, `Comment.yaml`).                         |
| `data/`         | Holds YAML or JSON files with records. The subfolders (`users/`, `tags/`, `posts/`, `comments/`) mirror the entities. |
| `mergeway.yaml` | Top‑level configuration referencing entity schemas and data globs.                                                    |

To see the contents of the dataset, you can run:

```bash
mergeway-cli entity list          # list all entity types
mergeway-cli entity show User     # inspect the User schema
mergeway-cli list --type Post     # list all post records
mergeway-cli get --type User 1    # fetch a single user by id
mergeway-cli validate             # validate schemas, data and references
```

These commands read `mergeway.yaml`, load each entity definition under `entities/`, and then validate all records under `data/`. Validation errors are printed in a structured format with the phase (format, schema or references) and the offending file and message.

## 3. Exploring and extending a workspace

You are not limited to the blog dataset. The Mergeway workspace guide demonstrates how to scaffold a brand‑new workspace, move inline records into separate files and add JSON sources:

1. **Create a workspace.** Make a directory and run `mergeway-cli init` to create a new `mergeway.yaml`. The file starts with an inline entity definition; you can customise it to match your data model.
2. **Use CLI commands.** List entities, show schemas, list records and validate the workspace.
3. **Externalise data.** As the table grows, move records into files under `data/` and update the `include` globs in `mergeway.yaml`. Re‑run `mergeway-cli list` and `mergeway-cli validate` to see the effect.
4. **Split schemas and add JSON.** Create an `entities/` directory and define additional entities (e.g. `Product`) that pull from JSON via JSONPath selectors. Add the JSON file under `data/` and include external schemas using the `include` key in `mergeway.yaml`.
5. **Fix reference errors.** When validation reports missing references (e.g. a `Product` record refers to a non‑existent category), update the relevant data file and re‑validate to ensure consistency.
6. **Export a snapshot.** Use `mergeway-cli export --format json --output <file>` to collect the full dataset into a single JSON snapshot.

By following these steps you can evolve the example repository or build your own dataset from scratch while preserving referential integrity.

## 4. Automating formatting and reviews with GitHub Actions

To keep data formatted and ensure the right people review changes, configure a GitHub Actions workflow and CODEOWNERS. The **Set up Mergeway with GitHub Actions** guide recommends adding a workflow file called `.github/workflows/mergeway-fmt.yml`:

```yaml
name: Mergeway Formatting

on:
  pull_request:
    branches: [main]
    paths:
      - "**/*.yaml"
      - "**/*.yml"
      - mergeway.yaml
  push:
    branches: [main]

jobs:
  mergeway-cli-fmt:
    runs-on: ubuntu-latest
    steps:
      - name: Check out repository
        uses: actions/checkout@v4

      - name: Install Go toolchain
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Install Mergeway CLI
        run: go install github.com/mergewayhq/mergeway-cli@latest

      - name: Lint Mergeway formatting
        run: mergeway-cli fmt --lint
```

This workflow runs on every pull request and push to `main`. It checks out the code, installs Go and Mergeway, then runs `mergeway-cli fmt --lint` to lint YAML and JSON files. Adjust the `paths` filters if your data lives outside of YAML files. To ensure only valid PRs are merged, add a **CODEOWNERS** file that assigns responsibility for specific paths. For example:

```
# mergeway schema is owned by the Data Platform team
mergeway.yaml @myorg/data-platform

# Data files are split by operational team
data/users/ @myorg/inventory
data/posts/ @myorg/content
```

GitHub will automatically request reviews from the specified teams when a pull request touches those files. Pinning a version of the CLI (e.g. `@v0.11.0`) or caching the binary makes CI runs reproducible.

## 5. Enforcing formatting locally with pre‑commit

Developers can catch formatting issues before pushing by using [pre‑commit](https://pre-commit.com/). The **Pre‑commit integration** guide outlines the steps:

1. **Install pre‑commit** once per workstation using `pipx install pre-commit`, `pip install pre-commit` or `brew install pre-commit`.

2. **Add a configuration file** `.pre-commit-config.yaml` with a local hook that runs `mergeway-cli fmt` on your data directories:

   ```yaml
   repos:
     - repo: local
       hooks:
         - id: mergeway-fmt
           name: mergeway fmt
           entry: mergeway-cli fmt
           language: system
           pass_filenames: false
           files: ^data/(users|tags|posts|comments)/.*\.(ya?ml|json)$
   ```

   The `entry` runs `mergeway-cli fmt` to rewrite any out‑of‑format records, `pass_filenames: false` tells Mergeway to discover files via `mergeway.yaml`, and the `files` regex limits the hook to data directories.

3. **Install the hook** into your Git repository:

   ```bash
   pre-commit install --hook-type pre-commit
   # optionally also install a pre‑push hook
   pre-commit install --hook-type pre-push
   ```

4. **Test the hook** by running it against all tracked files:

   ```bash
   pre-commit run mergeway-fmt --all-files
   ```

If the dataset already follows Mergeway’s canonical layout, the hook passes silently; otherwise it rewrites offending files or fails when using `--lint` mode. This keeps data consistent even before it reaches CI.

## 6. Next steps

With the example repository and these guides you can experiment with Mergeway, extend the schema to reflect your own domain and automate formatting and validation. Use Mergeway’s other subcommands—such as `mergeway-cli gen-erd` to generate an entity relationship diagram and `mergeway-cli export` to snapshot the dataset—to further explore the capabilities. For deeper reference, consult the CLI documentation for commands like `mergeway-cli init`, `mergeway-cli list`, `mergeway-cli validate` and others.

Mergeway encourages teams to treat their data like code: schemas live alongside source files, relationships are validated automatically and collaboration workflows ensure the right reviewers approve every change. Using the example repository as a starting point, you can adapt these patterns to your own projects.

