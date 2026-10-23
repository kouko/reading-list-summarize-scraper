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
+
## Re-run on 2026-10-24, at c576b8e — final post-fix acceptance

The earlier failure measurements above remain historical evidence, not the current verdict. The initial report incorrectly used re-tested on its first measurement; that first measurement's Re-run cells should have been “—”. This is now a genuine post-fix re-run, so every current report row is re-tested.

### Clean setup, scope and commands

Fresh detached worktree /private/tmp/rlss-accept-rerun-c576b8e created with `git worktree add --detach /private/tmp/rlss-accept-rerun-c576b8e c576b8e`. Followed README `make` anew: npm install26 packages, audit0 vulnerabilities, JS bundle rebuilt, Go binary compiled; exit0. No root binary/config/vault/profile reused. Actual live configs, output and Chrome profiles under /private/tmp/rlss-accept-rerun-runs, all reading-list sources/RSS/copy_to disabled and --no-summary.

Reviewed exact fix delta `git diff 9e5c01d..c576b8e -- internal/extract/browser.go internal/extract/browser_test.go`: only initial configured-pagination guard and its focused tests changed. Explicit single numeric page parameter greater than1 rejects; rel first/prev checks retained; forward traversal applies entry check only to initial page. No walker/storage/retry/CLI policy changes.

Commands executed:
- `python3 /private/tmp/rlss-accept-rerun.py` —20 independent live runs using clean binary.
- `TMPDIR=/private/tmp GOCACHE=/private/tmp/rlss-go-build go test ./internal/pipeline ./internal/extract ./cmd/rlss -run 'Test(ProcessItemWithFallback|ProcessBatchFailedPaginationPreservesFiles|ProcessItemVideoResume|ExtractionRejectsObservedWalls|IsSkipped|Stats_Report|ApplyFiltersDateOrder|Paginate_|BrowserExtractPaginated|Pagination_laterentry)' -count=1 -v > /private/tmp/rlss-accept-rerun-focused.log 2>&1` — all three selected packages PASS, no skipped selected browser tests.
- `python3 /private/tmp/rlss-accept-rerun-compare.py > /private/tmp/rlss-accept-rerun-compare.log` — compare newly saved outputs against independently observed per-page references, inspection retains advertising/comment exclusions rather than asserting perfect cleanliness.
- `python3 /private/tmp/rlss-accept-rerun-itmedia.py` — compares new output against independently fetched article-boundary paragraphs retained from the immediately preceding measurement; PC USER61/61 all exactly once in five-page order, Business32/32 and Mobile10/10 with unique article paragraphs in page order. Two identical Business paragraphs occur in different source pages as well; no whole-page duplicate claim is made from repeated prose.
- `python3 /private/tmp/rlss-accept-rerun-extra.py` — actual complete overwrite, empty process and dry-run all exit0. Synthetic old content replaced by full content; synthetic summary retained. Existing completed-item skip established with named pipeline test, since explicit URL forces processing.
- Full suite not run by acceptance tester. finalize-review will run full suite and earned probe, blocking attestation on failure; neither check was skipped.

### Actual twenty-run outcomes

All12 first-page positives from the earlier corpus table were re-executed: same34 source pages, each exit0/Success1/Partial0/Failed0 and one content file. Prior working publishers and additional examples all actually ran again. Exact commands/exits/filenames are in /private/tmp/rlss-accept-rerun-runs/runs.json and per-case run.json; per-case cli.log retains displayed messages.

| Negative control | Actual exit | New content files | New summary files |
|---|---:|---:|---:|
| ITmedia cap1 | 1 | 0 | 0 |
| ITmedia missing selector with standard next hint | 1 | 0 | 0 |
| ITmedia initial page3 | 1 | 0 | 0 |
| ITmedia initial page5 | 1 | 0 | 0 |
| PRESIDENT initial page3 | 1 | 0 | 0 |
| Yahoo initial page2 | 1 | 0 | 0 |
| PRESIDENT initial page2 | 1 | 0 | 0 |
| Toyokeizai initial page2 | 1 | 0 | 0 |

New defect reproductions now show “initial page number is greater than one”, retained headless failure/allowed headed retry messages, Total1 Success0 Partial0 Failed1, final command error. In both cases each permitted attempt rejects; no suffix is saved. Configured headed_on_block:false still suppresses verification-page retries, while these nonchallenge failures follow existing permitted retry policy.

Exact new negative commands:
`/private/tmp/rlss-accept-rerun-c576b8e/rlss --config /private/tmp/rlss-accept-rerun-runs/president-page2/config.yaml --no-summary url 'https://president.jp/articles/-/120600?page=2'`
`/private/tmp/rlss-accept-rerun-c576b8e/rlss --config /private/tmp/rlss-accept-rerun-runs/toyokeizai-page2/config.yaml --no-summary url 'https://toyokeizai.net/articles/-/859253?page=2'`

### Row-by-row re-run record

1: re-tested — actual five-page CLI, all61 article paragraphs once/in order, focused successful walker and duplicate guard PASS.
2: re-tested — all eight live negative controls fail/no files; later-failure/empty/unfollowable/loop tests PASS; output-preservation integration PASS. Both earlier important findings closed.
3: re-tested — all12 live positives, full selected fallback matrix, challenge-disabled settings, failed retries, explicit headed behavior, browser readiness/navigation errors and earlier-entry checks PASS. Whole affected named surface exercised.
4: re-tested — twelve article corpus/ten groups/34pages executed again, saved outputs compared per page and results recorded, new negatives explicitly fail.
5: re-tested — live complete configured article Success1/exit0/all five pages.
6: re-tested — fallback matrix and actual negative retry messages; retry body replaces initial partial body and only complete extraction writes.
7: re-tested — later failed/empty page and exhausted retry tests retain page cause, no partial success; existing-file byte comparison PASS.
8: re-tested — live cap1 fails/no output; loop/unfollowable/cap walker tests PASS.
9: re-tested — pagination wrapped challenge disabled and ordinary block-disabled tests PASS with no headed call; no substitute output.
10: re-tested — batch failed-pagination preservation/counting and stats tests PASS; unchanged batch/watch shared path. No real reading lists accessed.
11: re-tested — actual empty process exit0/no items; dry-run URL exit0/Skipped1/no extraction or writing.
12: re-tested — actual complete overwrite retaining summary; failed extraction preserves synthetic old content+summary byte-for-byte; completed and dry-run pipeline tests make no extractor calls.
13: re-tested — expanded corpus re-executed and comparisons inspected, no reliance on author's success.

### Current findings, decisions, data effects, skips and open questions

No open findings. Both important page2 false-success instances are closed by independent observed failures with no saved files; neither was dismissed. No new case exploration or product repair by acceptance tester.

Decision retained: explicit numeric later entry is known incomplete within configured pagination, even absent backwards link; fail rather than automatically return to first. Unsupported/unnumbered encodings without recognizable signals remain outside the guarantee. Full automatic pagination discovery and perfect ad removal remain unpromised.

Only new isolated files and synthetic old files touched. Failure preserves old content+summary; success retains overwrite; completed items retain skip. No real user data, migration or backup. No user-instructed skipped steps. No open user decision.
+
## Re-run on 2026-10-24, at 7e7069a — review-fix closure

### Fresh setup and actual execution

Fresh detached worktree /private/tmp/rlss-accept-reviewfix-7e7069a, created with `git worktree add --detach /private/tmp/rlss-accept-reviewfix-7e7069a 7e7069a`. README `make` succeeded anew:26 npm packages, audit0 vulnerabilities, rebuilt bundle and Go executable. All live commands used that executable and fresh /private/tmp/rlss-accept-reviewfix-runs configs/profiles/outputs, disabled reading-list sources, RSS, copy_to and summaries. No root executable, real profile, user notes or production config used.

Read W5/W6 and actual d3647b1..7e7069a diff. Existing content policy moved into shared extract validation and now validates each page before combining; pipeline uses the same policy. Five incomplete-sequence regressions now enter through Browser.ExtractPaginated. This broadens the affected checks to page validity, caller classification, retries, cached walls and existing-file protection; all were re-tested.

Commands and results:
- `python3 /private/tmp/rlss-accept-reviewfix-run.py`:20 actual runs. Exact arguments, statuses and output filenames in runs.json and each case run.json; displayed processing/retry/failure output in cli.log.
- `TMPDIR=/private/tmp GOCACHE=/private/tmp/rlss-go-build go test ./internal/pipeline ./internal/extract ./cmd/rlss -run 'Test(ProcessItemWithFallback|ProcessBatchFailedPaginationPreservesFiles|ProcessItemVideoResume|ExtractionRejectsObservedWalls|ProcessItemRejectsCachedWall|IsBlockedPage_|IsUnavailablePage_|IsSkipped|Stats_Report|ApplyFiltersDateOrder|Paginate_|BrowserExtractPaginated|Pagination_laterentry)' -count=1 -v > /private/tmp/rlss-accept-reviewfix-focused.log 2>&1`: PASS all three selected packages; selected browser cases ran, no skips.
- `python3 /private/tmp/rlss-accept-reviewfix-body.py`: independently fetched current source HTML and compared current saved output; first Nikkan comparison raced its not-yet-written output, so `python3 /private/tmp/rlss-accept-reviewfix-nikkan.py` completed those two current-page comparisons afterward. No failed comparison counted as a pass.
- `python3 /private/tmp/rlss-accept-reviewfix-compare.py`: new saved outputs versus independent per-page reference blocks. Dynamic ads/recommendations/comment registration differ; inspected as such, not claimed exact article matches. Current article-boundary HTTP checks separately confirm all61 PC USER,32 Business and10 Mobile paragraphs.
- `python3 /private/tmp/rlss-accept-reviewfix-order.py`: current HTTP article-boundary text against current saved content; PC USER61 paragraphs each once in five-page order, Business32/Mobile10 present with unique paragraphs ordered by page. Raw current HTML, private body comparisons and article prose remain temporary.
- `python3 /private/tmp/rlss-accept-reviewfix-extra.py`: complete overwrite, empty batch and dry-run all exit0; synthetic content overwritten only with complete article, summary retained; empty selection/no write and dry-run Skipped1 confirmed.
- Full package suite and earned probe remain finalize-review's responsibility; neither skipped. Acceptance tester did not run full suite or author a probe.

| Actual live surface | Cases | Outcome |
|---|---:|---|
| All original and expanded first-page articles | 12 articles /10 groups /34pages | Each exit0, Success1, one content file; body checks completed |
| ITmedia cap1 and missing selector | 2 | Each exit1, Success0, zero content/summary files |
| ITmedia pages3/5, PRESIDENT3, Yahoo2 | 4 | Each exit1, Success0, zero content/summary files |
| Previously defective PRESIDENT2 and Toyokeizai2 | 2 | Each exit1, Success0, zero content/summary files |

### Review findings closed by actual entry-level evidence

Fatal earlier-prose-masks-later-wall finding: `TestBrowserExtractPaginated/real_extractor_validates_later_page` PASS using actual embedded extractor JS. The test's login, challenge and article cases all executed: later recognized login/challenge returns error naming affected wall-end URL and policy reason with empty returned body, despite usable preceding prose; genuine article prose discussing login remains accepted. This is targeted actual browser-entry evidence, not a live publisher login-wall claim.

Important private-tests-bypass-entry finding: `TestBrowserExtractPaginated/incomplete_sequences` PASS for cap, loop, later-failure, unfollowable and empty. Each calls actual browser entry and asserts no body, affected page URL and reason. Prior direct-test names above belong to historical SHAs, not this final run.

Shared-policy preservation: moved classifier tests for challenge languages, normal articles, single-pattern mentions, maintenance/login walls and cached walls PASS; fallback matrix includes pagination wrapped challenge with headed retry disabled and complete retry replacing partial body. Failed batch pagination preserves both old content and summary byte-for-byte. No important finding dismissed.

### Current row-by-row verdicts

1: re-tested, works — current five-page CLI plus61 paragraphs once/in order.
2: re-tested, works — all8 live incomplete controls plus real-extractor later login/challenge, all5 browser-entry incomplete sequences, no-body/page-reason assertions and failed-output preservation.
3: re-tested, works — all12 live positives plus shared classifier/cache policy, full selected fallback settings matrix, readiness, navigation and storage checks PASS.
4: re-tested, works — current12/10/34 corpus, independent current HTML/body comparisons and failures recorded.
5: re-tested, works — actual complete configured article Success1/exit0.
6: re-tested, works — allowed retry logs and full selected fallback matrix, complete retry replaces partial attempt.
7: re-tested, works — actual browser-entry later failure/empty/login/challenge plus existing-file protection.
8: re-tested, works — actual cap rejection and browser-entry cap/loop/unfollowable checks.
9: re-tested, works — wrapped and ordinary challenge-disabled tests retain no-headed-retry policy; shared classifier checks PASS.
10: re-tested, works — selected batch failure counts, stats and preservation tests; unchanged shared watch-cycle behavior, no real reading-list access.
11: re-tested, works — actual empty batch and dry-run exit0/no extraction/write.
12: re-tested, works — failed extraction preserves synthetic old content/summary; complete overwrite and completed-item skip checked again.
13: re-tested, works — expanded examples actually executed again, source/body evidence inspected.

### Decisions, data effects, skips and open questions

All previous important entry findings remain closed; newly reported fatal page-wall masking and important entry-test coverage findings also closed at7e7069a. No open findings or user decisions, no skipped instructed steps. Preserve original validation policy by sharing it per page rather than inventing another wall classifier; genuine independent article prose still accepted. Fail known incomplete, do not auto-return to first; configured-pagination scope and unsupported/unrecognizable signal limitations remain, with no universal or perfect-ad-cleaning claim. Only isolated and synthetic data touched; existing failure/success/skip semantics remain as described in report.
