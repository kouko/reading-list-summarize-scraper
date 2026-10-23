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


## Re-run on 2026-10-23, at ca5acba848ca87ec26cd2f500119796534ff6499

Fix range checked: `ad2bb6b..ca5acba`. Only README, plan, video extraction source/tests and generated browser bundle changed. The fix adds bounded official native-player refresh with revalidated identity/language and replaces caption HTML parsing with strict SAX cue parsing. Social extraction, Go pipeline, command settings and cache/skip machinery are unchanged. This rerun independently retested all 18 rows; no verdict was carried over without a rerun. The first-run observations above remain historical, not the current live result.

### Setup, isolation and exact commands

- Created new clean worktree `git worktree add /private/tmp/rlss-original-acceptance-w3 ca5acba848ca87ec26cd2f500119796534ff6499`; followed README quick source build into `/private/tmp/rlss-original-w3-cli`. Independently ran `npm ci --ignore-scripts` and `npm run build:defuddle`; all passed, npm reported 26 packages / 0 vulnerabilities and 784.5kb browser bundle. Worktree stayed clean after rebuild.
- Reused only tester-authored temporary harness/fixture code, never another implementer's active tree, profile, output or live result. Fresh browser profile `/private/tmp/rlss-original-w3-profile`, fresh direct output `/private/tmp/rlss-original-w3-output`; separate fresh batch profile/output, plus dedicated synthetic cache output. No user data or paid provider was touched.
- CLI config: source lists disabled for direct runs; headless true; wait_after_load 2s; headless/headed timeout 40s; existing headed retry enabled; copy disabled. Every extraction command included `--no-summary`.
- Per-criterion commands in order: `node --test --test-name-pattern='A1' extraction.test.mjs`, then A2, A3, A4. Results 4/4, 2/2, 4/4, 3/3 respectively; some native refresh tests intentionally match multiple affected criteria. No complete package-suite command was run.
- Pipeline criteria: `GOCACHE=/private/tmp/rlss-go-build go test ./internal/pipeline -run 'TestProcessItemWithFallback|TestProcessItemVideoResume|TestExtractionRejectsObservedWalls' -v`; passed.
- Actual rebuilt browser command: `go run /private/tmp/rlss-original-browser.go > /private/tmp/rlss-original-w3-browser.log`; all 12 existing real-Chrome fixture cases passed. The harness injects the actual generated bundle and awaits `window.extractArticle()`; only network page/caption resources are mocked.
- Exact live CLI pattern: `/private/tmp/rlss-original-w3-cli url <public-url> --config /private/tmp/rlss-original-w3.yaml --no-summary`. Public URLs: YouTube `https://www.youtube.com/watch?v=dQw4w9WgXcQ` and `https://www.youtube.com/watch?v=rfscVS0vtbw`; Threads `https://www.threads.com/@zuck/post/DeZtp-TkQHh`; Medium `https://ev.medium.com/welcome-to-medium-9e53ca408c48`; X `https://x.com/jack/status/20`; Bilibili `https://www.bilibili.com/video/BV123`; unsupported `https://vimeo.com/123`.
- All independent live results below were reproduced by this tester through the CLI; supplied primary observations were leads only. No signed caption URLs, authentication data or full third-party bodies are retained.

### 1. Original-language subtitles — re-tested

- Real YouTube direct CLI succeeded in 4s, exit 0, Total 1 / Success 1 / Failed 0.
- Saved content `www_youtube_com/2026-10-23__0424974c__content.md` is 2562 characters including frontmatter, 63 nonblank lines after the provenance marker. Confirmed target `dQw4w9WgXcQ`, language en, source human and evidence original-audio-asr-association. Substantive cue text is present; no full lyrics are committed. No summary file was produced.
- Real RSS batch independently saved the same public video's verified transcript after a failed ambiguous-language item.
- Current built-browser verified-ASR and criterion tests passed, including native refresh identity/language conflicts and translations rejected.
- Bilibili still has no code success path (`extraction.ts:231`); actual bundle rejects with explicit unverified-language cause. Live malformed sample also exited 1, without claiming that it represents a real accessible Bilibili video.
- Current verdict: partly because the retained named-platform promise still includes Bilibili. First-run YouTube failure is resolved for the independently tested public video, not a universal availability guarantee.

### 2. Target social bodies — re-tested

- Actual Threads CLI exit 0; saved body 136 characters, one target post. Verified its expected introductory text and target URL; no reply list/navigation was included. Content checked locally, not copied into evidence.
- Actual Medium CLI exit 0 after headless rejection and headed retry; saved Markdown body 6825 characters, heading `Welcome to Medium` and expected article sections. Target URL matched; recommendation labels `More from`, `Recommended from Medium` and X login chrome were absent.
- Actual X CLI exit 1, Failed 1, no X content or summary output. The observed page/network/access failure does not establish a product defect or prove login is the cause.
- All three controlled-browser target cases and social criterion tests still passed.
- Current verdict: partly; Threads/Medium now have independent full CLI saved-body evidence, X success remains unverified.

### 3. No substitute success — re-tested

- Current actual CLI ambiguous-language YouTube and unavailable X remained failures with no newly saved substitute. Unsupported video direct CLI also exit 1.
- A3 native refreshed wrong identity, conflicting language/audio, translated URL, HTTP/empty/malformed JSON failures all passed rejection assertions. Existing wall/preview/description and wrong-target cases passed.
- Actual three-item batch and watch each reported Success 2 / Failed 1; only verified real-video content and synthetic ordinary article were saved.
- Current verdict: works. Full automated suite remains due before acceptance and blocks acceptance if it fails.

### 4. Real subtitle content checks — re-tested

- A4 tests passed: strict SAX output with Trusted Types rejecting HTML sinks; empty-human fallback to usable automatic track; IOS failure followed by ANDROID usable automatic track; empty, negative-timed, malformed, timestamp-only, DTD/entity and explicitly incomplete checks.
- Actual built-browser empty-human case returned automatic source and cue text; both-usable returned a validated human source; malformed nested markup rejected.
- Real target page CLI now successfully saves nonempty human source despite the live page's HTML assignment restrictions. This does not infer semantic transcription accuracy or complete spoken-word coverage.
- Current verdict: works; full suite remains a blocking later check.

### 5. Video success — re-tested

- Real single-URL exit 0 and actual batch save witnessed with verified English human subtitles and labels. Watch resumed that verified content.
- Bilibili remains unavailable by implementation; summary providers were not called, so complete summary-provider execution is not verified.
- Verdict: partly.

### 6. Empty human subtitles — re-tested

- Actual rebuilt browser case used automatic subtitles after an empty human track. Updated criterion also exercised the native refresh fallback to automatic after IOS failure.
- No real live video with exactly this supplied empty-human/usable-auto state and complete CLI save was established.
- Verdict: partly.

### 7. Multiple usable tracks — re-tested

- Actual rebuilt browser both-usable fixture returned substantive validated human text, label and target provenance, with no accuracy comparison.
- The real video's human selection proves that selected track usable; it does not establish that every other offered track was usable. Full live multi-track comparison remains unverified.
- Verdict: partly.

### 8. Original language unknown — re-tested

- Real `rfscVS0vtbw` single URL still failed in both modes with `Unverified original language: ambiguous or dubbed audio association`, exit 1; no new content.
- Actual browser missing-original-association fixture rejected; current native refresh tests rejected changed language/audio.
- Verdict: works.

### 9. No usable original track — re-tested

- Real old Rick-video empty-response failure no longer applies at this HEAD; it now succeeds.
- Current actual browser malformed-caption fixture and updated A1/A3/A4 tests reject empty/incomplete/translated/unusable responses, including exhaustion/failure of fresh-player requests. Pipeline rejection tests ensure description-only/wrong-target payloads are not saved.
- No real live target intentionally confirmed to have unusable original captions was substituted for these controlled cases; success of unrelated live video is not evidence for this failure case.
- Verdict: works for criterion tests and actual browser rejection; full suite remains blocking.

### 10. Other video platform — re-tested

- Real `vimeo.com/123` direct CLI exit 1 with unsupported-transcript error, no content saved; actual browser fixture also rejected.
- Verdict: works.

### 11. Social success — re-tested

- Full actual CLI success, exit 0, and saved target-body checks for Threads/Medium now witnessed.
- Current X public sample remains unsuccessful, with no existing logged-in profile borrowed; X positive CLI save remains not verified.
- Verdict: partly.

### 12. Login, challenge or subscription — re-tested

- Current built-browser subscription-preview fixture rejected. Pipeline selected tests again passed login/challenge/retry-setting variants without failed-content writes.
- Real Medium headless failed with missing/different canonical identity; headed retry produced expected complete target body and exit 0. Do not relabel this observed rejection as a proven Cloudflare/login cause.
- No manual login/member credential recovery was performed.
- Verdict: partly.

### 13. Wrong or empty social content — re-tested

- All current social criterion cases and actual browser wrong-target X / Medium preview failed as expected; actual live X failed without saving.
- Did not run every negative state as full CLI across all three platforms.
- Verdict: partly.

### 14. In progress — re-tested

- Current actual CLI emitted processing, fallback/retry and final item outcomes. Updated native-refresh tests passed bounded-client ordering and request signal/options checks; built browser exercised candidate fallback.
- 40s enclosing browser limits configured; no intentionally stalled request tested timeout enforcement itself.
- Verdict: partly.

### 15. Batch and watch failure — re-tested

- Temporary loopback RSS served three items ordered newest first: ambiguous `rfscVS0vtbw`, eligible `dQw4w9WgXcQ`, synthetic ordinary article. Existing `process --config /private/tmp/rlss-original-w3-batch.yaml --no-summary` exit 0, Total 3 / Success 2 / Failed 1. It continued past failure and saved exactly two content files.
- Existing `process ... --watch` completed one cycle with the same counts, resumed verified video/ordinary content and rechecked failed video, then slept 10 minutes. SIGINT during sleep returned exit 0.
- No second cycle or restored-access recovery was attempted. Loopback server was stopped afterward.
- Verdict: partly.

### 16. Cached incomplete video — re-tested

- Seeded a synthetic old description in dedicated temporary `www_youtube_com/2026-10-23__5ff986bd__content.md`.
- Existing batch CLI `process --config /private/tmp/rlss-original-w3-batch.yaml --output /private/tmp/rlss-original-w3-cache-output --no-summary --limit 1` logged `cached video requires fresh extraction`, retried actual ambiguous-language target, Failed 1. Old content bytes remained exactly the seeded description; no summary existed at this step.
- Selected pending/verified pipeline criterion tests passed again. No LLM call was made.
- Verdict: works.

### 17. Already completed — re-tested

- After step 16, seeded only a synthetic completed summary in the same isolated cache directory.
- Same existing batch CLI command now exit 0, Skipped 1 / Failed 0, no extraction log, in 0s; old description and synthetic summary retained.
- Selected completed-cache test passed again.
- Verdict: works.

### 18. No selected items / dry run — re-tested

- Existing CLI with direct isolated config sources disabled: `no items to process`, exit 0.
- Existing RSS process `--dry-run --no-summary`: Total 3 / Skipped 3 / Failed 0, no extraction. Loopback server only saw feed GET for the dry-run.
- Selected pipeline dry-run test passed with zero extractor calls.
- Verdict: works.

### Current anchored findings and resolution of earlier findings

- **important, unresolved — Bilibili scope contradiction**, `extraction.ts:231`, retained spec Design decision / Video success. Every Bilibili URL is rejected before evidence acquisition. This is literal absent success implementation, honestly documented in README, not a mere observed access restriction. Close with grounded original-language implementation and independent success, or explicitly revise the agreed supported scope with the user. No fabricated provenance is acceptable.
- **Earlier important real-YouTube-success gap: resolved for witnessed sample.** Exact existing CLI independently returned/saved original subtitles at current HEAD. This is sufficient to remove the previous blanket absence-of-live-success finding, while preserving normal network/availability limits.
- **Social verification gap narrowed; X remains not verified**, retained spec UI flows / Social success. Actual independent Threads/Medium CLI success now established. X current public target fails network/access/body retrieval; this remains an explicit verification limitation, not an invented adapter defect. Controlled X cases pass. Close positive verification using a legitimately accessible public or authorized isolated-profile target.
- No important finding was dismissed. Bilibili remains blocking for literal scope completion; this report does not declare the entire original promise ready.

### Re-run captured output

#### Actual rebuilt browser fixtures

```text
test_bundle_original_automatic_saved expected=true pass=true transport=<nil> result={"ok":true,"text":"<!-- rlss-video-v1 {\"platform\":\"youtube\",\"target\":\"abc\",\"language\":\"en\",\"source\":\"automatic\",\"evidence\":\"original-audio-asr-association\"} -->\n\nSubtitle language: en\nSubtitle source: automatic\n\nActual original spoken words."}
test_bundle_empty_human_automatic_used expected=true pass=true transport=<nil> result={"ok":true,"text":"<!-- rlss-video-v1 {\"platform\":\"youtube\",\"target\":\"abc\",\"language\":\"en\",\"source\":\"automatic\",\"evidence\":\"original-audio-asr-association\"} -->\n\nSubtitle language: en\nSubtitle source: automatic\n\nActual original spoken words."}
test_bundle_multiple_usable_validated expected=true pass=true transport=<nil> result={"ok":true,"text":"<!-- rlss-video-v1 {\"platform\":\"youtube\",\"target\":\"abc\",\"language\":\"en\",\"source\":\"human\",\"evidence\":\"original-audio-asr-association\"} -->\n\nSubtitle language: en\nSubtitle source: human\n\nActual original spoken words."}
test_bundle_unknown_original_rejected expected=false pass=true transport=<nil> result={"ok":false,"error":"Unverified original language: missing or conflicting original-audio automatic captions"}
test_bundle_malformed_cues_rejected expected=false pass=true transport=<nil> result={"ok":false,"error":"Unusable original-language subtitles: Unusable subtitles: malformed XML (1:44: unexpected close tag.); Unexpected token '<', \"<!doctype \"... is not valid JSON; Unexpected token '<', \"<!doctype \"... is not valid JSON"}
test_bundle_other_video_rejected expected=false pass=true transport=<nil> result={"ok":false,"error":"Unsupported video transcript extraction"}
test_bundle_bilibili_unverified_rejected expected=false pass=true transport=<nil> result={"ok":false,"error":"Unverified original language: Bilibili exposes subtitle language but no verified original-audio association"}
test_bundle_threads_target_selected expected=true pass=true transport=<nil> result={"ok":true,"text":"Original target body."}
test_bundle_medium_target_selected expected=true pass=true transport=<nil> result={"ok":true,"text":"Original target body."}
test_bundle_x_target_selected expected=true pass=true transport=<nil> result={"ok":true,"text":"Original target body."}
test_bundle_medium_preview_rejected expected=false pass=true transport=<nil> result={"ok":false,"error":"Inaccessible social target: subscription or preview"}
test_bundle_wrong_x_rejected expected=false pass=true transport=<nil> result={"ok":false,"error":"Inaccessible social target: main body unavailable"}
```

#### Actual live CLI results (trimmed)

```text
--- youtube ---
time=2026-10-23T14:30:17.807+08:00 level=INFO msg="processing item" index=1 total=1 url="https://www.youtube.com/watch?v=dQw4w9WgXcQ" title=""
time=2026-10-23T14:30:22.186+08:00 level=INFO msg="wrote content" path=/private/tmp/rlss-original-w3-output/www_youtube_com/2026-10-23__0424974c__content.md
Pipeline completed in 4s
  Total:   1
  Success: 1
  Skipped: 0
  Partial: 0
  Failed:  0
CLI_EXIT=0

--- threads ---
Pipeline completed in 5s
  Total:   1
  Success: 1
  Skipped: 0
  Partial: 0
  Failed:  0
time=2026-10-23T14:30:50.127+08:00 level=INFO msg="processing item" index=1 total=1 url=https://www.threads.com/@zuck/post/DeZtp-TkQHh title=""
time=2026-10-23T14:30:54.822+08:00 level=INFO msg="wrote content" path=/private/tmp/rlss-original-w3-output/www_threads_com/2026-10-23__f3772312__content.md
CLI_EXIT=0

--- medium ---
Pipeline completed in 7s
  Total:   1
  Success: 1
  Skipped: 0
  Partial: 0
  Failed:  0
time=2026-10-23T14:30:54.831+08:00 level=INFO msg="processing item" index=1 total=1 url=https://ev.medium.com/welcome-to-medium-9e53ca408c48 title=""
time=2026-10-23T14:30:57.639+08:00 level=WARN msg="headless extraction failed, retrying with headed mode" url=https://ev.medium.com/welcome-to-medium-9e53ca408c48 err="exception \"Uncaught (in promise) Error: Inaccessible social target: canonical identity missing or different\" (344:4148): Error: Inaccessible social target: canonical identity missing or different\n    at dt (<anonymous>:344:4148)\n    at Xi (<anonymous>:353:407)\n    at jo (<anonymous>:353:1942)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8"
time=2026-10-23T14:31:01.367+08:00 level=INFO msg="wrote content" path=/private/tmp/rlss-original-w3-output/ev_medium_com/2026-10-23__300cbc52__content.md
CLI_EXIT=0

--- x ---
Pipeline completed in 5s
  Total:   1
  Success: 0
  Skipped: 0
  Partial: 0
  Failed:  1
time=2026-10-23T14:31:01.374+08:00 level=INFO msg="processing item" index=1 total=1 url=https://x.com/jack/status/20 title=""
time=2026-10-23T14:31:02.256+08:00 level=WARN msg="headless extraction failed, retrying with headed mode" url=https://x.com/jack/status/20 err="page load error net::ERR_HTTP_RESPONSE_CODE_FAILURE"
time=2026-10-23T14:31:06.608+08:00 level=ERROR msg="item failed" url=https://x.com/jack/status/20 err="extract https://x.com/jack/status/20: headless attempt: page load error net::ERR_HTTP_RESPONSE_CODE_FAILURE; headed retry: exception \"Uncaught (in promise) Error: Inaccessible social target: main body unavailable\" (344:4148): Error: Inaccessible social target: main body unavailable\n    at dt (<anonymous>:344:4148)\n    at Xi (<anonymous>:353:1123)\n    at jo (<anonymous>:353:1942)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8"
CLI_EXIT=1

--- unknown ---
Pipeline completed in 8s
  Total:   1
  Success: 0
  Skipped: 0
  Partial: 0
  Failed:  1
time=2026-10-23T14:31:06.616+08:00 level=INFO msg="processing item" index=1 total=1 url="https://www.youtube.com/watch?v=rfscVS0vtbw" title=""
time=2026-10-23T14:31:10.758+08:00 level=WARN msg="headless extraction failed, retrying with headed mode" url="https://www.youtube.com/watch?v=rfscVS0vtbw" err="exception \"Uncaught (in promise) Error: Unverified original language: ambiguous or dubbed audio association\" (344:4148): Error: Unverified original language: ambiguous or dubbed audio association\n    at dt (<anonymous>:344:4148)\n    at Fo (<anonymous>:346:517)\n    at Vi (<anonymous>:346:1423)\n    at jo (<anonymous>:353:1391)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8"
time=2026-10-23T14:31:14.457+08:00 level=ERROR msg="item failed" url="https://www.youtube.com/watch?v=rfscVS0vtbw" err="extract https://www.youtube.com/watch?v=rfscVS0vtbw: headless attempt: exception \"Uncaught (in promise) Error: Unverified original language: ambiguous or dubbed audio association\" (344:4148): Error: Unverified original language: ambiguous or dubbed audio association\n    at dt (<anonymous>:344:4148)\n    at Fo (<anonymous>:346:517)\n    at Vi (<anonymous>:346:1423)\n    at jo (<anonymous>:353:1391)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8; headed retry: exception \"Uncaught (in promise) Error: Unverified original language: ambiguous or dubbed audio association\" (344:4148): Error: Unverified original language: ambiguous or dubbed audio association\n    at dt (<anonymous>:344:4148)\n    at Fo (<anonymous>:346:517)\n    at Vi (<anonymous>:346:1423)\n    at jo (<anonymous>:353:1391)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8"
CLI_EXIT=1

--- bilibili ---
Pipeline completed in 55s
  Total:   1
  Success: 0
  Skipped: 0
  Partial: 0
  Failed:  1
time=2026-10-23T14:31:14.465+08:00 level=INFO msg="processing item" index=1 total=1 url=https://www.bilibili.com/video/BV123 title=""
time=2026-10-23T14:31:29.011+08:00 level=WARN msg="headless extraction failed, retrying with headed mode" url=https://www.bilibili.com/video/BV123 err="exception \"Uncaught (in promise) Error: Unverified original language: Bilibili exposes subtitle language but no verified original-audio association\" (344:4148): Error: Unverified original language: Bilibili exposes subtitle language but no verified original-audio association\n    at dt (<anonymous>:344:4148)\n    at jo (<anonymous>:353:1455)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8"
time=2026-10-23T14:32:09.027+08:00 level=ERROR msg="item failed" url=https://www.bilibili.com/video/BV123 err="extract https://www.bilibili.com/video/BV123: headless attempt: exception \"Uncaught (in promise) Error: Unverified original language: Bilibili exposes subtitle language but no verified original-audio association\" (344:4148): Error: Unverified original language: Bilibili exposes subtitle language but no verified original-audio association\n    at dt (<anonymous>:344:4148)\n    at jo (<anonymous>:353:1455)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8; headed retry: context deadline exceeded"
CLI_EXIT=1

--- unsupported ---
Pipeline completed in 7s
  Total:   1
  Success: 0
  Skipped: 0
  Partial: 0
  Failed:  1
time=2026-10-23T14:32:09.036+08:00 level=INFO msg="processing item" index=1 total=1 url=https://vimeo.com/123 title=""
time=2026-10-23T14:32:12.394+08:00 level=WARN msg="headless extraction failed, retrying with headed mode" url=https://vimeo.com/123 err="exception \"Uncaught (in promise) Error: Unsupported video transcript extraction\" (344:4148): Error: Unsupported video transcript extraction\n    at dt (<anonymous>:344:4148)\n    at jo (<anonymous>:353:1783)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8"
time=2026-10-23T14:32:15.551+08:00 level=ERROR msg="item failed" url=https://vimeo.com/123 err="extract https://vimeo.com/123: headless attempt: exception \"Uncaught (in promise) Error: Unsupported video transcript extraction\" (344:4148): Error: Unsupported video transcript extraction\n    at dt (<anonymous>:344:4148)\n    at jo (<anonymous>:353:1783)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8; headed retry: exception \"Uncaught (in promise) Error: Unsupported video transcript extraction\" (344:4148): Error: Unsupported video transcript extraction\n    at dt (<anonymous>:344:4148)\n    at jo (<anonymous>:353:1783)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8"
CLI_EXIT=1
```

#### Actual batch, cache, completed and dry-run

```text
--- batch ---
time=2026-10-23T14:31:44.429+08:00 level=INFO msg="processing item" index=1 total=3 url="https://www.youtube.com/watch?v=rfscVS0vtbw" title="Ambiguous original audio"
time=2026-10-23T14:31:48.423+08:00 level=WARN msg="headless extraction failed, retrying with headed mode" url="https://www.youtube.com/watch?v=rfscVS0vtbw" err="exception \"Uncaught (in promise) Error: Unverified original language: ambiguous or dubbed audio association\" (344:4148): Error: Unverified original language: ambiguous or dubbed audio association\n    at dt (<anonymous>:344:4148)\n    at Fo (<anonymous>:346:517)\n    at Vi (<anonymous>:346:1423)\n    at jo (<anonymous>:353:1391)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8"
time=2026-10-23T14:31:52.272+08:00 level=ERROR msg="item failed" url="https://www.youtube.com/watch?v=rfscVS0vtbw" err="extract https://www.youtube.com/watch?v=rfscVS0vtbw: headless attempt: exception \"Uncaught (in promise) Error: Unverified original language: ambiguous or dubbed audio association\" (344:4148): Error: Unverified original language: ambiguous or dubbed audio association\n    at dt (<anonymous>:344:4148)\n    at Fo (<anonymous>:346:517)\n    at Vi (<anonymous>:346:1423)\n    at jo (<anonymous>:353:1391)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8; headed retry: exception \"Uncaught (in promise) Error: Unverified original language: ambiguous or dubbed audio association\" (344:4148): Error: Unverified original language: ambiguous or dubbed audio association\n    at dt (<anonymous>:344:4148)\n    at Fo (<anonymous>:346:517)\n    at Vi (<anonymous>:346:1423)\n    at jo (<anonymous>:353:1391)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8"
time=2026-10-23T14:31:52.272+08:00 level=INFO msg="processing item" index=2 total=3 url="https://www.youtube.com/watch?v=dQw4w9WgXcQ" title="Public original subtitles"
time=2026-10-23T14:31:56.191+08:00 level=INFO msg="wrote content" path=/private/tmp/rlss-original-w3-batch-output/www_youtube_com/2026-10-23__0424974c__content.md
time=2026-10-23T14:31:56.191+08:00 level=INFO msg="processing item" index=3 total=3 url=http://127.0.0.1:18643/article title="Ordinary test article"
time=2026-10-23T14:31:58.568+08:00 level=INFO msg="wrote content" path=/private/tmp/rlss-original-w3-batch-output/127_0_0_1/2026-10-23__3f11badb__content.md
Pipeline completed in 14s
  Total:   3
  Success: 2
  Skipped: 0
  Partial: 0
  Failed:  1
CLI_EXIT=0

--- cache ---
time=2026-10-23T14:32:02.511+08:00 level=INFO msg="processing item" index=1 total=1 url="https://www.youtube.com/watch?v=rfscVS0vtbw" title="Ambiguous original audio"
time=2026-10-23T14:32:02.511+08:00 level=INFO msg="resuming: content exists, skipping extraction" sha8=5ff986bd
time=2026-10-23T14:32:02.511+08:00 level=INFO msg="cached video requires fresh extraction" sha8=5ff986bd err="unverified original language: missing or invalid target-bound transcript provenance"
time=2026-10-23T14:32:06.173+08:00 level=WARN msg="headless extraction failed, retrying with headed mode" url="https://www.youtube.com/watch?v=rfscVS0vtbw" err="exception \"Uncaught (in promise) Error: Unverified original language: ambiguous or dubbed audio association\" (344:4148): Error: Unverified original language: ambiguous or dubbed audio association\n    at dt (<anonymous>:344:4148)\n    at Fo (<anonymous>:346:517)\n    at Vi (<anonymous>:346:1423)\n    at jo (<anonymous>:353:1391)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8"
time=2026-10-23T14:32:09.854+08:00 level=ERROR msg="item failed" url="https://www.youtube.com/watch?v=rfscVS0vtbw" err="extract https://www.youtube.com/watch?v=rfscVS0vtbw: headless attempt: exception \"Uncaught (in promise) Error: Unverified original language: ambiguous or dubbed audio association\" (344:4148): Error: Unverified original language: ambiguous or dubbed audio association\n    at dt (<anonymous>:344:4148)\n    at Fo (<anonymous>:346:517)\n    at Vi (<anonymous>:346:1423)\n    at jo (<anonymous>:353:1391)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8; headed retry: exception \"Uncaught (in promise) Error: Unverified original language: ambiguous or dubbed audio association\" (344:4148): Error: Unverified original language: ambiguous or dubbed audio association\n    at dt (<anonymous>:344:4148)\n    at Fo (<anonymous>:346:517)\n    at Vi (<anonymous>:346:1423)\n    at jo (<anonymous>:353:1391)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8"
Pipeline completed in 7s
  Total:   1
  Success: 0
  Skipped: 0
  Partial: 0
  Failed:  1
CLI_EXIT=0

--- completed ---
time=2026-10-23T14:32:52.755+08:00 level=INFO msg="processing item" index=1 total=1 url="https://www.youtube.com/watch?v=rfscVS0vtbw" title="Ambiguous original audio"
Pipeline completed in 0s
  Total:   1
  Success: 0
  Skipped: 1
  Partial: 0
  Failed:  0
CLI_EXIT=0

--- dry ---
time=2026-10-23T14:32:52.764+08:00 level=INFO msg="processing item" index=1 total=3 url="https://www.youtube.com/watch?v=rfscVS0vtbw" title="Ambiguous original audio"
time=2026-10-23T14:32:52.764+08:00 level=INFO msg="dry-run: would process" url="https://www.youtube.com/watch?v=rfscVS0vtbw" sha8=5ff986bd
time=2026-10-23T14:32:52.764+08:00 level=INFO msg="processing item" index=2 total=3 url="https://www.youtube.com/watch?v=dQw4w9WgXcQ" title="Public original subtitles"
time=2026-10-23T14:32:52.764+08:00 level=INFO msg="dry-run: would process" url="https://www.youtube.com/watch?v=dQw4w9WgXcQ" sha8=0424974c
time=2026-10-23T14:32:52.764+08:00 level=INFO msg="processing item" index=3 total=3 url=http://127.0.0.1:18643/article title="Ordinary test article"
time=2026-10-23T14:32:52.764+08:00 level=INFO msg="dry-run: would process" url=http://127.0.0.1:18643/article sha8=3f11badb
Pipeline completed in 0s
  Total:   3
  Success: 0
  Skipped: 3
  Partial: 0
  Failed:  0

--- empty ---
time=2026-10-23T14:32:52.771+08:00 level=INFO msg="no items to process"
```

#### Watch cycle (trimmed)

```text
time=2026-10-23T14:32:32.520+08:00 level=INFO msg="processing item" index=1 total=3 url="https://www.youtube.com/watch?v=rfscVS0vtbw" title="Ambiguous original audio"
time=2026-10-23T14:32:36.644+08:00 level=WARN msg="headless extraction failed, retrying with headed mode" url="https://www.youtube.com/watch?v=rfscVS0vtbw" err="exception \"Uncaught (in promise) Error: Unverified original language: ambiguous or dubbed audio association\" (344:4148): Error: Unverified original language: ambiguous or dubbed audio association\n    at dt (<anonymous>:344:4148)\n    at Fo (<anonymous>:346:517)\n    at Vi (<anonymous>:346:1423)\n    at jo (<anonymous>:353:1391)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8"
time=2026-10-23T14:32:40.472+08:00 level=ERROR msg="item failed" url="https://www.youtube.com/watch?v=rfscVS0vtbw" err="extract https://www.youtube.com/watch?v=rfscVS0vtbw: headless attempt: exception \"Uncaught (in promise) Error: Unverified original language: ambiguous or dubbed audio association\" (344:4148): Error: Unverified original language: ambiguous or dubbed audio association\n    at dt (<anonymous>:344:4148)\n    at Fo (<anonymous>:346:517)\n    at Vi (<anonymous>:346:1423)\n    at jo (<anonymous>:353:1391)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8; headed retry: exception \"Uncaught (in promise) Error: Unverified original language: ambiguous or dubbed audio association\" (344:4148): Error: Unverified original language: ambiguous or dubbed audio association\n    at dt (<anonymous>:344:4148)\n    at Fo (<anonymous>:346:517)\n    at Vi (<anonymous>:346:1423)\n    at jo (<anonymous>:353:1391)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8"
time=2026-10-23T14:32:40.472+08:00 level=INFO msg="processing item" index=2 total=3 url="https://www.youtube.com/watch?v=dQw4w9WgXcQ" title="Public original subtitles"
time=2026-10-23T14:32:40.473+08:00 level=INFO msg="resuming: content exists, skipping extraction" sha8=0424974c
time=2026-10-23T14:32:40.473+08:00 level=INFO msg="processing item" index=3 total=3 url=http://127.0.0.1:18643/article title="Ordinary test article"
time=2026-10-23T14:32:40.474+08:00 level=INFO msg="resuming: content exists, skipping extraction" sha8=3f11badb
time=2026-10-23T14:32:40.474+08:00 level=INFO msg="watch: cycle complete" report="Pipeline completed in 8s\n  Total:   3\n  Success: 2\n  Skipped: 0\n  Partial: 0\n  Failed:  1\n\nErrors:\n  - Ambiguous original audio: extract https://www.youtube.com/watch?v=rfscVS0vtbw: headless attempt: exception \"Uncaught (in promise) Error: Unverified original language: ambiguous or dubbed audio association\" (344:4148): Error: Unverified original language: ambiguous or dubbed audio association\n    at dt (<anonymous>:344:4148)\n    at Fo (<anonymous>:346:517)\n    at Vi (<anonymous>:346:1423)\n    at jo (<anonymous>:353:1391)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8; headed retry: exception \"Uncaught (in promise) Error: Unverified original language: ambiguous or dubbed audio association\" (344:4148): Error: Unverified original language: ambiguous or dubbed audio association\n    at dt (<anonymous>:344:4148)\n    at Fo (<anonymous>:346:517)\n    at Vi (<anonymous>:346:1423)\n    at jo (<anonymous>:353:1391)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8\n"
time=2026-10-23T14:32:40.474+08:00 level=INFO msg="watch: sleeping" minutes=10
```

#### Criterion tests

```text
A1
# Subtest: A1/A3 native player refresh recovers empty inline captions with bounded first-party requests
ok 1 - A1/A3 native player refresh recovers empty inline captions with bounded first-party requests
# Subtest: A1 verified original audio association
ok 2 - A1 verified original audio association
# Subtest: A1 unverified default language rejected
ok 3 - A1 unverified default language rejected
# Subtest: A1/A3 refreshed invalid identity, language and API failures stay unsuccessful
ok 4 - A1/A3 refreshed invalid identity, language and API failures stay unsuccessful
# tests 4
# pass 4
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
# Subtest: A1/A3 native player refresh recovers empty inline captions with bounded first-party requests
ok 1 - A1/A3 native player refresh recovers empty inline captions with bounded first-party requests
# Subtest: A1/A3 refreshed invalid identity, language and API failures stay unsuccessful
ok 2 - A1/A3 refreshed invalid identity, language and API failures stay unsuccessful
# Subtest: A3 verified social content succeeds without network fallback
ok 3 - A3 verified social content succeeds without network fallback
# Subtest: A3 login, preview and unsupported video fail
ok 4 - A3 login, preview and unsupported video fail
# tests 4
# pass 4
# fail 0

A4
# Subtest: A4 captions work when Trusted Types rejects HTML parsing sinks
ok 1 - A4 captions work when Trusted Types rejects HTML parsing sinks
# Subtest: A4 blank human track does not block usable automatic track
ok 2 - A4 blank human track does not block usable automatic track
# Subtest: A4 malformed, truncated and timestamp-only cues fail
ok 3 - A4 malformed, truncated and timestamp-only cues fail
# tests 3
# pass 3
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
ok  	github.com/kouko/reading-list-summarize-scraper/internal/pipeline	0.380s
```

No complete Go/npm package suite or workflow-helper test was run by this tester. Finalize-review is still scheduled by the primary reviewer; no steps skipped by instruction. Dispatch remains inherited overrides null, unobservable-main-profile; no subagents dispatched.

Commit classification: routine acceptance rerun with corrected observations; no product implementation or new product design decision. Exact final commit-message deterministic privacy scan is required, and no ambiguous private-party text requires a semantic judge.


## Re-run on 2026-10-23, at 595326adefaa8669732ebda43bc09eb35b76aaa5

Fix range checked in full: `515086f..595326a`. Only `extraction.ts` X branch, its test, generated bundle, and plan changed. The X branch now rejects an explicit show-more marker owned by the target post or long article; its cloned own-body filter removes quoted/nested posts before this check. Threads, Medium, all video functions, Go CLI/pipeline, settings and cache/watch/skip code are unchanged. The carried-over rows below were checked against that diff, not inferred from commit title.

### Clean setup and bounds

- New clean worktree: `git worktree add /private/tmp/rlss-original-acceptance-w4 595326adefaa8669732ebda43bc09eb35b76aaa5`.
- README quick Go build: `GOCACHE=/private/tmp/rlss-go-build go build -o /private/tmp/rlss-original-w4-cli ./cmd/rlss/`. Also `npm ci --ignore-scripts` and `npm run build:defuddle`. All passed; 26 locked packages / zero npm vulnerabilities, generated bundle 784.7kb. Worktree remained clean after setup/rebuild.
- Independent criterion commands: `node --test --test-name-pattern='A2' extraction.test.mjs` (3/3 pass), then `node --test --test-name-pattern='A3' extraction.test.mjs` (5/5 pass). Selected pipeline command `GOCACHE=/private/tmp/rlss-go-build go test ./internal/pipeline -run 'TestProcessItemWithFallback|TestExtractionRejectsObservedWalls' -v` passed.
- Actual shipped browser bundle: `go run /private/tmp/rlss-original-browser.go > /private/tmp/rlss-original-w4-browser.log` from clean worktree. Harness still injects rebuilt bundle and invokes actual `window.extractArticle()`; only navigation/caption network resources intercepted. Added four synthetic X cases named below. All 16 cases passed.
- Current direct CLI uses fresh `/private/tmp/rlss-original-w4-profile` and fresh output `/private/tmp/rlss-original-w4-output`, 35s browser timeouts, 2s load wait, existing headed retry, no sources or copying, always `--no-summary`. No user profiles, vault, root binary or running batches touched; no provider calls.
- Exact command pattern: `/private/tmp/rlss-original-w4-cli url <public-url> --config /private/tmp/rlss-original-w4.yaml --no-summary`. Independently reran full real CLI for Threads `https://www.threads.com/@zuck/post/DeZtp-TkQHh`, Medium `https://ev.medium.com/welcome-to-medium-9e53ca408c48`, X `https://x.com/jack/status/20`, and ambiguous-audio YouTube `https://www.youtube.com/watch?v=rfscVS0vtbw`.
- Threads and Medium again exit 0 and save 136 / 6825 character bodies with expected target introduction/title and no recommendation labels. No summaries saved. Public X and ambiguous-audio YouTube exit 1; zero X or failed-video output files. X failure is still an environmental/access observation, not proof of an adapter defect.
- No full Go/npm package suite, second watch recovery cycle, private credentials or workflow-helper work was added. No skipped-by-instruction step. Full suite/finalization remains owned by primary and cannot imply acceptance while unresolved important scope remains.

### Per-row rerun record

- 1: carried over — full diff confirms original-language candidate evidence, native requests, cue parser and video saving unchanged; prior real YouTube success and unresolved Bilibili limitation remain.
- 2: re-tested — actual built browser target/negative cases on all three platforms; full real CLI Threads/Medium saved bodies and X failure repeated.
- 3: re-tested — current A3 includes native-video failures and all social rejection surfaces; current actual built browser checks post/article preview, Medium wall, wrong X, unsupported video and ambiguous language; pipeline no-write/retry criteria pass; real failed video/X saved nothing.
- 4: carried over — strict SAX, cue validation and human/automatic candidate selection unchanged.
- 5: carried over — video implementation, CLI, pipeline and source labels unchanged; prior direct/batch success remains.
- 6: carried over — empty-human/native-refresh automatic fallback implementation unchanged.
- 7: carried over — multiple original-track ordering/validation unchanged.
- 8: re-tested — real ambiguous-language video direct CLI exit 1 with explicit original-language cause and no output.
- 9: carried over — video unavailable/unusable checks and fallback exhaustion unchanged.
- 10: carried over — unsupported-video branch and direct-command failure handling unchanged.
- 11: re-tested — current real Threads/Medium exit 0 and target-body saved; X exit 1, no new body.
- 12: re-tested — current Medium preview rejection and positive fallback, all login/challenge/retry criterion cases passed; no credential restoration attempted; remains partly.
- 13: re-tested — current actual built browser X post/article preview errors, wrong-target X, Medium preview, Threads positive/exclusion fixtures and criterion negatives passed. Quoted/irrelevant truncation markers did not block complete target; no full three-platform negative CLI matrix, remains partly.
- 14: carried over — progress, browser timeout and retry policy unchanged; this fix adds only extraction failures handled by the existing path.
- 15: carried over — no Go batch/watch code or failure accounting changed; new preview errors use the previously witnessed failed-extraction path. No next-cycle recovery claim.
- 16: carried over — pending-cache invalidation, video provenance and preserving old text on failure unchanged.
- 17: carried over — completed-summary skip behavior unchanged.
- 18: carried over — source selection, no-items and dry-run paths unchanged.

### Explicit preview behavior observed in real Chrome

- `test_bundle_x_post_preview_rejected`: substantive target post with its own `tweet-text-show-more-link` fails `Inaccessible social target: truncated post preview`.
- `test_bundle_x_article_preview_rejected`: canonical target long article with its own marker fails `Inaccessible social target: truncated article preview`.
- `test_bundle_x_quoted_preview_complete_target_accepted`: complete own target remains exactly `Complete own target body.`; quote's marker/body excluded.
- `test_bundle_x_complete_article_unrelated_preview_accepted`: complete canonical long article remains exactly `Complete target article.`; unrelated post's marker does not block it.
- These controlled page cases prove current generated-entry behavior, not current public-X availability or a witnessed full CLI preview scenario. New criterion regression asserts the exact returned text for both positives.

### Current finding status

- **Review's explicit-X-preview rejection issue: resolved for tested target-owned post and long-article markers**, `extraction.ts:179-195`; independently verified in current generated browser entry plus regression.
- **important, unresolved — Bilibili scope contradiction**, `extraction.ts:236`, retained spec supported-video decision / Video success. Still unconditional rejection with no positive implementation; prior partly verdict persists. Requires grounded success implementation or explicit scope revision with user.
- **important verification limitation — live X success**, retained spec Social success: current public target still failed; controlled X positive/negative cases pass. No claim that network/access failure is a code defect. Requires legitimate accessible-target positive CLI/save witness; A2 and Social success remain partly.
- Earlier real YouTube success gap stays resolved by independent W3 evidence because video implementation is unchanged. Current run does not make any broader availability promise.
- No important finding was dismissed and no acceptance/ready verdict is issued.

### Captured output — actual rebuilt browser

```text
test_bundle_original_automatic_saved expected=true pass=true transport=<nil> result={"ok":true,"text":"<!-- rlss-video-v1 {\"platform\":\"youtube\",\"target\":\"abc\",\"language\":\"en\",\"source\":\"automatic\",\"evidence\":\"original-audio-asr-association\"} -->\n\nSubtitle language: en\nSubtitle source: automatic\n\nActual original spoken words."}
test_bundle_empty_human_automatic_used expected=true pass=true transport=<nil> result={"ok":true,"text":"<!-- rlss-video-v1 {\"platform\":\"youtube\",\"target\":\"abc\",\"language\":\"en\",\"source\":\"automatic\",\"evidence\":\"original-audio-asr-association\"} -->\n\nSubtitle language: en\nSubtitle source: automatic\n\nActual original spoken words."}
test_bundle_multiple_usable_validated expected=true pass=true transport=<nil> result={"ok":true,"text":"<!-- rlss-video-v1 {\"platform\":\"youtube\",\"target\":\"abc\",\"language\":\"en\",\"source\":\"human\",\"evidence\":\"original-audio-asr-association\"} -->\n\nSubtitle language: en\nSubtitle source: human\n\nActual original spoken words."}
test_bundle_unknown_original_rejected expected=false pass=true transport=<nil> result={"ok":false,"error":"Unverified original language: missing or conflicting original-audio automatic captions"}
test_bundle_malformed_cues_rejected expected=false pass=true transport=<nil> result={"ok":false,"error":"Unusable original-language subtitles: Unusable subtitles: malformed XML (1:44: unexpected close tag.); Unexpected token '<', \"<!doctype \"... is not valid JSON; Unexpected token '<', \"<!doctype \"... is not valid JSON"}
test_bundle_other_video_rejected expected=false pass=true transport=<nil> result={"ok":false,"error":"Unsupported video transcript extraction"}
test_bundle_bilibili_unverified_rejected expected=false pass=true transport=<nil> result={"ok":false,"error":"Unverified original language: Bilibili exposes subtitle language but no verified original-audio association"}
test_bundle_threads_target_selected expected=true pass=true transport=<nil> result={"ok":true,"text":"Original target body."}
test_bundle_medium_target_selected expected=true pass=true transport=<nil> result={"ok":true,"text":"Original target body."}
test_bundle_x_target_selected expected=true pass=true transport=<nil> result={"ok":true,"text":"Original target body."}
test_bundle_medium_preview_rejected expected=false pass=true transport=<nil> result={"ok":false,"error":"Inaccessible social target: subscription or preview"}
test_bundle_wrong_x_rejected expected=false pass=true transport=<nil> result={"ok":false,"error":"Inaccessible social target: main body unavailable"}
test_bundle_x_post_preview_rejected expected=false pass=true transport=<nil> result={"ok":false,"error":"Inaccessible social target: truncated post preview"}
test_bundle_x_article_preview_rejected expected=false pass=true transport=<nil> result={"ok":false,"error":"Inaccessible social target: truncated article preview"}
test_bundle_x_quoted_preview_complete_target_accepted expected=true pass=true transport=<nil> result={"ok":true,"text":"Complete own target body."}
test_bundle_x_complete_article_unrelated_preview_accepted expected=true pass=true transport=<nil> result={"ok":true,"text":"Complete target article."}
```

### Captured output — all named social CLI surfaces and original-language failure

```text
--- threads ---
Pipeline completed in 5s
  Total:   1
  Success: 1
  Skipped: 0
  Partial: 0
  Failed:  0
time=2026-10-23T14:45:01.207+08:00 level=INFO msg="processing item" index=1 total=1 url=https://www.threads.com/@zuck/post/DeZtp-TkQHh title=""
time=2026-10-23T14:45:06.141+08:00 level=INFO msg="wrote content" path=/private/tmp/rlss-original-w4-output/www_threads_com/2026-10-23__f3772312__content.md
CLI_EXIT=0

--- medium ---
Pipeline completed in 7s
  Total:   1
  Success: 1
  Skipped: 0
  Partial: 0
  Failed:  0
time=2026-10-23T14:45:06.149+08:00 level=INFO msg="processing item" index=1 total=1 url=https://ev.medium.com/welcome-to-medium-9e53ca408c48 title=""
time=2026-10-23T14:45:08.883+08:00 level=WARN msg="headless extraction failed, retrying with headed mode" url=https://ev.medium.com/welcome-to-medium-9e53ca408c48 err="exception \"Uncaught (in promise) Error: Inaccessible social target: canonical identity missing or different\" (344:4148): Error: Inaccessible social target: canonical identity missing or different\n    at ut (<anonymous>:344:4148)\n    at Xi (<anonymous>:353:407)\n    at zo (<anonymous>:353:1942)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8"
time=2026-10-23T14:45:12.724+08:00 level=INFO msg="wrote content" path=/private/tmp/rlss-original-w4-output/ev_medium_com/2026-10-23__300cbc52__content.md
CLI_EXIT=0

--- x ---
Pipeline completed in 5s
  Total:   1
  Success: 0
  Skipped: 0
  Partial: 0
  Failed:  1
time=2026-10-23T14:45:12.731+08:00 level=INFO msg="processing item" index=1 total=1 url=https://x.com/jack/status/20 title=""
time=2026-10-23T14:45:13.420+08:00 level=WARN msg="headless extraction failed, retrying with headed mode" url=https://x.com/jack/status/20 err="page load error net::ERR_HTTP_RESPONSE_CODE_FAILURE"
time=2026-10-23T14:45:17.299+08:00 level=ERROR msg="item failed" url=https://x.com/jack/status/20 err="extract https://x.com/jack/status/20: headless attempt: page load error net::ERR_HTTP_RESPONSE_CODE_FAILURE; headed retry: exception \"Uncaught (in promise) Error: Inaccessible social target: main body unavailable\" (344:4148): Error: Inaccessible social target: main body unavailable\n    at ut (<anonymous>:344:4148)\n    at Xi (<anonymous>:353:1123)\n    at zo (<anonymous>:353:1942)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8"
CLI_EXIT=1

--- unknown ---
Pipeline completed in 8s
  Total:   1
  Success: 0
  Skipped: 0
  Partial: 0
  Failed:  1
time=2026-10-23T14:45:17.306+08:00 level=INFO msg="processing item" index=1 total=1 url="https://www.youtube.com/watch?v=rfscVS0vtbw" title=""
time=2026-10-23T14:45:21.240+08:00 level=WARN msg="headless extraction failed, retrying with headed mode" url="https://www.youtube.com/watch?v=rfscVS0vtbw" err="exception \"Uncaught (in promise) Error: Unverified original language: ambiguous or dubbed audio association\" (344:4148): Error: Unverified original language: ambiguous or dubbed audio association\n    at ut (<anonymous>:344:4148)\n    at Fo (<anonymous>:346:517)\n    at Gi (<anonymous>:346:1423)\n    at zo (<anonymous>:353:1391)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8"
time=2026-10-23T14:45:25.260+08:00 level=ERROR msg="item failed" url="https://www.youtube.com/watch?v=rfscVS0vtbw" err="extract https://www.youtube.com/watch?v=rfscVS0vtbw: headless attempt: exception \"Uncaught (in promise) Error: Unverified original language: ambiguous or dubbed audio association\" (344:4148): Error: Unverified original language: ambiguous or dubbed audio association\n    at ut (<anonymous>:344:4148)\n    at Fo (<anonymous>:346:517)\n    at Gi (<anonymous>:346:1423)\n    at zo (<anonymous>:353:1391)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8; headed retry: exception \"Uncaught (in promise) Error: Unverified original language: ambiguous or dubbed audio association\" (344:4148): Error: Unverified original language: ambiguous or dubbed audio association\n    at ut (<anonymous>:344:4148)\n    at Fo (<anonymous>:346:517)\n    at Gi (<anonymous>:346:1423)\n    at zo (<anonymous>:353:1391)\n    at window.extractArticle (<anonymous>:353:2118)\n    at <anonymous>:1:8"
CLI_EXIT=1
```

### Captured output — criterion checks only

```text
A2
# Subtest: A2 target bodies on three platforms exclude replies
ok 1 - A2 target bodies on three platforms exclude replies
# Subtest: A2/A3 explicit target previews fail while complete own bodies remain usable
ok 2 - A2/A3 explicit target previews fail while complete own bodies remain usable
# Subtest: A2 wrong target rejected
ok 3 - A2 wrong target rejected
# tests 3
# pass 3
# fail 0

A3
# Subtest: A1/A3 native player refresh recovers empty inline captions with bounded first-party requests
ok 1 - A1/A3 native player refresh recovers empty inline captions with bounded first-party requests
# Subtest: A1/A3 refreshed invalid identity, language and API failures stay unsuccessful
ok 2 - A1/A3 refreshed invalid identity, language and API failures stay unsuccessful
# Subtest: A2/A3 explicit target previews fail while complete own bodies remain usable
ok 3 - A2/A3 explicit target previews fail while complete own bodies remain usable
# Subtest: A3 verified social content succeeds without network fallback
ok 4 - A3 verified social content succeeds without network fallback
# Subtest: A3 login, preview and unsupported video fail
ok 5 - A3 login, preview and unsupported video fail
# tests 5
# pass 5
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
=== RUN   TestExtractionRejectsObservedWalls
--- PASS: TestExtractionRejectsObservedWalls (0.00s)
PASS
ok  	github.com/kouko/reading-list-summarize-scraper/internal/pipeline	0.372s
```

Dispatch unchanged: inherited overrides null, unobservable-main-profile; no agents spawned. Commit classification routine fix-verification evidence, no new product decision. Exact commit-message privacy scan required; no private-party ambiguity requires semantic judge. Only report/evidence changed by this tester.


## Re-run on 2026-10-23, at bf8b98ab8b0e91325f2b661600777b65efea0ac5

Range checked in full: `c2543e1..bf8b98a`. The shared extraction entry is now an error wrapper around the same successful extraction function. Only after failure, explicit challenge UI/title adds a fixed Cloudflare/CAPTCHA identity to the existing error; no raw page body is forwarded. Source, regression, generated bundle and plan are the only changed files. Go settings/policy, all successful body paths, target preview guards, video language/quality and caches are unchanged. Retained reasons below were checked against the actual diff.

### Fresh setup and bounded commands

- Fresh clean worktree `git worktree add /private/tmp/rlss-original-acceptance-w5 bf8b98ab8b0e91325f2b661600777b65efea0ac5`.
- Followed README with `npm ci --ignore-scripts`, `npm run build:defuddle`, `GOCACHE=/private/tmp/rlss-go-build go build -o /private/tmp/rlss-original-w5-cli ./cmd/rlss/`. Passed; 26 locked packages, zero npm vulnerabilities, actual browser bundle 785.3kb. Worktree remained clean after setup.
- Criterion only: `node --test --test-name-pattern='A3' extraction.test.mjs` → 6/6 pass. Consumer criteria only: `GOCACHE=/private/tmp/rlss-go-build go test ./internal/pipeline -run 'TestProcessItemWithFallback|TestExtractionRejectsObservedWalls' -v` → pass, including `specialized_challenge_disabled`, retry policy variants and no-write-on-failure cases.
- Actual generated browser entry: `go run /private/tmp/rlss-original-browser.go > /private/tmp/rlss-original-w5-browser.log` in fresh worktree. Uses actual current bundle and `window.extractArticle()`, not a substituted extractor. Only navigation/caption resources controlled. All 21 fixture cases passed, including Cloudflare failures for X/Medium/YouTube, CAPTCHA X, and a successful article mentioning those words. Other existing platform/preview/video cases also passed.
- New fixture page bodies are synthetic. Their text did not leak into returned errors. The same words in a complete target article remain successful. The wrapper does not promise every possible vendor challenge is recognized.
- Real CLI settings comparison: same public URL `https://ev.medium.com/welcome-to-medium-9e53ca408c48`, separate fresh profiles/output for each setting, headless true, load wait 2s, headless/headed bounds 30s, no reading-list sources or copying, always `--no-summary`. Exact command `/private/tmp/rlss-original-w5-cli url 'https://ev.medium.com/welcome-to-medium-9e53ca408c48' --config /private/tmp/rlss-original-w5-disabled.yaml --no-summary`, then enabled config at analogous path.
- No user profiles, vault, active batch, root binary or providers used. No full Go/npm package suite or watch-recovery cycle attempted. There were no skipped-by-instruction steps; finalization belongs to the primary and unresolved important Bilibili scope still prevents declaring the original promise complete.

### Independently reproduced live setting effect

- **headed_on_block false:** real CLI sees `Cloudflare security challenge: Inaccessible social target: canonical identity missing or different`; a single headless extraction attempt, **zero** `retrying with headed mode` logs, exit 1, Success 0 / Failed 1, zero Markdown files.
- **headed_on_block true:** same fixed challenge identity in initial failure; exactly one `retrying with headed mode` log. Headed retry also sees the challenge, exit 1, Success 0 / Failed 1, zero Markdown files.
- This independently verifies policy compatibility through the actual browser/CLI boundary. It does not claim true guarantees successful bypass or legitimate access recovery. Earlier W3/W4 Medium saved-body successes remain valid observations for those runs; this run's access result is unsuccessful.
- Existing consumer recognizes the fixed challenge identity, preserving configured suppression; no Go implementation change was necessary.
- Captured logs contain only fixed classification and existing diagnostic, no challenge body, cookie, token or signed caption URL.

### Per-row rerun record

- 1: carried over — original-language evidence, video selection and returned content are unchanged; wrapper changes errors only after a rejection.
- 2: carried over — all successful social-body extraction branches, target checks and saving are unchanged. Prior independent Threads/Medium saved-body evidence remains historical success; current public Medium access can still fail as observed below.
- 3: re-tested — A3 criteria across native-video and social error surfaces, actual rebuilt-browser challenge cases, and existing pipeline retry/no-write consumer tests all pass. Real Medium disabled/enabled CLI failures save no content.
- 4: carried over — human/automatic cue parsing and validation unchanged.
- 5: carried over — video success and saving unchanged; old confirmed YouTube success remains and Bilibili implementation gap remains.
- 6: carried over — empty-human fallback and refreshed automatic subtitles unchanged.
- 7: carried over — multiple original-track selection unchanged.
- 8: carried over — ambiguous original-audio rejection unchanged; wrapper only decorates explicit challenge pages.
- 9: carried over — original subtitle unusability/translation checks unchanged; selected A3 tests also exercised failure propagation.
- 10: carried over — supported/unsupported video routing and direct CLI error status unchanged.
- 11: carried over — successful social content return and save unchanged; prior Threads/Medium positive witness retained, not a guarantee of current access.
- 12: re-tested — real Medium explicit challenge with headed_on_block false has one headless failure and no headed retry; true has one headed retry. Both fail safely because challenge persists. Controlled Cloudflare/CAPTCHA and login/preview error cases pass. No credential recovery attempted.
- 13: carried over — target identity and truncation/preview rules unchanged; current actual browser incidental preview/negative controls also passed.
- 14: re-tested — actual CLI processing/failure logs and policy-dependent retry witnessed with retained challenge identity; no stalled request or timeout enforcement test added.
- 15: carried over — no Go batch/watch/failure accounting change, and consumer retry policy is directly retested; no recovery cycle expansion.
- 16: carried over — video cache invalidation/resume and writes unchanged.
- 17: carried over — completed-summary skipping unchanged.
- 18: carried over — source selection, no-items and dry-run unchanged.

### Current finding status

- **Challenge-setting compatibility issue resolved for witnessed case**, `extraction.ts:244-259`, consumer `TestProcessItemWithFallback/specialized_challenge_disabled`: actual public Medium explicit challenge now preserves identity and suppresses headed retry when disabled; enabled setting performs the existing retry.
- **important, unresolved — Bilibili scope contradiction**, `extraction.ts:236`, retained supported-video / Video success promise. The success path is absent, rather than merely failing current access. Grounded implementation and independent success or explicit user scope revision still required.
- **important verification limitation — X live success**, retained Social success. Prior actual public X failure and controlled positive tests remain; this wrapper fix does not prove current X availability or an extraction defect.
- No important finding dismissed. No readiness, acceptance or merge claim made.

### Current generated-browser challenge evidence

```text
test_bundle_x_cloudflare_failure_classified expected=false pass=true transport=<nil> result={"ok":false,"error":"Cloudflare security challenge: Inaccessible social target: main body unavailable"}
test_bundle_medium_cloudflare_failure_classified expected=false pass=true transport=<nil> result={"ok":false,"error":"Cloudflare security challenge: Inaccessible social target: canonical identity missing or different"}
test_bundle_youtube_cloudflare_failure_classified expected=false pass=true transport=<nil> result={"ok":false,"error":"Cloudflare security challenge: Unavailable subtitles: missing player data"}
test_bundle_x_captcha_failure_classified expected=false pass=true transport=<nil> result={"ok":false,"error":"CAPTCHA security challenge: Inaccessible social target: main body unavailable"}
test_bundle_x_article_mentions_preserved expected=true pass=true transport=<nil> result={"ok":true,"text":"Cloudflare CAPTCHA security challenge is this article topic."}
```

### Actual live CLI, retry disabled

```text
Pipeline completed in 3s
  Total:   1
  Success: 0
  Skipped: 0
  Partial: 0
  Failed:  1
time=2026-10-23T14:54:43.078+08:00 level=INFO msg="processing item" index=1 total=1 url=https://ev.medium.com/welcome-to-medium-9e53ca408c48 title=""
time=2026-10-23T14:54:45.955+08:00 level=ERROR msg="item failed" url=https://ev.medium.com/welcome-to-medium-9e53ca408c48 err="extract https://ev.medium.com/welcome-to-medium-9e53ca408c48: exception \"Uncaught (in promise) Error: Cloudflare security challenge: Inaccessible social target: canonical identity missing or different\" (344:4148): Error: Cloudflare security challenge: Inaccessible social target: canonical identity missing or different\n    at st (<anonymous>:344:4148)\n    at zo (<anonymous>:353:2622)"
CLI_EXIT=1
```

### Actual live CLI, retry enabled

```text
Pipeline completed in 6s
  Total:   1
  Success: 0
  Skipped: 0
  Partial: 0
  Failed:  1
time=2026-10-23T14:54:45.963+08:00 level=INFO msg="processing item" index=1 total=1 url=https://ev.medium.com/welcome-to-medium-9e53ca408c48 title=""
time=2026-10-23T14:54:48.778+08:00 level=WARN msg="headless extraction failed, retrying with headed mode" url=https://ev.medium.com/welcome-to-medium-9e53ca408c48 err="exception \"Uncaught (in promise) Error: Cloudflare security challenge: Inaccessible social target: canonical identity missing or different\" (344:4148): Error: Cloudflare security challenge: Inaccessible social target: canonical identity missing or different\n    at st (<anonymous>:344:4148)\n    at zo (<anonymous>:353:2622)"
time=2026-10-23T14:54:51.512+08:00 level=ERROR msg="item failed" url=https://ev.medium.com/welcome-to-medium-9e53ca408c48 err="extract https://ev.medium.com/welcome-to-medium-9e53ca408c48: headless attempt: exception \"Uncaught (in promise) Error: Cloudflare security challenge: Inaccessible social target: canonical identity missing or different\" (344:4148): Error: Cloudflare security challenge: Inaccessible social target: canonical identity missing or different\n    at st (<anonymous>:344:4148)\n    at zo (<anonymous>:353:2622); headed retry: exception \"Uncaught (in promise) Error: Cloudflare security challenge: Inaccessible social target: canonical identity missing or different\" (344:4148): Error: Cloudflare security challenge: Inaccessible social target: canonical identity missing or different\n    at st (<anonymous>:344:4148)\n    at zo (<anonymous>:353:2622)"
CLI_EXIT=1
```

### Criterion-only tests

```text
# Subtest: A1/A3 native player refresh recovers empty inline captions with bounded first-party requests
ok 1 - A1/A3 native player refresh recovers empty inline captions with bounded first-party requests
# Subtest: A1/A3 refreshed invalid identity, language and API failures stay unsuccessful
ok 2 - A1/A3 refreshed invalid identity, language and API failures stay unsuccessful
# Subtest: A2/A3 explicit target previews fail while complete own bodies remain usable
ok 3 - A2/A3 explicit target previews fail while complete own bodies remain usable
# Subtest: A3 verified social content succeeds without network fallback
ok 4 - A3 verified social content succeeds without network fallback
# Subtest: A3 challenge identity survives actual adapter failures without classifying article mentions
ok 5 - A3 challenge identity survives actual adapter failures without classifying article mentions
# Subtest: A3 login, preview and unsupported video fail
ok 6 - A3 login, preview and unsupported video fail
# tests 6
# pass 6
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
--- PASS: TestProcessItemWithFallback (0.02s)
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
=== RUN   TestExtractionRejectsObservedWalls
--- PASS: TestExtractionRejectsObservedWalls (0.00s)
PASS
ok  	github.com/kouko/reading-list-summarize-scraper/internal/pipeline	0.404s
```

Dispatch profile inherited overrides null, unobservable-main-profile; resolver initial completed redispatches 1 as supplied by the primary. The repeated task packet did not cause additional worktrees, product changes, identity/model switches or duplicated test runs; this completed one bounded W5 measurement. No agents spawned. Commit classification routine acceptance verification; exact final message privacy scan required and no ambiguous private-party text requires a semantic judge. Tester edits are report/evidence only.

