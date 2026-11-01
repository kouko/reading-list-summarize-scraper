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

## Re-run on 2026-11-01, at 3225870 — label/audio language consistency fix

- Read previous report/evidence and audited `git diff deb039cd7504c027223218759ad1e970e1e59a82..322587050a3133fb8e770adf9840c96830c7e244`: only `extraction.ts`, its regression test and regenerated embedded bundle changed. The change requires each recognized original label's language family to match its audio ID's language family, including repeated labels on the same ID. Conflicting claims remain eligible only as honestly labeled fallback. No downloader, Go consumer, cache, command, watch, empty-selection or dry-run control flow changed.
- NEW clean detached checkout `/private/tmp/rlss-youtube-acceptance-round2`, exact tested HEAD `322587050a3133fb8e770adf9840c96830c7e244`. README setup repeated: `npm ci`, `npm run build:defuddle`, `go build -o /private/tmp/rlss-acceptance-round2-cli ./cmd/rlss/` all succeeded; no tracked build drift. No implementation or probe edits made by this tester.
- `/private/tmp/rlss-acceptance-round2-study.ts` imports this NEW checkout's exact `extraction.ts` and own dependencies; freshly bundled to `/private/tmp/rlss-acceptance-round2-study.js`. Existing browser driver reused, no new harness. NEW logged-out profile `/private/tmp/rlss-acceptance-round2-profile`, NEW raw output `/private/tmp/rlss-acceptance-round2-live`.
- Fresh command: `RLSS_STUDY_METHODS=current /private/tmp/rlss-acceptance-study /private/tmp/rlss-acceptance-round2-study.js /private/tmp/rlss-acceptance-round2-live /private/tmp/rlss-acceptance-round2-profile headless rfscVS0vtbw aircAruvnKk dQw4w9WgXcQ DHjqpvDnNGE gBumdOWWMhY N0zhdMwD2Z8 aqz-KE-bpKQ`. All seven browserError values `<nil>`.
- Fresh independent XML audit `/private/tmp/verify-acceptance-round2-live.py`, pointed ONLY at the round-two output, checked all initial/refreshed player identities and exact downloaded cue/body equality. Raw audit `/private/tmp/rlss-acceptance-round2-cue-audit.jsonl`: 6 exact matches, 4 verified-original, 2 original-unconfirmed, 1 no-subtitles; 4,370 substantive cues / 324,483 characters. These are freshly observed values, not copied assertions from the previous run.

| Public target | Fresh source evidence / delivered result | Exact substantive cues / characters |
|---|---|---|
| rfscVS0vtbw | `en.4`, `English original`, exact caption association; en human verified-original metadata | 2935 / 262335 |
| aircAruvnKk | `en-US.4`, `English (US) original`, exact caption association; en human verified-original metadata | 286 / 18715 |
| dQw4w9WgXcQ | single-audio automatic association; en human verified-original | 60 / 2142 |
| DHjqpvDnNGE | single-audio automatic association; en automatic verified-original | 81 / 3149 |
| gBumdOWWMhY | insufficient original evidence; en human original-unconfirmed, validated-platform-caption | 463 / 19121 |
| N0zhdMwD2Z8 | insufficient original evidence; en human original-unconfirmed, validated-platform-caption | 545 / 19021 |
| aqz-KE-bpKQ | no eligible caption tracks, returned false, no body | none |

- Focused browser command repeated unchanged: `node --test --test-name-pattern='A1|A3 labeled|A3 refreshed|A4|wrong target|test_transcript|test_originalEvidence' extraction.test.mjs extraction-boundaries.test.mjs` → **17 pass / 0 fail**, log `/private/tmp/rlss-acceptance-round2-js.log`. In particular `A1 multi_audio_original and conflicting_original_metadata` now exercises English-label/ja-ID mismatch, Japanese-label/en-ID mismatch, contradictory repeated same-ID labels (all original-unconfirmed), plus consistent Japanese metadata (verified-original). Other selected tests cover manual/automatic recovery, translated/fallback labels, original-first refresh ordering, wrong identity, malformed/incomplete/timestamp-only cues, bounded native refresh and negated evidence.
- Focused Go command repeated: `go test ./internal/pipeline -run 'TestProcessItemVideo|TestProcessItemWithFallback|TestProcessBatchFailedPaginationPreservesFiles' -count=1 -v` → PASS, log `/private/tmp/rlss-acceptance-round2-go.log`. This covers ordinary fallback summary plumbing, misleading provenance refusal, retry settings and pending/verified/completed/dry-run cache cases; no real summary-provider call is claimed.
- Actual fresh-built CLI with NEW isolated config/profile/output: `/private/tmp/rlss-acceptance-round2-cli url '<public target URL>' --config /private/tmp/rlss-acceptance-round2-config.yaml --no-summary`. Neural-network target and Japanese fallback target each exited **0**, Success 1 / Failed 0 and wrote content. Saved visible labels are respectively `en / human / verified original` and `en / human / original language unconfirmed`. Logs `/private/tmp/rlss-acceptance-round2-cli-original.log` and `/private/tmp/rlss-acceptance-round2-cli-fallback.log`.
- Same CLI on aqz exited **1**, Success 0 / Failed 1, after one actual headless attempt and one headed retry, both no eligible tracks; log `/private/tmp/rlss-acceptance-round2-cli-failure.log`. Only the two prior successful target files exist in `/private/tmp/rlss-acceptance-round2-cli-output/`; no description was saved as a failed target's transcript. Existing success files retained.
- Rows **1–5 re-tested**: every Acceptance line exercised across its full named surface with the fresh seven-video metadata/source/cue audit and focused content/identity/summary/cache tests above. No audio-listening accuracy claim.
- Rows **6–12 re-tested**: actual successful-original, fallback and failing CLI flows; fresh human/automatic cases; conflicting-evidence, translation/original-first, empty-track, wrong-target, progress and bounded retry fixture checks all passed. No live translated-endpoint or real summary-provider claim is substituted for fixture evidence.
- Row **14 re-tested**: all four cache/resume states passed again, including unverified pending content rejection and existing completed-summary skipping.
- Row **13 carried over**: verified against the exact fix diff; original-language evidence classification cannot change the existing batch/watch loop, item continuation/counting, source refetch or retry-control mechanics. Prior two-cycle real-video watch evidence remains applicable. New fresh extraction checks separately establish the changed classification behavior.
- Row **15 carried over**: exact fix diff does not alter empty-selection or dry-run branches, which avoid extraction entirely. Earlier actual command results remain applicable.
- No user's existing notes, source lists, browser login state or production configuration read or modified. No credentials or real provider calls. No instructed skips, no dismissed findings, and no remaining acceptance findings. Full suite not run by this tester; `go test ./... && npm test` remains the finalizer-owned acceptance-blocking check.

## Re-run on 2026-11-01, at 8a8ccf2 — empty cue and Traditional Chinese fixes

- NEW clean detached checkout `/private/tmp/rlss-youtube-acceptance-round3` at `8a8ccf2b59ae76267c2b270b65f58d21272044cd`. Repeated README setup `npm ci`, `npm run build:defuddle`, `go build -o /private/tmp/rlss-acceptance-round3-cli ./cmd/rlss/`: all succeeded, no tracked build drift. Tester made no product or probe changes.
- Audited diff from `c5cbd26`: functional changes only normalize/skip empty cue text before timing checks and add exact affirmative `Chinese (Traditional) original` recognition subject to audio-language-family agreement. Nonempty cues retain timing checks. Existing conflicting/negated evidence safeguards remain. Other changed files are generated bundle, regression tests, plan task additions and prior attestation; none changes watch or empty-selection control flow. Read previous report/evidence and unchanged intent/spec; no instructed skips or dismissed important findings.
- Own helper `/private/tmp/rlss-acceptance-round3-study.ts` imports **this NEW clean checkout's** `extraction.ts` and own installed dependencies; bundled anew with its esbuild. No old study bundle or main-agent cohort result used as independent evidence.
- Command: `RLSS_STUDY_METHODS=current /private/tmp/rlss-acceptance-study /private/tmp/rlss-acceptance-round3-study.js /private/tmp/rlss-acceptance-round3-live /private/tmp/rlss-acceptance-round3-profile headless rfscVS0vtbw aircAruvnKk dQw4w9WgXcQ DHjqpvDnNGE gBumdOWWMhY N0zhdMwD2Z8 aqz-KE-bpKQ Wz4Qy4BU7ks gLcnqPTGT_w jOlPOxqFD5s o4D1J1YWjhg bjE6HYuEA1k cQLKVzbwN7I`. Fresh isolated logged-out profile; all 13 browserError values `<nil>`.
- Own XML reconstruction audit `/private/tmp/verify-acceptance-round3-live.py`, pointing only at these new raw files, passed exact requested-target/player checks and exact saved-body/downloaded-cue correspondence for all **12 successful transcripts**, **8,714 substantive cues / 434,236 characters**. Ten verified-original, two original-unconfirmed; one genuine no-subtitles result. Raw audit `/private/tmp/rlss-acceptance-round3-cue-audit.jsonl`; full captions and signed URLs remain local only.
- Fresh seven baseline outcomes: both collected multi-audio samples en human metadata-verified originals; dQw en human and DHj en automatic association-verified originals; gBumd/N0zh en human original-unconfirmed; aqz returned false, no tracks/body. Their fresh exact-match counts remain 2935/286/60/81/463/545 substantive cues respectively, totaling 4,370 cues / 324,483 characters. Public source links are in sections 1–4; these are rerun results, not merely carried earlier claims.

| Additional real public target | Fresh original evidence and delivered source | Exact substantive cues / characters |
|---|---|---|
| [Wz4Qy4BU7ks](https://www.youtube.com/watch?v=Wz4Qy4BU7ks) | single-audio ASR association; ja automatic verified-original | 366 / 13662 |
| [gLcnqPTGT_w](https://www.youtube.com/watch?v=gLcnqPTGT_w) | single-audio ASR association; ja automatic verified-original | 96 / 1971 |
| [jOlPOxqFD5s](https://www.youtube.com/watch?v=jOlPOxqFD5s) | single-audio ASR association; en automatic verified-original | 902 / 32890 |
| [o4D1J1YWjhg](https://www.youtube.com/watch?v=o4D1J1YWjhg) | single-audio ASR association; en automatic verified-original | 10 / 93 |
| [bjE6HYuEA1k](https://www.youtube.com/watch?v=bjE6HYuEA1k) | single-audio ASR association; en automatic verified-original | 855 / 32314 |
| [cQLKVzbwN7I](https://www.youtube.com/watch?v=cQLKVzbwN7I) | `zh-Hant.4` / `Chinese (Traditional) original`, exact caption association; zh-TW human metadata-verified original | 2115 / 28823 |

- Chinese source metadata has two audio IDs (`en-US.10`, `zh-Hant.4`), uniquely binds the affirmative original marker to `zh-Hant.4`, and associates caption index 0. Returned provenance originalLanguage `zh`, delivered/sourceLanguage `zh-TW`, source human, languageStatus verified-original. No platform translation flag. This is first-party metadata proof, not audio recognition.
- Focused browser command repeated: `node --test --test-name-pattern='A1|A3 labeled|A3 refreshed|A4|wrong target|test_transcript|test_originalEvidence' extraction.test.mjs extraction-boundaries.test.mjs` → **18 pass / 0 fail**; log `/private/tmp/rlss-acceptance-round3-js.log`. Includes new `A4 whitespace formatting cues do not require timing, but blank-only payloads fail`, retained malformed/truncated/nonempty timestamp/timing tests, Traditional Chinese positive/mismatched/negated/qualified markers within the multi-audio regression, and existing original-first/fallback/translated-endpoint/identity checks.
- Focused Go command repeated: `go test ./internal/pipeline -run 'TestProcessItemVideo|TestProcessItemWithFallback|TestProcessBatchFailedPaginationPreservesFiles' -count=1 -v` → PASS; log `/private/tmp/rlss-acceptance-round3-go.log`. Covers summary plumbing without a real provider, misleading provenance, retry settings and all pending/verified/completed/dry-run cache cases.
- Actual clean-built CLI with NEW config `/private/tmp/rlss-acceptance-round3-config.yaml`, profile `/private/tmp/rlss-acceptance-round3-cli-profile`, output `/private/tmp/rlss-acceptance-round3-cli-output`: `url '<public URL>' --config /private/tmp/rlss-acceptance-round3-config.yaml --no-summary` on Wz4Qy4BU7ks and cQLKVzbwN7I each exited **0**, wrote content, Success 1 / Failed 0. Saved visible labels respectively `ja / automatic / verified original` and `zh-TW / human / verified original`. Logs `/private/tmp/rlss-acceptance-round3-cli-pivot.log`, `/private/tmp/rlss-acceptance-round3-cli-chinese.log`.
- Same CLI on aqz exited **1**, Success 0 / Failed 1, after one headless and one headed no-eligible-tracks attempt; no target file created and both successful files retained. Log `/private/tmp/rlss-acceptance-round3-cli-failure.log`. No real summary provider was invoked.
- Report rows **1–5 re-tested in full**: fresh baseline plus six regression videos establish original/fallback identities, evidence, source labels and content correspondence; summary integration and all content/timing/completeness/cache protections covered by relevant passing fixtures. No word-for-word listening claim.
- Report rows **6–12 re-tested in full**: actual saved Japanese/Chinese originals and actual no-caption CLI failure, fresh human/automatic/unconfirmed cases and original-first/translation/conflicting-evidence/wrong-target/progress/bounded retry tests cover each named surface. Known translation and summary evidence remain explicitly fixture-only.
- Report row **14 re-tested**: pending/verified/completed/dry-run cache tests pass again.
- Report row **13 carried over** — reviewed exact functional delta: empty-cue parsing and recognized-label classification do not change batch/watch counting, continuation, refetch, scheduling or retry control. Prior two-cycle actual watch evidence remains valid; changed extraction behavior is freshly exercised above.
- Report row **15 carried over** — empty-selection/dry-run branches do not invoke the modified subtitle parser/selector and were not changed. Prior actual command evidence remains valid.
- Data result unchanged: only new isolated profiles/configs/output; no user's notes, real reading lists, login state or production configuration modified. No skipped steps, dismissed findings or remaining acceptance findings. Full package suite remains the finalizer-owned blocking check `go test ./... && npm test`; this tester did not run it.

## Re-run on 2026-11-01, at 2a14a42 — authorized extra shared-association repair round

- This is the **single additional round explicitly authorized by the user**, recorded in plan Risks; it does not reset the review episode or grant a general round-limit waiver. No step skipped or important finding dismissed.
- NEW clean detached checkout `/private/tmp/rlss-youtube-acceptance-extra` at `2a14a42928f05ff23f493ffc9b3d1d51289f59e4`. Repeated README setup `npm ci`, `npm run build:defuddle`, `go build -o /private/tmp/rlss-acceptance-extra-cli ./cmd/rlss/`: all succeeded, no tracked build drift. Read current intent/spec/plan and prior report/evidence; tester made no product or probe edits.
- Audited `0867fbb..2a14a4`: functional change confines the other-language-ASR veto to ASR-inferred evidence. Explicit consistent original metadata can coexist with other-language captions in a shared association. Audio-label language agreement, contradictory/negated explicit markers, endpoint identity/translation, cue body and timing checks remain intact. Other changes are generated bundle, regression tests, plan and attestation; no batch/watch or empty/dry-run control-flow changes.
- Own `/private/tmp/rlss-acceptance-extra-study.ts` imports NEW clean checkout's exact `extraction.ts` and own dependencies, bundled anew. The normal `current` extraction path uses real public responses with request logging only: **no preferred-language probe, skipped-language interceptor or fabricated caption response**. Main-agent cohort results were not used as independent evidence.
- Command: `RLSS_STUDY_METHODS=current /private/tmp/rlss-acceptance-study /private/tmp/rlss-acceptance-extra-study.js /private/tmp/rlss-acceptance-extra-live /private/tmp/rlss-acceptance-extra-profile headless rfscVS0vtbw aircAruvnKk dQw4w9WgXcQ DHjqpvDnNGE gBumdOWWMhY N0zhdMwD2Z8 aqz-KE-bpKQ 5r6_iMWRFNs Oi9DHk57KeQ Wz4Qy4BU7ks gLcnqPTGT_w jOlPOxqFD5s o4D1J1YWjhg bjE6HYuEA1k cQLKVzbwN7I`.
- All 15 actual browser runs returned browserError `<nil>`. Own fresh XML/body audit `/private/tmp/verify-acceptance-extra-live.py` points only to these new raw outputs and validates requested address, all player identities, returned provenance and exact downloaded-cue/body equality. **14 exact matching successes, 12 verified-original, 2 original-unconfirmed, 1 no-subtitles; 9,339 substantive cues / 442,320 characters**. Raw output/audit local only under `/private/tmp/rlss-acceptance-extra-live/` and `/private/tmp/rlss-acceptance-extra-cue-audit.jsonl`; no full captions, signed URLs or browser state committed.

| Newly repaired public target | Fresh platform evidence and NORMAL product result | Exact substantive cues / characters |
|---|---|---|
| [PanSci 5r6_iMWRFNs](https://www.youtube.com/watch?v=5r6_iMWRFNs) | `zh-Hant.4`, `Chinese (Traditional) original`; shared original association indices [0,1] contains manual zh-Hant plus en ASR; returned **zh-Hant human verified-original**, metadata evidence, no translation | 602 / 7213 |
| [OpenAI Oi9DHk57KeQ](https://www.youtube.com/watch?v=Oi9DHk57KeQ) | `en-US.4`, `English (US) original`; exact shared original association contains 21 ASR languages, including en and ar; returned **en automatic verified-original**, metadata evidence, no translation | 23 / 871 |

- PanSci original association's default caption index is 1 (English ASR), yet product correctly selects original-family manual Chinese. OpenAI association's platform indices start with English index 3 while the caption list begins with Arabic; actual saved output is English. Both successful cue matches came from HTTP 200 request 3 with source language matching provenance, no translated target. These observations verify original preference rather than merely nonempty fallback availability.
- All prior 13 public regression targets were independently rerun too: baseline multi-audio English originals, human/automatic English, two unconfirmed English fallbacks, five repaired automatic-caption examples and Traditional Chinese metadata-original all retain their previously verified source/status and exact cue correspondence. Aqz again returned false with no eligible tracks/body. Prior 12 successful bodies total 8,714 cues / 434,236 characters; new two add 625 / 8,084. No word-for-word audio-listening claim.
- Focused browser command repeated: `node --test --test-name-pattern='A1|A3 labeled|A3 refreshed|A4|wrong target|test_transcript|test_originalEvidence' extraction.test.mjs extraction-boundaries.test.mjs` → **18 pass / 0 fail**, log `/private/tmp/rlss-acceptance-extra-js.log`. Shared-association positive fixtures cover original inline selection, refreshed recovery and refusal to request other-language alternatives prematurely; invalid-host/translated candidates do not become original. Ambiguous multi-language ASR without explicit proof remains original-unconfirmed. Existing contradictory original labels, language mismatch, negation, malformed association, wrong identity, malformed/empty/timing/incomplete XML and fallback-translation tests remain passing.
- Focused Go command repeated: `go test ./internal/pipeline -run 'TestProcessItemVideo|TestProcessItemWithFallback|TestProcessBatchFailedPaginationPreservesFiles' -count=1 -v` → PASS, log `/private/tmp/rlss-acceptance-extra-go.log`. Normal configured summary plumbing, misleading provenance, cache pending/verified/completed/dry-run states and retry settings checked; no real summary-provider call claimed.
- Actual fresh CLI command `url '<public URL>' --config /private/tmp/rlss-acceptance-extra-config.yaml --no-summary` uses NEW isolated profile `/private/tmp/rlss-acceptance-extra-cli-profile` and output `/private/tmp/rlss-acceptance-extra-cli-output`. PanSci and OpenAI each exited **0**, Success 1 / Failed 0, wrote content; saved visible labels respectively `zh-Hant / human / verified original` and `en / automatic / verified original`. Logs `/private/tmp/rlss-acceptance-extra-cli-pansci.log`, `/private/tmp/rlss-acceptance-extra-cli-openai.log`.
- Same actual CLI on aqz exited **1**, Success 0 / Failed 1, with one headless and one headed no-eligible-tracks attempt; no failed target output created, both successful transcript files retained. Log `/private/tmp/rlss-acceptance-extra-cli-failure.log`.
- Report rows **1–5 re-tested in full** across the new 15-target source/evidence/cue audit and relevant summary/provenance/content/timing/completeness/cache fixtures. Rows **6–12 re-tested in full** via actual saved shared-association originals, genuine CLI failure, fresh human/automatic/unconfirmed cases and focused translation/conflicting-evidence/wrong-target/progress/bounded retry assertions. Row **14 re-tested** via all four cache/resume states. Known translation and summary-provider integration remain explicitly fixture evidence, not invented live-provider claims.
- Report row **13 carried over** — exact diff only changes original-evidence classification, not batch/watch counting, continuation, refetch, scheduling or retry controls; prior actual two-cycle watch evidence remains applicable. Row **15 carried over** — empty-selection/dry-run bypass the modified selector and are unchanged. Carried reasons were checked against the full fix diff.
- Data result unchanged: only isolated test profiles, config and output; no user's notes, real reading lists, login state or production config modified. No remaining acceptance findings. Full package suite not run by this tester; `go test ./... && npm test` remains the finalizer-owned blocking check, with package tests and finalization not skipped.
