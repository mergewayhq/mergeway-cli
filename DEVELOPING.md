# Developing Mergeway CLI

Use this guide when you are extending the CLI itself. For installation and everyday usage, stay with the main README.

## Prerequisites

- Go 1.24.x (the project targets Go 1.24)
- [`golangci-lint`](https://golangci-lint.run/) v1.55 or newer if you plan to run linting locally

## Project Layout

- `main.go` is the CLI entrypoint that wires flags into internal packages
- `internal/` contains shared packages that power metadata handling and integrity checks
- `examples/` stores sample configuration and data for demos and local tests
- `docs/` captures design notes and behavioral specs

## Common Tasks

### Formatting

- `make fmt` rewrites all Go sources with `gofmt`
- `make fmt-check` verifies formatting and is enforced in CI

### Linting

- `make lint` runs `golangci-lint run`

### Testing

- `make test` runs the unit suite
- `make race` re-runs tests with the race detector enabled
- `make coverage` writes the coverage profile to `coverage.out`
- Inspect coverage via `go tool cover -func=coverage.out` or open the HTML report with `go tool cover -html=coverage.out -o coverage.html`

### Building

- `make build` produces `bin/mw`

### Release Binaries

- `make release` or `scripts/build_binaries.sh` creates macOS and Linux archives in `dist/` for `amd64` and `arm64`

### Version Metadata

- `mw version` prints semantic version, commit, and build timestamp for any built binary
- Update `internal/version/version.txt` before cutting a release; build tooling injects Git metadata via `-ldflags`

### Publishing

- Commit your changes and bump `internal/version/version.txt`
- Tag the repo with `v<major>.<minor>.<patch>` (for example `git tag v0.1.0 && git push origin v0.1.0`)
- GitHub Actions builds signed binaries, uploads release assets, and publishes the release named after the tag

### All-in-one CI Check

- `make ci` aggregates fmt-check, lint, test, race, and coverage — matching the GitHub Actions workflow

## Troubleshooting

- Run `go env` to confirm your Go toolchain matches the supported version
- Use `GOFLAGS=-count=1` when tests depend on freshly generated data
- Treat `.github/workflows/ci.yml` as the ground truth for required checks
