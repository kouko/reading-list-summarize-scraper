# Independent acceptance evidence

Tried on 2026-10-31 (Asia/Taipei) by a fresh acceptance tester, starting from a
clean detached worktree at `a0e708edd112a1e6b553971eb8a6ca24947b31f4` and advancing
to final input `55cb5894ab8cb325318245c2a4972eb7a331b49a`.

## Setup, version and isolation

- `git worktree add --detach /private/tmp/rlss-accept-a0e708e a0e708e` succeeded.
- Followed README source setup: `npm ci`, `npm run build:defuddle`, and
  `go build -o /private/tmp/rlss-accept-bin ./cmd/rlss/` all exited 0.
  npm installed 26 packages, audited 27, and found 0 vulnerabilities.
- Advanced with `git merge --ff-only 55cb589`, then repeated all three setup
  commands successfully. The intervening diff only renames the unchanged earned
  identity probe, updates its two runners, and adjusts plan references: no product
  code, browser bundle, acceptance sentence or source-comparison behavior changed.
  Thus fresh live comparisons at a0e708e remain valid at 55cb589.
- Final focused identity runner:
  `node --test --test-reporter=junit test_facebook_conflictingPermalink_rejected.mjs`:
  exit 0; suite name `test_facebook_conflictingPermalink_rejected`, failures 0,
  overall pass 3/fail 0. Both conflicting numeric permalink and author cases passed.
- Temporary root: `/private/tmp/rlss-accept-fresh`. Public samples used its isolated
  `profile`; X used the previously authorized isolated
  `/private/tmp/rlss-thread-pagination-authclone-20261029/profile`, never a live profile.
  Source and CLI browser sessions ran sequentially, with separate fresh navigations
  on the same respective profile. Source capture waited six seconds after loading.
- Actual command for every live sample:
  `/private/tmp/rlss-accept-bin url URL --no-summary --config TEMP_CONFIG`.
  Safari/Chrome list reading, copy-to and LLM use were disabled. Config, executable,
  profiles and output were temporary. Production files and processes were untouched.
- Source JSON and full article Markdown remain temporary and uncommitted. The
  capture helper and comparison programs were inspected before reuse; comparisons
  consumed only this tester's fresh captures, never implementer results.

## 1. Complete specified Facebook post

- Source URLs: original slugged reading-list URL (loaded locally from the recorded
  original URL without republishing its long slug) and
  https://www.facebook.com/fatality1117/posts/10234719952645505/ .
- Both CLI runs exited 0 and reported Total 1, Success 1, Failed 0.
- Independently selected the public native message for post 10234719952645505,
  then corroborated it against exactly one owned rendered message in a separate
  fresh DOM reload for each URL. Replaced emoji images with their alt glyphs
  before comparing the native rendered message; both comparisons matched.
- Both outputs match all 148 authored lines in order, not merely length:
  4,400 Unicode characters; all six hashtags in the original sequence;
  glyphs `◾` 16/16, `▪` 2/2, `■` 8/8 (source/output). Whitespace normalization
  and Markdown unescaping were explicit; nonempty line order was checked separately.
- Each saved exactly one target-owned photo. Native attachment data and the owned
  DOM photo link agree on photo 10234719952085491; the saved image's asset path
  matches that owned DOM image. Avatars and comment images were not saved.
- The message has six hashtag links and no additional non-hashtag prose links.
  No comments, other post bodies, login controls or extra prose appear: the entire
  saved prose, after Markdown decoding, equals the independently owned message.
- Captured evidence: source-01/source-02, facebook-dom, 01/02-cli.log,
  comparison.json, facebook-dom-check.json and facebook-extra.json under temp root.

## 2. Complete Felix article

- Source: https://felixrieseberg.com/how-to-be-more-ambitious-a-tutorial-for-humans-with-models/ .
- Fresh source identified the visible article-owned gh-content body, independently
  collected rendered paragraphs, leaf list items, blockquotes and headings, and
  recorded all its links and images. Hidden PART labels were excluded by visibility.
- CLI exit 0; Total 1, Success 1, Failed 0. All 43 visible source units occur in
  the same order in saved Markdown, including the blockquote. Four heading strings
  match in order: Altitude, Depth, Range, We have been here before.
- All eight visible body link destinations and the one body image survive.
  Navigation/player controls, subscription interface and hidden PART labels are absent.
- Captured evidence: source-03/01-source.json, 03-cli.log and comparison.json.
  Historical counts were not used: 43 is the newly measured visible-unit count.

## 3. Incorrect or incomplete content is rejected; existing notes survive

- `node --test --test-name-pattern='facebook-|dated-body-' extraction.test.mjs`:
  4 passed/0 failed. Positive ownership and dated-body cases passed alongside
  wrong/ambiguous owner, conflicting source, truncation, missing media, lost units,
  login/challenge and dated sidebar negatives. Defuddle printed synthetic URL
  diagnostics; the assertions passed.
- Final identity probe also passed (setup section). Focused real-browser fixture
  `TestBrowserExtractRequestedArticle` rejected redirected targets and retained
  loading/target/body/incomplete category causes; its valid numeric, slug and
  confirmed alias cases succeeded, including matching source ports.
- Focused pipeline tests `TestProcessItemSocialCache`, `TestProcessItemFocalScope`,
  `TestProcessBatchFailedPaginationPreservesFiles`, `TestExtractionRejectsObservedWalls`
  passed: failed extraction never summarized or overwrote seeded body/summary files.
- These are focused checks, not a claim that this tester ran the complete package
  suite. `go test ./...` is the automated test suite that runs before acceptance
  and blocks it on failure; finalize-review owns that run.

## 4. Distinguishable failure reasons

- `TestProcessBatchScopedFailureCategory` passed all four cases:
  `Article load failed`, `Article target unverified`, `Article body unavailable`,
  `Article incomplete`. Each retained both attempt messages and the underlying
  wrapped cause, with Failed 1/Success 0/Partial 0.
- Real-browser scoped-category fixture passed, including challenge classification.

## 5. Existing X, ordinary, pagination and retry behavior

All 11 X URLs below were freshly reloaded for source capture, then independently
processed through the CLI. Every CLI exited 0 with Success 1/Failed 0. Plain post
source text came from target-ID-bound public native note data; rich article text
came from that target's owned rendered rich body plus its title. Other embedded
posts were excluded only after checking their distinct status IDs.

| Source URL | Authored text/order | Owned body headings / images | Plain authored line breaks |
|---|---|---|---|
| https://x.com/0xluffy_eth/status/2114705266759569624 | equal | plain post | 20/20 |
| https://x.com/silvanrec/status/2090280285623464099 | equal | 10 / 1, matched | — |
| https://x.com/thorstenball/status/2114386050638336082 | equal | 0 / 0, matched | — |
| https://x.com/yihui_indie/status/2115251581998563837 | equal | 4 / 8, matched | — |
| https://x.com/romandevz/status/2114655026190618718 | equal | plain post | 72/72 |
| https://x.com/kenichiota0711/status/2113959093220335871 | equal | 7 / 7, matched | — |
| https://x.com/AdrianPunk115/status/2114648961399349608 | equal | 9 / 9, matched | — |
| https://x.com/addyosmani/status/2115688494865465370 | equal | 0 / 3, matched | — |
| https://x.com/Hitomiiim/status/2114704821852991771 | equivalent URL serialization, see below | plain post | 13/13 |
| https://x.com/katelyn_lesse/status/2114542280535318557 | equal | 7 / 0, matched | — |
| https://x.com/shota7180/status/2113460983100530808 | equal | 10 / 24, matched | — |

- Compared full normalized text sequences, punctuation, hashtags and glyphs, not
  just counts; all eight rich-body heading-label sequences and every visible owned
  prose link destination matched. Titles were checked in addition to body headings.
  All 52 owned rich-body image/thumbnail references matched by stable asset identity.
- Kenichi's source contains 13 image references including three other-post cards;
  removing those verified cards leaves the same seven owned images and full owned
  prose saved by the CLI. Other post text and six card images were excluded.
- Exact caveat: Hitomiiim expands a native reference `http://amitaro.net` to
  `http://amitaro.net/`; the sole normalized-text edit is that slash. This is
  equivalent root-URL serialization, not literal 11/11 text equality.
- `https://21st.dev/` and `https://claude.ai/code` are serialized as protocol-relative
  Markdown targets; BizPeel root links omit the trailing slash. Resolve `//` against
  the HTTPS source before comparing; no destination is missing. These forms are
  disclosed rather than silently counted as literal equality.
- Existing rich X conversion emits some empty heading markers followed by the
  original heading text on a separate line. Heading labels and order match; this
  is a pre-existing formatting limitation. The baseline diff confirms bodyMarkdown,
  modernXBody and social routines are unchanged in this change. No repair was made.
- Focused final-revision browser tests passed:
  `go test ./internal/extract -run 'TestBrowserExtractRequestedArticle|TestBrowserExtractPaginated|TestPoolHeadedRecoveryLifecycle|TestPoolExplicitHeaded|TestPoolRecoveryPreservesPaginationProfile|TestPoolRecoveryBrowserFailure|TestMatchDomainRules|TestMatchPaginationRule' -count=1 -v`.
  Result: PASS, 7.087s. Real-browser fixtures exercised ordinary and complete
  paginated extraction, later-page failure, selector handling and navigation timing.
  Pool/domain checks retained profile, explicit headed selection and retry behavior.
- Focused pipeline command:
  `go test ./internal/pipeline -run 'TestProcessBatchScopedFailureCategory|TestProcessItemWithFallback|TestProcessBatchFailedPaginationPreservesFiles|TestExtractionRejectsObservedWalls|TestProcessItemFocalScope|TestProcessItemSocialCache|TestClassifyResult|TestStats_Partial' -count=1 -v`.
  Result: PASS, 0.689s. Blocked retry disabled/enabled, ordinary load/empty retry,
  explicit headed no-repeat and pagination recovery all passed.
- Captured evidence: source-04 through source-14, 04 through 14-cli.log,
  x-comparison.json, x-extra.json and browser-tests.log under temp root.

## 6. Source-backed per-sample verification

- Rows 1 and 2 record fresh source-vs-saved comparisons for every named surface;
  Facebook compares native message, independently rendered owned message and saved
  body; Felix compares 43 ordered visible units, heading strings, actual reference
  targets and image URL. No body was inferred complete from a success message,
  nonempty output, matching count, product hash or historical report.
- Every requested live sample was available in this run; no unavailable sample was
  replaced with a cached result. Source text and profile state were not committed.

## 7. Successful extraction flow

- Actual three scoped CLI runs: exit 0; `Total: 1`, `Success: 1`, `Failed: 0`.
  Their saved bodies were checked in rows 1–2, rather than trusting this counter.

## 8. Summary enabled flow

- `TestProcessItemSocialCache` passed scoped fresh Facebook/Felix variants: recording
  summary received the fresh verified extraction, and forced summary failure returned
  Partial with content preserved. `TestProcessItemFocalScope` passed generated and
  embedded summary success. Partial classification checks passed independently.
- This checks the integration contract with isolated summary fixtures; no production
  LLM account or real provider availability was exercised.

## 9. In-progress and retry flow

- Actual empty ordinary-page CLI showed processing, failed headless attempt and one
  headed retry. Focused pipeline/pool tests checked selected profile, successful
  recovery and no second retry for explicitly headed invocation.

## 10. Verification-page flow

- Focused adapter/browser/pipeline negatives rejected recognized verification-only
  bodies. `scoped challenge disabled`, `block disabled challenge`, and recovery cases
  confirmed headed_on_block false suppresses blocked retries while enabled retries
  remain available. Verification text was never accepted as authored content.

## 11. Loading-failure flow

- Focused browser scoped error/cause case and pipeline retry-failure cases passed;
  failed navigation retained its loading category and cause. Batch statistics
  reported Failed 1; preservation fixtures retained existing notes.
- Actual single-command failure exit handling was independently exercised with
  an empty-page fixture (row 13), not claimed to reproduce a live-site outage.

## 12. Target-unavailable/wrong flow

- Real-browser redirected-target fixture and final earned contradictory-permalink
  probe passed. Wrong target, conflicting author and wrong marker were rejected;
  valid numeric, slug and bound alias positives succeeded.

## 13. Body-unavailable flow

- Owned loopback fixture served an empty HTML body. Actual single-URL command
  exhausted its retry, reported Success 0/Failed 1 and exited 1, without new content.
  In-scope body-category propagation separately passed in rows 3–4.

## 14. Incomplete/conflicting flow

- Focused adapter negatives rejected disagreeing source messages, explicit previews,
  missing attachments and missing protected authored units. Pipeline preservation
  negatives confirmed old body and summary remain byte-for-byte intact.

## 15. Batch/watch item-failure flow

- Actual isolated CLI used an owned loopback RSS feed with an empty first article
  and a complete second article. Batch exited 0 (existing behavior), reporting
  Total 2/Success 1/Failed 1; the second item proceeded after the first failed.
- Watch ran two cycles against the same feed. Each cycle reported the failed item
  and continued; the following cycle attempted that failed URL again. SIGINT stopped
  only this tester's watch process while sleeping. Evidence: flow-batch.log and
  flow-watch.log; no user process was signalled.

## 16. Completed/pending notes flow

- Focused seeded-note cases passed: completed skip, forced refresh, fresh pending
  Facebook/Felix, failure preservation, ordinary cached resume and dry run.
  The actual single-URL live commands independently reprocessed their supplied URL.

## 17. No-work/dry-run flow

- Actual no-enabled-source process: exit 0, `no items to process`.
- Actual local-feed dry run: exit 0, listed both fixture items without extraction
  or note writes. Evidence: flow-empty.log and flow-dry.log. Output did not exist
  before the subsequent extraction runs; dry-run fixture checks also passed.

## 18. Development acceptance flow and decisions

- All six acceptance conditions and all twelve visible flows have a result above.
  Tests, fixtures, downloaded source and browsers were isolated from production.
- No important/fatal finding was dismissed by the orchestrator. No plan/intent
  skipped-by-instruction entry exists. Full suite belongs to finalize-review.
- Agent-decided behavior disclosed in the reader report: fail closed on uncertain
  ownership/completeness; narrow supported URL scope; recheck pending scoped caches
  while preserving completed notes; compare meaningful source sequence and URL
  identity while explicitly disclosing serialization differences. No new product
  choice or fix was made by the acceptance tester.
- First independent acceptance pass for this closing-review run. The final test-only
  revision was reviewed for carry-over validity and clean setup repeated; there was
  no intervening functional fix requiring another live-content pass.

---

## Re-run on 2026-10-31, at 6706a06 — independent acceptance, Round 2

### Fix, scope and clean setup

- Final input: `6706a060d4a0ec81fc112ff59386b6035d121ad4`. Earlier independent
  report/evidence landed as `b4a756e` (original detached acceptance commit 3a9bda8).
- Reviewed `git diff b4a756e..6706a06`: two plan updates, Facebook adapter changes,
  added cases within existing adapter tests, and rebuilt browser bundle. No X,
  dated article, browser wrapper, pipeline, command, configuration, batch or watch
  logic changed. The generated bundle change corresponds to the Facebook adapter.
- Round 1 review's fatal finding was fixed, not dismissed: another matching-post
  record without a complete message could explicitly contradict attachment extent
  or photo identity yet be ignored. The fix reconciles every supplied root/known
  nested story attachment record, rejects explicit conflicts, and retains positives
  when optional fields are simply absent. There are no dismissed important/fatal
  findings and no skipped-by-instruction steps.
- Created a NEW clean detached worktree:
  `git worktree add --detach /private/tmp/rlss-accept-round2-6706a06 6706a06`.
  Did not reuse the earlier acceptance worktree or its artifacts.
- README setup again: `npm ci` (26 packages, audit 27, vulnerabilities 0),
  `npm run build:defuddle` (809.1kb bundle),
  `go build -o /private/tmp/rlss-accept-round2-bin ./cmd/rlss/`; all exited 0.
- New config, logged-out profile, output and public-source captures are under
  `/private/tmp/rlss-accept-round2-fresh`. Safari/Chrome reading-list sources, copy-to
  and LLM calls disabled. Source capture and CLI execution were sequential separate
  browser sessions using that new profile, with fresh navigation for each.
  Public native/owned DOM capture waited six seconds; CLI wait remained three.
  Neither production state nor the earlier X clone was touched in this rerun.
- Full source/profiles remain temporary and uncommitted. Prior helpers were
  inspected, paths adapted, and comparisons rerun only against new captures.

### Focused commands actually executed

- `node --test --test-name-pattern='facebook-|dated-body-' extraction.test.mjs`:
  4 passed, 0 failed, exit 0 (211.616ms). Includes both supported URL forms and the
  positive duplicate with partial optional media fields. The existing negative
  aggregate now also executes 22 root/nested sparse-record attachment variants:
  explicit empty extent, conflicting media IDs/type/URL/image/story, invalid image
  dimensions, extra nodes and outstanding pagination. Existing wrong target,
  conflicting message, truncation, login-only, missing body and missing owned photo
  negatives remain covered. Dated-article unit-loss/ambiguity negatives also passed.
  Synthetic fixture URL diagnostics were emitted by Defuddle; assertions passed.
- `node --test --test-reporter=junit test_facebook_conflictingPermalink_rejected.mjs`:
  exit 0, overall pass 3/fail 0; named suite failures 0. Original identity probe
  remains unchanged; both conflicting numeric permalink/author cases passed.
- `go test ./internal/pipeline -run 'TestProcessBatchScopedFailureCategory|TestProcessItemWithFallback|TestProcessBatchFailedPaginationPreservesFiles|TestExtractionRejectsObservedWalls|TestProcessItemFocalScope|TestProcessItemSocialCache|TestClassifyResult|TestStats_Partial' -count=1 -v`:
  PASS, 0.671s. Fresh execution covered all four categories with both attempt causes,
  failure statistics, scoped/ordinary cache behavior, file preservation, summary
  input and Partial behavior, walls, headed selection and blocked/load/empty retry.
- After all live source sessions finished:
  `go test ./internal/extract -run 'TestBrowserExtractRequestedArticle|TestBrowserExtractSocialCompleteOnly' -count=1 -v`:
  PASS, 3.874s. Requested numeric/slug/confirmed-alias and ordinary positives,
  redirected-target negatives, scoped load/target/body/incomplete causes, challenge
  retention, pagination entry and complete-single-post boundaries passed.
- Package suite is NOT run by this tester. `go test ./...` remains the automated
  test suite that runs before acceptance and blocks it on failure, owned by
  finalize-review. Parent-reported suite results are not substituted for these runs.
- New logs: `/private/tmp/rlss-accept-round2-{adapter.log,identity.xml,pipeline.log,browser.log}`.

### Fresh actual-source comparisons

Each sample was freshly captured, then processed with
`/private/tmp/rlss-accept-round2-bin url URL --no-summary --config TEMP_CONFIG`.
All three exited 0 and reported Total 1 / Success 1 / Failed 0.

| Sample | Newly observed original vs saved body |
|---|---|
| Original slugged Facebook reading-list URL | Full normalized prose equal; all 148 nonempty/blank authored source lines checked in sequence, source message 4,400 Unicode characters; all six hashtags, ordered six hashtag destinations, 16 ◾, 2 ▪, 8 ■ retained; exactly one owned photo |
| https://www.facebook.com/fatality1117/posts/10234719952645505/ | Same independently checked full prose/line sequence, six hashtag destinations, every glyph and one owned photo; not inferred from the slugged result |
| https://felixrieseberg.com/how-to-be-more-ambitious-a-tutorial-for-humans-with-models/ | 43 visible owned units including blockquote in sequence; four heading labels in order; all eight body link destinations and one image; no hidden PART labels, navigation/player controls or subscription form |

- For EACH Facebook URL, native post message equals exactly one owned rendered
  message after emoji alt expansion; independently saved prose equals that native
  source. Line order is checked separately from full whitespace-normalized equality.
  Six hashtag-link paths match in order after URL decoding; there are no additional
  non-hashtag prose references. No comments or extra interface prose are present.
- Target native attachment/owned DOM photo identity is 10234719952085491 for both
  URLs; each saved image asset path matches the owned DOM image, excluding avatars
  and comments. Hashtags and glyph counts supplement full sequence checks rather
  than replacing them. Native records can omit optional fields; no contradiction
  was observed in either newly loaded public sample.
- Felix units were recollected from visible authored elements, including blockquote,
  paragraphs, leaf list items and headings. Hidden labels were excluded using actual
  rendered visibility. Every fresh unit/reference/image was compared to the new
  CLI output, not to historical counts or output.
- Captures/logs and comparison outputs are in new temp root:
  `source-01` through `source-03`, `01-cli.log` through `03-cli.log`, `comparison.json`,
  `facebook-dom-check.json`, `facebook-extra.json`, `linkcheck.json`.

### Every report row: full re-test or checked carry-over

- 1: re-tested — both Facebook URL forms across full ordered prose, paragraph/line
  sequence, native/owned identity, all hashtag targets/glyphs and owned image, with
  no extra prose. New positive CLI outputs and source comparisons above.
- 2: re-tested — full fresh Felix authored units, section strings/order, links,
  blockquote, image and excluded interface. Though its code did not change, row 6
  names both scoped sources, so Felix was freshly compared again.
- 3: re-tested — all wrong/ambiguous targets, contradictory/full-or-sparse native
  sources, preview/truncated text, missing body/photo, login/challenge and protected
  dated-unit negatives; real browser delivery plus seeded old body/summary preservation
  and no-summary-on-extraction-failure pipeline cases. New attachment variants passed.
- 4: re-tested — four named category cases retain initial/retry causes and Failed 1,
  Success 0; actual adapter conflict rejection and browser category/cause transport
  passed. New supplied-attachment contradiction paths reject rather than succeed.
- 5: carried over — checked the entire fix diff: only Facebook attachment evidence
  reconciliation changes product behavior. X parsing, general/dated extraction,
  pagination, browser-mode/profile rules and retry policy are unchanged. Prior fresh
  11-X comparisons and ordinary/pagination/retry tests remain valid, including their
  explicitly disclosed root-slash, protocol-relative-link and heading-format limits.
  No new X live run or broader regression claim is made in Round 2.
- 6: re-tested — all source surfaces for BOTH scoped samples were independently
  captured and compared again; neither count-only checks nor prior results stand in.
- 7: re-tested — numeric Facebook, original slug and Felix final CLI each exit 0,
  Success 1/Failed 0; saved bodies checked against fresh sources as above.
- 8: re-tested — full scoped fresh Facebook/Felix summary-input and summary-failure
  Partial cases plus generated/embedded success and file preservation. Isolated
  summary fixtures were used; production provider availability remains untested.
- 9: re-tested — fresh CLI processing messages; pipeline headless retry, retry
  success/failure, same-check rejection and explicit-headed no-repeat cases passed.
- 10: re-tested — adapter/browser recognized-wall negatives and pipeline blocked
  retry enabled/disabled cases passed; no verification body is accepted or saved.
- 11: re-tested — real-browser loading/category/cause fixture plus pipeline retry
  exhaustion, Failed classification and preserved seeded-file cases passed. The
  unchanged CLI error-exit boundary was also exercised by row 13 below.
- 12: re-tested — fresh requested target positives and real-browser wrong redirect,
  native conflicting identity/permalink/author negatives including earned probe.
- 13: re-tested — adapter missing-body/login-only negatives, scoped body-category
  propagation and actual single-URL owned empty-page command: exhausted retry,
  Failed 1/Success 0, exit 1, no new body. Reused the existing empty-page fixture
  scenario in the new environment, not a new probe or production page.
- 14: re-tested — all source-message and attachment conflict/truncation/extent
  negatives, including sparse duplicate records, plus Felix lost-unit/ambiguity and
  seeded old-content/summary preservation. Both successful scoped sources also
  pass, so the negatives do not replace testing complete bodies.
- 15: carried over — batch/watch loops, statistics and next-cycle retry routing are
  untouched by the diff. Prior actual two-item continuation and two-cycle watch
  run remain evidence, including its graceful stop of only the isolated process.
- 16: re-tested — completed skips, forced reprocess, pending scoped fresh extracts,
  failures preserving both existing files, ordinary cached resume and dry-run tests;
  actual scoped single-URL invocations freshly reprocess all three samples.
- 17: carried over — no-work/dry-run return before invoking the changed Facebook
  adapter; command/source-selection paths unchanged. Earlier real CLI results
  (exit 0, no extraction or new notes) remain valid; focused dry-run cases also passed.
- 18: re-tested — clean setup, all affected source surfaces and categories recorded;
  remaining rows checked against actual fix diff with explicit reasons. Fatal review
  finding history is preserved, none dismissed. No production writes or skipped steps.

The extra row-13 CLI result is captured in
`/private/tmp/rlss-accept-round2-fresh/flow-single-failure.log`.
All runs finished; this tester did not fix product/test code, run a full suite,
or leave a browser/watch process running.

# Prior implementer evidence appendix — not independent acceptance

The following pre-existing material is retained for provenance. Its results are
not substituted for any fresh independent observation above.

# Final embedded integration evidence

Date: 2026-10-31 (Asia/Taipei). Initial implementation input: `3f8baaa4e64f`;
Facebook identity follow-up input: `8c37cac`.
This records implementer evidence, not an independent acceptance verdict.

## Build and focused checks

- `npm run build:defuddle`: passed; regenerated the committed browser bundle.
- `env GOCACHE=/private/tmp/rlss-go-build go build -o /private/tmp/rlss-complete-articles-20261031 ./cmd/rlss`: exit 0 and executable produced. Go printed a denied optional module stat-cache write outside the workspace; no dependency or production binary was changed.
- `node --test --test-name-pattern='dated|facebook' extraction.test.mjs`: 4 passed, 0 failed. Defuddle printed fixture URL diagnostics; assertions passed.
- Memory `regenerate-index` and `validate`: passed.

## Isolation and method

All actual CLI runs used `url URL --no-summary --config TEMP_CONFIG`, temporary
output under `/private/tmp/rlss-w4-20261031/output`, disabled Safari/Chrome list
sources, watch and copy-to, and no LLM. Scoped public samples used a fresh logged-out
CLI profile. Independent source reloads used separate fresh headless and headed
profiles and public DOM/native fields only. Browser commands required sandbox
escalation after isolated Chrome startup failed inside the sandbox.

X regression runs used the previously authorized isolated session clone,
sequentially: final CLI first, independent source browser afterward. No live user
profile, production config, production binary or existing notes were modified.
Full source captures, signed media URLs and browser state remain uncommitted.

## Facebook

Both the original slugged reading-list URL and
`https://www.facebook.com/fatality1117/posts/10234719952645505/` saved successfully.
Independent headless and headed source reloads matched both saved bodies:

- Native authored text: 4,400 Unicode characters and 148 lines, equal after
  restoring the six hashtag Markdown link labels and normalizing whitespace.
- Six hashtags and all symbols retained: 16 `◾`, two `▪` (18 small bullet glyphs),
  plus eight `■`. These categories explain the earlier 18-glyph shorthand.
- One native target photo, identity `10234719952085491`, corresponds to the owned
  rendered photo and saved image path. Signed CDN parameters vary per reload;
  stable photo identity and asset path agree.
- Equality of the authored body excludes comments, reactions and login controls;
  source-owned interface images were excluded from the one-photo count.

## Felix

The Felix and X evidence below was collected against the initial implementation;
those extraction paths were unchanged by the Facebook identity repair. They were
not reloaded during this bounded follow-up; independent acceptance will reload
them against the final tree.

`https://felixrieseberg.com/how-to-be-more-ambitious-a-tutorial-for-humans-with-models/`
saved successfully. Fresh headless and headed source reloads retained all 42
paragraph/heading/list/caption selector units in order, plus the independently
captured blockquote (43 visible units total), four headings, eight authored
links and one original image URL. The source helper's 42-unit selector omits
blockquote, so it must not be reported as a 43-unit selector result.

Comparison removes Markdown emphasis/link syntax and layout whitespace, preserving
prose, punctuation, order and destinations. Heading comparison uses visible text;
CSS-hidden `PART` labels are excluded. The headed source helper also observed four
empty section-anchor controls; these are interface links, not missing authored
links. Generated image-alt captions and emphasis formatting are layout differences.

## X regression sources

All 11 original X URLs were re-exercised with the final CLI and independently
reloaded afterward; every CLI run saved successfully and every source was
available. Comparison results are in local `x-comparison.json`.

Ten saved bodies equal the native authored source after Markdown and whitespace
normalization. Target `2114704821852991771` has one explicit URL serialization
difference: `http://amitaro.net` becomes `http://amitaro.net/`; all other authored
text and all 13 line breaks match. The destination is equivalent, but this is
not counted as literal URL equality. A focused Node WHATWG `new URL` check found
equal serialized `href` and equal scheme, host, port, path, query and fragment for
only this pair; no broader URL normalization was applied. The other two plain posts retain all 20
and 72 authored line breaks respectively.

All eight rich articles retain matching owned text and 52 owned image/thumbnail
references. Target `2113959093220335871` excludes three independently identified
other-post cards and their six images; these are not counted as owned content.
Saved rich heading counts include the article title, explaining the additional
one heading compared with source body headings. No cached source report was used
as a substitute for a fresh comparison, and no unavailable source was marked pass.

## Growth and limits

Against confirmed-intent commit `fc19803`, tree `8c37cac` handwritten production changes
add 300 and remove 8 lines (net +292); tests add 232 and remove 1 (net +231);
regenerated bundle adds 149 and removes 143 (net +6 minified physical lines).
These counts exclude documentation and classify root `.test.mjs`, `_test.go`, and
browser extraction tests as tests. The production
estimate of 100–200 lines was exceeded, including request binding and failure
classification across browser and pipeline boundaries. Physical generated lines
are not a useful complexity measure and are reported separately.

`git diff --check` reports upstream template-literal trailing whitespace in the
generated bundle at lines 241 and 336. These are generated string contents and
were retained without manual bundle edits.

These checks cover the studied public Facebook post and Felix article, not every
site, Facebook URL shape, attachment type or full X thread. Cached earlier proofs
were not substituted for final CLI verification. Independent closing review and
full package verification remain the parent station's responsibility.

Local audit artifacts: `/private/tmp/rlss-w4-20261031/` contains command logs,
isolated configs/output, public-source captures and comparison scripts/results.

## Facebook identity follow-up

Repair `4b52a3f` rejects a matching native post ID when its explicit permalink
contradicts the requested post or author. Its regenerated bundle was included in
`8c37cac`, from which the final isolated CLI was rebuilt with
`env GOCACHE=/private/tmp/rlss-go-build go build -o /private/tmp/rlss-w4-final-20261031 ./cmd/rlss`
(exit 0; the same optional module stat-cache warning remained).

Both numeric and original slugged URLs were rerun with `--no-summary`, a fresh
temporary profile/config and temporary output. Each saved successfully. Fresh
independent logged-out headless and headed source reloads again matched both
saved authored bodies after restoring six hashtag labels and whitespace layout:
4,400 characters, 148 lines, six hashtags, 16 `◾`, two `▪`, eight `■`, one native
owned photo `10234719952085491` and the same stable saved image asset path.
No prior source capture was substituted for these reloads.

The earned identity probe went RED at `83df5fa`, then GREEN after `4b52a3f`, and
was graduated to root `facebook-identity.test.mjs` in `8c37cac`. Its original and
maintained file blobs are identical (`94602b5707815c79969c9d010fbf2704498ab653`).
`node --test facebook-identity.test.mjs` passed the parent and both contradictory
permalink subtests (3 tests total) during this follow-up.

Documentation defect class: ambiguous comparison wording and claims exceeding
observed behavior. Evidence and memory were corrected after searching those
artifacts plus the retained plan, spec and README; no other same-class wording
was found. The memory now captures expected units before cleanup and verifies
the resulting extraction before saving; it records the observed Felix empty
result without describing it as a nonempty result. No spec/hash was changed.

Follow-up audit artifacts: `/private/tmp/rlss-w4-final-20261031-evidence/` contains
fresh source captures, both CLI logs/output, `comparison.json` and
`media-check.json`. No full suite was run here; the parent station owns it.
