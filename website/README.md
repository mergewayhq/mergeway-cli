# Mergeway website

This directory contains the future Mergeway website. It is a Hugo site using Hugo Book as a Hugo module. During the phased migration, `website/` is independent of the existing mdBook site in `docs/`; the mdBook build and deployment remain authoritative until the production-cutover phase.

## Pinned toolchain

- Hugo Extended `0.161.1`, supplied by the repository's locked Nix/devenv environment.
- Hugo Book `v13`, resolved to the exact pseudo-version recorded in `go.mod` and `go.sum`.
- Lychee `0.23.0` and curl, supplied by the repository's locked Nix/devenv environment, for deterministic internal-link validation through a loopback preview.

Hugo Extended is required because Hugo Book compiles SCSS. Enter the development environment with `devenv shell` (or the repository's usual Nix workflow) before running the commands below. Go is also required for Hugo module resolution.

## Build and preview

From the repository root:

```sh
make website-build
make website-check
make website-serve
```

The production-like build excludes draft and future-dated content, uses the production base URL, promotes Hugo warnings to failures, and writes generated output to `website/public/`. `website-check` also asserts required routes, legacy aliases, search terms, and navigation order; rejects links into the legacy mdBook source tree; runs the first-repository tutorial against a temporary copy of `examples/full`; and runs Lychee against every generated HTML route through a loopback preview server. External links are excluded so the result does not depend on network availability or third-party state. The preview server prints its local URL (normally `http://localhost:1313/`) and preserves local preview links.

To remove only generated website output:

```sh
make website-clean
```

The existing site uses separate commands and files:

```sh
make docs-build
make docs-serve
```

The Hugo content tree is a migration target, not a second copy of `docs/`. Do not link from Hugo pages into `docs/src`; migrate and verify a page in its assigned phase before linking to its Hugo route.

## Navigation and authoring

Hugo Book generates the primary sidebar from `content/docs`. Every section has a non-empty `_index.md`, and `weight` values use increments of ten. Home appears before the file tree, Components is the first tree section, and GitHub, Releases, Contributing, and the license appear in the auxiliary menu after it.

Baseline front matter and page-mode guidance live in [the documentation authoring page](content/docs/project/documentation.md). Hugo archetypes for `tutorial`, `how-to`, `reference`, and `explanation` pages are in `archetypes/`. These templates are drafts by default so unfinished pages cannot appear as empty navigation nodes.

The site intentionally uses Hugo Book's bundled `favicon.svg` placeholder until project identity assets are reviewed. Site metadata identifies the repository and MIT license without adding a theme override.

## Updating dependencies

Review Hugo and Hugo Book compatibility before changing either pin. To update Hugo, update the locked Nix/devenv inputs and the supported version above. To update Hugo Book, run this from `website/` with the reviewed release tag:

```sh
hugo mod get github.com/alex-shpak/hugo-book@<release>
hugo mod tidy
hugo mod graph
```

Commit both `go.mod` and `go.sum`, then run both site builds. Never use an unqualified `@main`, `@master`, or floating update command.

## Theme overrides

The site has three narrow Hugo Book overrides:

- `layouts/_partials/docs/inject/content-before.html` uses the theme's supported content-injection point to add breadcrumbs, which Hugo Book v13 does not provide.
- `layouts/baseof.html` replaces the theme's deprecated Hugo language accessors with their current equivalents so warning-strict builds remain clean.
- `layouts/_partials/docs/brand.html` replaces the theme's deprecated cross-site accessor; this single-language site always links the brand to its own home page.

No styling or component templates are overridden. Empty `assets/` and `static/` override directories are intentionally omitted.
