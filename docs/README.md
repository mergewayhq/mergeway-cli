# Mergeway Documentation

This directory hosts the [mdBook](https://rust-lang.github.io/mdBook/) project for Mergeway. The goal is to keep the docs lightweight so you can ship metadata changes without wading through ceremony.

## Tooling

Install mdBook (v0.4.40+) however you prefer, for example:

```bash
cargo install mdbook
```

Useful commands:

- `make docs-build` – build static output into `docs/book/`.
- `make docs-serve` – run `mdbook serve` for local preview.

## Layout

```
docs/
├── book.toml
└── arch/                   # architecture considerations (not part of the book)
└── src/
    ├── SUMMARY.md
    ├── README.md           # landing page
    ├── installation.md     # install commands + source build
    ├── getting-started.md  # first validation walkthrough
    └── concepts.md         # terminology
```

Add new pages under `docs/src/` and register them in `SUMMARY.md`.

## Legacy Content

- `docs/arch/cli-behavior.md` → `src/reference/cli/`
- `docs/arch/schema-spec.md` → `src/reference/schema-spec.md`

## Maintenance

- Follow the writing guidance in [`docs/STYLEGUIDE.md`](STYLEGUIDE.md) and reuse snippets from [`docs/templates/`](templates/).
- Every feature PR should either update docs or state why no change is needed.
- Once a quarter, skim Overview, Installation, Getting Started, Concepts, Reference, and Troubleshooting to ensure they still match the CLI. Track follow-up work in GitHub issues labeled `docs-review`.

## License

Documentation and supporting tooling are distributed under the same [MIT License](../LICENSE.md) as the CLI.
