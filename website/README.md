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
make website-release-candidate
make website-serve
```

The production-like build excludes draft and future-dated content, uses the production base URL, promotes Hugo warnings to failures, and writes generated output to `website/public/`. `website-check` also asserts required routes, legacy aliases, search terms, and navigation order; rejects links into the legacy mdBook source tree; checks the first-repository tutorial and normative specification fixture; derives CLI help coverage and executes CLI workflows against temporary copies of `examples/full`; verifies advanced repository fixtures, migration-ledger coverage, canonical-page rendering, and duplicate-content hygiene; verifies MCP and LSP protocol sessions against real binaries; checks basic HTML/accessibility structure and release-artifact hygiene; crawls every retained or redirected legacy route against the built artifact; and runs Lychee against every generated HTML route. External links are excluded so the result does not depend on third-party state. `website-release-candidate` runs that complete check and creates the ignored, deterministic `mergeway-website-release-candidate.tar.gz` archive uploaded by CI. The preview server prints its local URL (normally `http://localhost:1313/`) and preserves local preview links.

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

The disposition of every legacy public mdBook source and relevant architecture note is recorded in [the migration ledger](migration-ledger.md). The exact published-route compatibility contract is [the legacy route inventory](legacy-routes.tsv), and [the release-candidate runbook](release-candidate.md) defines preview, deployment, rollback, and content-freeze procedure. Update these records when a source or public URL changes.

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

The site has six narrow Hugo Book overrides:

- `layouts/_partials/docs/inject/content-before.html` uses the theme's supported content-injection point to add breadcrumbs, which Hugo Book v13 does not provide.
- `layouts/_partials/docs/brand.html` keeps the brand linked to this single-language site's home page and uses a non-heading container so navigation chrome does not precede the page's `h1` in the document outline.
- `layouts/404.html` keeps Hugo Book's not-found presentation but repairs its skipped heading level and gives the return link a descriptive name.
- `layouts/_partials/docs/toc.html` removes Hugo's fixed `TableOfContents` ID because the responsive theme renders the table of contents twice, which would otherwise create duplicate IDs.
- `layouts/_partials/docs/header.html` renders the repeated mobile page label as styled text rather than a heading before the page's canonical `h1`.
- `assets/styles/custom.css` guarantees a visible keyboard focus indicator and disables non-essential motion when the operating system requests reduced motion.

No other styling or component templates are overridden.
