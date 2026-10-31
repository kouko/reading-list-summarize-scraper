# Complete Facebook and dated articles — plan
intent: 2026-10-31-complete-facebook-and-dated-articles@fc198038870a9f5a2af6b2b9f5e8575c4fd8e4ae
spec: docs/loom/2026-10-31-complete-facebook-and-dated-articles/spec.md@dbe267a
charter: 1.1

## Task DAG

### Wave 0 — Dated article parser boundary
**W0-01 Preserve complete dated article bodies** after: none acceptance: 2
- Files: extraction.ts (entry: extractArticle), extraction.test.mjs
- Test: A2 positive: dated-body-complete; negative: dated-body-unit-loss.
- Risk: agent-decided — REQ-2/3 narrowly protect owned prose on a copy; preserve and widen extraction.test.mjs coverage for metadata cleanup, hidden interface removal and source immutability.

### Wave 1 — Facebook adapter boundary
**W1-01 Extract verified Facebook post bodies and images** after: W0-01 acceptance: 1
- Files: facebook.ts (entry: Facebook extraction), extraction.ts (entry: extractArticle), extraction.test.mjs
- Test: A1 positive: facebook-owned-complete; negative: facebook-source-conflict.
- Risk: agent-decided — REQ-1/3 require bounded public data and owned text/media correspondence; preserve extraction.test.mjs video/social rejection coverage and widen Facebook identity, glyph and attachment cases.

### Wave 2 — Browser delivery boundary
**W2-01 Bind delivered content to requested targets and classify failures** after: W1-01 acceptance: 3, 4, 5
- Files: inject.ts (entry: window.extractArticle), internal/extract/browser.go (entry: Extract), internal/extract/browser_test.go
- Test: A3 positive: requested-target-delivered; negative: redirected-target-rejected. A4 positive: scoped-error-category; negative: cause-lost. A5 positive: browser-entries-retained; boundary: pagination-entry-retained.
- Risk: agent-decided — REQ-3/4/5 bind both ordinary and paginated browser calls; preserve and widen browser_test.go target, timeout, social and pagination coverage without changing experimental thread routing.

### Wave 3 — Pipeline cache and save boundary
**W3-01 Revalidate scoped caches before summary and preserve notes on failure** after: W2-01 acceptance: 3, 4, 5
- Files: internal/pipeline/runner.go (entry: ProcessItem), internal/pipeline/extraction.go (entry: extractWithFallback), internal/pipeline/pipeline_test.go, internal/pipeline/extraction_test.go
- Test: A3 positive: scoped-cache-fresh; negative: scoped-cache-failure-preserves. A4 positive: failure-category-retained; negative: retry-cause-lost. A5 positive: ordinary-cache-resume; boundary: blocked-retry-disabled.
- Risk: agent-decided — REQ-3/4/5 preserve pipeline_test.go and extraction_test.go resume, completed-skip, retry and save-boundary coverage; widen only in-scope unverified cache handling, retaining unrelated ordinary resumes.

### Wave 4 — Embedded integration and actual-source evidence
**W4-01 Verify final embedded output against source and document behavior** after: W3-01 acceptance: 6
- Files: embed/defuddle.min.js, README.md, docs/loom/2026-10-31-complete-facebook-and-dated-articles/evidence/acceptance-test-evidence.md, docs/loom/memory/
- Test: A6 positive: source-compared-final-cli; boundary: unavailable-source-disclosed.
- Risk: agent-decided — REQ-6 uses isolated profiles/config/output and final bundle; keep full copyrighted captures and private browser state uncommitted, record actual-source changes and unavailable cases honestly.

### Wave 5 — Adversarial identity regression
**W5-01 Reject contradictory Facebook source identities** after: W4-01 acceptance: 1, 3
- Files: facebook.ts, extraction.test.mjs, embed/defuddle.min.js
- Test: A3 negative: committed facebook-identity.test.mjs adversarial probe must change from RED to GREEN; existing Facebook alias positives remain green.
- Risk: agent-decided — Sweep requested/native/permalink/actor identities and duplicate public records as one contradiction class; preserve verified pfbid aliases and media ownership checks.

**W5-02 Graduate the earned identity probe into maintained tests** after: W5-01 acceptance: 3
- Files: docs/loom/2026-10-31-complete-facebook-and-dated-articles/evidence/probes/facebook-identity.test.mjs (move only), facebook-identity.test.mjs (intermediate move only), test_facebook_conflictingPermalink_rejected.mjs, extraction_test.go, package.json
- Test: A3 negative: unchanged earned probe runs through npm test and go test -run '^TestExtraction$' .
- Risk: agent-decided — Move the program unchanged into existing runners. Match its filename to its existing test name: Node 22 JUnit omits testcase.file, so reviewers must identify genuine passing output without modifying the program or synthesizing evidence.

### Wave 6 — Cross-record attachment consistency
**W6-01 Reject contradictory supplied attachment evidence across target records** after: W5-02 acceptance: 1, 3
- Files: facebook.ts (entry: facebook), extraction.test.mjs, embed/defuddle.min.js
- Test: A3 negative: extend facebook-source-conflict with supplied duplicate-record attachment extent/media contradictions, observe RED then GREEN; existing complete Facebook extraction remains green.
- Risk: agent-decided — Matching records without full messages bypass attachment reconciliation. Reconcile supplied attachment extent/media across matching records and nested stories; absent optional fields assert no contradiction. Preserve native/DOM ownership and original media checks.

### Wave 7 — Native suite evidence binding
**W7-01 Bind adversarial execution to the existing Go suite entry** after: W6-01 acceptance: 3
- Files: extraction_test.go (existing TestExtraction runner, concern comment only), test_facebook_conflictingPermalink_rejected.mjs (read-only retained program)
- Test: A3 negative: go test -json . -run '^TestExtraction$' executes the unchanged earned JavaScript regression through its maintained Go entry.
- Risk: agent-decided — Loom's Go collector recognizes only selected _test.go files, not child JavaScript programs. Reuse the existing Go entry as execution evidence; the adversary adds its concern annotation. Preserve the earned JavaScript program, runner behavior and standing suite.

## Simplicity check
- none found

## Questions asked
① — what — 以上是否就是你要的範圍？
② — behaviour — 上述執行後的行為是否符合你的預期？

## Risks
1. agent-decided — Source shapes can change; acceptance must independently reload both scoped sources and all 11 X regression samples, with no production profile or note writes.
2. agent-decided — Facebook's text-only research does not prove attachment completeness; final implementation must verify target-owned media extent and reject missing or ambiguous attachments.
3. agent-decided — Both scoped pending caches require fresh extraction; unrelated ordinary resume and completed-summary skips remain. Additional schema/configuration or a parser-switching framework is unnecessary.
4. agent-decided — Separate production code, tests and generated bundle growth; the spec's 100–200 production-line estimate is unverified and will be recomputed from the final tree.
5. agent-decided — Spec review fixed two internal clarifications without changing any of the six confirmed requirement sentences; retain the original approval question record and refresh the spec content binding before planning.
6. agent-decided — Full package verification is go test ./...; final Build also runs committed independent adversarial programs before closing review. No step is skipped by instruction.
