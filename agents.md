# Repository Context & Agent Best Practices

## Project Overview

- **Language:** Go (module `github.com/mergewayhq/mergeway-cli`, currently built against Go 1.24.x)
- **Purpose:** File-based database with CLI (`mw`) supporting CRUD, validation, and configuration export.
- **Structure Highlights:**
  - `internal/config`: Modular configuration loader (`model.go`, `raw.go`, `load.go`, `normalize.go`).
  - `internal/data`: Modular store layer (`model.go`, `store_ops.go`, `load.go`, `fileio.go`, `util.go`).
  - `internal/validation`: Modular validation pipeline (`types.go`, `validate.go`, `collect.go`, `schema.go`, `field.go`, `references.go`, `parse.go`, `util.go`).
  - `internal/cli`: CLI entrypoints split by command (`root.go`, `init.go`, `type.go`, `object.go`, `validate.go`, `config_cmd.go`, `helpers.go`).
  - `examples/`: Reference datasets used by tests and e2e workflows.
  - `scripts/`, `Makefile`, `.github/workflows/ci.yml`: Tooling for fmt, lint, test, release binaries.

## Tooling & Commands

- **Formatting:** `make fmt`, `make fmt-check`
- **Lint:** `make lint` (`golangci-lint run`). Installed in `pre-commit` (`pre-commit install`) so lint fires before every commit.
- **Tests:** `make test`, `make race`, `make coverage`; coverage inspect via `go tool cover -func=coverage.out` or HTML report.
- **CI:** GitHub Action at `.github/workflows/ci.yml` mirrors the make targets.
- **Release:** `make release` to produce binaries under `dist/` for linux/darwin (amd64/arm64).

## Agent Workflow Guidelines

1. **Stay Modular:** Match the existing refactoring style—prefer smaller files with clear responsibilities over monoliths; reuse the established file breakdowns when adding new behaviors.
2. **Respect Existing Patterns:** Use helper packages (`config`, `data`, `validation`) rather than duplicating logic.
3. **Cache Constraints:** When invoking Go tooling in this sandbox, set `GOMODCACHE` and `GOCACHE` to local `.cache/...` directories to avoid permission issues.
4. **No Network Installs:** `go mod tidy` may fail due to blocked proxy access; note the failure rather than retrying endlessly.
5. **Tests & Lint:** Always run `go test ./...` (with the cached env vars) and `make fmt-check` after structural changes. Keep the `pre-commit` hook (`golangci-lint`) installed so commits fail fast on lint errors.
6. **Docs Sync:** If you add/modify examples or CLI behavior, update docs (`README.md`, `docs/*.md`) and ensure e2e scripts still reflect reality. Any CLI-visible flag/command change must be mirrored in `docs/src/cli-reference/`.
7. **Clean Diff:** Remove generated outputs (`bin/`, `dist/`, `coverage.out`, etc.) before finalizing work; `.gitignore` already covers these.
8. **Communication:** Document notable changes in `agents.md`, README, and docs when workflow or architecture updates occur.

## Testing Data & Examples

- Example dataset under `examples/` validated via unit tests in `internal/data/examples_test.go` and `internal/validation/examples_test.go`.
- `e2e_test/mw_e2e_test.sh` exercises CLI flows; keep it up-to-date with new features.

## Future Enhancements to Watch

- API integration for repo changes (TODO in README).
- Potential Go 1.25 upgrade once toolchain available (currently locked to 1.24.x due to environment restrictions).

Following these practices ensures consistent, maintainable contributions aligned with the project’s structure.
