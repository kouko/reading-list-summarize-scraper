# Recover the first page — independent acceptance evidence


## Re-run on 2026-10-25, at d8e89a7c1acfd3bbbc04e42abe316a9e68776458

**Current result: all four Acceptance conditions and all eleven UI rows work. All three previously recorded important findings are resolved by actual CLI evidence.** Historical failures and the previous rerun above remain intact.

### Setup, fix range and source snapshot
- Reviewed fix range 64d2db3..d8e89a7: supported explicit canonical page positions must match actual position in recovery and forward validation; page-less article canonical remains allowed.
- New clean detached checkout `/private/tmp/rlss-acceptance-d8e89a7` at the full SHA above. README `npm ci` passed (26 packages, zero audit findings); `env GOCACHE=/private/tmp/rlss-go-build go build -o /private/tmp/rlss-acceptance-d8e89a7-bin ./cmd/rlss/` passed.
- Fresh actual-CLI runner `/private/tmp/rlss-acceptance-run-d8e89a7.py`; new isolated per-case config/profile/output under `/private/tmp/rlss-acceptance-live-d8e89a7/`. Maximum two CLI processes, headless30/headed60/wait2, headed_on_block false, max_pages10 (ITmedia final page5 cap5), no summary/reading-list/RSS/copy for live corpus. No real vault/profile/root executable touched.
- Source observations explicitly reused from **2026-10-25 12:17:13 Asia/Taipei**: 34 current-at-capture HTTP source pages, 279 corroborated first-page candidate blocks, with Mynavi biography distinction as previously disclosed. This final run did not refresh source HTTP. All eighteen CLI outputs and all body/order/uniqueness/original-URL checks are new.
- Analyzer `/private/tmp/rlss-acceptance-analyze-d8e89a7.py` re-read the newly saved outputs. Source URLs/anchor hashes are as above; new result hashes/page counts/offsets are in current `sanitized-results.json`.
- Same focused command from section4 run at new checkout; output `/private/tmp/rlss-acceptance-focused-d8e89a7.log`: exit0, all selected tests pass; browser test4.40s, pipeline package0.739s. It covers canonical contradictory recovery, HTTP404, H1/title challenge, valid security-heading article, empty/unusable/nav failures, subresource404, retry replacement/lifecycle/profile, wrapped challenge suppression, original URL/index and existing-file protection.
- Full package suite is **not skipped**; finalize-review will execute `env GOCACHE=/private/tmp/rlss-go-build go test ./...` and block acceptance on failure. No future suite outcome is asserted here.

### 1. Complete verified later entries — re-tested — works
| New case | Exit | Seconds | Source blocks present/checked | Once-only page anchors in order (offsets) |
|---|---|---|---|---|
| pcuser | 0 | 20.31 | 61/61 | 482, 5088, 10463, 16653, 22246 |
| president | 0 | 24.41 | 38/38 | 195, 1661, 2517, 3563, 4749, 6110 |
| gendai | 0 | 7.29 | 14/14 | 194, 1432 |
| bunshun | 0 | 13.19 | 14/14 | 189, 1565 |
| jbpress | 0 | 11.42 | 21/21 | 385, 1631 |
| toyokeizai | 0 | 12.66 | 22/22 | 324, 1018, 1964 |
| yahoo | 0 | 7.61 | 5/5 | 477, 1512 |
| itmedia-business | 0 | 9.83 | 32/32 | 2181, 16259, 30790 |
| itmedia-mobile | 0 | 9.55 | 10/10 | 1119, 4760, 9255 |
| ascii | 0 | 9.32 | 34/34 | 356, 3899 |
| mynavi | 0 | 7.12 | 26/26 | 251, 2020 |
| nikkan | 0 | 11.34 | 2/2 | 416, 3057 |
| itmedia-page3 | 0 | 23.9 | 61/61 | 1653, 7147, 11350, 17540, 23133 |
| itmedia-page5 | 0 | 33.04 | 61/61 | 1653, 8317, 13692, 19882, 25525 |
| president-page2 | 0 | 29.7 | 38/38 | 202, 1668, 2524, 3570, 4756, 6117 |
| president-page3 | 0 | 31.55 | 38/38 | 202, 1668, 2524, 3570, 4756, 6117 |
| yahoo-page2 | 0 | 10.16 | 5/5 | 484, 1519 |
| toyokeizai-page2 | 0 | 17.8 | 22/22 | 321, 1015, 1961 |

All eighteen print Success1/Failed0 and preserve supplied URL in frontmatter and SHA8 filename. Exact command per row: `/private/tmp/rlss-acceptance-d8e89a7-bin --config /private/tmp/rlss-acceptance-live-d8e89a7/<id>/config.yaml --no-summary url '<URL in initial section1 table>'`; new `run.json` and `cli.log` capture each actual invocation.

### 2. Start at first page, correct order, no repeated merge — re-tested — works
All 279/279 source-snapshot candidates in twelve first-page cases and 225/225 across six recovered entries appear. Every recovery candidate occurs once; all page anchors occur once and increase in order. Output identity is unchanged in all18. New contradictory-first control now rejects instead of saving the repeated second-page suffix.

### 3. Conservative rejection — re-tested — works
All same local flows from the previous run were executed, without new case expansion, using `python3 /private/tmp/rlss-acceptance-flows-d8e89a7.py` and production binary. Every local fixture remains synthetic. Full config/commands and before/after SHA256 maps: `/private/tmp/rlss-acceptance-live-d8e89a7/flows/results.json`; each case has `config.yaml` and `cli.log`.

| Actual local case | Exit | Files | Old hashes unchanged | Headed retry | Result |
|---|---|---|---|---|---|
| wrong-article | 1 | 0 | true | true | total1 success0 skipped0 partial0 failed1 |
| missing-page | 1 | 0 | true | true | total1 success0 skipped0 partial0 failed1 |
| wall-disabled | 1 | 0 | true | false | total1 success0 skipped0 partial0 failed1 |
| title-wall-disabled | 1 | 0 | true | false | total1 success0 skipped0 partial0 failed1 |
| loop | 1 | 0 | true | true | total1 success0 skipped0 partial0 failed1 |
| cap | 1 | 0 | true | true | total1 success0 skipped0 partial0 failed1 |
| failure-preserves | 1 | 2 | true | true | total1 success0 skipped0 partial0 failed1 |
| missing-preserves | 1 | 2 | true | true | total1 success0 skipped0 partial0 failed1 |
| wall-preserves | 1 | 2 | true | false | total1 success0 skipped0 partial0 failed1 |
| title-wall-preserves | 1 | 2 | true | false | total1 success0 skipped0 partial0 failed1 |
| contradictory-first | 1 | 0 | true | true | total1 success0 skipped0 partial0 failed1 |
| success-overwrites | 0 | 2 | false | false | total1 success1 skipped0 partial0 failed0 |
| ordinary | 0 | 1 | false | false | total1 success1 skipped0 partial0 failed0 |
| batch-continues | 0 | 1 | false | true | total2 success1 skipped0 partial0 failed1 |
| completed-skips | 0 | 2 | true | true | total2 success0 skipped1 partial0 failed1 |
| dry-run | 0 | 0 | true | false | total2 success0 skipped2 partial0 failed0 |
| empty-batch | 0 | 0 | true | false | empty; no items |

- **Resolved canonical contradiction:** exact same HTTP200 fixture as previous failure: requested /contradict?page=2, nominated /contradict?page=1 explicitly canonical to page2, both bodies second-page prose, nominated page1 next to page2. Current exact command: `/private/tmp/rlss-acceptance-d8e89a7-bin --config /private/tmp/rlss-acceptance-live-d8e89a7/flows/contradictory-first/config.yaml --no-summary url 'http://127.0.0.1:52495/contradict?page=2'`. Actual exit1, Success0/Failed1, zero files; initial and headed attempts both report `article identity mismatch: canonical on http://127.0.0.1:52495/contradict?page=1`. Source reproduction unchanged; no relaxed fixture.
- **Resolved HTTP404 remains resolved:** actual /missing?page=2 returns Python send_error404; current CLI fails with HTTP404 after allowed headed retry, no partial output. Seeded body+summary remain byte-identical.
- **Resolved discarded challenge heading remains resolved:** exact original H1 CAPTCHA response plus title-only CAPTCHA variant both fail without headed retry or new files; seeded body+summary remain byte-identical. Current focused suite accepts independent legitimate article prose beside a security heading.
- Wrong-article first link, loop and cap still fail/no files; focused empty/unusable/nav cases pass. Nothing was repaired by the acceptance agent.

### 4. Preserve established behavior — re-tested — works
All12 first-page controls, retry settings, profile lifecycle, original URL identity, old-file preservation, explicit successful overwrite, completed skipping, continued batch processing, ordinary extraction and empty/dry-run pass. Actual cap+HTTP404+both challenge seeded-file cases all retain both old files.

### 5. Later-page success UI — re-tested — works
All six current real runs Success1/exit0, complete source-snapshot prose/order/uniqueness/originalURL checks pass.

### 6. First-page success UI — re-tested — works
All twelve current real runs Success1/exit0 and source-snapshot checks pass.

### 7. Progress/retry UI — re-tested — works
Actual wrong-article/HTTP404/loop/cap/canonical contradiction logs show processing and permitted headed retry from the supplied URL. Focused complete-retry replacement, lifecycle and pagination-profile tests pass. Successful complete retry is measured at existing test seam rather than claimed as a live network-recovery observation.

### 8. Uncertain first page UI — re-tested — works
Both wrong-article first link and explicit canonical page-position contradiction fail after retry, nonzero, no output.

### 9. Missing/empty/wall UI — re-tested — works
Actual required HTTP404 and both original-document challenge variants fail/no partial body and preserve old files; focused empty/unusable/navigation failures pass.

### 10. Loop/limit UI — re-tested — works
Actual loop and cap1 return Failed1/exit1 with no files; cap seeded files unchanged.

### 11. Disabled challenge retry UI — re-tested — works
Actual H1/title challenge runs and their seeded-file counterparts fail with no headed retry. Wrapped-challenge and explicit-headed focused tests pass.

### 12. Existing-output UI — re-tested — works
Four actual seeded failure cases retain body+summary hashes exactly; successful complete extraction overwrites body (summary retained under no-summary); completed batch URL skips with unchanged output hashes. Original URL/index tests pass.

### 13. Batch/watch UI — re-tested — works
Actual two-item RSS: Success1/Failed1 with one article saved, processing continues; completed rerun Skipped1/Failed1, existing files unchanged. Existing batch exit0 preserved. Original-URL retry asserted by focused tests. Watch uses same per-cycle batch path; no persistent watch was launched.

### 14. Ordinary no-rule UI — re-tested — works
Actual no-rule command Success1/exit0, one ordinary article saved.

### 15. Empty/dry-run UI — re-tested — works
Actual dry-run Total2/Skipped2/Failed0, exit0, no article files; empty RSS logs no items, exit0, no article files.

### Final finding disposition
All three important findings (HTTP404 false success, discarded original challenge heading, explicit canonical position contradiction) are **resolved** by actual rerun evidence. No important finding was dismissed. All4 Acceptance and11 UI rows re-tested fully; none carried over. Only report/evidence changed. Model oracle inherited/null; one completed redispatch, no overrides or child agents in this acceptance run. Full suite remains the later finalize-review gate.


## Re-run on 2026-10-25, at 2633afb7a01c9b0c8ecf8a2cbb085a009a5aad59

The earlier results above are historical observations at 8439560, including the two real failures. This section supersedes their current verdicts, not their historical record. No row was carried over: per-page browser fetch changed, so all four Acceptance conditions and eleven flows were re-tested in full. Fix diff reviewed: dc1558a..2633afb (HTTP document status, original DOM challenge evidence, document-title evidence). No important finding was dismissed.

### Clean setup and oracle
- New detached checkout `/private/tmp/rlss-acceptance-2633afb` created directly from 2633afb. README `npm ci` passed (26 packages; zero audit findings); `env GOCACHE=/private/tmp/rlss-go-build go build -o /private/tmp/rlss-acceptance-2633afb-bin ./cmd/rlss/` passed.
- Fresh profiles/configs/outputs: `/private/tmp/rlss-acceptance-live-2633afb/`. Actual real runner: `/private/tmp/rlss-acceptance-run-2633afb.py`; at most two concurrent CLI processes, same isolation/timeouts/selectors/caps as initial run. All eighteen CLI runs are new; no old output was compared as the new product result.
- Oracle reuse explicitly authorized: source HTTP captures finalized **2026-10-25 12:17:13 Asia/Taipei** in `/private/tmp/rlss-acceptance-live-8439560/current-source-private.json`. These 34 source pages and 279 corroborated candidate blocks are the source snapshot; this rerun did not fetch a newer HTTP snapshot. The Mynavi author-profile/image-marker distinction remains as disclosed above.
- Analyzer `/private/tmp/rlss-acceptance-analyze-2633afb.py` re-read every newly saved body against that source snapshot; per-page source URLs/hashes are unchanged, new offsets/counts are listed below. No raw articles or real profiles committed.
- Focused test command is identical to section 4, executed at new checkout with output `/private/tmp/rlss-acceptance-focused-2633afb.log`; exit 0, all selected tests pass (browser test 4.52s, pipeline package 0.781s). This includes actual extractor HTTP404, H1-only challenge, title-only challenge and valid article containing a security heading, as well as subresource 404 separation.
- Full `env GOCACHE=/private/tmp/rlss-go-build go test ./...` remains assigned to finalize-review; package-tests was not skipped. Its future result is not claimed here.

### 1. Complete later entries — re-tested — works
All six later entries and all twelve first-page controls exited 0, Success 1 / Failed 0. Each case preserves original URL frontmatter and SHA8 filename identity. Each tuple below is checked blocks, all-present count, then ordered page-anchor offsets; every page anchor occurs exactly once.

| Case | Seconds | Present/checked | Unique page-anchor offsets |
|---|---|---|---|
| pcuser | 20.28 | 61/61 | 1652, 7146, 11349, 17539, 23182 |
| president | 25.33 | 38/38 | 195, 1661, 2517, 3563, 4749, 6110 |
| gendai | 7.66 | 14/14 | 194, 1432 |
| bunshun | 13.56 | 14/14 | 189, 1419 |
| jbpress | 6.04 | 21/21 | 385, 1631 |
| toyokeizai | 12.05 | 22/22 | 324, 1018, 1964 |
| yahoo | 6.58 | 5/5 | 477, 1512 |
| itmedia-business | 10.95 | 32/32 | 2183, 16262, 30861 |
| itmedia-mobile | 10.71 | 10/10 | 1119, 4760, 9255 |
| ascii | 7.17 | 34/34 | 356, 3905 |
| mynavi | 7.07 | 26/26 | 251, 2020 |
| nikkan | 8.65 | 2/2 | 416, 3057 |
| itmedia-page3 | 23.08 | 61/61 | 1653, 8317, 13692, 19950, 25611 |
| itmedia-page5 | 32.86 | 61/61 | 1653, 7215, 10742, 17000, 22711 |
| president-page2 | 30.22 | 38/38 | 202, 1668, 2524, 3570, 4756, 6117 |
| president-page3 | 34.61 | 38/38 | 202, 1668, 2524, 3570, 4756, 6117 |
| yahoo-page2 | 15.01 | 5/5 | 484, 1519 |
| toyokeizai-page2 | 14.54 | 22/22 | 321, 1015, 1961 |

Commands are `/private/tmp/rlss-acceptance-2633afb-bin --config /private/tmp/rlss-acceptance-live-2633afb/<id>/config.yaml --no-summary url '<URL from section 1 table>'`. Each case has a new `run.json`, `cli.log` and saved body. Current `sanitized-results.json` retains every new file hash and each page's source URL/hash, presence count and anchor offset.

### 2. Order and uniqueness — re-tested — works for the six verified entries
279/279 source-snapshot blocks in twelve first-page runs and 225/225 across six recovery runs. Every recovered candidate block appears once; page anchors are once and ordered from page one. Original URL identity holds for all eighteen. The additional contradictory-page control in row 3 exposes a separate conservative-rejection failure; the verified real entries themselves passed.

### 3. Reject uncertain/incomplete/wall content — re-tested — fails
The original HTTP404 and H1 challenge findings are **fixed by actual CLI evidence**, and the added title-only challenge is rejected correctly. However, the additional same-article contradictory page-position control still falsely succeeds and saves a duplicated suffix.

Actual local runner: `python3 /private/tmp/rlss-acceptance-flows-2633afb.py`. All local responses are synthetic. Shared config: host 127.0.0.1, selector `a.next`, cap 10 unless specified, headless 30s, headed 60s, wait 2s, headed_on_block false, isolated per-case profile/output. The reproduction source remains on disk.

| Local case | Exit | Saved files | Before/after hashes equal | Headed retry observed | Visible counts |
|---|---|---|---|---|---|
| wrong-article | 1 | 0 | true | true | total 1, success 0, skipped 0, partial 0, failed 1 |
| missing-page | 1 | 0 | true | true | total 1, success 0, skipped 0, partial 0, failed 1 |
| wall-disabled | 1 | 0 | true | false | total 1, success 0, skipped 0, partial 0, failed 1 |
| title-wall-disabled | 1 | 0 | true | false | total 1, success 0, skipped 0, partial 0, failed 1 |
| loop | 1 | 0 | true | true | total 1, success 0, skipped 0, partial 0, failed 1 |
| cap | 1 | 0 | true | true | total 1, success 0, skipped 0, partial 0, failed 1 |
| failure-preserves | 1 | 2 | true | true | total 1, success 0, skipped 0, partial 0, failed 1 |
| missing-preserves | 1 | 2 | true | true | total 1, success 0, skipped 0, partial 0, failed 1 |
| wall-preserves | 1 | 2 | true | false | total 1, success 0, skipped 0, partial 0, failed 1 |
| title-wall-preserves | 1 | 2 | true | false | total 1, success 0, skipped 0, partial 0, failed 1 |
| contradictory-first | 0 | 1 | false | false | total 1, success 1, skipped 0, partial 0, failed 0 |
| success-overwrites | 0 | 2 | false | false | total 1, success 1, skipped 0, partial 0, failed 0 |
| ordinary | 0 | 1 | false | false | total 1, success 1, skipped 0, partial 0, failed 0 |
| batch-continues | 0 | 1 | false | true | total 2, success 1, skipped 0, partial 0, failed 1 |
| completed-skips | 0 | 2 | true | true | total 2, success 0, skipped 1, partial 0, failed 1 |
| dry-run | 0 | 0 | true | false | total 2, success 0, skipped 2, partial 0, failed 0 |
| empty-batch | 0 | 0 | true | false | empty batch; no items |

**Original findings now fixed:**
- `missing-page`: HTTP404 required page now reports HTTP 404, tries allowed headed retry, exits 1, Success 0 / Failed 1, no files. Same `send_error(404)` response as original reproduction.
- `wall-disabled`: exact original H1 HTML from section 3 now fails as Cloudflare/CAPTCHA, exits 1, no headed retry, no files.
- `title-wall-disabled`: HTTP200 `<html><head><title>Cloudflare CAPTCHA</title></head><body><p>Verify you are human</p></body></html>` likewise fails, no headed retry, no files.
- `missing-preserves`, `wall-preserves`, `title-wall-preserves`: each seeded synthetic content and summary retained byte-for-byte; first allows retry, last two do not.

**New important finding: contradictory first-page evidence accepted.**
- Exact command: `/private/tmp/rlss-acceptance-2633afb-bin --config /private/tmp/rlss-acceptance-live-2633afb/flows/contradictory-first/config.yaml --no-summary url 'http://127.0.0.1:63656/contradict?page=2'`.
- All documents HTTP200. Original requested page two: synthetic Page 2 body and no next link. Nominated page one: `<link rel="canonical" href="/contradict?page=2">`, the same Page 2 body, and `<a class="next" href="/contradict?page=2">Next page</a>`. There is no page-one prose. This is explicitly page-specific canonical evidence, not an ordinary page-less article canonical.
- Actual result: exit 0, Success 1, Failed 0; saved `flows/contradictory-first/output/127_0_0_1/2026-10-25__02e218b2__content.md`; “Page 1” appears 0 times, “Page 2” 24 times (12 from each response). The complete source snapshot for this synthetic article is defined in the local response generator, independent of product output.
- Log: `/private/tmp/rlss-acceptance-live-2633afb/flows/contradictory-first/cli.log`. Full config/command/hash maps: case config and `flows/results.json`.
- Anchor: `internal/extract/pagination.go:compatibleCanonical` and `internal/extract/browser.go:ExtractPaginated` recovery/forward metadata validation. Canonical article identity comparison discards page position, so contradictory earlier-page evidence is not rejected. Fix must reject supported explicit canonical page-position contradictions during both recovery and forward reading while preserving normal page-less canonical behavior; rerun affected whole criteria.

### 4. Existing first-page/retry/output behavior — re-tested — works
All twelve first-page controls pass. Focused retry/lifecycle/profile/wall-wrapper/original-URL/index tests pass. Real loop/cap failures retry as before; recognized challenges do not when disabled. Seeded cap, missing-page, H1 challenge and title challenge preserve old body/summary byte-for-byte; successful explicit overwrite and completed batch skipping pass. New false success in row 3 remains a separate completeness failure.

### 5. Later-page UI — re-tested — works
Six actual recovery runs print Success 1, exit 0, validated source body/order/identity.

### 6. First-page UI — re-tested — works
Twelve actual first-page runs print Success 1, exit 0, source-snapshot prose and page order checks pass.

### 7. In progress and retry UI — re-tested — works
Actual wrong-article, HTTP404, loop and cap logs show processing and permitted headed retry from original URL. Named focused complete-retry replacement, lifecycle and pagination-profile tests pass. Successful complete retry is asserted at existing test seam, not invented as a live network-recovery observation.

### 8. Uncertain first page UI — re-tested — partly
Wrong-article first link still fails after retry with identity mismatch and no output. New same-article canonical page-position contradiction succeeds incorrectly as detailed in row 3.

### 9. Missing/empty/wall UI — re-tested — works
Original missing-page plus both H1/title challenge cases fail without output; all three preserve seeded old files. Focused empty/unusable/nav tests pass. Exact historical reproductions remain above.

### 10. Loop/limit UI — re-tested — works
Actual loop and cap1 fail, exit 1, Failed 1, no files. Cap old-file case retains both files.

### 11. Disabled challenge retry UI — re-tested — works
Both actual H1/title-only challenge variants fail with no headed retry and no partial output, including seeded-file cases. Wrapped challenge, explicit headed and lifecycle focused tests pass.

### 12. Existing-output UI — re-tested — works
Actual cap/missing/H1/title failures retain both seeded files byte-for-byte. Actual complete success overwrites body while retaining existing summary in no-summary mode. Completed ordinary RSS item is skipped and before/after hashes remain equal. Detailed hashes in current local `flows/results.json`.

### 13. Batch/watch UI — re-tested — works
Actual two-item RSS batch: Success 1, Failed 1, one article saved; completed rerun: Skipped 1, Failed 1 with no old-file changes. Retry retains original URL through focused assertions. Existing batch exit0 despite one failed item is preserved. Watch is covered via same per-cycle batch path; no persistent watch process was launched.

### 14. Ordinary no-rule UI — re-tested — works
Fresh actual no-pagination-rule command exits 0, Success 1, saves one ordinary synthetic article.

### 15. Empty/dry-run UI — re-tested — works
Actual dry run Total2 / Skipped2 / Failed0, exit0, no article files; empty RSS logs no items, exit0, no article files.

### Current findings and scope
- Prior important HTTP404 false-success finding: fixed, verified with actual CLI and seeded existing files.
- Prior important discarded-heading challenge finding: fixed, verified with exact original H1 case, additional title-only case, disabled retry and seeded files.
- New **important** canonical page-position contradiction false success: open. No dismissal.
- All four Acceptance and all eleven UI rows re-tested, none carried over. Product unchanged by acceptance agent. Only report/evidence modified; original data and running processes preserved. No child agent/override/redispatch. Full suite left for finalize-review.

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

