# Reliable ITmedia pagination — acceptance test evidence

Tried independently on 2026-10-24 at 21fa256cbb8744a358c1d18f217f92d9dd428326.

## Clean setup and scope

Fresh detached worktree: /private/tmp/rlss-accept-final-21fa256. Created with:
`git worktree add --detach /private/tmp/rlss-accept-final-21fa256 21fa256cbb8744a358c1d18f217f92d9dd428326`.
Followed README full-build route: `make`; npm installed 26 packages, audit zero vulnerabilities, bundle rebuilt, Go binary compiled. Setup exit 0. Only this clean binary was executed. Observational DOM helper was subsequently added only to the temporary worktree; no product repair was performed.

Live configs, profiles and outputs: /private/tmp/rlss-accept-final-runs and /private/tmp/rlss-accept-exact-later. Safari, Chrome reading lists, RSS, copy_to disabled; --no-summary; per-case isolated user_data_dir. No real browser profiles, running user app, production config, notes, LLM or credentials used.

Commands:
- `python3 /private/tmp/rlss-accept-final-run.py`: adapts existing runner solely to clean binary/new root; 12 positive corpus cases, cap, wrong selector, four additional later entries.
- `python3 /private/tmp/rlss-accept-exact-later.py`: exactly four documented later-entry examples, fresh root.
- Every live invocation has shape `/private/tmp/rlss-accept-final-21fa256/rlss --config <isolated-case>/config.yaml --no-summary url <public-url>`. Exact argument arrays, exits and output filenames recorded in each case run.json.
- `TMPDIR=/private/tmp GOCACHE=/private/tmp/rlss-go-build go test ./internal/pipeline ./internal/extract ./cmd/rlss -run 'Test(ProcessItemWithFallback|ProcessBatchFailedPaginationPreservesFiles|ProcessItemVideoResume|ExtractionRejectsObservedWalls|IsSkipped|Stats_Report|ApplyFiltersDateOrder|Paginate_|BrowserExtractPaginated|Pagination_laterentry)' -count=1 -v > /private/tmp/rlss-accept-focused.log 2>&1`: PASS all three selected packages. An earlier identical focused run used the platform temporary directory; no real data was touched. Repeated with TMPDIR explicitly set for strict isolation.
- Full `go test ./...` deliberately left to finalize-review, which will execute it and block attestation on failure. Package-tests were not skipped. No new regression probes were authored.

## Live corpus results

All twelve first-page CLI commands exited 0, Success 1, Partial 0, Failed 0, one content file each. Independent current HTTP retrieval covered all34 pages; original independent per-page observations were also compared against these newly saved files. No author success claim was accepted without a new execution.

| Case | Pages | Matched reference blocks / scanned blocks | Result |
|---|---:|---:|---|
| ITmedia PC USER | 5 | 60/70 | article body verified separately 61/61 |
| PRESIDENT | 6 | 38/38 | works |
| Gendai | 2 | 14/14 | works |
| Bunshun | 2 | 14/14 | works |
| JBpress | 2 | 21/21 | works |
| Toyokeizai | 3 | 22/22 | works |
| Yahoo | 2 | 5/5 | works |
| ITmedia Business | 3 | 31/33 | article body verified separately 32/32 |
| ITmedia Mobile | 3 | 14/20 | article body verified separately 10/10 |
| ASCII | 2 | 34/34 | works |
| Mynavi | 2 | 27/27 | works |
| Nikkan Gendai | 2 | 2/2 | works |

Public URLs and configured selectors are retained in news-pagination-cases.json. Current raw HTTP pages, saved content and private comparisons remain temporary, not copied into history. Commands `python3 /private/tmp/rlss-accept-final-compare.py` and `python3 /private/tmp/rlss-accept-final-body.py`; outputs body-comparison-private.json and fresh-body.json under isolated run root. Generic scans include advertising, comments-registration prompts and recommendations: unmatched items were inspected, never silently counted as article loss or passed as exact cleanliness. ITmedia uses cmsBodyMainStart/End article boundaries in freshly fetched HTML; paragraph normalization only removes Markdown formatting, whitespace and Unicode presentation differences. Counts by PC USER page:13,9,12,9,18. All61 present. Business:10,10,12 all32 present. Mobile:2,3,5 all10 present. Source page sequence and saved paragraph positions establish order; walker duplicate-continuation tests independently passed.

## Important finding: known later entry still returns success

Severity: important. Anchor: internal/extract/browser.go:100–109.
Class: known incomplete initial entry returns success. Two live instances:
1. `https://president.jp/articles/-/120600?page=2`, selector `a.next`. CLI exit0, Success1, wrote one file; saved reference article paragraphs per source page: page1 0/6, page2 7/7, page3 6/6, page4 6/6, page5 6/6, page6 7/7. Browser DOM has canonical page2 and `<a class="prev" href="/articles/-/120600">前ページ</a>`. Guard ignores absent page parameter on first-page link. This is explicit earlier-page evidence, not an unsupported-signal excuse.
2. `https://toyokeizai.net/articles/-/859253?page=2`, selector `a.lead-heading__link`. CLI exit0, Success1, wrote one file; saved page1 0/6, page2 8/8, page3 8/8. Browser DOM canonical page2, title identifies page2, next target page3. No explicit backwards link was observed; numeric current page2 in configured pagination nevertheless identifies a later entry. Both cases violate Acceptance2.

Reproduction commands:
`/private/tmp/rlss-accept-final-21fa256/rlss --config /private/tmp/rlss-accept-final-runs/president-later/config.yaml --no-summary url 'https://president.jp/articles/-/120600?page=2'`
`/private/tmp/rlss-accept-final-21fa256/rlss --config /private/tmp/rlss-accept-final-runs/toyokeizai-later/config.yaml --no-summary url 'https://toyokeizai.net/articles/-/859253?page=2'`
Logs and saved output filenames: corresponding cli.log/run.json. Saved prose range is source pages2–6 /2–3 respectively, quantified above without committing copyrighted article text.

Independent browser observation: adapted existing inspector into temporary cmd/accept-dom; `GOCACHE=/private/tmp/rlss-go-build go run ./cmd/accept-dom /private/tmp/rlss-accept-dom-cases.json > /private/tmp/rlss-accept-dom.log`. DOM captures /private/tmp/rlss-accept-dom-president/inspection.json and /private/tmp/rlss-accept-dom-toyokeizai/inspection.json; fresh profiles and direct per-page navigation. Raw body/links stay local. No finding was dismissed. Required closure: reject justified numeric later entry without saving a suffix, retain first-page forward traversal, re-test both cases and affected criteria; no automatic return to first page requested.

## 1. 已驗證的 ITmedia 五頁文章可取得全部五頁的目標正文，順序正確且不重複合併頁面。

- How I tried it / what came back: Actual clean CLI five-page run, current independent HTTP61/61 article paragraphs and per-page order checks above; TestPaginate_FollowsToEnd and duplicate loop guard PASS.
- Evidence: isolated logs and named tests above; full automated suite runs before acceptance and blocks on failure, its result is not claimed here.

## 2. 已知文章尚未完整取得時，不回報完整擷取成功，也不把部分正文保存為成功結果。

- How I tried it / what came back: Cap1 and nonexistent selector both exit1/no files. Exact later cases ITmedia pages3 and5, PRESIDENT page3, Yahoo page2 each exit1/no files. Additional ITmedia page2 and Mynavi page2 also reject. However PRESIDENT2/Toyokeizai2 save suffixes as detailed above: fails.
- Evidence: isolated logs and named tests above; full automated suite runs before acceptance and blocks on failure, its result is not claimed here.

## 3. 既有的瀏覽器重試設定維持效果，先前可成功取得的其他新聞分頁案例仍能正確取得正文。

- How I tried it / what came back: All original seven publisher cases re-executed successfully. TestProcessItemWithFallback including pagination_retry_replaces_partial_body, pagination_wrapped_challenge_disabled, block_disabled_challenge/navigation/empty PASS. Browser readiness tests PASS with pending subresource and DOMContentLoaded.
- Evidence: isolated logs and named tests above; full automated suite runs before acceptance and blocks on failure, its result is not claimed here.

## 4. 收集更多實際日本新聞網站的分頁文章案例，逐項記錄頁數、正文核對及成功或失敗結果。

- How I tried it / what came back: Corpus table above records12 articles/10 groups/34pages. Current independent HTTP and reference comparison performed; additional three publishers and two ITmedia articles were actually executed. Known incomplete outcomes recorded honestly.
- Evidence: isolated logs and named tests above; full automated suite runs before acceptance and blocks on failure, its result is not claimed here.

## 5. 完整的已設定分頁文章

- How I tried it / what came back: Actual five-page CLI exit0, Success1, one content file; 61 verified article paragraphs.
- Evidence: isolated logs and named tests above; full automated suite runs before acceptance and blocks on failure, its result is not claimed here.

## 6. 處理進度與失敗後重試

- How I tried it / what came back: TestProcessItemWithFallback/pagination_retry_replaces_partial_body PASS; headed retry log visible; written body equals complete retry, excludes failed attempt.
- Evidence: isolated logs and named tests above; full automated suite runs before acceptance and blocks on failure, its result is not claimed here.

## 7. 後續頁面失敗或無法使用

- How I tried it / what came back: TestPaginate_LaterFailureRejectsPriorPages, EmptyPage and TestProcessBatchFailedPaginationPreservesFiles PASS. Returned reason names failed page, no partial saving; nonzero single URL exit actually observed for cap/selector/later-entry failures.
- Evidence: isolated logs and named tests above; full automated suite runs before acceptance and blocks on failure, its result is not claimed here.

## 8. 頁數上限或重複頁面仍有後續

- How I tried it / what came back: TestPaginate_MaxPagesCap, VisitedLoopGuard, StopsOnUnfollowableNext PASS. Live cap1 exit1, no output. Existing one retry runs and then failure.
- Evidence: isolated logs and named tests above; full automated suite runs before acceptance and blocks on failure, its result is not claimed here.

## 9. 關閉驗證頁的有視窗重試

- How I tried it / what came back: TestProcessItemWithFallback/pagination_wrapped_challenge_disabled and block_disabled_challenge PASS: no headed call, no substitute output. Navigation and empty errors still retry as before.
- Evidence: isolated logs and named tests above; full automated suite runs before acceptance and blocks on failure, its result is not claimed here.

## 10. 批次與持續監看

- How I tried it / what came back: TestProcessBatchFailedPaginationPreservesFiles and Stats_Report tests PASS. Existing loop continues per-item; no new watch policy. Empty batch actual command exit0; shared batch handling exercised by focused tests rather than waiting for a live reading-list/watch cycle.
- Evidence: isolated logs and named tests above; full automated suite runs before acceptance and blocks on failure, its result is not claimed here.

## 11. 空白清單與預覽

- How I tried it / what came back: python3 /private/tmp/rlss-accept-final-extra.py: actual process with disabled sources exit0/no items; --dry-run url exit0, Total1 Skipped1 Success0, no extraction/write.
- Evidence: isolated logs and named tests above; full automated suite runs before acceptance and blocks on failure, its result is not claimed here.

## 12. 已有原文與摘要

- How I tried it / what came back: TestProcessBatchFailedPaginationPreservesFiles PASS compares old content+summary byte-for-byte. TestProcessItemVideoResume/completed and dry-run PASS no extractor calls. Actual --force url with synthetic old content+summary exit0 overwrites complete content and retains summary. URL command always forces processing, so URL skip was not expected; an initial tester expectation otherwise was corrected after reading actual command contract.
- Evidence: isolated logs and named tests above; full automated suite runs before acceptance and blocks on failure, its result is not claimed here.

## 13. 擴充實站驗證

- How I tried it / what came back: Read committed expanded corpus plus independent re-execution and current body checks above. Article access restrictions never relabelled as successes; no new site-specific extractor promised.
- Evidence: isolated logs and named tests above; full automated suite runs before acceptance and blocks on failure, its result is not claimed here.

## Re-run on 2026-10-24, at 21fa256

All thirteen rows re-tested independently. No carried-over verdict relied on previous author success. Fresh build succeeded. No user-requested skips, no dismissed important findings, no product fixes. No open user decision; Acceptance2 remains failed pending repair. Only configured pagination is covered; unrecognized/unnumbered signals and perfect ad removal remain outside guarantees.

## Data effects and decisions

Only synthetic existing files and newly created isolated artifacts touched. Failed extraction preserved content and summary in focused integration tests. Successful explicit URL retains overwrite; completed batch items retain skip. No real user data or backup/migration. Agent decisions retained: fail known incomplete rather than auto-return first; wait DOMContentLoaded and usable body rather than all resources; preserve caller retry policy. Main agent agreed both discovered page2 cases require repair; neither dismissed.
