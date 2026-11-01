# Original-first YouTube subtitles — plan
intent: 2026-11-01-verify-youtube-original-audio-evidence@88e1abe
spec: docs/loom/2026-11-01-verify-youtube-original-audio-evidence/spec.md@b8925cb
charter: 1.1

## Task DAG
### Wave 0 — Original-first extraction
**W0-01 Original evidence and labeled fallback selection** after: none acceptance: 1, 2, 3
- Files: extraction.ts (entry: extractArticle), extraction.test.mjs, extraction-boundaries.test.mjs
- Test: A1 positive: multi_audio_original; boundary: conflicting_original_metadata. A2 positive: single_audio_human_auto; boundary: empty_human_auto_recovery. A3 positive: labeled_unknown_and_translation; negative: invalid_translation_endpoint.
- Risk: agent-decided — extend existing selector/downloader per REQ-1–3; extraction.test.mjs coverage widens language eligibility while preserving original-first preference, target and XML checks; revise only obsolete strict-language rejection assertions.

### Wave 1 — Consumer and reproducible evidence
**W1-01 Validate and preserve original/fallback provenance** after: W0-01 acceptance: 3, 4
- Files: internal/pipeline/extraction.go (entry: validateTargetContent), internal/pipeline/extraction_test.go, internal/pipeline/runner.go (entry: ProcessItem), README.md
- Test: A3 positive: fallback_saved_and_summarized; negative: misleading_language_status. A4 positive: valid_old_and_new_provenance; negative: wrong_target_empty_or_unmarked_cache.
- Risk: agent-decided — update producer/consumer together per REQ-3–4; internal/pipeline/extraction_test.go coverage widens accepted evidence without admitting unmarked caches; runner changes only if existing behavior needs integration correction.

**W1-02 Verify actual video evidence and subtitle correspondence** after: W1-01 acceptance: 5
- Files: docs/loom/2026-11-01-verify-youtube-original-audio-evidence/evidence/youtube-verification.md
- Test: A5 positive: live_target_language_source_cue_match; boundary: unavailable_source_reported_honestly.
- Risk: agent-decided — isolated public-source runs per REQ-5; keep raw captions and signed URLs outside Git, record bounded source metadata and outcomes; stale captures are replay evidence, never fresh live success.

## Simplicity check
- none found

## Questions asked
① — what — 是否確認這個範圍？ 同意也代表授權測試與審查通過後推送分支、建立待合併的 PR；合併仍另行確認，你也可取消自動發布。
② — behaviour — 確認以上行為嗎？
② — behaviour — 確認以上行為後，我就接續實作與驗證。

## Risks
1. agent-decided — original evidence is undocumented platform data; missing/contradictory signals permit only labeled fallback, never a verified-original claim.
2. agent-decided — original preference spans bounded native retries, so an available inline fallback must wait until verified-original attempts finish; no new unbounded network loop.
3. agent-decided — language relaxation changes an extraction-to-cache contract; update validation and source labels together and retain old valid provenance compatibility.
4. agent-decided — no authorization for paid services, external reviewers, data migration or browser profile changes; use reversible local changes and isolated profiles only.
