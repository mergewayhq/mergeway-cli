# CLI Behavior

## Purpose

Capture the expected surface of the `mw` command-line tool used to manage the file-backed database.

## Global Flags

- `--root <path>`: filesystem root containing `mergeway.yaml` (defaults to current directory).
- `--config <path>`: override the configuration entry point (defaults to `<root>/mergeway.yaml`).
- `--format <yaml|json>`: preferred payload format for command output (default `yaml`).
- `--fail-fast`: when supplied to validation commands, stop on the first error instead of aggregating.
- `--yes/-y`: auto-confirm prompts for destructive operations.

## Command Groups

### 1. Initialization

`mw init [--root <path>]`

- Ensure `mergeway.yaml` exists (creating it from a commented template when missing). You can build out additional folders manually if your workflow benefits from them.
- Never overwrite existing files unless `--force` is specified (future enhancement).

### 2. Type Introspection

- `mw type list`
  - Lists type identifiers defined in the configuration.
- `mw type show <type>`
  - Prints the effective schema for a type, including identifier metadata, field definitions, and file patterns pulled from the entity definition.
  - Accepts `--format` to output JSON or YAML.

### 3. Object CRUD

All object-focused commands require `--type <type>` unless the object payload embeds `type`.

- `mw list --type <type> [--filter <expr>]`
  - Streams object identifiers, optionally filtered by simple expressions (e.g., `status=active`).
- `mw get <id> --type <type>`
  - Emits the object document in the chosen format.
- `mw create --type <type> --file <path> [--id <id>]`
  - Creates a new object from a payload file or STDIN.
  - Validates format/schema before writing unless `--skip-validate` is provided.
  - Generates an identifier when the type definition flags `identifier.generated: true` and `--id` is omitted.
- `mw update --type <type> --id <id> --file <path>`
  - Replaces the stored document with the provided payload (use `--merge` for deep merges).
  - Supports partial updates with `--merge` to deep-merge fields instead of replacing wholesale.
- `mw delete <id> --type <type>`
  - Removes an object file or deletes the entry from a multi-object file; prompts for confirmation unless `--yes` is present.

### 4. Batch Operations (Optional Enhancements)

- `mw apply --dir <path>`
  - Applies all object definitions within a directory, respecting CRUD semantics per object (`state: present|absent`).
  - Useful for automation; designed to run validation before persisting changes.

### 5. Validation

`mw validate [<path>...]`

- Walks objects referenced by the configuration (or specific paths) and runs the three validation phases.
- Options:
  - `--phase <format|schema|references>` may be repeated to scope validation to selected phases (defaults to all).
  - `--fail-fast` triggers early exit on first error; otherwise the CLI aggregates and reports every violation with file context.
  - `--summary` prints a concise success/failure report; default output enumerates errors grouped by phase.

### 6. Configuration Utilities

- `mw config lint`
  - Validates configuration files only (ensures includes resolve, schema definitions are consistent, etc.).
- `mw config export --type <type> [--format json]`
  - Emits derived JSON Schema for the requested type, enabling tooling integration.

### 7. Metadata

- `mw version`
  - Prints CLI build information (semantic version, git commit, build timestamp) in the selected format.

## Input and Output Formats

- Commands that accept payloads (`create`, `update`, `apply`) read from `--file` or STDIN; the CLI infers format from extension or `--format` flag.
- Output defaults to YAML; specify `--format json` for JSON.
- Validation errors render as structured YAML/JSON objects with keys: `phase`, `type`, `id`, `file`, `message`.

## Validation Workflow

1. CLI loads configuration (resolving includes/globs).
2. For each type, the CLI locates object files using `include`.
3. Validation phases execute sequentially: format → schema → references.
4. Errors collated unless `--fail-fast` is set.

## Extensibility Considerations

- Subcommands should be implemented as discrete modules to allow future plug-ins (e.g., migrations once versioning is added).
- All commands must operate offline on local files to honor the mergeway-cli scope; remote backends are out-of-scope but future-compatible via the `--root` abstraction.
