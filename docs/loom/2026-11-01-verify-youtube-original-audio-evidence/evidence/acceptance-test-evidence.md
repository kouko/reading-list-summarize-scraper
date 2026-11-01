# YouTube original audio evidence — acceptance test evidence

Tried on 2026-11-01, clean detached worktree `/private/tmp/rlss-youtube-acceptance-fresh`, commit `f5c46a4de2fddc74497ee1bbc1864cf981b272ba`. No implementation or probes were changed. No earlier acceptance result is carried over.

## Setup and shared commands

- `git worktree add --detach /private/tmp/rlss-youtube-acceptance-fresh f5c46a4de2fddc74497ee1bbc1864cf981b272ba` succeeded after permission to register its shared git metadata.
- README full-build equivalents: `npm ci`, `npm run build:defuddle`, `go build -o /private/tmp/rlss-acceptance-cli ./cmd/rlss/` all succeeded. Go initially encountered sandbox denial for the existing build cache; the authorized retry succeeded. Fresh build left no tracked worktree changes.
- Study helper `/private/tmp/rlss-acceptance-study.ts` imports **this exact clean worktree's** `extraction.ts` and its own installed Defuddle dependencies, replacing the older helper's absolute imports. Bundled with this worktree's esbuild and compiled the existing Go browser driver. No old extraction bundle was used.
- Fresh live command: `RLSS_STUDY_METHODS=current /private/tmp/rlss-acceptance-study /private/tmp/rlss-acceptance-study.js /private/tmp/rlss-acceptance-live /private/tmp/rlss-acceptance-profile-20261101 headless rfscVS0vtbw aircAruvnKk dQw4w9WgXcQ DHjqpvDnNGE gBumdOWWMhY N0zhdMwD2Z8 aqz-KE-bpKQ`.
- All seven browser operations returned browserError `<nil>`. Full raw responses and cue text remain outside Git under `/private/tmp/rlss-acceptance-live/`; no signed endpoints, cookies or full captions are committed.
- Audit: `/private/tmp/verify-acceptance-live.py` adapts the existing parser only to this new output directory. It checks requested address and every captured player identity, parses actual downloaded XML, independently reconstructs substantive cue text and requires exact equality with saved body. All six successful bodies matched: **4,370 substantive cues, 324,483 characters**. Four verified-original, two original-unconfirmed, one no-subtitles. This checks downloaded text correspondence, not word-for-word audio accuracy or uninterrupted speech coverage.
- Focused browser command: `node --test --test-name-pattern='A1|A3 labeled|A3 refreshed|A4|wrong target|test_transcript|test_originalEvidence' extraction.test.mjs extraction-boundaries.test.mjs` → **17 pass, 0 fail**. Log: `/private/tmp/rlss-acceptance-js.log`.
- Focused pipeline command: `go test ./internal/pipeline -run 'TestProcessItemVideo|TestProcessItemWithFallback|TestProcessBatchFailedPaginationPreservesFiles' -count=1 -v` → PASS. Log: `/private/tmp/rlss-acceptance-go.log`.
- Full suite is intentionally owned by the final review gate: `go test ./... && npm test`. It runs before acceptance and blocks acceptance on failure; this witness did not run it or claim its later result. Package tests and finalization are not skipped.

## 1. Verified multi-audio originals

- Public targets: [Python](https://www.youtube.com/watch?v=rfscVS0vtbw), [neural networks](https://www.youtube.com/watch?v=aircAruvnKk).
- Fresh inline source evidence respectively: unique `en.4` / `English original` and `en-US.4` / `English (US) original` streaming audio metadata; exactly corresponding caption-audio associations present, alongside multiple other audio IDs. Original evidence is metadata plus association, not the numeric suffix alone.
- Returned provenance for both: language `en`, source `human`, status `verified-original`, evidence `original-audio-metadata`, source/original language `en`, no platform translation.
- Exact downloaded-cue/body matches: Python 2,935 substantive cues / 262,335 characters (request 3); neural networks 286 / 18,715 (request 4). HTTP 200, requested language matches provenance, no translation parameter.

## 2. Existing human and automatic captions

- [Human-caption target](https://www.youtube.com/watch?v=dQw4w9WgXcQ): `en`, `human`, `verified-original`, `original-audio-asr-association`; source association contains the manual and automatic English captions. Exact match 60 substantive cues / 2,142 characters, request 4, HTTP 200.
- [Automatic-caption target](https://www.youtube.com/watch?v=DHjqpvDnNGE): `en`, `automatic`, `verified-original`, `original-audio-asr-association`; single caption association, `asr` endpoint. Exact match 81 substantive cues / 3,149 characters, request 3, HTTP 200.
- Focused `A4 blank human track does not block usable automatic track` passed.

## 3. Honestly labeled fallback and normal summarization

- [Japanese talk](https://www.youtube.com/watch?v=gBumdOWWMhY) and [Chinese talk](https://www.youtube.com/watch?v=N0zhdMwD2Z8) have no sufficient original-language proof in captured player metadata. Both returned `en`, `human`, `original-unconfirmed`, `validated-platform-caption`, no platformTranslation flag. Neither is claimed to be original Japanese/Chinese or independently proven translation.
- Exact matches respectively: 463 cues / 19,121 characters (request 11), 545 / 19,021 (request 6), HTTP 200.
- `A3 labeled_unknown_and_translation; invalid_translation_endpoint`, `A3 refreshed original takes precedence over inline fallback without initial proof` and conflict/negated-original tests passed.
- `TestProcessItemVideoFallbackSummary` and `TestProcessItemVideoMisleadingProvenance` passed: the existing fixture checks normal configured summary plumbing and refusal of misleading provenance. These are fixture results, not claims of live translation availability or a real summary provider call.

## 4. Identity, substantive content, timing, completeness, retries and caches

- All captured initial/refreshed player IDs match the requested target; exact body/cue correspondence is detailed above.
- Focused tests passed: `A2 wrong target rejected`; `A1/A3 refreshed invalid identity, language and API failures stay unsuccessful`; `A4 malformed, truncated and timestamp-only cues fail`; `test_transcript_mismatchedmarkup_rejectsrepair`; `A4 captions work when Trusted Types rejects HTML parsing sinks`.
- `TestProcessItemWithFallback` passed its retries/settings table; `TestProcessBatchFailedPaginationPreservesFiles` passed failed counting, bounded initial/headed attempts and unchanged existing output; `TestProcessItemVideoResume` passed pending/verified/completed/dry-run.
- [No-caption target](https://www.youtube.com/watch?v=aqz-KE-bpKQ): captured players have no caption tracks, returned false with `Unusable original-language subtitles: no eligible tracks`; no content body returned.
- Complete package suite remains the final blocking check described above.

## 5. Actual source-backed verification

- Fresh public targets, actual platform metadata, source kind, delivered language and exact XML correspondence were independently checked in sections 1–4. Totals: six matches, seven target checks. No success was inferred solely from nonempty output.
- Conservative explicit-marker support is finite; unknown labels do not prove original language. Public platform assertions are not independent recognition of audio language.

## 6. UI flow: verified multi-audio original

- Fresh-built CLI: `/private/tmp/rlss-acceptance-cli url 'https://www.youtube.com/watch?v=aircAruvnKk' --config /private/tmp/rlss-acceptance-config.yaml --no-summary` → exit **0**, `Success: 1`, `Failed: 0`, `wrote content`. Output is isolated under `/private/tmp/rlss-acceptance-cli-output/`; embedded bundle was rebuilt first. Log `/private/tmp/rlss-acceptance-cli-success.log`.
- Summary explicitly disabled for this live run; fixture summary evidence is section 3.

## 7. UI flow: existing single-audio original

- Live human/automatic results in section 2 plus focused empty-human-to-automatic recovery test passed.

## 8. UI flow: unknown or conflicting original language

- Two real unconfirmed cases in section 3; conflicting metadata, negated markers, refreshed conflicting identities and summary handling covered by named passing focused tests. No real provider claim.

## 9. UI flow: only other-language subtitles usable

- Named original-first, unknown/translation, unusable-caption and refreshed-original tests passed, including explicit delivered/source language and translated-endpoint checks. Real fallback examples are section 3; known-translation evidence is fixture-only.

## 10. UI flow: no usable captions

- Fresh live no-caption extraction result is section 4. Fresh CLI on the same target exited **1**, `Success: 0`, `Failed: 1`, after one headless attempt and one headed retry, both reporting no eligible tracks. Only the earlier successful neural-network transcript exists in the isolated output directory; no no-caption target transcript was created. Log: `/private/tmp/rlss-acceptance-cli-failure.log`. Malformed/incomplete cue tests passed.

## 11. UI flow: wrong target

- Passing focused `A2 wrong target rejected` and `A1/A3 refreshed invalid identity, language and API failures stay unsuccessful` exercise wrong player/endpoint identities without substituting fake live claims.

## 12. UI flow: in progress

- Actual CLI log prints `processing item`, `wrote content`, counts and duration. Browser logged bounded request lists; original-first retry ordering and `TestProcessItemWithFallback` check bounded retry/settings behavior. No interactive screen was introduced.

## 13. UI flow: batch/watch failure and retry

- `TestProcessBatchFailedPaginationPreservesFiles` and `TestProcessItemWithFallback` passed counting, retry and preservation behavior. Fresh no-caption run is an actual failure.
- **Partly**: a watch cycle with mixed successful/failed videos, continuation and later fresh retry was not executed; the one-item batch fixture does not establish that whole flow. No failing behavior was observed, but this gap remains unverified.

## 14. UI flow: cached content or completed summary

- `TestProcessItemVideoResume/pending`, `/verified`, `/completed`, `/dry-run` all passed. Pending unverified cache retried; valid cache resumed; completed item skipped with no extraction. Existing files were fixture-owned temporary files, never the user's notes.

## 15. UI flow: empty selection / dry run

- Actual CLI `process --config /private/tmp/rlss-acceptance-config.yaml --no-summary` with both reading-list sources disabled → exit 0, `no items to process`.
- Actual CLI `url 'https://www.youtube.com/watch?v=rfscVS0vtbw' --config /private/tmp/rlss-acceptance-config.yaml --no-summary --dry-run` → exit 0, `dry-run: would process`, Total 1, Success 0, Skipped 1, Failed 0. No browser launched or transcript/summary saved for this request.
- Logs `/private/tmp/rlss-acceptance-empty.log`, `/private/tmp/rlss-acceptance-dry.log`. A first unquoted URL was rejected by zsh globbing before CLI execution; the quoted command above is the measured result.

## Data, decisions and skipped steps

- New isolated profiles and outputs only; no user notes, login state or production config modified. No account credentials, paid services, mirrors, proxies, audio uploads or real summary provider were used.
- No instructed skips in intent Constraints or plan Risks. No dismissed important-or-higher findings supplied by the station.
- The report discloses the metadata-confidence ceiling, honest English fallback behavior and fixture-only summary evidence. The initial mixed-watch gap was resolved by the re-run below.

## Re-run on 2026-11-01, at add8784 (functional code f5c46a4)

- Diff audit: `git diff f5c46a4..HEAD --name-only` lists only this acceptance report and evidence file. Product code and committed browser bundle are unchanged. Reused the independently clean-built CLI; no new product code or probes and no full suite run.
- Report rows 1–12 and 14–15: **carried over** — the only intervening changes were the two acceptance documents; none can affect their measured product behavior. Existing no-user-data result remains true.
- Report row 13: **re-tested in full** — real mixed-result watch cycles exercised failure counting, continuation, subsequent list refetch and fresh retry of the failed video.
- New synthetic Safari source list `/private/tmp/rlss-acceptance-watch-Bookmarks.plist` used the existing Safari test's plist shape and custom `safari.plist_path`. It contained real public [no-caption aqz](https://www.youtube.com/watch?v=aqz-KE-bpKQ) first by DateAdded, then [automatic-caption DHj](https://www.youtube.com/watch?v=DHjqpvDnNGE). Only the reading-list source was synthetic; all player and subtitle responses came from actual YouTube public requests.
- Config `/private/tmp/rlss-acceptance-watch.yaml`: Safari points only at that temporary plist; Chrome reading-list disabled; new extraction profile `/private/tmp/rlss-acceptance-watch-profile`; new output `/private/tmp/rlss-acceptance-watch-output`; watch enabled, interval 1 minute; delay min/max 0; skip_existing false; no-summary passed. No real Safari/Chrome reading list, notes, login state or production config was modified.
- Command: `/private/tmp/rlss-acceptance-cli process --config /private/tmp/rlss-acceptance-watch.yaml --no-summary`; raw local log `/private/tmp/rlss-acceptance-watch.log`.
- Cycle 1, local UTC+08 timestamps: 09:55:44 fetched 2 and processed aqz first; 09:55:55 `item failed` after one headless and one headed attempt; then `processing item index=2` DHj; 09:56:00 `wrote content`; `watch: cycle complete` reported Total 2, Success 1, Skipped 0, Partial 0, Failed 1; then sleep 1 minute.
- During that sleep, changed **only the temporary source list's failed-item title** to `Acceptance no captions cycle two refreshed`. No video URL or product response was changed.
- Cycle 2: 09:57:00 refetched 2 items, and first processing log contains the changed title, proving source refresh; aqz again made fresh headless/headed attempts and failed at 09:57:09; DHj then resumed the previously validated content cache successfully. Cycle report again Total 2, Success 1, Skipped 0, Partial 0, Failed 1. This retained-cache path is expected; the failed item was not turned into a cached success.
- Exactly one transcript file exists in the isolated output: DHj. The failed aqz produced none across either cycle. No actual summary provider was used.
- Identified the exact acceptance child PID 80243, sent `SIGTERM` only to it after the second `watch: sleeping` log. At 09:57:25 it logged `watch: signal during sleep, exiting`, and the command exited **0**. No production process was signaled.
- Finding closure: the initial important mixed-watch verification gap is **closed** by both observed cycles. No new failed behavior or remaining acceptance finding.
