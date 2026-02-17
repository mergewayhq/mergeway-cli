# Go CLI Build & Test Expectations

## Toolchain

- Target Go version: 1.25.x (keep CI images and local environments aligned).
- Single Go module rooted at the repository (`go.mod` in repo root) with module-aware builds (`GO111MODULE=on`).
- Use semantic import versioning conventions; avoid replace directives unless absolutely necessary.

## Source Layout

```
repo-root/
  internal/
    config/
    data/
    validation/
    cli/
  pkg/                # optional for exported APIs if external consumption becomes necessary
  go.mod
  go.sum
  main.go
```

- Keep the top-level `main.go` minimal—parse flags and hand off to internal packages.
- Group core logic under `internal/` by domain (configuration loading, storage, validation pipeline, command orchestration).
- Add `pkg/` only when symbols need to be consumed by other modules; otherwise prefer `internal/` for encapsulation.

## Build Workflow

Use first-party Go commands (no Makefile wrappers expected):

- Format: `go fmt ./...`
- Static analysis: `golangci-lint run` (configured to enable `staticcheck` and other desired linters).
- Build: `go build .`
- Dependency hygiene: `go mod tidy` before committing changes that affect dependencies.

Document these commands in the README so contributors follow the canonical workflow.

## Linting & Analysis

- Maintain a `.golangci.yml` configuration enabling at least `staticcheck`, `govet`, `gosimple`, `unused`, and formatting checks.
- Run `golangci-lint run` locally and in CI; fail the pipeline on any lint errors.
- Keep linter runtime manageable by caching build artifacts (configure CI cache accordingly).

## Testing Strategy

- Unit tests: `go test ./...` (table-driven tests for config parsing, storage helpers, validation logic).
- Race detection: `go test -race ./...` in CI at least once per pipeline.
- Integration tests: leverage `t.TempDir()` to create temporary repository structures and execute command flows via package APIs or `exec.Command` wrappers.
- Fixtures: store canonical configs/data under `testdata/`, mirroring representative scenarios from `docs/examples/` (single objects, repeated fields, cross-type references).
- Coverage target: aim for ≥ 80% across `internal/config`, `internal/data`, `internal/validation`, and CLI orchestration packages.

## Continuous Integration Expectations

Typical CI pipeline stages:

1. `go fmt ./...` (fail if diffs are detected via `git diff --exit-code`).
2. `golangci-lint run` (with `staticcheck` enabled).
3. `go test -race ./...`.
4. `go test -cover ./...` and publish coverage results.

Use `actions/setup-go` (or equivalent) pinned to Go 1.25.x and enable module download caching.

## Release Artifacts

- Build static binaries for Linux (`amd64`, optionally `arm64`) and macOS (`amd64`, `arm64`) using environment matrices (`GOOS`, `GOARCH`).
- Output binaries as `dist/mergeway-cli_<os>_<arch>`; strip debug symbols for release builds.
- Windows support can be added later when requirements expand.

## Dependency Management

- Rely solely on Go modules (`go mod`); avoid vendoring until policy changes.
- Keep dependencies minimal; prefer the standard library. Introduce third-party packages only when necessary and include rationale in PR descriptions.
- Run `go mod tidy` and `go mod vendor` (if policy changes) in CI to guard against drift.

## Logging & Observability

- Use the standard library’s `log/slog` for structured logging.
- Provide a `--verbose` or `--log-level` flag that toggles `slog` handler levels; default to informational output with timestamps.
- Tests should assert on critical log messages when behavior depends on logging side-effects.

## Developer Experience

- Recommend installing `golangci-lint` locally (document installation instructions in README).
- Encourage the use of `direnv` or `.tool-versions` (asdf) to pin Go 1.25.x, but keep enforcement optional unless team standardizes on a tool.
- Provide VS Code/GoLand settings snippets if beneficial (e.g., enabling `staticcheck`).
