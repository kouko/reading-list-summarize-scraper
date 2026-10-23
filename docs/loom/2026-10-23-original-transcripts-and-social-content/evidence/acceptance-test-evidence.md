# Original transcripts and social content — acceptance test evidence

Tried on 2026-10-23 in a clean worktree at a031287017215340a880153f7f92f2299bedbec3 (base 412deb870e5975a16e238075d48c26242700c8d8).

## Setup and isolation

- Created with `git worktree add /private/tmp/rlss-original-acceptance a031287017215340a880153f7f92f2299bedbec3`. It started clean.
- Followed README source-build instructions using its quick Go build, output redirected to `/private/tmp/rlss-original-acceptance-cli`; also ran `npm ci --ignore-scripts` and `npm run build:defuddle`. All exited 0. npm installed 26 locked packages and reported 0 vulnerabilities. Browser bundle rebuilt to 783.5kb; git status remained clean, proving no tracked bundle difference.
- No root binary, existing browser profile, reading list, vault, running batch, or summarization provider was used. CLI extraction ran with `--no-summary`, temporary browser data, temporary output, and disabled copy settings.
- Test Chrome used a fresh temporary profile. A Go chromedp harness navigated actual first-party URL shapes, fulfilled navigation/caption requests through DevTools Fetch interception, injected the rebuilt `embed/defuddle.min.js`, and awaited the actual public `window.extractArticle()` entry. It never substituted an extraction function or fabricated Bilibili original-language evidence.
- Controlled fixtures are synthetic, not observed live platform responses. They prove shipped-browser processing behavior only. No live successful video/social extraction was witnessed.
- Exact browser run: `go run /private/tmp/rlss-original-browser.go > /private/tmp/rlss-original-browser-results.log`, from clean worktree. Harness used `chromedp.Navigate(tc.url)`, `chromedp.Evaluate(string(bundle), nil)`, and an awaited `window.extractArticle().then(x=>JSON.stringify({ok:true,text:x})).catch(e=>JSON.stringify({ok:false,error:e.message}))`. Browser transport was successful for all 12 cases.
- Fixtures: target ID abc; English ASR caption track with exactly one caption-audio association; optional English human track; actual XML cues and empty/malformed variants. Social fixtures used exact target permalink/canonical/JSON-LD identity and separate reply content. Captions were served only at the platform timedtext endpoint, navigation only at intended first-party address. No real caption signed URL was saved.
- Per-criterion commands ran in Acceptance order: `node --test --test-name-pattern='A1' extraction.test.mjs`, then A2, A3, A4. Each ran 2 tests, 2 pass, 0 fail. These are criterion tests, not a complete package-suite invocation.
- Pipeline criterion command: `GOCACHE=/private/tmp/rlss-go-build go test ./internal/pipeline -run 'TestProcessItemWithFallback|TestProcessItemVideoResume|TestExtractionRejectsObservedWalls' -v`. All selected tests/subtests passed.
- The complete package suite was deliberately left to finalize-review, which is scheduled to run and block acceptance if it fails. No instruction skipped any step. No dismissed important finding was handed to this tester.
- Dispatch: inherited overrides null; unobservable-main-profile. No subagents/outside reviewers dispatched.

## 1. 影片字幕使用與影片原始語言一致的內容，不以翻譯字幕替代。

- How tried: A1 criterion tests; actual built bundle with verified ASR audio association and valid cues; real CLI `url 'https://www.youtube.com/watch?v=dQw4w9WgXcQ' --config /private/tmp/rlss-original-acceptance.yaml --no-summary`.
- Result: controlled actual browser returned original English text with target-bound provenance and source. Criterion tests rejected unverified defaults, ambiguous audio, and translated ASR association.
- Live: both actual headless and headed attempts failed with unusable subtitle XML missing a root element; CLI counted Success 0 / Failed 1 and printed `Error: 1 item(s) failed`. Its initial shell wrapper attempted assigning zsh readonly variable status after the command, so that wrapper did not preserve the CLI numeric exit; this does not affect the captured failure report. No output files were created.
- Prior first-party diagnostic `/private/tmp/rlss-original-transcripts-live-findings.md` independently records public metadata eligibility and HTTP 200 with zero bytes for unchanged, srv3 and json3 caption requests. That observation is supplied diagnostic evidence, not a second test performed here. No reliable transcript DOM fallback was established.
- Bilibili: actual browser entry always rejects at `extraction.ts:199`. README documents this limitation, but spec design decision and Video success flow still name Bilibili as initially supported. This is an explicit conservative implementation limitation and a scope contradiction, not a claimed temporary successful path.
- Verdict: partly.

## 2. Threads、Medium 與 X／Twitter 能取得對應目標文章的正文。

- How tried: A2 criterion tests; real Chrome bundle for all three synthetic target pages; real public X `url https://x.com/jack/status/20 --config /private/tmp/rlss-original-acceptance.yaml --no-summary`.
- Result: actual bundle returned exactly `Original target body.` for each platform, excluding reply text. Tests additionally cover X long article and Threads DOM/JSON-LD shapes.
- Live X: initial page load failed; headed retry reported `Inaccessible social target: main body unavailable`; exit 1, Success 0 / Failed 1. No output saved. This is legitimate access failure, not proof that a genuinely accessible post is broken.
- No successful live Threads or Medium target was established, and no authenticated profile was borrowed. Their current real-site success remains unverified.
- Verdict: partly.

## 3. 未取得目標正文或原語字幕時，不將其他頁面或不符合要求的內容回報為成功。

- How tried: A3 tests, actual browser unknown language / wrong X / Medium paywall / Vimeo / Bilibili fixtures; actual CLI live YouTube and X, mixed RSS batch and watch cycle.
- Result: all selected rejection cases passed. No video description or wall was newly saved or summarized. Mixed actual batch: Failed 1 unsupported video, Success 1 synthetic ordinary article; other items continued. Batch exit 0 is existing batch behavior, not false per-item success.
- Verdict: works, with the full automated suite still due before acceptance.

## 4. 人工及自動字幕均經實際內容檢查，空白或沒有有效字幕文字的內容不回報為成功。

- How tried: A4 tests, actual browser empty-human/usable-ASR, both-usable, malformed XML. Caption text was actual cue XML `<transcript><text start="0" dur="2">Actual original spoken words.</text></transcript>`; empty human fixture returned `<transcript></transcript>`.
- Result: blank human track led to automatic source; both usable led to labeled human source without an accuracy claim; mismatched nested XML rejected. Tests rejected negative timing, incomplete XML, timestamp-only text, external DTD and unknown XML entity.
- Verdict: works for content validation, not a claim of live transcript availability or audio/semantic accuracy. Full automated suite remains a blocking pre-acceptance check.

## 5. Video success

- Actual bundle returned provenance, language, source and cue text. Pipeline criterion `TestProcessItemWithFallback/video_verified` saved verified text into temporary content file.
- Real CLI YouTube did not succeed, Bilibili has no successful path, summaries deliberately disabled, and no full single-URL/batch/watch success with real transcript was observed.
- Verdict: partly.

## 6. Empty human subtitles

- Actual Chrome `test_bundle_empty_human_automatic_used` returned automatic source with substantive original text; A4 test passed.
- Full CLI save with such a real video not witnessed.
- Verdict: partly.

## 7. Multiple usable tracks

- Actual Chrome `test_bundle_multiple_usable_validated` returned human source and valid text, both supplied tracks having substantive cues.
- No accuracy comparison was printed. Full CLI save/live case not witnessed.
- Verdict: partly.

## 8. Original language unknown

- Actual browser missing-ASR fixture rejected. Real CLI public `rfscVS0vtbw` rejected ambiguous/dubbed audio in both browser modes, exit 1, Failed 1; no output files.
- Command: `/private/tmp/rlss-original-acceptance-cli url 'https://www.youtube.com/watch?v=rfscVS0vtbw' --config /private/tmp/rlss-original-acceptance.yaml --no-summary`.
- Verdict: works.

## 9. No usable original track

- Actual CLI `dQw4w9WgXcQ` headless/headed failed on empty response; actual browser malformed fixture failed; A1/A4 tests cover translation-only/unusable variants.
- No new transcript, description or summary saved.
- Verdict: works.

## 10. Other video platform

- Actual bundle Vimeo rejection and existing CLI batch/watch Vimeo rejection witnessed; no saved video content.
- Did not separately run the single-URL Vimeo exit-status path.
- Verdict: partly.

## 11. Social success

- Actual browser bundle returned target body for Threads, Medium and X; criterion tests prove replies excluded.
- No successful live platform target or complete CLI success/save/exit witnessed.
- Verdict: partly.

## 12. Login, challenge or subscription

- Actual browser Medium preview rejected. Live X unavailable-body failure went through existing headed retry. Pipeline criterion tests covered challenge setting, login/paywall and retry failures, with no failed-content writes.
- Did not automate login, borrow authenticated profiles, or restore private/member access.
- Verdict: partly.

## 13. Wrong or empty social content

- Actual browser wrong-target X rejected, A2 tests rejected wrong X and quoted Threads target; live X no-body case exit 1 without output.
- Full single-URL save/exit behavior for every missing/wrong/preview variant on all three platforms was not run.
- Verdict: partly.

## 14. In progress

- Existing CLI processing, retry and final failed-item logs observed. Empty-human case demonstrated next-candidate fallback.
- Configured isolated CLI timeouts were 15s (single URL) / 10s (batch) and 1s post-load wait. No intentionally stalled endpoint tested actual timeout enforcement.
- Verdict: partly.

## 15. Batch and watch failure

- Started a temporary loopback server at `127.0.0.1:18643` serving synthetic RSS with `https://vimeo.com/123` followed by a synthetic ordinary article. Existing `process --config /private/tmp/rlss-original-batch.yaml --no-summary` returned exit 0, Total 2 / Success 1 / Failed 1; it saved only the ordinary article.
- Existing `process --config /private/tmp/rlss-original-batch.yaml --no-summary --watch` completed one cycle: same Failed 1 / Success 1, ordinary content cache resumed, then `watch: sleeping minutes=10`. Sent SIGINT during sleep; exit 0. Did not wait for a second cycle or prove later recovery.
- Test server stopped afterward; no existing batch touched.
- Verdict: partly.

## 16. Cached incomplete video

- Selected `TestProcessItemVideoResume/pending` and `/verified`: unverified saved text triggers fresh extraction and surfaces failure; verified current-target cache resumes without browser extraction.
- Selected pipeline tests also reject description-only, metadata-only and wrong-target video payloads without writes.
- These are pipeline boundary tests with an extractor stub, not a witnessed full CLI cache/browser replay or summary-provider call.
- Verdict: works for the criterion tests; full suite remains blocking before acceptance.

## 17. Already completed

- Selected `TestProcessItemVideoResume/completed` returned existing skip with no extraction call; summary file seeded only in test temporary directory.
- No user's old note was used or revalidated.
- Verdict: works for the criterion test; full suite remains blocking before acceptance.

## 18. No selected items / dry run

- Existing CLI process with both reading-list sources disabled returned `no items to process`, exit 0.
- Existing CLI process of two-item synthetic feed with `--dry-run --no-summary` returned Total 2 / Skipped 2 / Failed 0 in 0s; logs only `dry-run: would process`.
- No extraction requests appeared during dry-run; loopback server saw feed retrieval only. Pipeline `TestProcessItemVideoResume/dry-run` passed with zero extractor calls.
- Verdict: works.

## Anchored unresolved findings

1. important — `extraction.ts:199`; spec Design decision and UI flows/Video success: **Bilibili has no success path (scope contradiction)**. Every matching URL rejects before fetching subtitle/audio evidence. README honestly documents this; it does not fulfill the retained spec's named-platform success promise. Close by grounding and implementing original-language evidence plus independently witnessed success, or explicitly renegotiating the retained scope with the user. Do not invent provenance.
2. important — `extraction.ts:94-110`; this evidence sections 1/5/9: **Real video success remains unwitnessed (availability limitation)**. Public eligible YouTube failed with empty native caption responses through the real CLI; fixtures passed but do not demonstrate live availability. No established reliable DOM transcript alternative exists. Close with a genuine eligible first-party video returning usable original captions through the shipped CLI, or an explicitly accepted narrower capability. Do not label the HTTP 200 empty response as successful content.
3. important — spec UI flows/Social success; this evidence sections 2/11: **Live social success remains unverified (validation gap)**. Controlled bodies passed across all three adapters; public X failed access/body extraction, and no successful live Threads/Medium target was established. These are not proofs of extraction defects, but the requested real target outcome is not yet witnessed. Close with independently captured legitimately accessible targets through existing CLI and saved body checks; otherwise retain the partly verdict.

## Captured output

### Actual built browser entry, controlled first-party boundary fixtures

```text
test_bundle_original_automatic_saved expected=true pass=true transport=<nil> result={"ok":true,"text":"<!-- rlss-video-v1 {\"platform\":\"youtube\",\"target\":\"abc\",\"language\":\"en\",\"source\":\"automatic\",\"evidence\":\"original-audio-asr-association\"} -->\n\nSubtitle language: en\nSubtitle source: automatic\n\nActual original spoken words."}
test_bundle_empty_human_automatic_used expected=true pass=true transport=<nil> result={"ok":true,"text":"<!-- rlss-video-v1 {\"platform\":\"youtube\",\"target\":\"abc\",\"language\":\"en\",\"source\":\"automatic\",\"evidence\":\"original-audio-asr-association\"} -->\n\nSubtitle language: en\nSubtitle source: automatic\n\nActual original spoken words."}
test_bundle_multiple_usable_validated expected=true pass=true transport=<nil> result={"ok":true,"text":"<!-- rlss-video-v1 {\"platform\":\"youtube\",\"target\":\"abc\",\"language\":\"en\",\"source\":\"human\",\"evidence\":\"original-audio-asr-association\"} -->\n\nSubtitle language: en\nSubtitle source: human\n\nActual original spoken words."}
test_bundle_unknown_original_rejected expected=false pass=true transport=<nil> result={"ok":false,"error":"Unverified original language: missing or conflicting original-audio automatic captions"}
test_bundle_malformed_cues_rejected expected=false pass=true transport=<nil> result={"ok":false,"error":"Unusable original-language subtitles: Unusable subtitles: malformed XML (1:44: unexpected close tag.)"}
test_bundle_other_video_rejected expected=false pass=true transport=<nil> result={"ok":false,"error":"Unsupported video transcript extraction"}
test_bundle_bilibili_unverified_rejected expected=false pass=true transport=<nil> result={"ok":false,"error":"Unverified original language: Bilibili exposes subtitle language but no verified original-audio association"}
test_bundle_threads_target_selected expected=true pass=true transport=<nil> result={"ok":true,"text":"Original target body."}
test_bundle_medium_target_selected expected=true pass=true transport=<nil> result={"ok":true,"text":"Original target body."}
test_bundle_x_target_selected expected=true pass=true transport=<nil> result={"ok":true,"text":"Original target body."}
test_bundle_medium_preview_rejected expected=false pass=true transport=<nil> result={"ok":false,"error":"Inaccessible social target: subscription or preview"}
test_bundle_wrong_x_rejected expected=false pass=true transport=<nil> result={"ok":false,"error":"Inaccessible social target: main body unavailable"}
```

### Real CLI live-video failure (trimmed)

```text
time=2026-10-23T14:19:08.239+08:00 level=INFO msg="processing item" index=1 total=1 url="https://www.youtube.com/watch?v=dQw4w9WgXcQ" title=""
time=2026-10-23T14:19:11.538+08:00 level=WARN msg="headless extraction failed, retrying with headed mode" url="https://www.youtube.com/watch?v=dQw4w9WgXcQ" err="exception \"Uncaught (in promise) Error: Unusable original-language subtitles: Unusable subtitles: malformed XML (1:0: document must contain a root element.); Unusable subtitles: malformed XML (1:0: document must contain a root element.)\" (344:4148): Error: Unusable original-language subtitles: Unusable subtitles: malformed XML (1:0: document must contain a root element.); Unusable subtitles: malformed XML (1:0: document must contain a root element.)\n    at ft (<anonymous>:344:4148)\n    at Wi (<anonymous>:351:63)"
time=2026-10-23T14:19:14.780+08:00 level=ERROR msg="item failed" url="https://www.youtube.com/watch?v=dQw4w9WgXcQ" err="extract https://www.youtube.com/watch?v=dQw4w9WgXcQ: headless attempt: exception \"Uncaught (in promise) Error: Unusable original-language subtitles: Unusable subtitles: malformed XML (1:0: document must contain a root element.); Unusable subtitles: malformed XML (1:0: document must contain a root element.)\" (344:4148): Error: Unusable original-language subtitles: Unusable subtitles: malformed XML (1:0: document must contain a root element.); Unusable subtitles: malformed XML (1:0: document must contain a root element.)\n    at ft (<anonymous>:344:4148)\n    at Wi (<anonymous>:351:63); headed retry: exception \"Uncaught (in promise) Error: Unusable original-language subtitles: Unusable subtitles: malformed XML (1:0: document must contain a root element.); Unusable subtitles: malformed XML (1:0: document must contain a root element.)\" (344:4148): Error: Unusable original-language subtitles: Unusable subtitles: malformed XML (1:0: document must contain a root element.); Unusable subtitles: malformed XML (1:0: document must contain a root element.)\n    at ft (<anonymous>:344:4148)\n    at Wi (<anonymous>:351:63)"
  Total:   1
  Success: 0
  Failed:  1
Error: 1 item(s) failed
```

### Real CLI unknown-original failure (trimmed)

```text
time=2026-10-23T14:20:26.960+08:00 level=INFO msg="processing item" index=1 total=1 url="https://www.youtube.com/watch?v=rfscVS0vtbw" title=""
time=2026-10-23T14:20:29.780+08:00 level=WARN msg="headless extraction failed, retrying with headed mode" url="https://www.youtube.com/watch?v=rfscVS0vtbw" err="exception \"Uncaught (in promise) Error: Unverified original language: ambiguous or dubbed audio association\" (344:4148): Error: Unverified original language: ambiguous or dubbed audio association\n    at ft (<anonymous>:344:4148)\n    at Wi (<anonymous>:346:689)\n    at Uo (<anonymous>:353:1391)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8"
time=2026-10-23T14:20:32.981+08:00 level=ERROR msg="item failed" url="https://www.youtube.com/watch?v=rfscVS0vtbw" err="extract https://www.youtube.com/watch?v=rfscVS0vtbw: headless attempt: exception \"Uncaught (in promise) Error: Unverified original language: ambiguous or dubbed audio association\" (344:4148): Error: Unverified original language: ambiguous or dubbed audio association\n    at ft (<anonymous>:344:4148)\n    at Wi (<anonymous>:346:689)\n    at Uo (<anonymous>:353:1391)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8; headed retry: exception \"Uncaught (in promise) Error: Unverified original language: ambiguous or dubbed audio association\" (344:4148): Error: Unverified original language: ambiguous or dubbed audio association\n    at ft (<anonymous>:344:4148)\n    at Wi (<anonymous>:346:689)\n    at Uo (<anonymous>:353:1391)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8"
  Total:   1
  Success: 0
  Failed:  1
```

### Real CLI public X failure (trimmed)

```text
time=2026-10-23T14:19:34.393+08:00 level=INFO msg="processing item" index=1 total=1 url=https://x.com/jack/status/20 title=""
time=2026-10-23T14:19:35.681+08:00 level=WARN msg="headless extraction failed, retrying with headed mode" url=https://x.com/jack/status/20 err="page load error net::ERR_HTTP_RESPONSE_CODE_FAILURE"
time=2026-10-23T14:19:38.492+08:00 level=ERROR msg="item failed" url=https://x.com/jack/status/20 err="extract https://x.com/jack/status/20: headless attempt: page load error net::ERR_HTTP_RESPONSE_CODE_FAILURE; headed retry: exception \"Uncaught (in promise) Error: Inaccessible social target: main body unavailable\" (344:4148): Error: Inaccessible social target: main body unavailable\n    at ft (<anonymous>:344:4148)\n    at Gi (<anonymous>:353:1123)\n    at Uo (<anonymous>:353:1942)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8"
  Total:   1
  Success: 0
  Failed:  1
Error: 1 item(s) failed
```

### Actual mixed CLI batch (trimmed)

```text
time=2026-10-23T14:20:24.315+08:00 level=INFO msg="fetched items" source=rss count=2
time=2026-10-23T14:20:24.315+08:00 level=INFO msg="processing item" index=1 total=2 url=https://vimeo.com/123 title="Unavailable video"
time=2026-10-23T14:20:30.055+08:00 level=INFO msg="processing item" index=2 total=2 url=http://127.0.0.1:18643/article title="Ordinary test article"
time=2026-10-23T14:20:31.601+08:00 level=INFO msg="wrote content" path=/private/tmp/rlss-original-batch-output/127_0_0_1/2026-10-23__3f11badb__content.md
time=2026-10-23T14:20:31.601+08:00 level=INFO msg="no summarizer configured, skipping summary" sha8=3f11badb
Pipeline completed in 7s
  Total:   2
  Success: 1
  Skipped: 0
  Partial: 0
  Failed:  1
```

### Actual CLI watch (trimmed)

```text
time=2026-10-23T14:20:52.461+08:00 level=INFO msg="fetched items" source=rss count=2
time=2026-10-23T14:20:52.461+08:00 level=INFO msg="processing item" index=1 total=2 url=https://vimeo.com/123 title="Unavailable video"
time=2026-10-23T14:20:56.787+08:00 level=INFO msg="processing item" index=2 total=2 url=http://127.0.0.1:18643/article title="Ordinary test article"
time=2026-10-23T14:20:56.787+08:00 level=INFO msg="resuming: content exists, skipping extraction" sha8=3f11badb
time=2026-10-23T14:20:56.787+08:00 level=INFO msg="watch: cycle complete" report="Pipeline completed in 4s\n  Total:   2\n  Success: 1\n  Skipped: 0\n  Partial: 0\n  Failed:  1\n\nErrors:\n  - Unavailable video: extract https://vimeo.com/123: headless attempt: exception \"Uncaught (in promise) Error: Unsupported video transcript extraction\" (344:4148): Error: Unsupported video transcript extraction\n    at ft (<anonymous>:344:4148)\n    at Uo (<anonymous>:353:1783)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8; headed retry: exception \"Uncaught (in promise) Error: Unsupported video transcript extraction\" (344:4148): Error: Unsupported video transcript extraction\n    at ft (<anonymous>:344:4148)\n    at Uo (<anonymous>:353:1783)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8\n"
time=2026-10-23T14:20:56.787+08:00 level=INFO msg="watch: sleeping" minutes=10
```

### Actual CLI empty batch and dry-run

```text
time=2026-10-23T14:19:19.403+08:00 level=INFO msg="no items to process"
time=2026-10-23T14:20:50.615+08:00 level=INFO msg="fetched items" source=rss count=2
time=2026-10-23T14:20:50.615+08:00 level=INFO msg="processing item" index=1 total=2 url=https://vimeo.com/123 title="Unavailable video"
time=2026-10-23T14:20:50.615+08:00 level=INFO msg="dry-run: would process" url=https://vimeo.com/123 sha8=7eb4e7d7
time=2026-10-23T14:20:50.615+08:00 level=INFO msg="processing item" index=2 total=2 url=http://127.0.0.1:18643/article title="Ordinary test article"
time=2026-10-23T14:20:50.615+08:00 level=INFO msg="dry-run: would process" url=http://127.0.0.1:18643/article sha8=3f11badb
Pipeline completed in 0s
  Total:   2
  Success: 0
  Skipped: 2
  Partial: 0
  Failed:  0
```

### Criterion test results

```text
A1
# Subtest: A1 verified original audio association
ok 1 - A1 verified original audio association
# Subtest: A1 unverified default language rejected
ok 2 - A1 unverified default language rejected
# tests 2
# pass 2
# fail 0

A2
# Subtest: A2 target bodies on three platforms exclude replies
ok 1 - A2 target bodies on three platforms exclude replies
# Subtest: A2 wrong target rejected
ok 2 - A2 wrong target rejected
# tests 2
# pass 2
# fail 0

A3
# Subtest: A3 verified social content succeeds without network fallback
ok 1 - A3 verified social content succeeds without network fallback
# Subtest: A3 login, preview and unsupported video fail
ok 2 - A3 login, preview and unsupported video fail
# tests 2
# pass 2
# fail 0

A4
# Subtest: A4 blank human track does not block usable automatic track
ok 1 - A4 blank human track does not block usable automatic track
# Subtest: A4 malformed, truncated and timestamp-only cues fail
ok 2 - A4 malformed, truncated and timestamp-only cues fail
# tests 2
# pass 2
# fail 0
=== RUN   TestProcessItemWithFallback
=== RUN   TestProcessItemWithFallback/success
=== RUN   TestProcessItemWithFallback/video_verified
=== RUN   TestProcessItemWithFallback/video_description
=== RUN   TestProcessItemWithFallback/video_metadata_only
=== RUN   TestProcessItemWithFallback/video_wrong_target
=== RUN   TestProcessItemWithFallback/social_body
=== RUN   TestProcessItemWithFallback/specialized_challenge_disabled
=== RUN   TestProcessItemWithFallback/load_error
=== RUN   TestProcessItemWithFallback/empty
=== RUN   TestProcessItemWithFallback/challenge
=== RUN   TestProcessItemWithFallback/login_wall
=== RUN   TestProcessItemWithFallback/headed_failure
=== RUN   TestProcessItemWithFallback/headed_challenge
=== RUN   TestProcessItemWithFallback/retry_failure
=== RUN   TestProcessItemWithFallback/retry_empty
=== RUN   TestProcessItemWithFallback/retry_challenge
=== RUN   TestProcessItemWithFallback/retry_login_wall
=== RUN   TestProcessItemWithFallback/block_disabled_challenge
=== RUN   TestProcessItemWithFallback/block_disabled_navigation
=== RUN   TestProcessItemWithFallback/block_disabled_empty
--- PASS: TestProcessItemWithFallback (0.01s)
    --- PASS: TestProcessItemWithFallback/success (0.00s)
    --- PASS: TestProcessItemWithFallback/video_verified (0.00s)
    --- PASS: TestProcessItemWithFallback/video_description (0.00s)
    --- PASS: TestProcessItemWithFallback/video_metadata_only (0.00s)
    --- PASS: TestProcessItemWithFallback/video_wrong_target (0.00s)
    --- PASS: TestProcessItemWithFallback/social_body (0.00s)
    --- PASS: TestProcessItemWithFallback/specialized_challenge_disabled (0.00s)
    --- PASS: TestProcessItemWithFallback/load_error (0.00s)
    --- PASS: TestProcessItemWithFallback/empty (0.00s)
    --- PASS: TestProcessItemWithFallback/challenge (0.00s)
    --- PASS: TestProcessItemWithFallback/login_wall (0.00s)
    --- PASS: TestProcessItemWithFallback/headed_failure (0.00s)
    --- PASS: TestProcessItemWithFallback/headed_challenge (0.00s)
    --- PASS: TestProcessItemWithFallback/retry_failure (0.00s)
    --- PASS: TestProcessItemWithFallback/retry_empty (0.00s)
    --- PASS: TestProcessItemWithFallback/retry_challenge (0.00s)
    --- PASS: TestProcessItemWithFallback/retry_login_wall (0.00s)
    --- PASS: TestProcessItemWithFallback/block_disabled_challenge (0.00s)
    --- PASS: TestProcessItemWithFallback/block_disabled_navigation (0.00s)
    --- PASS: TestProcessItemWithFallback/block_disabled_empty (0.00s)
=== RUN   TestProcessItemVideoResume
=== RUN   TestProcessItemVideoResume/pending
=== RUN   TestProcessItemVideoResume/verified
=== RUN   TestProcessItemVideoResume/completed
=== RUN   TestProcessItemVideoResume/dry-run
--- PASS: TestProcessItemVideoResume (0.00s)
    --- PASS: TestProcessItemVideoResume/pending (0.00s)
    --- PASS: TestProcessItemVideoResume/verified (0.00s)
    --- PASS: TestProcessItemVideoResume/completed (0.00s)
    --- PASS: TestProcessItemVideoResume/dry-run (0.00s)
=== RUN   TestExtractionRejectsObservedWalls
--- PASS: TestExtractionRejectsObservedWalls (0.00s)
PASS
ok  	github.com/kouko/reading-list-summarize-scraper/internal/pipeline	(cached)
```

A3 ordinary-article synthetic DOM test also logged an upstream Invalid URL warning caused by undefined linkedom location, but its substantive-content assertion and test passed. The actual Chrome/CLI ordinary article fixture independently succeeded; this warning is not evidence of a user-visible failure.

## Data and publication checks

### Later diagnostic supplied by the primary reviewer

After these acceptance runs, the primary reviewer reported that the official native IOS/ANDROID player API could issue fresh same-target caption URLs with nonempty responses for the public Rick video (3641 / 11536 bytes), while this HEAD still uses the page caption URLs. This result was not independently reproduced by this tester and does not change the failed CLI observation at this HEAD. A further reported live-page Trusted Types restriction may reject the existing caption DOM parser assignments once nonempty data reaches them; the controlled pages did not apply that security policy. These are leads for the next bounded fix, not successful acceptance evidence. No native-client signed URL or private response was committed. The next fix must be retested against the affected rows.

All captured URLs are public sample targets or synthetic fixtures. No signed caption endpoint, cookie, auth token, account name, private article body or browser snapshot is retained. Temporary pipeline tests create their own files; pending-cache extraction failure does not migrate/delete old notes, while successful re-extraction can overwrite content by existing pipeline semantics. Completed summary skipping is unchanged. No added backup is claimed.

Git-memory classification for this report commit: routine review artifact, no new product decision; the diagnosed availability and scope limits are recorded here rather than presented as a new chosen implementation. Exact commit-message privacy scan must pass before commit. No ambiguous private-party identifying text appears, so no semantic judge dispatch is required.

