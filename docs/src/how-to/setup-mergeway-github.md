# Set Up Mergeway with GitHub

Goal: wire Mergeway into GitHub so formatting stays consistent and ownership of datasets is clearly distributed across teams.

Throughout this guide we will reference **GrainBox Market**, a fictional marketplace that stores product data in `data/products/` and category lookups in `data/categories/`. The Data Platform team maintains `mergeway.yaml`, while Inventory Operations and Category Management own their respective folders.

## Prerequisites

- A repository that already contains a valid `mergeway.yaml` and the files it references.
- GitHub Actions enabled for the repository.
- Permission to configure GitHub teams (or at least invite individual maintainers) so CODEOWNERS can route reviews correctly.

## 1. Add a Mergeway Workflow

Create `.github/workflows/mergeway-fmt.yml` to ensure every pull request keeps GrainBox data formatted:

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
  mw-fmt:
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
        run: mw fmt --lint
```

This job fails fast whenever a record under `data/products/` or `data/categories/` is out of format, ensuring reviewers only see clean diffs. Adjust the `paths` filters if your workspace stores data outside YAML.

## 2. Explain the Failure Mode to Contributors

`mw fmt --lint` prints each offending file, so GrainBox developers fix CI failures locally with `mw fmt --in-place`. Capture that reminder in your pull-request template or CONTRIBUTING guide so the workflow feels helpful rather than mysterious.

## 3. Assign Ownership with CODEOWNERS

Formatting alone is not enough—you also want the right people reviewing Mergeway changes. GitHub’s [CODEOWNERS](https://docs.github.com/articles/about-code-owners) file routes pull requests to specific teams or individuals based on path globs.

Create `.github/CODEOWNERS` with entries for both GrainBox teams and any shared files:

```
# Mergeway schema is owned by Data Platform
mergeway.yaml @grainbox/data-platform

# Data files are split by operational team
data/products/ @grainbox/inventory-ops
data/categories/ @grainbox/category-mgmt
```

Key considerations:

- Teams must exist inside your GitHub organization (for example `@org/team-slug`). If you do not have teams yet, either create them under **Settings → Teams** or list specific people such as `@alice` and `@bob`.
- You can mix teams and individuals to cover overlapping areas—for instance, keep `mergeway.yaml` owned by `@grainbox/data-platform` and also list `@lead-architect` for extra oversight.
- CODEOWNERS applies to all pull requests, so combining it with the workflow guarantees every Mergeway change is reviewed by someone who understands that slice of the dataset.

## 4. Keep Versions Predictable

GrainBox pins versions for long-lived branches by replacing `@latest` with a tag (e.g., `@v0.11.0`) or by caching the binary. Matching versions between local machines and CI prevents "works on my laptop" formatting diffs.

Once the workflow and CODEOWNERS file land in the default branch, any pull request that touches Mergeway files will:

1. Trigger the formatting check so contributors fix issues before merging.
2. Automatically request reviewers from the right team, ensuring accountability for each dataset.

That combination keeps Mergeway-managed data healthy as your GitHub organization grows.
