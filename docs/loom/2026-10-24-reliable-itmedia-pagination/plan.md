# Reliable ITmedia pagination — plan
intent: 2026-10-24-reliable-itmedia-pagination@c721bfb
spec: docs/loom/2026-10-24-reliable-itmedia-pagination/spec.md@c61ed4a
charter: 1.1

## Task DAG
**W1-01 Correct configured pagination completion and readiness** after: none acceptance: 1, 2, 3
- Files: internal/extract/browser.go (entry: Browser.ExtractPaginated), internal/extract/browser_test.go, internal/extract/pagination.go, internal/extract/pagination_test.go
- Test: A1 positive: all-pages-in-order; boundary: duplicate-continuation. A2 positive: complete-terminal-page; negative: later-page-failure. A3 positive: ready-body-pending-subresource; negative: navigation-error. Retain existing pagination tests, updating fail-soft expectations to the confirmed contract.
- Risk: agent-decided — Shared walker affects configured sites. Preserve ordinary extraction and caller retry policy; use grounded pinned navigation APIs, bounded document readiness, per-page validity and explicit continuation errors. Keep original failure classification in returned errors.

**W2-01 Verify retry and output integration** after: W1-01 acceptance: 2, 3
- Files: internal/pipeline/extraction_test.go, internal/pipeline/runner_test.go, internal/pipeline/extraction.go (entry: extractWithFallback), internal/pipeline/runner.go (entry: Pipeline.ProcessBatch)
- Test: A2 positive: complete-retry-saves-only-complete-body; negative: failed-pagination-preserves-existing-file. A3 positive: retry-continues-pagination; negative: challenge-disabled-no-headed-retry. Extend existing entry-level tests; alter production caller only if an observed integration defect requires it.
- Risk: agent-decided — Avoid parallel policy copies: lower-level pagination must return the original failure cause so existing fallback remains authoritative. No partial content may leak through retry, storage or summary paths.

**W3-01 Expand and run real Japanese news examples** after: W2-01 acceptance: 1, 3, 4
- Files: docs/loom/2026-10-24-reliable-itmedia-pagination/evidence/news-pagination-cases.json, docs/loom/2026-10-24-reliable-itmedia-pagination/evidence/real-site-validation.md
- Test: A1 positive: demonstrated-itmedia-five-pages; negative: real-itmedia-cap-incomplete. A3 positive: prior-working-publishers; negative: inaccessible-page-not-counted-success. A4 positive: new-publishers-and-itmedia-articles; negative: unverified-candidate-not-counted-pass.
- Risk: agent-decided — Aim for twelve articles across ten publishers, retaining original seven plus three new publishers and two more ITmedia articles. Use isolated live CLI and independent per-page observations; report restrictions and incomplete evidence honestly. Commit sanitized metadata only.

**W4-01 Graduate the earned later-entry regression unchanged** after: W3-01 acceptance: 2
- Files: docs/loom/2026-10-24-reliable-itmedia-pagination/evidence/probes/pagination_entry_probe_test.go, internal/extract/pagination_entry_probe_test.go
- Test: A2 positive: reject-known-incomplete-entry; negative: later-entry-returns-suffix. The independent probe owner moves its committed RED program byte-for-byte into the normal Go package suite before the product fix.
- Risk: agent-decided — User-requested non-first-page testing and the independent adversary exposed the same omission. Only the probe owner moves its program; neither its assertion nor the demonstrated failure is weakened.

**W4-02 Reject known missing earlier pages** after: W4-01 acceptance: 2
- Files: internal/extract/browser.go (entry: Browser.ExtractPaginated), internal/extract/browser_test.go
- Test: A2 positive: first-page-traversal-with-earlier-links-on-later-pages; negative: initial-earlier-page-evidence. Retain the graduated regression unchanged; cover standard first/previous metadata and same-article lower numbered page navigation observed on the four real later-entry examples.
- Risk: agent-decided — Class: known incomplete initial entry returns success. Search all changed files and initial/next navigation surfaces. Reject justified earlier-page evidence only at initial entry; do not add automatic first-page recovery, new settings or site-specific extractors.

**W5-01 Validate each page before combining article content** after: W4-02 acceptance: 2, 3
- Files: internal/extract/browser.go (entry: Browser.ExtractPaginated), internal/extract/browser_test.go, internal/extract/content.go (entry: ValidateContent), internal/pipeline/extraction.go, internal/pipeline/runner.go, internal/pipeline/pipeline_test.go, internal/pipeline/extraction_test.go
- Test: A2 positive: article-discussing-login-is-retained; negative: real-extractor-later-login-wall. A3 positive: shared-validation-retains-retry-classification; negative: later-challenge-with-headed-retry-disabled. Extend the browser entry fixture using the real injected extractor; retain the graduated probe unchanged.
- Risk: agent-decided — Class: earlier prose masks a recognized later-page wall. Search changed files, TypeScript bridge, browser/walker, validation, retry and storage. Share existing validation before combining pages; avoid duplicate patterns or dependency cycles. Preserve current policy and ordinary extraction.

**W6-01 Exercise incomplete sequences through the browser entry** after: W5-01 acceptance: 2, 3
- Files: internal/extract/browser_test.go (entry: Browser.ExtractPaginated), internal/extract/pagination_test.go
- Test: A2 positive: complete-sequence-retained; negative: table-of-incomplete-sequences. Move the changed private-walker cap, loop, later failure, unusable continuation and empty-page tests into one table-driven subtest of the existing browser entry test.
- Risk: agent-decided — Class: changed tests bypass the declared browser entry. Search changed tests and affected Acceptance surfaces. Remove corresponding direct tests and newly unused helpers/imports; retain unrelated tests and the independent probe. Change tests only.

**W7-01 Isolate Chrome fixture startup and graceful cleanup** after: W6-01 acceptance: 2, 3
- Files: internal/extract/browser_test.go (entry: Browser.ExtractPaginated), internal/extract/pagination_entry_probe_test.go (independent owner only)
- Test: A2 positive: actual-browser-ready-before-negative-cases; negative: startup-error-cannot-pass-rejection. A3 positive: existing-navigation-deadlines-retained; negative: profile-cleanup-race. Reuse existing regressions that failed CI; retain all extraction assertions and navigation deadlines.
- Risk: agent-decided — Class: cold startup consumes navigation timeout and abrupt cleanup races profile removal. Search changed files and browser lifecycle surfaces. Initialize one test-owned browser and close it gracefully. The independent probe owner updates setup only; production remains unchanged.

## Simplicity check
- none found

## Questions asked
① — what — 確認以這個範圍進行嗎？同意後，審查完成會自動推送並更新／建立 PR；合併仍另行確認，你也可在發布前取消自動推送。
② — behaviour — 以上行為符合你的期待嗎？

## Risks
1. agent-decided — Root principles require correctness over retaining partial bodies. The same class covers late errors, unusable pages, caps, loops and contradictory standard next hints. Missing selectors without independent continuation evidence remain an explicitly documented limitation.
2. agent-decided — Browser load completion and actual article readiness differ. Validate the pinned API's navigation error text and document lifecycle; prove that a pending ancillary resource does not block a ready article without accepting stale or empty documents.
3. user-decided — Preserve existing headed retry configuration and headed_on_block: false. The user confirmed no automatic pagination discovery and no batch rewrite of old notes; successful saves retain existing overwrite behavior.
4. agent-decided — New examples verify the common fix, not a promise to implement every publisher's special extractor. No paid account, proxy, provider review, real browser-profile reuse or existing-data modification is authorized.
5. agent-decided — A probe that catches a defect must graduate into the Go package suite before review finalization; the probe owner retains its content. Existing source extraction tests and earlier earned JavaScript regressions remain in the complete suite.
6. agent-decided — Non-first entry is another instance of confirmed incomplete-success failure, not a new automatic-discovery feature. Explicit earlier-page evidence requires failure with no partial save; insufficient evidence remains a documented limitation. W4 records the newly demonstrated boundary.
7. agent-decided — CI 39995109168 exposed cold startup consuming fixture deadlines and asynchronous profile cleanup. W7 changes tests only; independent acceptance at 7e7069a retains its product results. Complete tests and the existing probe must pass again before terminal review.
