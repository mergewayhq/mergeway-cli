# Mergeway website

This directory contains the future Mergeway website. It is a Hugo site using Hugo Book as a Hugo module. During the phased migration, `website/` is independent of the existing mdBook site in `docs/`; the mdBook build and deployment remain authoritative until the production-cutover phase.

## Pinned toolchain

- Hugo Extended `0.155.3`, supplied by the repository's locked Nix/devenv environment.
- Hugo Book `v13`, resolved to the exact pseudo-version recorded in `go.mod` and `go.sum`.

Hugo Extended is required because Hugo Book compiles SCSS. Enter the development environment with `devenv shell` (or the repository's usual Nix workflow) before running the commands below. Go is also required for Hugo module resolution.

## Build and preview

From the repository root:

```sh
make website-build
make website-serve
```

The production-like build excludes draft and future-dated content, uses the production base URL, and writes generated output to `website/public/`. The preview server prints its local URL (normally `http://localhost:1313/`) and preserves local preview links.

To remove only generated website output:

```sh
make website-clean
```

The existing site uses separate commands and files:

```sh
make docs-build
make docs-serve
```

Do not move or duplicate mdBook content from `docs/` into `website/` during the foundation phase.

## Updating dependencies

Review Hugo and Hugo Book compatibility before changing either pin. To update Hugo, update the locked Nix/devenv inputs and the supported version above. To update Hugo Book, run this from `website/` with the reviewed release tag:

```sh
hugo mod get github.com/alex-shpak/hugo-book@<release>
hugo mod tidy
hugo mod graph
```

Commit both `go.mod` and `go.sum`, then run both site builds. Never use an unqualified `@main`, `@master`, or floating update command.

## Theme overrides

Hugo Book v13 does not provide breadcrumbs. `layouts/_partials/docs/inject/content-before.html` uses the theme's supported content-injection point to render Hugo's page ancestors when `BookBreadcrumbs` is enabled. This is the only theme override in the minimal site. Empty `assets/` and `static/` override directories are intentionally omitted.
