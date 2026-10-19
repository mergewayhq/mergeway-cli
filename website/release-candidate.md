# Website release-candidate runbook

Status: ready for independent review; production remains unchanged  
Content freeze began: 2026-10-19  
Production base URL: `https://mergewayhq.github.io/`

This runbook governs the Phase 11 artifact. Until Phase 12 is approved, the `build-and-deploy` job continues publishing mdBook and the Hugo artifact is review-only.

## Frozen inputs and route inventory

The live-route snapshot is the complete set of HTML files in `origin/gh-pages` commit `6ca7115862b16dcd50bda0b53c855e12f73beef2`, deployed from source commit `6a8b671c4d8e2320f08ae828fff03ad6274c7ddf`. The root page was also fetched from `https://mergewayhq.github.io/` on 2026-10-19 and matched that mdBook deployment. `legacy-routes.tsv` records all 36 published HTML routes, their reviewed destinations, and three useful source-form aliases found in the deployed navigation/source.

`print.html` and `toc.html` are intentionally retired. They are mdBook application views rather than independent content pages: Hugo provides per-page browser printing and its persistent documentation navigation. The generated 404 route and home page are retained; all other published content routes receive Hugo aliases.

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
4. Change only the publish input of `build-and-deploy` from `docs/book` to the checked `website/public` output (or the unpacked approved artifact); retain the existing GitHub Pages branch and action until post-cutover verification is complete.
5. Deploy from `main`, then crawl every non-retired entry in `legacy-routes.tsv`, the canonical route set, internal links, and anchors on the public host.
6. Recheck the accessibility matrix on the public host. If any release blocker appears, execute rollback before accepting unrelated changes.

Phase 11 does not authorize these production mutations.

## Rollback

The last known-good mdBook site is `origin/gh-pages` commit `6ca7115862b16dcd50bda0b53c855e12f73beef2`. Verify its recoverability without changing production:

```sh
git cat-file -e 6ca7115862b16dcd50bda0b53c855e12f73beef2^{commit}
git archive --format=tar 6ca7115862b16dcd50bda0b53c855e12f73beef2 | tar -tf - | sort
```

For an actual rollback, first record the failed Hugo deployment tip. A repository administrator then restores the known-good tree to `gh-pages` using a force-with-lease scoped to that recorded failed tip, or reruns the unchanged mdBook publisher from source commit `6a8b671c4d8e2320f08ae828fff03ad6274c7ddf`. Afterward, verify `/`, representative CLI/guide routes, all static assets, and the legacy route inventory on the public host. Never force-push without the expected remote tip and an incident record.

## Approval record

- Content parity: pending independent reviewer
- URL mapping: pending independent reviewer
- Accessibility: pending independent reviewer
- Rollback readiness: pending independent reviewer

The implementation evidence may record automated and implementer review results, but the review gate remains open until an independent reviewer fills this approval record or provides equivalent pull-request approval.
