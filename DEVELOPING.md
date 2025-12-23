# Developing Mergeway CLI

Use this guide when you are extending the CLI itself. For installation and everyday usage, stay with the main README.

## Prerequisites

### Required

- **Go 1.24.x** (the project targets Go 1.24.x)
- **devenv** (optional, but recommended for a consistent development environment via `devenv shell`)

### Optional

- [`golangci-lint`](https://golangci-lint.run/) v2.4.0 or newer if you plan to run linting locally (CI uses v2.4.0)
- [mdbook](https://rust-lang.github.io/mdBook/) for building and serving documentation locally

## Development Environment

### Using devenv (Recommended)

The project provides a devenv.sh shell which sets up a complete development environment with all required dependencies:

```bash
# Enter the development shell
devenv shell

# Or use direnv for automatic environment activation
# (ensure .envrc is already present)
direnv allow
```

The devenv.sh environment includes:
- Go 1.24.x
- golangci-lint
- mdbook and mdbook-gitinfo
- graphviz (for ERD generation)
- pre-commit
- shellcheck

### Without devenv

Install the prerequisites manually:
- Install Go 1.24.x
- Install golangci-lint v2.4.0 or newer
- Install mdbook if working on documentation
- Install graphviz if using the `gen-erd` command

## Project Layout

- `main.go` is the CLI entrypoint that wires flags into internal packages
- `internal/` contains shared packages that power metadata handling and integrity checks
- `pkg/` contains public packages
- `examples/` stores sample configuration and data for demos and local tests
- `docs/` captures design notes and behavioral specs
- `e2e_test/` contains end-to-end tests

## Common Tasks

### Formatting

- `make fmt` rewrites all Go sources with `gofmt`
- `make fmt-check` verifies formatting and is enforced in CI

### Linting

- `make lint` runs `golangci-lint run` (CI uses golangci-lint v2.4.0)

### Testing

- `make test` runs the unit suite
- `make race` re-runs tests with the race detector enabled
- `make coverage` generates a coverage report
  - Coverage profile is written to `coverage.out`
  - HTML report is automatically generated at `coverage.html`

### Building

- `make build` produces `bin/mw` with version metadata injected via `-ldflags`
- The build includes Git commit hash, semantic version, and build timestamp

### Documentation

- `make docs-build` builds the documentation site using `mdbook`
- `make docs-serve` starts a local server to preview documentation

### All-in-one CI Check

- `make ci` aggregates fmt-check, lint, test, race, and coverage — matching the GitHub Actions workflow

### Clean

- `make clean` removes build artifacts (`bin/`, `dist/`, `coverage.out`, `.cache/`)

## Version Metadata

Version information is managed in `internal/version/version.go`:

- The `Number` variable contains the semantic version (e.g., `x.y.z-dev`) for the flake build
- Update this variable before cutting a release
- Build tooling injects Git commit hash and build timestamp via `-ldflags`
- Run `mw version` to see the current version, commit, and build date

## Publishing

The project uses [GoReleaser](https://goreleaser.com/) for automated releases:

1. Update the version in `internal/version/version.go` (e.g., change `x.y.1-dev` to `x.y.2-dev`)
2. Commit your changes
3. Tag the repo with `v<major>.<minor>.<patch>`:
   ```bash
   git tag vX.Y.Z
   git push origin vX.Y.Z
   ```
4. GitHub Actions automatically:
   - Builds binaries for Linux, macOS, and Windows on `amd64` and `arm64`
   - Creates archives (`.tar.gz` for Unix, `.zip` for Windows)
   - Generates checksums
   - Publishes the release with all assets
   - The version number is overwritten via `ldflags` in the release build, using the tag version

See `.github/workflows/release.yml` and `.goreleaser.yaml` for configuration details.
