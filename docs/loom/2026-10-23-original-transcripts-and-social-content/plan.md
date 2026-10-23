# Original transcripts and social content — plan
intent: 2026-10-23-original-transcripts-and-social-content@7565660
spec: docs/loom/2026-10-23-original-transcripts-and-social-content/spec.md@10e7241
charter: 1.1

## Task DAG
### Wave 0 — browser-side extraction policy
**W0-01 Validate first-party transcripts and social bodies** after: none acceptance: 1, 2, 3, 4
- Files: inject.ts (entry: window.extractArticle), extraction.ts (entry: extractArticle), extraction.test.mjs, package.json, package-lock.json, embed/defuddle.min.js
- Test: A1 positive: verified-original; negative: translated-unknown. A2 positive: three-platform-bodies; negative: wrong-target. A3 positive: verified-content; negative: login-preview-description. A4 positive: usable-auto-after-empty-human; boundary: malformed-incomplete-cues.
- Risk: agent-decided — REQ-1–4 require constrained first-party data and real content validation; pin 0.19.5 and test the shipped module entry with real DOM fixtures and network-boundary responses. Preserve the existing string-facing browser contract; ordinary articles and pagination remain compatible.

### Wave 1 — saved-content and pipeline integration
**W1-01 Enforce verified video content through saving and resume** after: W0-01 acceptance: 1, 2, 3, 4
- Files: internal/pipeline/runner.go (entry: Pipeline.ProcessItem), internal/pipeline/extraction.go, internal/pipeline/extraction_test.go, README.md, .github/workflows/build.yml
- Test: A1 positive: saved-language-source; negative: cached-unverified-video. A2 positive: saved-social-body; negative: wall-not-saved. A3 positive: continue-after-failure; boundary: completed-skip-dry-run. A4 positive: validated-transcript-saved; negative: unusable-transcript-not-summarized.
- Risk: agent-decided — REQ-1/REQ-3 preserve structured video provenance inside the existing string content boundary; validate target-bound provenance before saving or resuming. Preserve and extend internal/pipeline/extraction_test.go coverage for headed_on_block, retries, ordinary caches, and failed writes; no migration or automatic completed-note rewrite.

### Wave 2 — adversarial boundary corrections
**W2-01 Reject quoted identity and malformed caption markup** after: W1-01 acceptance: 2, 4
- Files: extraction.ts (entry: extractArticle), extraction.test.mjs, extraction-boundaries.test.mjs, package.json, package-lock.json, embed/defuddle.min.js
- Test: A2 positive: own-post-permalink; negative: quoted-target-permalink. A4 positive: valid-caption-markup; negative: mismatched-inner-markup. Run the existing adversarial program RED before the fix, retain its bytes unchanged as extraction-boundaries.test.mjs, and include it in npm test.
- Risk: agent-decided — reject descendant identity mistaken for ownership and HTML repair mistaken for XML validity. Search changed files and social/caption surfaces; preserve valid content, regenerate the bundle, and graduate the adversary's regression unchanged into CI.

### Wave 3 — fresh first-party caption URLs
**W3-01 Refresh unusable inline caption URLs through the native player API** after: W2-01 acceptance: 1, 3, 4
- Files: extraction.ts (entry: extractArticle), extraction.test.mjs, embed/defuddle.min.js, README.md
- Test: A1 positive: native-player-original-caption; negative: refreshed-wrong-target-or-ambiguous-language. A3 positive: empty-inline-recovers; negative: native-api-failure-remains-failed. A4 positive: refreshed-usable-auto-after-empty-human; negative: refreshed-empty-or-malformed-cues. Boundary: captions parse on pages enforcing Trusted Types without HTML string assignment.
- Risk: agent-decided — live page URLs return zero bytes while upstream's first-party iOS/Android player requests return usable same-video captions. Refresh only through fixed official endpoints, reapply all identity/original-language/cue checks, and preserve timeouts and existing recovery settings.

### Wave 4 — explicit social truncation
**W4-01 Reject target-owned X preview text** after: W3-01 acceptance: 2, 3
- Files: extraction.ts (entry: extractArticle), extraction.test.mjs, embed/defuddle.min.js
- Test: A2 positive: complete-target-with-unrelated-quoted-preview; negative: target-owned-show-more. A3 positive: complete-canonical-read-view; negative: explicit-truncated-target-body.
- Risk: agent-decided — both closing reviewers found explicit X truncation ignored before success. Reject remaining target-owned expansion markers; preserve complete bodies and ignore unrelated quoted markers. Bilibili scope and live X access remain unresolved, not dismissed.

### Wave 5 — preserve challenge retry policy
**W5-01 Retain challenge identity through specialized extraction errors** after: W4-01 acceptance: 3
- Files: extraction.ts (entry: extractArticle), extraction.test.mjs, embed/defuddle.min.js
- Test: A3 positive: ordinary-access-error-keeps-normal-retry; negative: specialized-challenge-retains-block-signal. Boundary: X, Medium and YouTube challenge pages; existing ProcessItem specialized-challenge-disabled case confirms the retry-policy consumer.
- Risk: agent-decided — live Medium's Cloudflare page becomes a generic canonical error, hiding the signal used by headed_on_block:false. Preserve a bounded challenge classification on extraction failure without returning page text or treating legitimate article mentions as challenges.

## Simplicity check
- none found

## Questions asked
① — what — 你能接受原語自動字幕嗎？
① — what — 字幕來源還有前一題待確認：沒有人工字幕時，你能接受平台提供的原語自動字幕嗎？
① — what — 確認這次就先實作以上完整範圍嗎？確認後會完成測試、審查並建立 PR；合併另行確認，發布前也可取消。
② — behaviour — 以上行為符合預期嗎？

## Risks
1. Original-audio evidence can be absent, ambiguous, or tied to a dubbed track; reject uncertainty instead of guessing from locale/default captions. No promise of audio-verified word accuracy or complete speech coverage.
2. New platform requests remain first-party and existing-profile-bound. Disable external aggregation/proxy fallbacks; no paid services, new accounts, audio uploads, private fixture publication, or changes to the running batch.
3. Fixture-backed tests exercise actual module entries; fake only network/resource boundaries. Independent acceptance also runs the built browser bundle and existing CLI pipeline in isolated output directories; live-site access limitations must remain explicit.
4. Add only a small DOM test dependency if required, reusing esbuild and Node's built-in test runner. Run npm test in existing CI and alongside the standing Go suite; finalization also executes the committed JavaScript adversarial command.
5. The prior general article data-date defect is known and not fixed by upstream 0.19.5; this scope does not claim its correction. The existing extraction mode/profile settings and completed-summary skip behavior remain unchanged.
