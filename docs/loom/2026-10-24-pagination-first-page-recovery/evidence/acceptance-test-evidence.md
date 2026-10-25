# Recover the first page — independent acceptance evidence

Tried on 2026-10-25 Asia/Taipei against commit 8439560c8a5392d03e2557659a4945ca87ea0199.

## Setup and scope
- Fresh detached checkout: `git worktree add --detach /private/tmp/rlss-acceptance-8439560 8439560c8a5392d03e2557659a4945ca87ea0199`.
- Followed README: `npm ci` passed (26 packages, zero audit findings); `env GOCACHE=/private/tmp/rlss-go-build go build -o /private/tmp/rlss-acceptance-8439560-bin ./cmd/rlss/` passed.
- Runner: `/private/tmp/rlss-acceptance-run.py`; original live captures: `/private/tmp/rlss-acceptance-live-8439560/`. No old CLI outputs were reused.
- At most two CLI processes, fresh per-case config/profile/output, no summary, Safari/Chrome reading-list sources/RSS/copy disabled for real corpus. Headless 30s, headed 60s, wait 2s, headed_on_block false, max_pages 10 except ITmedia page-five max_pages 5.
- Exact real-case invocation: `/private/tmp/rlss-acceptance-8439560-bin --config /private/tmp/rlss-acceptance-live-8439560/<id>/config.yaml --no-summary url '<URL>'`. Per-case `run.json` retains command, return code, duration, output filenames; `cli.log` retains stdout/stderr.
- Profiles, executable, article bodies, source HTML and configurations remain temporary; only this sanitized evidence and the report are committed. The original checkout, vault, real profiles, root binary and running batch were untouched.
- No product edits. No dismissed important-or-worse findings were supplied. No child agents, model overrides or redispatches.

## 1. All six verified later entries obtain the same complete target body
All six exited 0 and printed Success 1 / Failed 0. The following table includes their twelve first-page controls. Every row passed current-source prose presence, page order, page-anchor uniqueness and original URL frontmatter / SHA8 filename identity. “Blocks” counts source-backed candidate prose, not a claim of complete semantic article coverage.

| Case | Exit | Seconds | Current source blocks present/checked | Pages | URL | Selector |
|---|---|---|---|---|---|---|
| pcuser | 0 | 22.02 | 61/61 | 5 | https://www.itmedia.co.jp/pcuser/articles/2606/08/news096.html | `link[rel=next]` |
| president | 0 | 23.6 | 38/38 | 6 | https://president.jp/articles/-/120600 | `a.next` |
| gendai | 0 | 6.59 | 14/14 | 2 | https://gendai.media/articles/-/172201 | `a[data-testid="pagination-next-page-link"]` |
| bunshun | 0 | 12.28 | 14/14 | 2 | https://bunshun.jp/articles/-/84967 | `a.goto-next` |
| jbpress | 0 | 8.35 | 21/21 | 2 | https://jbpress.ismedia.jp/articles/-/97650 | `link[rel=next]` |
| toyokeizai | 0 | 9.53 | 22/22 | 3 | https://toyokeizai.net/articles/-/859253?display=b | `a.lead-heading__link` |
| yahoo | 0 | 6.52 | 5/5 | 2 | https://news.yahoo.co.jp/articles/04860bcd34a7f82d074799f216f7f76dbfd12129 | `a[data-cl-params*="_cl_link:next;"]` |
| itmedia-business | 0 | 11.13 | 32/32 | 3 | https://www.itmedia.co.jp/business/articles/2610/23/news006.html | `link[rel~=next]` |
| itmedia-mobile | 0 | 9.19 | 10/10 | 3 | https://www.itmedia.co.jp/mobile/articles/2610/23/news053.html | `link[rel~=next]` |
| ascii | 0 | 6.67 | 34/34 | 2 | https://ascii.jp/elem/000/004/431/4431424/ | `link[rel~=next]` |
| mynavi | 0 | 6.89 | 26/26 | 2 | https://news.mynavi.jp/article/20160130-businessnote/ | `link[rel~=next]` |
| nikkan | 0 | 8.22 | 2/2 | 2 | https://www.nikkan-gendai.com/articles/view/money/394477 | `link[rel~=next]` |
| itmedia-page3 | 0 | 22.83 | 61/61 | 5 | https://www.itmedia.co.jp/pcuser/articles/2606/08/news096_3.html | `link[rel=next]` |
| itmedia-page5 | 0 | 34.11 | 61/61 | 5 | https://www.itmedia.co.jp/pcuser/articles/2606/08/news096_5.html | `link[rel=next]` |
| president-page2 | 0 | 23.67 | 38/38 | 6 | https://president.jp/articles/-/120600?page=2 | `a.next` |
| president-page3 | 0 | 25.8 | 38/38 | 6 | https://president.jp/articles/-/120600?page=3 | `a.next` |
| yahoo-page2 | 0 | 10.78 | 5/5 | 2 | https://news.yahoo.co.jp/articles/04860bcd34a7f82d074799f216f7f76dbfd12129?page=2 | `a[data-cl-params*="_cl_link:next;"]` |
| toyokeizai-page2 | 0 | 13.26 | 22/22 | 3 | https://toyokeizai.net/articles/-/859253?page=2 | `a.lead-heading__link` |

The six recovered bodies contain the same current-source prose as their first-page controls: 61 blocks each for two ITmedia entries, 38 each for two PRESIDENT entries, 5 for Yahoo and 22 for Toyokeizai (225/225 across six runs). Identity preserved in all 18 runs. SHA256 of each complete saved file and all checks are in temporary `sanitized-results.json`.

## 2. Start at page one, keep order, combine each page once
- Observer/analyzer: `/private/tmp/rlss-acceptance-source.py` and `/private/tmp/rlss-acceptance-analyze.py`.
- Independently fetched all 34 source-page URLs via current HTTP. ITmedia paragraph candidates were derived from current cmsBodyMainStart/End p/h2/h3 blocks longer than 35 characters. Other publishers used independently observed per-page prose candidates at least 60 normalized characters, but each counted candidate had to match this run's current HTTP page text. An old candidate alone never counted as source evidence.
- Normalization removes Markdown link markup and formatting/whitespace and applies NFKC; this allows meaningful prose comparison without treating navigation or dynamic ad noise as article identity.
- Every later-entry candidate block appeared exactly once. At least one independently source-backed unique anchor for each page appeared in increasing output offsets. No extracted discovery-page output was counted as a source oracle.
- First-page corpus: 279/279 currently corroborated candidate blocks present. Mynavi page two's previous 290-character normalized candidate included `![](…/images/profile.jpg)` plus an author biography; the original normalizer retained the image URL. Removing the image marker confirmed all 214 normalized biography characters in current HTTP. This author-panel candidate is disclosed separately and excluded from 279; no unconfirmed “280/280 article paragraphs” claim is made.
- ITmedia business has author-profile passages repeated in source pages one and three. The two related profile candidates appear twice in output; page-specific article anchors remain single and ordered. This is source repetition, not duplicate navigation/merging.

Per-page evidence below gives URL, matching blocks, anchor occurrence count, byte-independent normalized-character offset and SHA256 prefix. Same source-page URL/anchor across runs is deliberate, proving first-page/later-entry agreement without retaining raw articles.

### pcuser
| Page | Source URL | Blocks | Anchor count | Offset | Anchor SHA256 prefix |
|---|---|---|---|---|---|
| 1 | https://www.itmedia.co.jp/pcuser/articles/2606/08/news096.html | 13/13 | 1 | 1652 | 1441ea7c610b8df4 |
| 2 | https://www.itmedia.co.jp/pcuser/articles/2606/08/news096_2.html | 9/9 | 1 | 8316 | b90d0a1d00ad808b |
| 3 | https://www.itmedia.co.jp/pcuser/articles/2606/08/news096_3.html | 12/12 | 1 | 13691 | a1812ce7d7372de4 |
| 4 | https://www.itmedia.co.jp/pcuser/articles/2606/08/news096_4.html | 9/9 | 1 | 19881 | a9bf71c5a7e097ec |
| 5 | https://www.itmedia.co.jp/pcuser/articles/2606/08/news096_5.html | 18/18 | 1 | 25474 | caf28ba397790336 |
### president
| Page | Source URL | Blocks | Anchor count | Offset | Anchor SHA256 prefix |
|---|---|---|---|---|---|
| 1 | https://president.jp/articles/-/120600 | 6/6 | 1 | 195 | f32178f6cf9f7584 |
| 2 | https://president.jp/articles/-/120600?page=2 | 7/7 | 1 | 1661 | 0d8d2a7fe1e2168e |
| 3 | https://president.jp/articles/-/120600?page=3 | 6/6 | 1 | 2517 | 4dca554647226aa0 |
| 4 | https://president.jp/articles/-/120600?page=4 | 6/6 | 1 | 3563 | 817bf74e4ed295f9 |
| 5 | https://president.jp/articles/-/120600?page=5 | 6/6 | 1 | 4749 | dfaa5166ecdcb276 |
| 6 | https://president.jp/articles/-/120600?page=6 | 7/7 | 1 | 6110 | aa66a83fe3f71864 |
### gendai
| Page | Source URL | Blocks | Anchor count | Offset | Anchor SHA256 prefix |
|---|---|---|---|---|---|
| 1 | https://gendai.media/articles/-/172201 | 7/7 | 1 | 194 | e28b0ba2cd424993 |
| 2 | https://gendai.media/articles/-/172201?page=2 | 7/7 | 1 | 1432 | 7bbba3a27d16ef31 |
### bunshun
| Page | Source URL | Blocks | Anchor count | Offset | Anchor SHA256 prefix |
|---|---|---|---|---|---|
| 1 | https://bunshun.jp/articles/-/84967 | 6/6 | 1 | 189 | e1d4b69cf61173c5 |
| 2 | https://bunshun.jp/articles/-/84967?page=2 | 8/8 | 1 | 1419 | c7383c905ee3fff6 |
### jbpress
| Page | Source URL | Blocks | Anchor count | Offset | Anchor SHA256 prefix |
|---|---|---|---|---|---|
| 1 | https://jbpress.ismedia.jp/articles/-/97650 | 12/12 | 1 | 385 | e574bdb420b72615 |
| 2 | https://jbpress.ismedia.jp/articles/-/97650?page=2 | 9/9 | 1 | 1631 | 744388e1d948d01e |
### toyokeizai
| Page | Source URL | Blocks | Anchor count | Offset | Anchor SHA256 prefix |
|---|---|---|---|---|---|
| 1 | https://toyokeizai.net/articles/-/859253?display=b | 6/6 | 1 | 324 | da34f44995ba422a |
| 2 | https://toyokeizai.net/articles/-/859253?page=2 | 8/8 | 1 | 1018 | 356c9f4df8157d2a |
| 3 | https://toyokeizai.net/articles/-/859253?page=3 | 8/8 | 1 | 1964 | ffc29a6f917f64f9 |
### yahoo
| Page | Source URL | Blocks | Anchor count | Offset | Anchor SHA256 prefix |
|---|---|---|---|---|---|
| 1 | https://news.yahoo.co.jp/articles/04860bcd34a7f82d074799f216f7f76dbfd12129 | 4/4 | 1 | 477 | cef2513b63ca9d72 |
| 2 | https://news.yahoo.co.jp/articles/04860bcd34a7f82d074799f216f7f76dbfd12129?page=2 | 1/1 | 1 | 1512 | 53a2134ed61ecb37 |
### itmedia-business
| Page | Source URL | Blocks | Anchor count | Offset | Anchor SHA256 prefix |
|---|---|---|---|---|---|
| 1 | https://www.itmedia.co.jp/business/articles/2610/23/news006.html | 10/10 | 1 | 2183 | 59053c6c394510fd |
| 2 | https://www.itmedia.co.jp/business/articles/2610/23/news006_2.html | 10/10 | 1 | 16262 | d265bc3a2455bf09 |
| 3 | https://www.itmedia.co.jp/business/articles/2610/23/news006_3.html | 12/12 | 1 | 30796 | 3df8d6e15e21a8ee |
### itmedia-mobile
| Page | Source URL | Blocks | Anchor count | Offset | Anchor SHA256 prefix |
|---|---|---|---|---|---|
| 1 | https://www.itmedia.co.jp/mobile/articles/2610/23/news053.html | 2/2 | 1 | 1119 | 618ad62e5f45fe46 |
| 2 | https://www.itmedia.co.jp/mobile/articles/2610/23/news053_2.html | 3/3 | 1 | 4760 | f534f4628c74155a |
| 3 | https://www.itmedia.co.jp/mobile/articles/2610/23/news053_3.html | 5/5 | 1 | 9264 | aef55f51863b3b6a |
### ascii
| Page | Source URL | Blocks | Anchor count | Offset | Anchor SHA256 prefix |
|---|---|---|---|---|---|
| 1 | https://ascii.jp/elem/000/004/431/4431424/ | 16/16 | 1 | 356 | 181445ec5cd902a0 |
| 2 | https://ascii.jp/elem/000/004/431/4431424/2/ | 18/18 | 1 | 3905 | 36322cf326e8af24 |
### mynavi
| Page | Source URL | Blocks | Anchor count | Offset | Anchor SHA256 prefix |
|---|---|---|---|---|---|
| 1 | https://news.mynavi.jp/article/20160130-businessnote/ | 11/11 | 1 | 251 | 63acfeecc010d767 |
| 2 | https://news.mynavi.jp/article/20160130-businessnote/2 | 15/15 | 1 | 2020 | acfee148f97f160c |
### nikkan
| Page | Source URL | Blocks | Anchor count | Offset | Anchor SHA256 prefix |
|---|---|---|---|---|---|
| 1 | https://www.nikkan-gendai.com/articles/view/money/394477 | 1/1 | 1 | 416 | 8c4d1981733ea964 |
| 2 | https://www.nikkan-gendai.com/articles/view/money/394477/2 | 1/1 | 1 | 3057 | 9c9927c2ccd502d3 |
### itmedia-page3
| Page | Source URL | Blocks | Anchor count | Offset | Anchor SHA256 prefix |
|---|---|---|---|---|---|
| 1 | https://www.itmedia.co.jp/pcuser/articles/2606/08/news096.html | 13/13 | 1 | 1653 | 1441ea7c610b8df4 |
| 2 | https://www.itmedia.co.jp/pcuser/articles/2606/08/news096_2.html | 9/9 | 1 | 8317 | b90d0a1d00ad808b |
| 3 | https://www.itmedia.co.jp/pcuser/articles/2606/08/news096_3.html | 12/12 | 1 | 13692 | a1812ce7d7372de4 |
| 4 | https://www.itmedia.co.jp/pcuser/articles/2606/08/news096_4.html | 9/9 | 1 | 19882 | a9bf71c5a7e097ec |
| 5 | https://www.itmedia.co.jp/pcuser/articles/2606/08/news096_5.html | 18/18 | 1 | 25425 | caf28ba397790336 |
### itmedia-page5
| Page | Source URL | Blocks | Anchor count | Offset | Anchor SHA256 prefix |
|---|---|---|---|---|---|
| 1 | https://www.itmedia.co.jp/pcuser/articles/2606/08/news096.html | 13/13 | 1 | 483 | 1441ea7c610b8df4 |
| 2 | https://www.itmedia.co.jp/pcuser/articles/2606/08/news096_2.html | 9/9 | 1 | 5975 | b90d0a1d00ad808b |
| 3 | https://www.itmedia.co.jp/pcuser/articles/2606/08/news096_3.html | 12/12 | 1 | 11350 | a1812ce7d7372de4 |
| 4 | https://www.itmedia.co.jp/pcuser/articles/2606/08/news096_4.html | 9/9 | 1 | 17540 | a9bf71c5a7e097ec |
| 5 | https://www.itmedia.co.jp/pcuser/articles/2606/08/news096_5.html | 18/18 | 1 | 23183 | caf28ba397790336 |
### president-page2
| Page | Source URL | Blocks | Anchor count | Offset | Anchor SHA256 prefix |
|---|---|---|---|---|---|
| 1 | https://president.jp/articles/-/120600 | 6/6 | 1 | 202 | f32178f6cf9f7584 |
| 2 | https://president.jp/articles/-/120600?page=2 | 7/7 | 1 | 1668 | 0d8d2a7fe1e2168e |
| 3 | https://president.jp/articles/-/120600?page=3 | 6/6 | 1 | 2524 | 4dca554647226aa0 |
| 4 | https://president.jp/articles/-/120600?page=4 | 6/6 | 1 | 3570 | 817bf74e4ed295f9 |
| 5 | https://president.jp/articles/-/120600?page=5 | 6/6 | 1 | 4756 | dfaa5166ecdcb276 |
| 6 | https://president.jp/articles/-/120600?page=6 | 7/7 | 1 | 6117 | aa66a83fe3f71864 |
### president-page3
| Page | Source URL | Blocks | Anchor count | Offset | Anchor SHA256 prefix |
|---|---|---|---|---|---|
| 1 | https://president.jp/articles/-/120600 | 6/6 | 1 | 202 | f32178f6cf9f7584 |
| 2 | https://president.jp/articles/-/120600?page=2 | 7/7 | 1 | 1668 | 0d8d2a7fe1e2168e |
| 3 | https://president.jp/articles/-/120600?page=3 | 6/6 | 1 | 2524 | 4dca554647226aa0 |
| 4 | https://president.jp/articles/-/120600?page=4 | 6/6 | 1 | 3570 | 817bf74e4ed295f9 |
| 5 | https://president.jp/articles/-/120600?page=5 | 6/6 | 1 | 4756 | dfaa5166ecdcb276 |
| 6 | https://president.jp/articles/-/120600?page=6 | 7/7 | 1 | 6117 | aa66a83fe3f71864 |
### yahoo-page2
| Page | Source URL | Blocks | Anchor count | Offset | Anchor SHA256 prefix |
|---|---|---|---|---|---|
| 1 | https://news.yahoo.co.jp/articles/04860bcd34a7f82d074799f216f7f76dbfd12129 | 4/4 | 1 | 484 | cef2513b63ca9d72 |
| 2 | https://news.yahoo.co.jp/articles/04860bcd34a7f82d074799f216f7f76dbfd12129?page=2 | 1/1 | 1 | 1519 | 53a2134ed61ecb37 |
### toyokeizai-page2
| Page | Source URL | Blocks | Anchor count | Offset | Anchor SHA256 prefix |
|---|---|---|---|---|---|
| 1 | https://toyokeizai.net/articles/-/859253?display=b | 6/6 | 1 | 321 | da34f44995ba422a |
| 2 | https://toyokeizai.net/articles/-/859253?page=2 | 8/8 | 1 | 1015 | 356c9f4df8157d2a |
| 3 | https://toyokeizai.net/articles/-/859253?page=3 | 8/8 | 1 | 1961 | ffc29a6f917f64f9 |

## 3. Reject uncertain / missing / incomplete / wall content
**FAIL: two independent actual-CLI controls saved invalid page content and reported complete success.**
- Reproducer: `python3 /private/tmp/rlss-acceptance-flows.py` in the clean checkout. It starts a loopback HTTP server on an ephemeral port and runs the production binary, never an injected extractor.
- Exact failed-run commands:
  - `/private/tmp/rlss-acceptance-8439560-bin --config /private/tmp/rlss-acceptance-live-8439560/flows/missing-page/config.yaml --no-summary url http://127.0.0.1:57775/missing`
  - `/private/tmp/rlss-acceptance-8439560-bin --config /private/tmp/rlss-acceptance-live-8439560/flows/wall-disabled/config.yaml --no-summary url http://127.0.0.1:57775/wall`
- Shared config: domain 127.0.0.1, next_page_selector `a.next`, max_pages 10, headed_on_block false, timeouts/wait as above. The first response is HTTP 200, title/H1 “Independent acceptance article”, twelve repeated synthetic prose sentences, then `<a class="next" href="/missing?page=2">Next page</a>` (or /wall). The selector matched and the second page WAS visited; this is not a disappearing-next-link finding.
- Missing-page response: the handler explicitly calls `send_error(404)` at /missing?page=2. Python's standard response title/H1 is “Error response”, body includes “Error code: 404”, “Message: Not Found.” and “Error code explanation: 404 - Nothing matches the given URI.” Actual saved content appends those three messages after first-page prose. Exit 0, Success 1, Failed 0; one content file saved. Log: `flows/missing-page/cli.log`; output: `flows/missing-page/output/127_0_0_1/2026-10-25__d6261d42__content.md`.
- Wall response: HTTP 200 with exact synthetic HTML `<html><body><h1>Cloudflare CAPTCHA</h1><p>Verify you are human</p></body></html>`. Actual saved content appends “Verify you are human”; the challenge H1 disappeared during extraction. Exit 0, Success 1, Failed 0; one content file saved. Log: `flows/wall-disabled/cli.log`; output: `flows/wall-disabled/output/127_0_0_1/2026-10-25__18f93495__content.md`.
- Both actual logs show content written and no retry. Both findings are important: missing required content / a challenge became a successful article. These fail Acceptance 3 and UI flow 9; the title-only challenge also limits UI flow 11.
- Passing actual negative controls: wrong-article /wrong?page=2 has `<link rel="first" href="/other">`; it fails twice with article identity mismatch, exits 1, Failed 1 and no files. Loop and max_pages=1 controls also exit 1, Failed 1 and no files. The error logs retain their reasons.
- Focused browser tests additionally cover empty required content, unusable continuation, navigation failure, explicit earlier evidence and wrong identity. They passed, but do not override either actual-CLI failure.

## 4. First pages, retry policy and failure-time existing-file protection
- All twelve current first-page CLI controls passed; 279/279 source-backed blocks, ordered unique page anchors.
- Focused command run in clean checkout:
  `env GOCACHE=/private/tmp/rlss-go-build go test ./internal/extract ./internal/pipeline -run '^(TestBrowserExtractPaginated|TestPagination_laterentry_rejectsincomplete|TestPoolHeadedRecoveryLifecycle|TestPoolExplicitHeaded|TestPoolRecoveryPreservesPaginationProfile|TestPoolRecoveryBrowserFailure|TestProcessItemWithFallback|TestProcessBatchFailedPaginationPreservesFiles|TestProcessItemVideoResume|TestExtractionRejectsObservedWalls)$' -count=1 -v`
- Captured output: `/private/tmp/rlss-acceptance-focused.log`; exit 0, extract 4.708s and pipeline 0.361s. All selected subtests passed. They verify reset/reuse lifecycle, retry replacement of partial content, challenge retry suppression (including wrapped pagination errors), original requested URL on each attempt and in output/index, old content/summary preservation, completed skip and dry run.
- The full package suite was deliberately left to finalize-review, not skipped by user instruction. It must run `env GOCACHE=/private/tmp/rlss-go-build go test ./...`; the automated test suite runs before the change is accepted and blocks it on failure. No future suite pass is claimed here.

## 5. UI — Later-page success
Actual six runs in section 1; all Success 1, exit 0, current source body/order/uniqueness and original URL identity pass.

## 6. UI — First-page success
Actual twelve runs in sections 1–2; all Success 1, exit 0 and current source checks pass.

## 7. UI — In progress and retry
Actual wrong-article, loop, cap and failure-preservation logs show processing and “headless extraction failed, retrying with headed mode”; original supplied URL is retained. Named focused tests `TestProcessItemWithFallback/pagination_retry_replaces_partial_body`, `TestPoolHeadedRecoveryLifecycle` and `TestPoolRecoveryPreservesPaginationProfile` pass and cover a successful complete retry. Successful retry is tested at the established injected lifecycle seam, not claimed as a live network failure/recovery observation.

## 8. UI — Uncertain article or first page
Actual wrong-article fixture: first link points to /other. Exit 1, Failed 1, article identity mismatch after headed retry, zero output files.

## 9. UI — Missing or unusable page
Fails for the two actual responses in section 3. Focused empty/navigation/error tests pass but did not detect these HTTP/document-body cases.

## 10. UI — Loop or exhausted limit
Actual `flows/loop` and `flows/cap`: exit 1, Failed 1, zero files. Loop reports repeated continuation; cap reports continuation beyond maxPages 1. Both expose allowed retry messages.

## 11. UI — Challenge retry disabled
Partly. `TestProcessItemWithFallback/pagination_wrapped_challenge_disabled`, `/block_disabled_challenge`, `/specialized_challenge_disabled` pass: detected challenges do not retry/write. Actual title-only challenge in section 3 is not recognized and is saved as success, so the visible safety result is incomplete.

## 12. UI — Existing output
- Actual `flows/failure-preserves`: cap=1, seeded synthetic content/summary, exit 1, Failed 1, exact before/after SHA256 maps equal (two files). Result maps in `flows/results.json`.
- Actual `flows/success-overwrites`: complete two-page synthetic article with seeded files; exit 0, Success 1; content replaced, summary retained (no summarizer requested), two files remain. Existing behavior preserved.
- Actual `flows/completed-skips`: RSS batch with failing wrong-article plus completed ordinary URL; Success 0, Skipped 1, Failed 1, existing file hashes unchanged. No automatic old-note migration.
- Important limit: the two erroneously successful failures in section 3 write invalid content, so failure-file protection cannot protect data where failure was not recognized.

## 13. UI — Batch or watch
Actual `flows/batch-continues`: local RSS contains wrong-article recovery then ordinary article. Total 2, Success 1, Failed 1, exactly one article output; processing continued. Batch command returns 0, which is existing batch behavior; individual failure remains visible. Normal completed batch then skips the completed URL. Original-URL retry is additionally asserted by `TestProcessBatchFailedPaginationPreservesFiles`. Watch delegates each cycle through the same batch path; a persistent watch loop was not separately launched.

## 14. UI — No matching pagination rule
Actual `flows/ordinary` has no domain pagination rule. Exit 0, Success 1, one synthetic article saved through ordinary extraction.

## 15. UI — Empty batch or dry run
Actual `flows/dry-run`: same two-item RSS, exit 0, Total 2 / Skipped 2 / Failed 0, zero article files. Actual `flows/empty-batch`: empty RSS, exit 0, “fetched items count=0” / “no items to process”, zero article files.

## Findings handed to closing review
1. **important** — `internal/extract/browser.go:135` and per-page validation at line 164: HTTP 404 required page is extracted/merged and reported as complete success. Close by rejecting missing/error documents before accumulating output and re-running the entire affected criteria with old-file protection.
2. **important** — `internal/extract/browser.go:158` / per-page validation at line 164: challenge markers confined to an H1 disappear before content validation; title-only CAPTCHA document prose is saved as a successful page. Close by validating sufficient original-document wall evidence and re-running affected criteria, including challenge retry suppression and no partial output.

