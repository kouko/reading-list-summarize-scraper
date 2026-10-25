# Recover the first page of a configured article — plan
intent: 2026-10-24-pagination-first-page-recovery@e23184b
spec: docs/loom/2026-10-24-pagination-first-page-recovery/spec.md@3bb9a27
charter: 1.1

## Task DAG
**W1-01 Recover and validate the complete article through the browser entry** after: none acceptance: 1, 2, 3
- Files: internal/extract/browser.go (entry: Browser.ExtractPaginated), internal/extract/pagination.go, internal/extract/browser_test.go, internal/extract/pagination_test.go
- Test: A1 positive: query-and-suffix-later-entry-recovers; negative: different-article-first-link. A2 positive: ordered-unique-complete-body; boundary: original-entry-not-in-chain. A3 positive: bounded-recovery-at-forward-cap; negative: unavailable-earlier-page.
- Risk: agent-decided — Apply REQ-1–3 identity and phase bounds without a new crawler. Existing browser_test.go coverage widens from rejecting later entries to complete recovery or justified failure; retain every readiness, wall and incomplete-sequence case. Keep the independent probe unchanged.

**W2-01 Verify output identity and retry preservation** after: W1-01 acceptance: 3, 4
- Files: internal/pipeline/extraction_test.go (entry: Pipeline.ProcessItem), internal/extract/pool_test.go (entry: Pool.ExtractURL)
- Test: A3 positive: recovered-body-only-after-complete-retry; negative: failed-recovery-preserves-old-output. A4 positive: supplied-url-remains-output-identity; negative: challenge-disabled-no-headed-retry.
- Risk: agent-decided — Apply REQ-3–4 through existing module entries. Reuse current fallback and file-protection tests; extend only missing recovery integration assertions. Verify both pool routes preserve the original URL; change production callers only for an observed defect.

**W3-01 Compare real recovered articles with first-page baselines** after: W2-01 acceptance: 1, 2, 3, 4
- Files: docs/loom/2026-10-24-pagination-first-page-recovery/evidence/real-site-validation.md, docs/loom/2026-10-24-pagination-first-page-recovery/evidence/recovery-cases.json
- Test: A1 positive: six-real-later-entries; negative: unverifiable-case-not-success. A2 positive: source-paragraphs-once-in-order; boundary: same-article-corpus. A3 positive: real-first-page-match; negative: low-cap-no-output. A4 positive: twelve-first-page-regressions; negative: failed-output-not-overwritten.
- Risk: agent-decided — Apply REQ-1–4 using isolated CLI and independent source-body checks. Include ITmedia3/5, PRESIDENT2/3, Yahoo2 and Toyokeizai2, plus the existing twelve first-page articles. Record actual restrictions honestly; keep raw copyrighted text and profiles temporary.

**W4-01 Preserve the inherited review record as historical evidence** after: W3-01 acceptance: 4
- Files: docs/loom/2026-10-24-reliable-itmedia-pagination/attestation.json, docs/loom/2026-10-24-pagination-first-page-recovery/evidence/pr22-attestation.json
- Test: A4 positive: cumulative-branch-identifies-current-change; boundary: archived-record-bytes-unchanged.
- Risk: agent-decided — Publication rejected two recognized attestations in this cumulative branch. Move the inherited generated record without editing its bytes; retain PR22's branch, runtime behavior and current generated record. Closing review must judge the new digest before regenerating evidence.

## Simplicity check
- none found

## Questions asked
① — what — 這個範圍符合你的需求嗎？確認後，審查完成會自動推送並建立／更新 PR；合併仍另行確認，你也可在發布前取消自動推送。
② — behaviour — 以上行為符合你的期待嗎？

## Risks
1. agent-decided — Same-article identity is the principal risk: labels and URL guesses alone cannot establish correctness. The recovered forward chain must include the requested page; missing evidence fails without output.
2. agent-decided — Recovery and forward reading have separate allowances using the existing effective page limit; discovery cannot consume the complete article's allowance. Retain per-page timeouts and reset both phases on retry.
3. user-decided — Preserve existing headless/headed retry settings, headed_on_block: false, output identity, successful overwrite, completed-item skip and failure-time file protection; no existing notes are migrated.
4. agent-decided — The baseline is PR22 commit 9d0da2f, retained in this isolated branch without merging or modifying that PR. Publishing remains authorized by this intent; merge requires final acceptance.
5. agent-decided — Production behavior outside configured pagination stays unchanged. Use the production-default timeout in browser fixtures, warm test-owned Chrome and await graceful cleanup; do not reintroduce arbitrary two-second CI deadlines.
