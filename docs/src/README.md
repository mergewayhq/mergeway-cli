# Mergeway Overview

Mergeway is a lightweight CLI that keeps metadata honest by treating schemas as code. Instead of juggling spreadsheets or custom scripts, you describe entities in YAML/JSON, run a quick validation, and catch broken references before they reach production.

## What the CLI Does

- Stores entity definitions and relationships in version-controlled files.
- Validates schemas and records so required fields and references stay consistent.
- Generates simple reports you can attach to pull requests or issues.

## Why Teams Use Mergeway

- **Fast feedback:** One command surfaces missing fields, enum mismatches, or invalid references.
- **Git-native:** Changes live in branches and pull requests, making reviews trivial.
- **Lightweight:** No server component—just a binary that runs locally or in CI.

## Where to Go Next

1. [Install Mergeway](installation/README.md) (or build from source).
2. Follow the [First Validation guide](get-started/quickstart.md).
3. Review the [Concepts](concepts/README.md) and [Schema Format](schema-spec.md) when you define entities.
4. Browse through the [CLI Reference](cli-reference/README.md) handy for the command syntax.

Updates land in the [Changelog](releases/README.md). File GitHub issues for questions, bugs, or requests.
