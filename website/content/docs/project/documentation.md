---
title: Documentation authoring
description: Conventions and templates for writing pages in the Mergeway website.
type: docs
weight: 10
bookToC: true
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

From the repository root, run `make website-check`. It performs the production build with warnings promoted to failures, asserts the required routes and navigation order, rejects legacy mdBook-relative links, and validates generated internal links.
