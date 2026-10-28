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

### Wave 4 — Contradictory membership regression

**W4-01 Reject contradictory article membership and retain the earned regression** after: W3-01 acceptance: 2, 3
- Files: internal/extract/threads.go (collectSocial), internal/extract/threads_test.go, docs/loom/2026-10-28-extract-complete-author-threads/evidence/probes/conflicting_root_test.go, internal/extract/probes/conflicting_root_test.go
- Test: A2 positive: consistent-source-membership; negative: foreign-root-member. A3 positive: verified-complete-membership; negative: contradictory-reply-role.
- Risk: agent-decided — REQ-2–3: reject explicit ancestry or reply-role contradictions instead of trusting the sequence alone; widen threads_test.go coverage. Move the earned adversarial program unchanged into the permanent Go suite so its package remains isolated.

### Wave 5 — Saved member provenance

**W5-01 Retain a source permalink with every verified member body** after: W4-01 acceptance: 2
- Files: internal/extract/threads.go (collectSocial), internal/extract/threads_test.go, internal/extract/browser_test.go
- Test: A2 positive: verified-member-permalinks; negative: failed-thread-no-output.
- Risk: agent-decided — REQ-2: append each member's verified permalink without changing its text or following linked content; preserve and widen threads_test.go and browser_test.go coverage. Keep input storage identity, ordering, deduplication and incomplete-result rejection.

### Wave 6 — Reachable X completeness evidence

**W6-01 Establish X direct-reply extent from exact source counts** after: W5-01 acceptance: 1, 3, 4
- Files: extraction.ts (socialSnapshot X producer), extraction.test.mjs, embed/defuddle.min.js
- Test: A1 positive: exact-direct-child-count; A3 negative: missing-direct-child; A4 positive: bounded-complete-X-source.
- Risk: agent-decided — REQ-1,3–4: require exact nonnegative source counts and unique explicit direct children within one snapshot. Unknown or mismatched counts, display-group endings and timeline termination remain insufficient. Existing incomplete live samples remain failures.

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
5. agent-decided — Acceptance F1 remains open: the observed public X guest timeline termination does not establish exhaustive author membership, and Threads AI-colleague direct pagination remains pending. Preserve failed acceptance until permitted complete source evidence is available.
6. agent-decided — X selects partial conversation branches (https://help.x.com/en/resources/recommender-systems/conversations-recommendations); counts measure direct replies (https://help.x.com/en/using-x/x-conversations). Dashboard reports eleven direct replies but exposes two explicit direct children. Termination and module tip IDs do not prove author completeness.
