# Recover Felix and X article content — acceptance test evidence

Tried independently on 2026-10-27, at 6a8380ae1efbf2fe4e0c020f27a09cc624df6801, in a new detached worktree `/private/tmp/rlss-accept-fresh-20261027`. No implementation changes were made. No important findings were dismissed by the station.

## Setup and isolation

- Created a new worktree at the supplied commit, followed README's from-source setup: `npm ci`, `npm run build:defuddle`, and `GOCACHE=/private/tmp/rlss-accept-gocache go build -o /private/tmp/rlss-accept-bin ./cmd/rlss/`. Build passed; embedded-bundle regeneration left the tracked tree unchanged. The default Go build cache was sandbox-restricted, so only the cache location changed.
- Independently compiled the supplied observational Go harness from this worktree's Go module. It captures DOM and document metadata **before** injecting this clean copy's rebuilt `embed/defuddle.min.js`, then invokes the actual `window.extractArticle()` entry point. Every capture used a new temporary Chrome user-data directory. Public browser reads used escalation because sandboxed browser operation is unavailable.
- Captures: `/private/tmp/rlss-accept-live/01` through `/09`, each containing `page.html`, `document.json`, `extraction.json`, `summary.json`, and a temporary profile. Source copyright text stays in temporary files and is not committed.
- Isolated CLI config: `/private/tmp/rlss-accept-config.yaml`; output `/private/tmp/rlss-accept-cli-output`; explicit nonexistent temporary profile root, empty profile selection, no account or domain profile overrides. Profile resolver reported unavailable Local State, ensuring no main profile was used. Production executable, configuration, notes, browser profiles and live process were untouched.
- Local comparison helper `/private/tmp/rlss-accept-compare.cjs` and temporary pipeline witness `/private/tmp/rlss-accept-witness-test.go` remain outside version control. They were removed from the clean worktree after testing; only these two documentation artifacts were added to the change branch.

## 1. Complete target X bodies

- Command shape: `/private/tmp/rlss-accept-observer -url '<exact-url>' -headed -wait 12 -out /private/tmp/rlss-accept-live/NN -script /private/tmp/rlss-accept-fresh-20261027/embed/defuddle.min.js`.
- Independently inspected captured source containers and identities: ordinary posts use the target article's `text-lead` body; X articles use the target article's `data-article-part=body` / `.x-article-body`. Source permalinks match every requested numeric ID and author (case-insensitive author spelling). Actual destination remained the requested status URL, including its query; two authors use canonical capitalization.
- Compared every nonempty source body unit without a minimum-length threshold, including short paragraphs, headings, list items and quotations. Composite lists and quotations are compared as complete units; therefore units are **not** a paragraph count. Source paragraph counts are also provided. No paragraph is silently dropped because it is short. First and last authored text are included in coverage. Existing article titles are retained; ordinary posts have no article section headings.
- Normalization removes Markdown link/image syntax while retaining labels, unescapes Markdown punctuation, removes emphasis/heading and line-leading quote formatting, and ignores whitespace layout. It does not remove missing substantive text. This avoids treating wrapped links, escaped numbered lists, emphasis or quote prefixes as content loss.
- Excluded nested social-card UI, quote/reply containers, scripts/styles and aria-hidden decorative labels from source-body expectations. Kenichi has three embedded tweet frames: their social-card UI is outside the target-authored body contract; adjacent authored descriptions remain covered. No login/signup/navigation/reply text was present in extracted output; explicit login-contamination checks were false for all nine sources. All source body units and authored headings below matched. Full source bodies are deliberately absent from this committed evidence.

| # | Exact public source | Source kind | Source p elements | Matched/all text units | Matched/all headings | Extracted JS string length | Verdict |
|---|---|---|---:|---:|---:|---:|---|
| 1 | https://x.com/romandevz/status/2114655026190618718?s=12 | post | 0 | 1/1 | 0/0 | 2460 | works |
| 2 | https://x.com/hitomiiim/status/2114704821852991771?s=12 | post | 0 | 1/1 | 0/0 | 1099 | works |
| 3 | https://x.com/adrianpunk115/status/2114648961399349608?s=12 | article | 35 | 44/44 | 9/9 | 3705 | works |
| 4 | https://x.com/0xluffy_eth/status/2114705266759569624?s=12 | post | 0 | 1/1 | 0/0 | 795 | works |
| 5 | https://x.com/kenichiota0711/status/2113959093220335871?s=12 | article | 28 | 38/38 | 7/7 | 5802 | works |
| 6 | https://x.com/katelyn_lesse/status/2114542280535318557?s=12 | article | 12 | 19/19 | 7/7 | 6680 | works |
| 7 | https://x.com/thorstenball/status/2114386050638336082?s=12 | article | 16 | 16/16 | 0/0 | 3963 | works |
| 8 | https://x.com/shota7180/status/2113460983100530808?s=12 | article | 270 | 288/288 | 10/10 | 17777 | works |
| 9 | https://felixrieseberg.com/how-to-be-more-ambitious-a-tutorial-for-humans-with-models/ | ordinary article | 21 | 43/43 | 4/4 | 13616 | works |

Lengths use JavaScript UTF-16 string length, not Go rune counts; differences in harness rune counts are expected for emoji. Comparison results are captured at `/private/tmp/rlss-accept-comparison.jsonl`.

## 2. Complete Felix article

- Same observer command without `-headed`, same fresh profile and rebuilt embedded bundle. Navigation, observation and extraction errors were all nil. Canonical identity matches the requested Felix URL.
- Source root was `article .gh-content`, including the substantive dated container. All 43 nonempty paragraph/list/heading/quotation units matched; 21 source paragraph elements and all four authored section headings were covered. Decorative aria-hidden PART labels were excluded; actual section names were retained. No source text or section was removed due to date metadata.
- Actual CLI command: `/private/tmp/rlss-accept-bin url '<Felix exact URL above>' --no-summary --config /private/tmp/rlss-accept-config.yaml --verbose`. Exit 0; `Total: 1, Success: 1, Skipped: 0, Partial: 0, Failed: 0`; content only. It succeeded headlessly without retry. Saved content includes the observer's entire extracted Markdown byte-for-byte. Extracted body SHA256: `765e143ec003df907a6865c9e540eb001781df44b94bd1ae9e7e9e501b1d093c`.

## 3. Safe failure and data preservation

- Actual negative command: `/private/tmp/rlss-accept-bin url 'https://x.com/romandevz/status/1?s=12' --headed --no-summary --config /private/tmp/rlss-accept-config.yaml --verbose`. Exit 1; `Inaccessible social target: main body unavailable`; `Success: 0, Failed: 1`; no output content for this URL. Captured log `/private/tmp/rlss-accept-cli-failure.log`.
- Focused source/embedded-boundary regressions: `node --test --test-name-pattern='x-public|x-preview|ordinary-dat|target bodies|wrong target|explicit target previews|challenge identity|login, preview' extraction.test.mjs`: 9 tests, 9 passed, 0 failed. Covers complete new public posts/articles, legacy bodies excluding replies, wrong targets, explicit previews, ambiguous/incomplete bodies, quoted bodies and verification/login rejection. Log `/private/tmp/rlss-accept-js.log`.
- `GOCACHE=/private/tmp/rlss-accept-gocache go test ./internal/pipeline -run 'TestProcessItemWithFallback|TestProcessBatchFailedPaginationPreservesFiles|TestExtractionRejectsObservedWalls|TestProcessItemRejectsCachedWall|TestIsSkipped|TestClassifyResult|TestStats_Partial' -v`: PASS. Failed attempts do not write a new body or summary; failed extraction with existing content+summary preserves both byte-for-byte. Log `/private/tmp/rlss-accept-go.log`.
- Independent temporary probe `test_pipeline_failed_target_preserves_body_summary_and_watch_retry` (`TestWitnessPreserveAndRetry`): wrote a pair of synthetic existing files, ran a target-unavailable extractor through the actual pipeline, reset context and rebuilt index for another cycle, reran, and asserted both files remain byte-identical each time. Both cycles returned Failed 1; PASS. No production files were involved.

## 4. Existing successful extraction, retries and settings

- Above 9 focused JS regressions cover legacy X and ordinary articles alongside new body layouts; full package suite intentionally left to finalize-review, which runs before acceptance and blocks it on failure. The station confirmed package-tests is not skipped and finalize-review will run. Full suite command owned by that check: `go test ./...` (not run by this acceptance witness).
- Live actual CLI: `/private/tmp/rlss-accept-bin url 'https://x.com/romandevz/status/2114655026190618718?s=12' --no-summary --config /private/tmp/rlss-accept-config.yaml --verbose`. At 20:22:40 +08:00 initial headless navigation failed with `net::ERR_HTTP_RESPONSE_CODE_FAILURE`; log explicitly says `headless extraction failed, retrying with headed mode`. At 20:22:54 content was saved; exit 0, Success 1, Failed 0. Saved body exactly includes the independently observed entire body. SHA256 `3ef8838dcb867bd5f0688025af51db14b8cfa9e97d086c1f3a4e93b19b4b152b`. Log `/private/tmp/rlss-accept-cli-x.log`.
- `TestProcessItemWithFallback` passed success, explicit-headed no-retry, failed/empty headless retry, enabled challenge retry, disabled challenge no-retry, disabled-switch navigation/empty retry, and failed retry validation cases. This checks false-setting effects with deterministic fixtures, rather than falsely claiming a live verification challenge occurred.

## 5. Complete-source UI flow

Both real CLI commands in sections 2/4 ran through the shipped Go embedded entry, wrote only `__content.md`, showed Success 1 / Failed 0, and exited 0. X's saved body and Felix's saved body both match independently captured source-validated extraction. All nine browser-entry outputs passed actual-source completeness comparison in section 1.

## 6. Summarization-enabled UI flow

Independent temporary probe `test_pipeline_summary_failure_partial_and_batch_continues` (`TestWitnessSummaryBatch`) called the actual pipeline with a complete synthetic article, real temporary output storage, and an isolated deterministic Summarizer implementation. Success mode returned Success 1; provider-error mode returned Partial 1, with content retained; each batch also contained a preceding extraction failure. Both modes passed. No live provider, credential or user configuration was accessed. This proves pipeline behavior, not availability of an external provider.

## 7. In-progress UI flow

Section 4's live CLI logs show initial processing, headless failure, headed retry, content write, and final success. Eight explicit headed observer invocations also succeeded using fresh profiles. Section 2's Felix live CLI remained headless. No extra retry occurred for explicit-headed failures (focused pipeline test).

## 8. Verification restriction UI flow

Focused `TestProcessItemWithFallback` cases PASS: specialized challenge disabled, wrapped pagination challenge disabled, block-disabled challenge, ordinary challenge enabled, block-disabled navigation, block-disabled empty. These validate both false and true settings and ordinary-error retry retention. Live X navigation failure recovery independently verifies real browser retry integration; no claim is made that the live source served a verification page.

## 9. Missing/unverified-body UI flow

Section 3 real inaccessible URL exited 1, displayed Failed 1 and the unavailable-main-body category, and wrote no body. Focused preview/wrong-target/wall regressions and independent preservation probe cover incomplete or mismatched bodies and existing-file retention. Retry after another watch-cycle reset again returned an honest failure while preserving the files.

## 10. Batch/watch UI flow

Independent `TestWitnessSummaryBatch` submitted bad then good URLs through real `ProcessBatch`. It asserted Failed 1 plus Success 1 or Partial 1 according to the summary provider's response: the first failure did not stop the second item. `TestWitnessPreserveAndRetry` reset pipeline context and rebuilt its index before a second failed-target cycle. `TestStats_PartialKeptSeparateFromSuccessAndFailed` and `TestClassifyResult` passed all four result categories. The scheduler's real waiting interval was not exercised; cycle processing and retry behavior were.

## 11. Already-processed UI flow

`GOCACHE=/private/tmp/rlss-accept-gocache go test ./internal/pipeline -run 'TestProcessItemVideoResume|TestStats_Report|TestValidateSummaryText' -v`: PASS, including completed and dry-run cases that assert no extractor calls and skipped result. Log `/private/tmp/rlss-accept-flows.log`. Single URL commands exercised forced processing; no bulk data migration or rewrite occurs. Existing pending content retains its documented resume path.

## 12. Empty/dry-run UI flow

- `/private/tmp/rlss-accept-bin url 'https://example.com/dry-run-only' --dry-run --no-summary --config /private/tmp/rlss-accept-config.yaml`: exit 0, `dry-run: would process`, Total 1 / Skipped 1 / Success 0 / Failed 0; no body for that URL.
- `/private/tmp/rlss-accept-bin process --no-summary --config /private/tmp/rlss-accept-config.yaml`: exit 0, `no items to process`; all sources disabled in isolated configuration. Logs `/private/tmp/rlss-accept-cli-dry.log` and `...-empty.log`.

## 13. Isolated verification UI flow

All nine exact URLs were independently fetched with new temporary browser profiles and source DOM captured before extraction. Temporary config selected an isolated missing Local State and no inherited personal profile; output stayed under `/private/tmp`. No command inspected or modified production notes/configuration, replaced the production executable or controlled the live watch process. Full article captures and browser data are excluded from Git. Final clean test worktree contains no tracked changes after removal of temporary probes; rebuilding the committed bundle produced no diff.

## Independent temporary probe run

Command `GOCACHE=/private/tmp/rlss-accept-gocache go test ./internal/pipeline -run TestWitness -v`: PASS, both named witness tests. Probe identifiers/docstrings:

- `test_pipeline_summary_failure_partial_and_batch_continues`: exercises real storage and classification without an external provider.
- `test_pipeline_failed_target_preserves_body_summary_and_watch_retry`: tests an existing pair byte for byte and retries after a cycle reset.

Log `/private/tmp/rlss-accept-witness.log`. Probe source is retained only in `/private/tmp/rlss-accept-witness-test.go`; no production fix or test modification is committed by this witness.

## Decisions, skips and limitations

- Fresh anonymous browser profiles test public access; later source edits, access changes or outages can change results.
- Markdown formatting normalization is necessary for faithful body comparison; no minimum paragraph length filter was used.
- A deterministic summary provider tests success/error branching without invoking the user's service. A full timed watch daemon run was unnecessary to the changed extraction contract; real cycle processing/reset and batch continuation were exercised.
- No instructed step was skipped. The complete package suite belongs to finalize-review, not this witness. No important dismissed findings were supplied. No acceptance failures or unverified samples remain in this run.
