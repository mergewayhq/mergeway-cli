---
title: Documentation authoring
description: Conventions and templates for writing pages in the Mergeway website.
type: docs
weight: 10
---

# Documentation authoring

The website uses Hugo Book and organizes canonical documentation by reader need. Put pages under exactly one of `tutorials`, `how-to`, `reference`, or `explanation`; component hubs orient readers and link to those canonical pages.

## Required front matter

Every page must define a useful `title` and `description`, use `type: docs`, and set a `weight` in increments of ten. Section indexes follow the same convention. Use `bookToC: true` unless a short page has no useful second-level headings. New substantive pages should also declare their `mode`.

Weights determine file-tree order. Leave gaps rather than renumbering a whole section when inserting a page.

## Page modes

- A **tutorial** gives a learner one reproducible path, observable results, and recovery guidance.
- A **how-to** solves one named task for a reader who already understands the basics.
- A **reference** page states exact, released behavior without workflow digressions.
- An **explanation** page develops concepts, rationale, and relationships.

Start a page with the corresponding Hugo archetype, for example:

```sh
cd website
hugo new content docs/tutorials/my-tutorial.md --kind tutorial
```

Replace every placeholder before setting `draft: false`. Use site-root links for stable website routes and relative links for nearby pages. Never link into the legacy `docs/src` tree from this site.

## Validation

From the repository root, run `make website-check`. It performs the production build with warnings promoted to failures, asserts required routes, aliases, search terms, and navigation order, rejects legacy mdBook-relative links, checks executable examples and migration coverage, reports orphan/duplicate content, and validates generated internal links.

Use the narrower commands while drafting:

```sh
make website-build
make website-check-content
make website-serve
```

`website/content/docs/` is the canonical Hugo source. `website/migration-ledger.md` records the disposition of every legacy public page and relevant architecture note. The legacy `docs/` mdBook remains buildable during the transition; do not maintain a second canonical copy there.

Every behavioral page should link to a useful page in another mode when the relationship exists: guides link to reference, reference links to tasks or explanation, and explanation links to an executable workflow. See [Contributing](../contributing/) for the repository-wide review process.
