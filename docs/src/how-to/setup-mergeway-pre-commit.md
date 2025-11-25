# Enforce Mergeway Formatting with pre-commit

Goal: run `mw fmt` automatically before every commit so contributors push consistently formatted GrainBox Market data.

We will keep using the fictional GrainBox Market repository from the GitHub how-to: schemas live in `mergeway.yaml`, product data sits under `data/products/`, and category lookups live in `data/categories/`.

## Prerequisites

- The Mergeway CLI (`mw`) is already installed on developer machines and in your `PATH`.
- Python 3.8+ is available (pre-commit ships via `pipx`, `pip`, or Homebrew).
- Your repo includes the Mergeway workspace you want to protect.

## 1. Install pre-commit Locally

Pick the method that matches your tooling:

```bash
pipx install pre-commit      # recommended
# or
pip install pre-commit       # inside a virtualenv
# or
brew install pre-commit      # macOS
```

Developers only need to do this once per workstation.

## 2. Configure the Hook

Add a `.pre-commit-config.yaml` file in the repo root (or extend your existing config) with a local hook that invokes `mw fmt`:

```yaml
repos:
  - repo: local
    hooks:
      - id: mergeway-fmt
        name: mergeway fmt
        entry: mw fmt --in-place
        language: system
        pass_filenames: false
        files: ^data/(products|categories)/.*\.(ya?ml|json)$
```

Why these settings?

- `entry: mw fmt --in-place` rewrites any out-of-format records before the commit proceeds.
- `pass_filenames: false` lets Mergeway discover files from `mergeway.yaml` rather than only the files staged by Git—useful when your workspace spans multiple folders.
- `files` narrows execution to the GrainBox data directories so unrelated commits (docs, code) skip the hook. Adjust the regex for your layout or remove the key to run on everything.

If you prefer CI-style failures, swap `--in-place` for `--lint`. The hook will then block the commit and print offending files without mutating them.

## 3. Install the Git Hook

Tell pre-commit to write the hook into `.git/hooks/pre-commit`:

```bash
pre-commit install --hook-type pre-commit
```

To cover pushes from automation as well, you may also install it as a `pre-push` hook:

```bash
pre-commit install --hook-type pre-push
```

Each contributor only needs to run these commands once per clone.

## 4. Test the Setup

Run the hook against every tracked file to confirm it formats data as expected:

```bash
pre-commit run mergeway-fmt --all-files
```

- If the repo already follows Mergeway’s canonical layout, the command prints `mergeway-fmt..................................Passed`.
- If output shows `Failed`, inspect the listed files, rerun `mw fmt --in-place <file>` manually if needed, then stage the changes.

Developers now get immediate feedback before commits ever leave their machines, and CI stays clean because repositories reach GitHub with consistent Mergeway formatting.
