# Website release-candidate runbook

Status: ready for independent review; production remains unchanged  
Content freeze began: 2026-10-19  
Production base URL: `https://mergewayhq.github.io/`

This runbook governs the Phase 11 artifact. Until Phase 12 is approved, the `build-and-deploy` job continues publishing mdBook and the Hugo artifact is review-only.

## Frozen inputs and route inventory

Two public surfaces exist and must not be conflated:

- The organization site at `https://mergewayhq.github.io/` is built by `mergewayhq/mergewayhq.github.io`. Its sitemap was fetched on 2026-10-19 and contains 43 routes. Its source tip was `3e4b657b673e3a910f02b94dc428a310bb0fb9fb`; the live response identifies Hugo 0.154.2 and the repository's Pages workflow mounts CLI docs at `/cli/`.
- The mdBook project site at `https://mergewayhq.github.io/mergeway-cli/` is the 36-HTML-file artifact in this repository's `origin/gh-pages` commit `0240a15aa3214f136ba2022c15eee17e2b756bb8`, deployed from source commit `f69b6af9915deb9deee7b1925b266ef0bb45c86a`.

`legacy-routes.tsv` records both inventories, every reviewed replacement, and three useful source-form aliases found in old navigation/source. Deployment-level static aliases handle the current `/cli/` tree and project-prefixed `/mergeway-cli/` tree; Hugo aliases handle historical unprefixed `.html` and trailing-slash forms.

The empty blog/taxonomy indexes, legacy `llms` exports, and mdBook's aggregate print, table-of-contents, and project-specific 404 views are intentionally retired. The three `mergeway-python` routes have no equivalent in this repository and are explicitly marked out of scope rather than silently claimed as migrated. Independent parity review must accept that disposition or require the organization-site cutover to retain them before Phase 12. All CLI content routes receive aliases.

During parity review, change content only to fix a release-blocking defect. A correction must update its relevant executable check and the Phase 11 evidence. New features, new workflows, broad rewrites, dependency updates, and navigation restructuring wait until the freeze is lifted after cutover review.

## Build and review the candidate

Enter the locked devenv environment and run:

```sh
make website-release-candidate
tar -tzf mergeway-website-release-candidate.tar.gz | sort
```

This uses the CI production base URL and produces a deterministic archive whose root is the publish directory. On a pull request or `main` build, download the Actions artifact named `mergeway-website-release-candidate-<commit>`; do not use the mdBook deployment artifact as the Hugo preview.

For a local artifact-only preview, extract into a temporary directory and use a static server. Do not use `hugo server` for final parity review because that rebuilds pages with a preview base URL.

## Accessibility and responsive review

Automated checks require one English page title, production canonical, main landmark, h1-first non-skipping heading outline, unique IDs, image alternatives, visible-focus CSS, reduced-motion CSS, and unique canonical-page titles/URLs. They complement, rather than replace, this keyboard and visual review matrix:

| Page class | Desktop | Mobile |
| --- | --- | --- |
| Home and documentation index | `/`, `/docs/` | `/`, `/docs/` |
| Component hubs | all six under `/docs/components/` | all six under `/docs/components/` |
| Diátaxis samples | first-repository tutorial, install how-to, specification overview reference, metadata-as-code explanation | same four |
| Long/deep reference | specification fields, CLI create, MCP tools, LSP server reference | same four |

For each viewport, keyboard through the menu toggle, search, sidebar, breadcrumbs, in-page links, previous/next links, and auxiliary links. Confirm focus is never hidden, controls have an understandable name, landmarks/headings expose the page structure, text and focus remain distinguishable in light and dark schemes, images have an appropriate alternative, code/tables scroll instead of clipping, and no essential information depends on animation. Repeat with reduced motion enabled. Record release-blocking findings in the Phase 11 evidence; do not waive them silently.

## Phase 12 deployment procedure

1. Require a green `website-check` job and reviewer approval of parity, redirects, accessibility, and rollback readiness for the exact commit.
2. Download the CI artifact and verify its SHA-256 against the Actions log or locally rebuilt archive.
3. Preserve the current `gh-pages` tip and the approved Hugo artifact identifier in the cutover record.
4. Make the production change in `mergewayhq/mergewayhq.github.io`, whose Pages workflow owns the root host. Replace its generated `public` input with the unpacked approved artifact, or integrate this exact artifact as its root publish input. A change only to this repository's `gh-pages` publisher affects `/mergeway-cli/` and cannot cut over the organization root.
5. Preserve or explicitly approve retirement of the organization site's Python pages and utility exports. Do not let an unrelated rebuild of this repository supply mutable documentation during the approved artifact deployment.
6. Deploy through the organization repository's `actions/deploy-pages` workflow, then crawl every non-retired entry in `legacy-routes.tsv`, the canonical route set, internal links, and anchors on the public host.
7. Recheck the accessibility matrix on the public host. If any release blocker appears, execute rollback before accepting unrelated changes.

Phase 11 does not authorize these production mutations.

## Rollback

The last known-good organization-site source is `mergewayhq/mergewayhq.github.io` commit `3e4b657b673e3a910f02b94dc428a310bb0fb9fb`. The last known-good mdBook project artifact is this repository's `origin/gh-pages` commit `0240a15aa3214f136ba2022c15eee17e2b756bb8`. Verify both without changing production:

```sh
git ls-remote https://github.com/mergewayhq/mergewayhq.github.io.git refs/heads/main
git cat-file -e 0240a15aa3214f136ba2022c15eee17e2b756bb8^{commit}
git archive --format=tar 0240a15aa3214f136ba2022c15eee17e2b756bb8 | tar -tf - | sort
```

For an actual root-site rollback, first record the failed organization-site deployment and source tip. A repository administrator reverts the cutover commit back to `3e4b657b673e3a910f02b94dc428a310bb0fb9fb` (or its reviewed successor) and reruns that repository's unchanged `actions/deploy-pages` workflow. The `/mergeway-cli/` project site remains separately recoverable by rerunning this repository's mdBook publisher from `f69b6af9915deb9deee7b1925b266ef0bb45c86a`; do not force-push either branch without the expected remote tip and an incident record. Afterward, verify `/`, `/cli/`, `/python/`, representative guide/reference routes, `/mergeway-cli/`, and static assets.

## Approval record

- Content parity: pending independent reviewer
- URL mapping: pending independent reviewer
- Accessibility: pending independent reviewer
- Rollback readiness: pending independent reviewer

The implementation evidence may record automated and implementer review results, but the review gate remains open until an independent reviewer fills this approval record or provides equivalent pull-request approval.
