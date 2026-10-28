# Extract complete author threads — plan
intent: 2026-10-28-extract-complete-author-threads@183b82f
spec: docs/loom/2026-10-28-extract-complete-author-threads/spec.md@5b4361e
charter: 1.1

## Task DAG

### Wave 1 — First-party source interpretation

**W1-01 Parse verified social post relationships and complete bodies** after: none acceptance: 1, 2, 3, 4
- Files: extraction.ts (extractArticle and browser-facing social snapshot entry), extraction.test.mjs, extraction-boundaries.test.mjs, extraction_test.go, embed/defuddle.min.js
- Test: A1 positive: first-middle-source-equivalence; boundary: missing-source-member. A2 positive: ordered-distinct-author-posts; negative: reader-quote-recommendation. A3 positive: verified-complete-body; negative: unresolved-preview-boundary. A4 positive: standalone-and-existing-extraction; negative: challenge-body.
- Risk: agent-decided — REQ-1–4: reuse extraction entry and existing harness; widen extraction.test.mjs and preserve extraction-boundaries.test.mjs coverage. Incorporate only necessary reviewed modern-X selectors from PR24, leaving that branch and unrelated Felix changes untouched.

### Wave 2 — Bounded browser collection

**W2-01 Recover beginning and collect verified continuations through browser navigation** after: W1-01 acceptance: 1, 2, 3, 4
- Files: internal/extract/browser.go (Browser.Extract), internal/extract/threads.go, internal/extract/threads_test.go, internal/extract/browser_test.go
- Test: A1 positive: first-middle-complete-walk; boundary: unavailable-member. A2 positive: stable-id-order; negative: unrelated-branch. A3 positive: complete-only-return; negative: cursor-loop-cap. A4 positive: existing-browser-retry-contract; boundary: ordinary-article.
- Risk: agent-decided — REQ-1–4: use existing browser/session and extraction timeout, not a separate crawler or API client; require explicit beginning/end evidence. Preserve and widen browser_test.go; missing identity, ambiguity or unresolved limits fail closed.

### Wave 3 — Processing integration

**W3-01 Revalidate social cached bodies before summarizing and preserve failure artifacts** after: W2-01 acceptance: 3, 4
- Files: internal/pipeline/runner.go (ProcessItem), internal/pipeline/extraction.go, internal/pipeline/pipeline_test.go, internal/pipeline/extraction_test.go, README.md
- Test: A3 positive: verified-social-resume; negative: failed-refresh-preserves-files. A4 positive: completed-skip-and-force; boundary: ordinary-cache-summary-partial.
- Risk: agent-decided — REQ-3–4: bypass content-only social cache without changing storage identity or completed batch skips; preserve and widen pipeline_test.go and extraction_test.go coverage. Keep headed_on_block behavior and existing summary failure classification.

## Simplicity check
- none found

## Questions asked
① — what — 以上就是這次要實作的範圍，對嗎？
② — behaviour — 以下行為符合你的預期嗎？

## Risks
1. agent-decided — Public site schemas are unstable. Interpret explicit relationship fields, never adjacency alone; sanitize permanent fixtures and retain raw observations only in temporary storage. Live acceptance checks all sixteen first/middle entries across eight groups.
2. agent-decided — Navigation may expose previews, unavailable members or incomplete lists; use bounded retries and fail without replacing existing outputs. A successful short extraction is insufficient evidence of completeness.
3. agent-decided — Reuse the existing JavaScript test harness, register it in the declared Go package suite, and regenerate the embedded bundle deterministically. Run tasks sequentially, then independent adversary, complete package tests and independent acceptance.
4. agent-decided — No paid service, new account, bulk rewrite or PR24 merge is authorized or necessary; use reversible isolated development and leave production profiles, configuration, notes and running process untouched.
