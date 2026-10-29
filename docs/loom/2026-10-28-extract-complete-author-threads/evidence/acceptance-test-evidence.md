# Complete author threads — acceptance test evidence

Tried on 2026-10-28 in a fresh detached worktree at 26578a1aa625fc4718b0f0f8c89e226ba87a79d5.
Raw source HTML, full captions and logs remain only in /private/tmp/rlss-acceptance-live/. Other temporary outputs: /private/tmp/rlss-acceptance-extra-output/ and /private/tmp/rlss-acceptance-batch-output/.

## Setup and independent method

- Fresh worktree: `git worktree add --detach /private/tmp/rlss-acceptance-fresh 26578a1aa625fc4718b0f0f8c89e226ba87a79d5`.
- README-supported quick build: `env GOCACHE=/private/tmp/rlss-acceptance-go-build go build -o /private/tmp/rlss-acceptance-bin ./cmd/rlss/`; exit 0. A restricted module-cache stat warning did not prevent the fresh build.
- README test dependencies: `npm ci --cache /private/tmp/rlss-acceptance-npm-cache`; restricted first attempt had npm exit-handler error. Authorized unrestricted retry succeeded: 26 project packages, zero vulnerabilities. No global/conda installation.
- Sandbox Chrome startup failed; authorized guest-only runs then used `CHROME_CONFIG_HOME=/private/tmp/rlss-acceptance-live/chrome-config` and temporary user-data directories. No account profile or credentials were copied.
- All 16 cases ran `/private/tmp/rlss-acceptance-bin url <URL> --config /private/tmp/rlss-acceptance-config.yaml --no-summary`, with default headless-first/headed retry, isolated 30-second extraction settings and three-second render wait.
- Each URL also had an independent fresh headed guest source visit. The observer saved original page.html BEFORE injection. Neither its extraction result nor feature code supplied expected membership or text.
- Independent temporary JSON parsing selected Threads root `data.media` plus `text_post_app_info.self_thread.posts.edges`, checked explicit post_ids and final `page_info.has_next_page=false/end_cursor=null`, and compared all source captions with output after removing whitespace. Exact saved-body hashes compare first/middle without normalization.
- Reader/direct-reply, quote and related-post records were separate. Entire expected-caption equality proves no additional text in positives. A recommendation repeating a root caption is not evidence of another included member; identical expected sequence appears only once.
- X original HTML supplies visible IDs/previews but does not prove exhaustive membership. Same-author links below are observations, including possible recommendations/resources, not an asserted complete sequence. No login restriction is inferred without an observed prompt.
- This is the first independent acceptance run, with no carried-over rows. No product fixes or production writes were made. No important/fatal finding was dismissed.

## 1. 已確認的三組 X 與五組 Threads 樣本均能交付完整文章；同一組樣本的首篇與中段網址均對應完整且順序一致的正文。

All 16 actual attempts completed: eight successful exits across four Threads groups, and eight failed exits across three X groups plus Threads AI colleague. Only four groups passed full independent caption/order comparison and exact first/middle saved-body equivalence.

| Group / entry | Input URL | Exit | Outcome |
|---|---|---:|---|
| X dashboard skills / first | https://x.com/SuguruKun_ai/status/2039372810053967953 | 1 | Failed 1; complete reply extent is unknown |
| X dashboard skills / middle | https://x.com/SuguruKun_ai/status/2039372826311098624 | 1 | Failed 1; complete reply extent is unknown |
| X Japan MCP tools / first | https://x.com/SuguruKun_ai/status/2038872366743502989 | 1 | Failed 1; complete reply extent is unknown |
| X Japan MCP tools / middle | https://x.com/SuguruKun_ai/status/2038872379066327076 | 1 | Failed 1; complete reply extent is unknown |
| X Google I/O / first | https://x.com/Khazix0918/status/2056894400320708671 | 1 | Failed 1; complete reply extent is unknown |
| X Google I/O / middle | https://x.com/Khazix0918/status/2056894667413958907 | 1 | Failed 1; ambiguous post identity |
| Threads Open Design / first | https://www.threads.com/@aiposthub/post/DYliXKSkRBP | 0 | Success 1; Failed 0; source comparison below |
| Threads Open Design / middle | https://www.threads.com/@aiposthub/post/DYliX_qkbFu | 0 | Success 1; Failed 0; source comparison below |
| Threads AI colleague / first | https://www.threads.com/@ci.fullstack/post/DW4Dy96FBAP | 1 | Failed 1; direct reply pagination incomplete |
| Threads AI colleague / middle | https://www.threads.com/@ci.fullstack/post/DW4D1bfFPUj | 1 | Failed 1; direct reply pagination incomplete |
| Threads pstack / first | https://www.threads.com/@ci.fullstack/post/De13wZUiRHk | 0 | Success 1; Failed 0; source comparison below |
| Threads pstack / middle | https://www.threads.com/@ci.fullstack/post/De138E9CYfv | 0 | Success 1; Failed 0; source comparison below |
| Threads Claude Code files / first | https://www.threads.com/@gakikun_claudecode/post/DeJwYehANl- | 0 | Success 1; Failed 0; source comparison below |
| Threads Claude Code files / middle | https://www.threads.com/@gakikun_claudecode/post/DeJwZRmAIRk | 0 | Success 1; Failed 0; source comparison below |
| Threads AI adoption gap / first | https://www.threads.com/@lcj.hankliu/post/DXB4kzEFLdn | 0 | Success 1; Failed 0; source comparison below |
| Threads AI adoption gap / middle | https://www.threads.com/@lcj.hankliu/post/DXB4lahlFpz | 0 | Success 1; Failed 0; source comparison below |

### Original first-party source observations

#### X dashboard skills / first
- Original HTML SHA-256: 5caac958c915e92cb9c14fe19041e459f53ae8925326d5c1473b9c3bd3669db6.
- Observed same-author IDs, display order: 2039372810053967953, 2039372826311098624, 2039372841913864418, 2039373227672453129, 2026502754617995428, 2039373561148981507, 2037187073611079704, 2039375498183430212. These do not prove membership/end.
- Article containers: 9; preview-bearing containers: 2. No saved article exists for complete content/exclusion/equivalence comparison.

#### X dashboard skills / middle
- Original HTML SHA-256: ab684ee14f89131fb06729ed59c88bb40e7a3ad0ee770beff79f173b15380228.
- Observed same-author IDs, display order: 2039372810053967953, 2039372826311098624, 2039372841913864418, 2039373227672453129, 2026502754617995428, 2039373561148981507, 2037187073611079704, 2039375498183430212. These do not prove membership/end.
- Article containers: 8; preview-bearing containers: 2. No saved article exists for complete content/exclusion/equivalence comparison.

#### X Japan MCP tools / first
- Original HTML SHA-256: 871b0d5f868a96675b1e0bbc48f31baebbcd7802824d97b55c150399e516e87c.
- Observed same-author IDs, display order: 2038872366743502989, 2038872379066327076, 2038915434263773485, 2038924176921362899, 2038596640144331239. These do not prove membership/end.
- Article containers: 6; preview-bearing containers: 2. No saved article exists for complete content/exclusion/equivalence comparison.

#### X Japan MCP tools / middle
- Original HTML SHA-256: 750a2e7e18fc2c40859e6e61e492a04521150c908cfa779ffc63d5f38f3e344a.
- Observed same-author IDs, display order: 2038872366743502989, 2038872379066327076, 2038915434263773485, 2039378557970067874, 2039372810053967953. These do not prove membership/end.
- Article containers: 6; preview-bearing containers: 1. No saved article exists for complete content/exclusion/equivalence comparison.

#### X Google I/O / first
- Original HTML SHA-256: 3b2248035d4a1fef94f30b3ccfd6e14ccb7e604a9d13c0545366f24bc2d0db46.
- Observed same-author IDs, display order: 2056894400320708671, 2056894667413958907, 2056894891205296229, 2056895055617724832. These do not prove membership/end.
- Article containers: 5; preview-bearing containers: 3. No saved article exists for complete content/exclusion/equivalence comparison.

#### X Google I/O / middle
- Original HTML SHA-256: ace5fa562d38669cf54323a1da9815a5317e30880f0080bf89f7b711004068c4.
- Observed same-author IDs, display order: 2056894400320708671, 2056894667413958907, 2056894891205296229. These do not prove membership/end.
- Article containers: 3; preview-bearing containers: 1. No saved article exists for complete content/exclusion/equivalence comparison.

#### Threads Open Design / first
- Original HTML SHA-256: 2aa8d818ba171ec15238d8d4150db8e5f95033fd675a27d2c0ed42624d99f857.
- Source self-thread connection: edges=3, has_next_page=false, end_cursor=null.

#### Threads Open Design / middle
- Original HTML SHA-256: d03f0833e361ea912920448870e4ba3770b37c311e09c7e7817c75ed84b899ba.
- Source self-thread connection: edges=2, has_next_page=false, end_cursor=null.

#### Threads AI colleague / first
- Original HTML SHA-256: 9f9139d453e740da6d23f809a09b0891c496df78c2ea85fd4115f6a09e4df7f1.
- Source self-thread connection: edges=1, has_next_page=false, end_cursor=null.
- Source shows root, one self-thread continuation and substantive direct author replies. Complete direct-reply extent and narrative membership remain unresolved. Both entries fail; no full article comparison is claimed.

#### Threads AI colleague / middle
- Original HTML SHA-256: 6cdfeae550995217c4103a3a0e10cf1bb6ad86bb2322a61cb5c6d27fb5eee9fd.
- Source self-thread connection: edges=0, has_next_page=false, end_cursor=null.
- Source shows root, one self-thread continuation and substantive direct author replies. Complete direct-reply extent and narrative membership remain unresolved. Both entries fail; no full article comparison is claimed.

#### Threads pstack / first
- Original HTML SHA-256: d69a70695fd2c5891a00413b84c88ecec9f54ccfd658e3e10757eb7b06de86c1.
- Source self-thread connection: edges=5, has_next_page=false, end_cursor=null.

#### Threads pstack / middle
- Original HTML SHA-256: ae2d5a237e2aa4e597da2d1fb6b06066bb0d622d11266e081a677db9d3f0c59c.
- Source self-thread connection: edges=4, has_next_page=false, end_cursor=null.

#### Threads Claude Code files / first
- Original HTML SHA-256: e18ebc55af04058f38c4202546e827a1fd6e1243a1885697e65a30740c44899d.
- Source self-thread connection: edges=4, has_next_page=false, end_cursor=null.

#### Threads Claude Code files / middle
- Original HTML SHA-256: c064dd9fd4791737b246f1bf6a44b502c9a551c0f227d237357bc6f0b2dfd66a.
- Source self-thread connection: edges=3, has_next_page=false, end_cursor=null.

#### Threads AI adoption gap / first
- Original HTML SHA-256: 699574668ab28a40186be308ed767912c030817847f1f605320966eae0ed3a4e.
- Source self-thread connection: edges=4, has_next_page=false, end_cursor=null.

#### Threads AI adoption gap / middle
- Original HTML SHA-256: d46bfed40bcd01a477e0b4daf1ee3c230ab5e41553bf1ffd8797ece8dab2ad80.
- Source self-thread connection: edges=3, has_next_page=false, end_cursor=null.

### Complete comparison — Threads Open Design

| Ordered source member | Caption characters | Caption SHA-256 |
|---|---:|---|
| DYliXKSkRBP | 104 | 29bf53c104af66d874cce999d4a53671077da04ee7319c507ea33d7536e38d44 |
| DYliX_qkbFu | 117 | 0210938d0c0359e799abf40bf34dffd1c8ecce4930f698f8f143896e4c15c449 |
| DYliYzlkWqZ | 150 | 6496799f5cda48821062cd8bc1268244f0a59e453add276bee3846d329758cf5 |
| DYliZnNEc_e | 162 | 3ddb23551d1a577ea373b7bacc50c9c38ef29f6ace8c58a3d978dda925254a9e |
- Root is data.media; subsequent ordered members are self_thread.posts.edges. Explicit source post_ids and final page information agree with the captured member sequence.
- Entire source-caption sequence equals saved body after whitespace removal: first=true; middle=true.
- Member positions: 0, 89, 185, 309. The entire normalized body is exactly one ordered copy of these captions.
- Exact saved-body SHA-256: first bf688df684e87428d6dfd431277d45c5e82e3f20db4f4ef87bf82d1387a2fef3; middle bf688df684e87428d6dfd431277d45c5e82e3f20db4f4ef87bf82d1387a2fef3.
- Member source hyperlinks in either saved body: 0; frontmatter retains only the supplied URL. Finding F2.

### Complete comparison — Threads pstack

| Ordered source member | Caption characters | Caption SHA-256 |
|---|---:|---|
| De13wZUiRHk | 474 | 46e4b4315e36937f276af0ebb050547eece3a8b3e89ca30a69ba95b3bcf5a471 |
| De138E9CYfv | 374 | b3e3df2c3bfda3668e43febb75e9738b990a49b6c17ffa3aca4126707c58d19c |
| De14EhCCTYx | 417 | cc0653af5f5a2f728f623c9d627b9e812bce066a18e4e51333c3114a3ccb35e7 |
| De14Nm8CQet | 305 | 05b4f73da4731cba6c76803fb1c91fa792c4c8be15644db5bf9b3302c8ff407e |
| De14XtcCbhF | 293 | 74ce4a9064b4ab6e39760703e1a0b44a0cf204a731a0ca559cdd9eda858adc61 |
| De14gsRCUBW | 267 | 9db7566a957b1487e6f8206c4d61493d76895829eb13c6eecc2b2027d771bca3 |
- Root is data.media; subsequent ordered members are self_thread.posts.edges. Explicit source post_ids and final page information agree with the captured member sequence.
- Entire source-caption sequence equals saved body after whitespace removal: first=true; middle=true.
- Member positions: 0, 407, 727, 1083, 1361, 1611. The entire normalized body is exactly one ordered copy of these captions.
- Exact saved-body SHA-256: first 1937eb1bd2d827a65063523811ed41bf2f9fc6c83eda5b37b81dd3be3a549ce0; middle 1937eb1bd2d827a65063523811ed41bf2f9fc6c83eda5b37b81dd3be3a549ce0.
- Member source hyperlinks in either saved body: 0; frontmatter retains only the supplied URL. Finding F2.

### Complete comparison — Threads Claude Code files

| Ordered source member | Caption characters | Caption SHA-256 |
|---|---:|---|
| DeJwYehANl- | 241 | 4b3ef3a43cb928251c9558f0bb65121eb7366d1630bec679cca2ffbd46d5046f |
| DeJwZRmAIRk | 108 | a018f68957cfffa7da0f0a4b7ded498b5524f9aa71de5a77298df4a53afa50ff |
| DeJwaRQAOqg | 219 | 4644ecd27066575a2afcc730895d221776003359b1482aed064f68049ca0ce5a |
| DeJwbaaAHrQ | 270 | 0a9d0bc5364d54208d8b11164817255ce376466c208da0002dbf187e4840dc6d |
| DeJwcQqAI_L | 190 | 812741cc48ec1bcb9fe235c2e1634fd7d4298ab2e69d0b2d1a6028f6e8565a00 |
- Root is data.media; subsequent ordered members are self_thread.posts.edges. Explicit source post_ids and final page information agree with the captured member sequence.
- Entire source-caption sequence equals saved body after whitespace removal: first=true; middle=true.
- Member positions: 0, 220, 327, 536, 793. The entire normalized body is exactly one ordered copy of these captions.
- Exact saved-body SHA-256: first 4e6f70dd8960f4a52ddf7d99795e4179129cb0719f07c67d29e52a36dd24b08c; middle 4e6f70dd8960f4a52ddf7d99795e4179129cb0719f07c67d29e52a36dd24b08c.
- Member source hyperlinks in either saved body: 0; frontmatter retains only the supplied URL. Finding F2.

### Complete comparison — Threads AI adoption gap

| Ordered source member | Caption characters | Caption SHA-256 |
|---|---:|---|
| DXB4kzEFLdn | 111 | d63059268f0f30034caf0ed85e32eb89af6358833f4a13f41e4c7da2e8e8f02d |
| DXB4lahlFpz | 112 | f7e66abcd9d5841b950f4ea2de5e206669fe1e2ae8e9b5ec08f489fe92d0b7aa |
| DXB4l4UFNpk | 103 | 1b3c7fb26f49737aac2c5f7f23dfed0a136bf46cf2bc6ba56a24e64725694f7e |
| DXB4mQelO62 | 107 | 366b6c8ca27a9df188abb5b54181c8e9e184ffb1327ae737e86ef30b4d90bb6c |
| DXB4mpVFFrq | 113 | ce5cb80f1b3d7b8fb26830bfb79b97c8624ca25f3263567819c426934842e5e3 |
- Root is data.media; subsequent ordered members are self_thread.posts.edges. Explicit source post_ids and final page information agree with the captured member sequence.
- Entire source-caption sequence equals saved body after whitespace removal: first=true; middle=true.
- Member positions: 0, 97, 196, 289, 386. The entire normalized body is exactly one ordered copy of these captions.
- Exact saved-body SHA-256: first 31e9e3e6793453cbffa57d6843d1dcf3db82e0f50ae0e46945cd2100ec719866; middle 31e9e3e6793453cbffa57d6843d1dcf3db82e0f50ae0e46945cd2100ec719866.
- Member source hyperlinks in either saved body: 0; frontmatter retains only the supplied URL. Finding F2.

## 2. 組合內容只包含同一篇文章的作者正文與續文，保留原有內容順序，同一篇貼文不重複收錄，不混入讀者留言、作者回覆讀者的對話、引用貼文或推薦文章。

Four full-caption comparisons above establish exact content/order/membership in positive groups. AI adoption gap's author-to-reader reply DXCOiu5FDu3 is outside the source self-thread and absent from output. Claude Code related posts DeqP13ijIWg, DeEQ-A-kkhg, Debw9CoEXqD and DeVU58uAvRV are absent. A different recommendation repeats the root text; exact complete-sequence equality rules out including a second copy.

Google source independently exposes both distinct seventh-section IDs 2056894891205296229 and 2056895055617724832. No saved article exists, so their full inclusion/order remains unverified. No failed group is awarded exclusion or completeness success.

Focused command: `node --test --test-name-pattern='thread A|A2 target bodies|A2/A3 explicit target|A2 wrong target|A3 verified social|A3 challenge|A3 login' extraction.test.mjs`.
Result: 10 tests passed, zero failed/skipped. Covers ordered distinct members, first/middle snapshots, reader branches, missing members, previews, mismatched targets, challenges/login. An Invalid URL diagnostic from an incomplete test page was emitted; TAP remained 10/10 passing.

Focused command: `env GOCACHE=/private/tmp/rlss-acceptance-go-build go test ./internal/extract -run 'TestCompleteSocialThread|TestBrowserExtractSocialCompleteOnly' -v`.
Relationship tests passed: contradictory reply identity, unknown boundaries, missing members, unrelated branches, cursor/cap, unknown reply extent. Browser criterion initially failed at sandbox Chrome startup only. Authorized repeat:
`env CHROME_CONFIG_HOME=/private/tmp/rlss-acceptance-live/chrome-config GOCACHE=/private/tmp/rlss-acceptance-go-build go test ./internal/extract -run 'TestBrowserExtractSocialCompleteOnly' -v`.
PASS (0.698s). These tests supplement actual live comparisons, not replace them.

## 3. 無法確認目標、續文關係或完整正文時，不把缺段、預覽、登入或驗證畫面交付為完整文章的成功結果；擷取失敗時保留既有正文與摘要。

All eight unsuccessful entries exit 1, Failed 1, and save no partial body. Separate actual explicit URL processing with opt-in temporary sentinel old content and summary failed with complete reply extent unknown; both hashes were unchanged:
- Old content SHA-256: 09d1d872e21d3566dd20b6b8e411631f8a7c14b63c741e073393c313b892fcfa.
- Old summary SHA-256: ee8555e944df7ea005a37d46131e27168eaf285b70cf5e6047708ae36c24bdf9.
- Raw log: /private/tmp/rlss-acceptance-preservation.log. Actual CLI used --no-summary; no external summary call occurred.

Focused command:
`env GOCACHE=/private/tmp/rlss-acceptance-go-build go test ./internal/pipeline -run 'Test(ProcessItemSocialCache|ProcessItemWithFallback|ProcessBatchFailedPaginationPreservesFiles|ExtractionRejectsObservedWalls|IsSkipped|Stats_Report|Stats_Partial)' -v`.
PASS (0.368s). Includes verified-social-resume, failed-refresh-preserves-files, completed-skip/force, dry-run, ordinary-cache-summary-partial, wall rejection, fallback retries and result counts. Recording summary receiver proves failed extraction makes zero summary calls. JavaScript tests in section 2 reject unavailable/wrong/preview/login targets.

The full package suite was not run by this witness. The automated test suite runs before the change is accepted and blocks acceptance on failure; finalize-review owns that fresh execution.

## 4. 既有單篇貼文、長文及一般文章仍正常擷取，原有瀏覽器重試與驗證頁設定效果維持正常。

Actual command: `/private/tmp/rlss-acceptance-bin url https://go.dev/blog/strings --config /private/tmp/rlss-acceptance-extra.yaml --no-summary`.
Exit 0, Success 1, Failed 0; temporary body saved. An initial attempt shared a concurrent temporary profile and hit SingletonLock; repeated with a separate fresh temporary profile and succeeded. No production profile was involved.

Focused tests above pass headless-first, explicit-headed single attempt, ordinary navigation/empty retry, verification on/off and failed retry rejection. Fixture-level single-post target extraction passed. No positive live independent standalone social post or social long article was verified; criterion remains partly.

## 5. First-post success
All eight first inputs ran. Four complete source-matched Threads bodies succeeded, four failed. All positive bodies lack per-member source links, so the full UI flow is partly.

## 6. Middle-post success
All eight middle inputs ran. Four positive saved bodies exactly match first-post body hashes. Eight saved files preserve distinct supplied URLs and URL-derived output identities. No other URL's note was merged/deleted/renamed. Four groups remain unresolved.

## 7. Correct scope and repetition
Walked positive complete comparisons and failed groups as section 2. Expected full-body equality proves no extra or repeated source text in positives. The two distinct similar X seventh-section posts remain unverified in delivered output.

## 8. In progress and ordinary retry
Actual logs show processing item index=1 total=1 and headless extraction failed, retrying with headed mode. Focused fallback tests in section 3 pass default, explicit-headed and ordinary-failure behavior.

## 9. Verification restriction
Focused fallback cases specialized challenge disabled, challenge, block disabled challenge, block disabled navigation and block disabled empty passed. Recognized blocked pages do not call headed mode with restriction disabled; enabled challenges may retry; load/empty failures retry either way.

## 10. Missing target or body
Focused target mismatch, preview and login rejection pass. Actual failed entries have nonzero exit; no partial article is saved. Section 3 establishes unchanged sentinel content and summary.

## 11. Incomplete or ambiguous thread
Actual failures report unknown reply extent, direct reply pagination incomplete, or ambiguous post identity; Failed 1 and exit 1. Complete-only browser and relationship tests pass. No incomplete source list is reclassified as success.

## 12. Old content without summary
Verified-social-resume test re-extracts instead of immediately summarizing old social text. Failed refresh preserves files and makes zero summary calls. Ordinary content-only cache retains prior resume behavior. Focused tests passed.

## 13. Summary enabled
The focused recording receiver receives newly verified content; old social cache cannot bypass live extraction; extraction failure makes zero calls; summary failure remains Partial. No real external provider was invoked and no live multi-member article through an external summary provider was verified. This row is partly.

## 14. Completed notes and explicit reprocessing
Completed-skip and completed-force focused tests pass. Actual explicit reprocessing of temporary existing body/summary revalidated the live URL and preserved both on failure (section 3). No bulk migration or automatic backfill occurred.

## 15. Batch and watch
Ran /private/tmp/rlss-acceptance-batch.py: a temporary local RSS feed holds failed dashboard URL first, https://go.dev/blog/strings second. Invoked process --watch --no-summary with isolated output/profile.
Observed first-item failure, then second-item success:
- watch: cycle complete; Total 2, Success 1, Skipped 0, Partial 0, Failed 1.
- watch: sleeping minutes=1.
- watch: signal during sleep, exiting.
Exit 0 after interrupting only the newly created acceptance process. Raw log /private/tmp/rlss-acceptance-batch-watch.log.
Later retry cycles were not run; Skipped/Partial reporting is separately settled by focused tests. UI row remains partly.

## 16. Empty or dry run
`/private/tmp/rlss-acceptance-bin process --config /private/tmp/rlss-acceptance-extra.yaml --no-summary`: exit 0, no items to process.
`/private/tmp/rlss-acceptance-bin url https://example.com --config /private/tmp/rlss-acceptance-extra.yaml --dry-run --no-summary`: exit 0; Total 1, Success 0, Skipped 1, Partial 0, Failed 0. No extraction or new body/summary file.

## 17. Existing standalone content
Actual ordinary article succeeds (section 4). Fixture single-post target extraction passes. No independent live social standalone/long-article positive claim is made. Known incomplete social threads fail rather than fall back to single-post success.

## 18. Isolated verification
Fresh worktree, executable, configurations, browser user-data, Chrome config root and outputs are temporary. All 16 entries actually ran and have independent headed original-page captures. Permanent evidence records only prescribed public sample URLs/IDs, fields, counts, hashes and results; no raw HTML/full captions, account state, personal reading-list inventory or secret.

## Findings

### Finding F1 — important — eight required entry points fail
Three X groups and Threads AI colleague fail from both first and middle URLs. Acceptance 1 fails even though failure is safer than returning incomplete content. X complete extent/identity and Threads direct-reply pagination remain unresolved.
Fix: establish complete permitted first-party membership and full captions, then repeat both entries and complete source/order/exclusion checks for every affected group. If access cannot establish completeness, retain failed acceptance rather than weaken the contract or claim completion.

### Finding F2 — important — member source links absent
All eight positive output bodies have zero source hyperlinks; frontmatter only retains the supplied URL. This violates REQ-2 and First-post success's promised source links.
Fix: retain a permalink per verified member without expanding quotes/recommendations; rerun positive pairs for source equality, order/exclusions and separate URL storage identity.

## Verification limitations and review handoff
Live external summary delivery, a later watch retry cycle, and live independent standalone social/long-article positives were not verified. Related mechanical tests pass where stated, and report rows 4, 13, 15, 17 disclose these limits.
No package-tests or finalize-review skip was instructed. Build's reported full-suite PASS is context, not independent acceptance evidence. Finalize-review owns the fresh blocking full suite once otherwise ready. No important finding was dismissed, and this witness made no product fix.

## Re-run on 2026-10-28, at f67fa5b5573608daa93fc33e43f3f96d95173569

This dated section supersedes the baseline verdicts above. Fix range reviewed: b2b4555..f67fa5b. It changes saved member source links and X snapshot direct-reply extent evidence, so every report row was re-tested, including all sixteen actual first/middle entry points. No row is carried over and no important finding is dismissed. No product code was edited by this witness.

### Clean setup and commands

- New detached worktree: `git worktree add --detach /private/tmp/rlss-author-acceptance-rerun f67fa5b5573608daa93fc33e43f3f96d95173569`.
- README quick build: `env GOCACHE=/private/tmp/rlss-acceptance-rerun-go go build -o /private/tmp/rlss-acceptance-rerun-bin ./cmd/rlss/`; exit 0.
- README dependencies: `npm ci --cache /private/tmp/rlss-acceptance-rerun-npm`; exit 0, 26 packages, zero vulnerabilities. Both operations were authorized isolated setup; no global installation.
- Actual sixteen-entry driver: `python3 /private/tmp/rlss-rerun.py`. Each entry runs `/private/tmp/rlss-acceptance-rerun-bin url <confirmed URL> --config <case/config.yaml> --no-summary`, default headless-first/headed retry, 30-second attempt settings, three-second render wait, its own guest directory/output. Chrome config root: `/private/tmp/rlss-acceptance-rerun-live/chrome-config`.
- Raw actual logs and output: `/private/tmp/rlss-acceptance-rerun-live/{01..08}-{first,middle}/`. Cases use the same public URLs recorded in baseline section 1 and the confirmed spec.
- Four positive groups received new independent headed guest original-page captures, saved before injection under each first-entry `source-headed/page.html`. Oracle independently parses original JSON media root and explicit ordered self-thread edges/post_ids; source self-thread page info is terminal in all four. Product snapshots/extraction results do not supply expected captions or membership.
- Full actual comparison: `python3 /private/tmp/rlss-rerun-compare.py`; compare every original caption in order against saved Markdown after removing only whitespace and appended `[Source](...)` markers. Exact body hashes independently compare first/middle. Check one permalink per source member, exact ordered source codes, no extra body text, and distinct supplied URL frontmatter/storage.

### 1. Complete articles from both entry points — re-tested

| Confirmed group | First exit | Middle exit | Current outcome |
|---|---:|---:|---|
| X dashboard skills | 1 | 1 | complete reply extent is unknown; no output body |
| X Japan MCP tools | 1 | 1 | complete reply extent is unknown; no output body |
| X Google I/O | 1 | 1 | complete reply extent is unknown; no output body |
| Threads Open Design | 0 | 0 | complete caption equality, source links and exact pair identity pass |
| Threads AI colleague | 1 | 1 | direct reply pagination incomplete; no output body |
| Threads pstack | 0 | 0 | complete caption equality, source links and exact pair identity pass |
| Threads Claude Code files | 0 | 0 | complete caption equality, source links and exact pair identity pass |
| Threads AI adoption gap | 0 | 0 | complete caption equality, source links and exact pair identity pass |

Eight exits succeeded and eight failed. F1 remains important and open; returning honest failure does not satisfy Acceptance 1. Earlier Google middle identity failure did not recur, but full extraction still fails.

| Positive group | Source members / links per output | Exact first and middle body SHA-256 | New original HTML SHA-256 |
|---|---:|---|---|
| Open Design | 4 / 4 | aab776546687aad3d07ebe4f901dfd28a444d1f40a62fed1b1eabdbe63d1687b | 70620d28a0afdb5631e98cdf5cce0cb854a7e96ef342865d102ce28273f75dbe |
| pstack | 6 / 6 | 840ae6fdfe38d8c1fd18086d24f954b8f47e668c51998183a4933b460b48e46e | 61e67acdfcc780a75035e1b88027ed64d353e3467cab8a982c3737c6f0938864 |
| Claude Code files | 5 / 5 | 8f46290a6b74d9df1c2aa5d5cc11d2067a47084ebed0763a1e23c632ed182085 | 562a98f3d923a86b13af06228c60261f231a2f5d6eb73b578f6c62e268a4e873 |
| AI adoption gap | 5 / 5 | 8a37778517ffb465224a197856ea3126b0fd68f833bea1a976f153b870ba555e | 692eb552c4e4a394d9a52b94e724d7687f5b9de4af30827fe4d3b42871e28168 |

All new source caption hashes and ordered member codes equal the baseline per-member tables. This was recomputed from new original HTML, not assumed unchanged. For every positive output, normalized complete-caption equality is true; source link codes exactly match every ordered member once. Eight separate output identities retain the input URLs. Source links use verified threads.com aliases rather than requiring www. F2 is closed for all eight positive entry points; no positive live X claim is made.

### 2. Membership, order and exclusions — re-tested

The four entire-body comparisons above contain precisely one ordered copy of every author caption after removing only appended provenance links. Thus reader replies, reader-directed author conversations, recommendations and duplicate root recommendations add no text. Author references are unchanged by exact equality. Both distinct Google seventh-section IDs remain unverified in delivered content because both Google entry points still fail.

Focused JavaScript command: `node --test --test-name-pattern='thread A|A2 target bodies|A2/A3 explicit target|A2 wrong target|A3 verified social|A3 challenge|A3 login' extraction.test.mjs`; PASS, 10 tests, zero failures/skips. Covers current X exact-count producer, duplicate observations counted once, reader child counting, unknown/invalid/mismatched/conflicting counts, terminal zero count, plus source ordering, middle ancestry, reader exclusions, previews, target mismatch and walls. The existing incomplete-fixture Invalid URL diagnostic appeared; TAP is 10/10 PASS.

`node --test extraction-boundaries.test.mjs`: PASS, 2/2; includes quoted permalink rejection and transcript identity boundary.

`env CHROME_CONFIG_HOME=/private/tmp/rlss-acceptance-rerun-tests-chrome GOCACHE=/private/tmp/rlss-acceptance-rerun-go go test ./internal/extract -run 'TestCompleteSocialThread|TestBrowserExtractSocialCompleteOnly' -v`: PASS, package 1.268s. Covers full browser assembly/source links, contradictory identity, unknown beginning/extent, local reply sequences, late sibling, missing member, unrelated branch, loop/cap, and incomplete-only rejection.

### 3. Failure honesty and file preservation — re-tested

All eight failed actual entries save no new body/summary. Separate actual explicit dashboard reprocessing (`python3 /private/tmp/rlss-rerun-extra.py`) uses temporary old content and summary. Exit 1, complete reply extent unknown; exact before/after hashes unchanged:
- Content: c6f945cd02c4767cefdc70f7f17b8afce8b7ad1a2edb6952f1de979e612695d4.
- Summary: 95d622c7618d10e594b100c1d8bca7bd77e820ff3f4ab13d1fd0226b3cebb609.
- Evidence: `/private/tmp/rlss-acceptance-rerun-extra/preserve/result.json` and `cli.log`.

Focused pipeline command: `env GOCACHE=/private/tmp/rlss-acceptance-rerun-go go test ./internal/pipeline -run 'Test(ProcessItemSocialCache|ProcessItemWithFallback|ProcessBatchFailedPaginationPreservesFiles|ExtractionRejectsObservedWalls|IsSkipped|Stats_Report|Stats_Partial)' -v`; PASS, package 1.296s. Includes zero summary calls after failed refresh, old body preservation, completion/force/dry-run, cache identity and existing ordinary Partial behavior, walls and retry settings.

The witness did not run the complete package suite. The automated test suite runs before acceptance and blocks it on failure; `go test ./...` belongs to finalize-review when otherwise eligible. No package-tests/finalize-review skip was instructed; Build's earlier full-suite result is context, not this witness's result.

### 4. Existing standalone, long and ordinary content, retry settings — re-tested

Fresh actual ordinary article `https://go.dev/blog/strings` exits 0, Success 1, Failed 0. Body file SHA-256: 29dbfcfbfca50014f6bdf9e5be4c845ad5b93457e4f29daf46a4d500558d2887. Raw evidence `/private/tmp/rlss-acceptance-rerun-extra/ordinary/`.

Tried previous user URL `https://x.com/Sukiea1008/status/2112084040362897441` in another fresh guest: exit 1, headed result article end is unknown, no body. This does not establish a standalone source or a successful standalone/long-article regression. No live social standalone/long positive is claimed. Focused tests above verify existing single-target fixture behavior and all retry settings. Acceptance remains partly.

### 5. First-post success — re-tested

All eight first URLs actually run. Four groups fully source-match with 4/6/5/5 ordered source links; four fail. No missing-link finding remains in successful articles; full flow remains partly because four required first inputs fail.

### 6. Middle-post success — re-tested

All eight middle URLs actually run. Four successful pair body hashes match exactly and retain distinct supplied URL frontmatter/storage; four fail. No other note was merged, renamed or deleted. Whole flow remains partly.

### 7. Correct scope and repetition — re-tested

Entire-caption equality and exact ordered source link codes establish correct scope for all eight successful outputs. Focused exclusion/contradictory relationship tests pass. Distinct similar Google sections cannot be certified in saved output because no Google article is produced; partly.

### 8. Progress and ordinary retry — re-tested

Actual logs show item index/total and headless failure followed by headed retry for every X sample. Focused default fallback, explicit-headed single attempt, navigation/empty retry tests pass.

### 9. Verification restriction — re-tested

Focused pipeline cases pass challenge disabled/enabled, specialized challenge disabled, navigation and empty bodies with headed_on_block false. The false setting blocks recognized verification retries, while unrelated navigation/empty failures retain retry.

### 10. Missing target/body — re-tested

Current JavaScript target, preview, quoted-permalink and login/challenge cases reject; actual failed commands exit nonzero without new bodies. Actual failure preservation in section 3 and boundary receiver tests prove no replacement/summary from rejected bodies.

### 11. Incomplete/ambiguous thread — re-tested

Eight actual confirmed entries fail unknown extent or incomplete direct pagination without partial output. Contradictory identity, missing members, preview, unknown beginning/end, loop/cap and late sibling tests pass. Temporary old body/summary remain unchanged.

### 12. Old body without summary — re-tested

Focused verified-social-resume revalidates before summarizing, failed-refresh preserves files/makes zero summary calls, ordinary-cache-summary-partial retains ordinary behavior. PASS. No external provider used.

### 13. Summary enabled — re-tested

Focused recording receiver verifies newly validated body input, zero summary calls after extraction failure and existing Partial behavior after summary failure. No actual external summary provider was invoked; no live complete multi-member external-summary result was verified. Partly remains honest.

### 14. Completed notes/explicit reprocessing — re-tested

Focused completed-skip/completed-force/dry-run cases pass. Actual explicit reprocessing of temporary completed content/summary performs fresh extraction and preserves both on failure. No production data or bulk backfill touched.

### 15. Batch/watch — re-tested

`python3 /private/tmp/rlss-rerun-batch.py` serves a temporary localhost RSS feed with failed dashboard first, ordinary article second. Fresh actual watch operation completes TWO cycles, each Total 2 / Success 1 / Failed 1 / Skipped 0 / Partial 0. Second cycle actually retries the failed dashboard, then resumes ordinary cached content. Signal sent only to witness-created process during sleep; exit 0. New evidence `/private/tmp/rlss-acceptance-rerun-watch/cli.log`. Focused stats tests independently verify Skipped/Partial classification. This row now works.

### 16. Empty/dry run — re-tested

Fresh executable `process --config /private/tmp/rlss-acceptance-rerun-extra/ordinary/config.yaml --no-summary`: exit 0, no items to process. `url https://example.com --config <same isolated config> --dry-run --no-summary`: exit 0, Total 1 / Skipped 1 / Success 0 / Failed 0 / Partial 0. No new file is created.

### 17. Existing standalone content — re-tested

Actual ordinary article passes; single-target and rich-body fixture boundaries pass. The previous-user X candidate failed unknown end, so actual independent standalone social and social long article success remain unverified; partly.

### 18. Isolated verification — re-tested

All sixteen affected entries run using the new independent executable/worktree, individual guest profiles and isolated output/configuration. Four positive group original sources independently recaptured before injection, complete captions/order/exclusions/links checked against all eight actual outputs. Failed groups receive no invented content or completion claim; baseline source observations remain background only. Production settings, notes, accounts, executable and running processes were not changed. Raw source/full captions remain exclusively temporary; only prescribed public sample IDs/URLs, counts, hashes and outcomes are committed.

### Current findings and remaining limits

- **F1 — important — remains open**, anchor `spec.md REQ-1 / acceptance-test-report.md row 1`: eight mandatory entry points still fail (six X, two Threads AI colleague). Close by establishing complete permitted source membership/full bodies and passing independent full comparisons for both entry points of all remaining groups; do not weaken completeness or label safe failure successful extraction.
- **F2 — closed**, anchor `spec.md REQ-2 / First-post success`: all eight positive entry points now contain exactly one verified source link per member with unchanged ordered captions and identical first/middle output. No live X completion is inferred from this closure.
- Actual standalone social/long article success and actual external provider summary delivery remain unverified and disclosed as partly. Watch's later-cycle limitation is closed by two actual cycles. No required step was skipped; no important/fatal finding was dismissed.

## Re-run on 2026-10-28, at b9af8215846d972cd8fdd2b6f212d74e0e9f3f4d

This independent witness started a new detached worktree at `/private/tmp/rlss-author-final-acceptance`. Reviewed fix range `98b535f..b9af821`: reply-count consistency, late observed children, removal of a redundant early leaf check, and real public browser-entry regression coverage. All four Acceptance lines and all fourteen UI flows are affected and re-tested in full; no carried-over row and no product repair. This section supersedes earlier current verdicts, especially Acceptance 3: preservation on observed failures does not prove rejection of every uncertain narrative member.

### Setup, focused checks and independent method

- README setup: `npm ci --cache /private/tmp/rlss-final-accept-npm`, exit 0, 26 packages, zero vulnerabilities. `env GOCACHE=/private/tmp/rlss-final-accept-go go build -o /private/tmp/rlss-final-accept-bin ./cmd/rlss/`, exit 0.
- All actual entry commands use the new executable, temporary config, output and guest user-data directories under `/private/tmp/rlss-final-accept-live`; Chrome config root is also temporary. Command driver: `python3 /private/tmp/rlss-final-accept.py`. Every URL uses `url <confirmed URL> --config <case/config.yaml> --no-summary`, default headless-first/headed retry, 30-second attempt limits and three-second render wait. Case JSON and CLI logs retain outcomes locally.
- Independent observer: `python3 /private/tmp/rlss-final-accept-source.py`, public source visits for all four successful groups, headed fresh guests, four-second wait. Observer saves original `page.html` before injecting the current bundle; its extracted result is not the oracle. Original JSON parsing uses root `data.media` and ordered explicit `self_thread.posts.edges`, not product snapshots. Separate `python3 /private/tmp/rlss-final-connections.py` checks source post identities and final pagination; all four self-thread connections terminate with `has_next_page=false`, `end_cursor=null`.
- `python3 /private/tmp/rlss-final-accept-compare.py` independently compares every source caption against each saved body, removing only whitespace and appended provenance markers. It checks exact ordered member-link codes, one link each, no additional text, exact first/middle body hashes, and separate input URL/storage identity. Raw original HTML, captions, unrelated reply records and browser data remain exclusively temporary.
- Focused JavaScript command: `node --test --test-name-pattern='thread A|A2 target bodies|A2/A3 explicit target|A2 wrong target|A3 verified social|A3 challenge|A3 login' extraction.test.mjs`: 10/10 PASS. Existing incomplete-fixture Invalid URL diagnostic is present; zero failed/skipped tests. `node --test extraction-boundaries.test.mjs`: 2/2 PASS.
- Focused collector/public-browser command: `env CHROME_CONFIG_HOME=/private/tmp/rlss-final-accept-test-chrome GOCACHE=/private/tmp/rlss-final-accept-go go test ./internal/extract -run 'TestCompleteSocialThread|TestBrowserExtractSocialCompleteOnly' -v`: PASS, package 1.453s. Includes changed explicit counts, zero with author or reader child, late reader exceeding count, body navigation, unknown extent, missing members, loops/caps and complete-only public dispatch.
- Focused pipeline command: `env GOCACHE=/private/tmp/rlss-final-accept-go go test ./internal/pipeline -run 'Test(ProcessItemSocialCache|ProcessItemWithFallback|ProcessBatchFailedPaginationPreservesFiles|ExtractionRejectsObservedWalls|IsSkipped|Stats_Report|Stats_Partial)' -v`: PASS, package 0.645s. Includes cache revalidation, recording summary receiver, zero summary calls on rejected refresh, completed skip/force, ordinary Partial behavior, failure preservation and verification retry settings.
- No full package suite ran in this witness checkout. Neither package-tests nor finalize-review is skipped. The automated test suite runs before acceptance and blocks it on failure; `go test ./...` remains the finalize-review command when verdicts permit. Main-agent earlier suite evidence is context only.

### 3. Failure honesty and existing files — re-tested

`python3 /private/tmp/rlss-final-accept-extra.py` explicitly reprocesses dashboard first URL with temporary sentinel old body/summary. Exit 1, complete reply extent unknown; both before/after hashes unchanged: content `c6f945cd02c4767cefdc70f7f17b8afce8b7ad1a2edb6952f1de979e612695d4`, summary `95d622c7618d10e594b100c1d8bca7bd77e820ff3f4ab13d1fd0226b3cebb609`. Current focused tests also pass conflicting-count and late-direct-child rejection. These prove observed failures and preservation, not universal narrative-membership rejection. A3 is partly because resource-only author-reply membership remains unresolved.

### 4. Existing content and retry behavior — re-tested

Actual `https://go.dev/blog/strings`: exit 0, body file SHA-256 `29dbfcfbfca50014f6bdf9e5be4c845ad5b93457e4f29daf46a4d500558d2887`. Actual standalone candidate `https://x.com/Sukiea1008/status/2112084040362897441`: exit 1, headed result complete reply extent unknown, no body. No live independent standalone-social or social-long-article positive is claimed. Focused retry and single-target tests pass; A4 remains partly.

### 8. Progress and ordinary retry — re-tested

Actual X logs show item progress and headless failure followed by headed retry. Focused navigation/empty fallback and explicitly headed single-attempt cases pass. No production browser or process was used.

### 9. Verification restriction — re-tested

Focused tests pass verification setting false/true, specialized verification disabled, and navigation/empty retries with the setting false. Disabled verification retry does not disable ordinary failure retries.

### 10. Missing target/body — re-tested

Focused target mismatch, previews, quoted permalink, login/challenge and incomplete public-entry cases reject; actual failures exit nonzero without new body. Actual sentinel refresh preserves both old files. This does not assert that every ambiguous resource reply is detected.

### 11. Incomplete/ambiguous thread — re-tested

Actual incomplete-extent/pagination cases fail without output; focused missing member, contradictory identity, count conflict, late child, loop/cap and unknown boundary cases pass. Narrative membership of same-author resource replies is still not established, so the whole row is partly.

### 12. Old body without summary — re-tested

Focused verified-social-resume refreshes live content before the recording receiver; failed refresh preserves existing files and invokes no summary; ordinary cached content retains existing resume behavior. PASS.

### 13. Summary enabled — re-tested

Focused recording receiver receives newly validated body, extraction failure makes no summary call and summary failure stays Partial. No real external provider was authorized or invoked, and no actual live complete multi-member external-summary result was verified; partly.

### 14. Completed notes and explicit reprocessing — re-tested

Completed skip/force focused cases pass. Actual explicit temporary completed-note reprocessing performs fresh extraction and preserves body/summary on failure. No bulk backfill or production note operation occurs.

### 15. Batch/watch — re-tested

`python3 /private/tmp/rlss-final-accept-batch.py` runs a temporary localhost RSS feed (failed dashboard then ordinary Go article) through actual `process --watch --no-summary`. TWO cycles each produce Total 2 / Success 1 / Failed 1 / Skipped 0 / Partial 0. Second cycle retries failed dashboard and then continues with the ordinary cached body. Signal sent only to this witness-created process during sleep; exit 0. Raw log `/private/tmp/rlss-final-accept-watch/cli.log`. Focused stats tests separately verify Skipped/Partial reporting.

### 16. Empty/dry run — re-tested

New binary `process --config /private/tmp/rlss-final-accept-extra/ordinary/config.yaml --no-summary`: exit 0, no items. `url https://example.com --config <same temporary config> --dry-run --no-summary`: exit 0, Total 1 / Skipped 1 / Success 0 / Failed 0 / Partial 0, no new file.

### 17. Existing standalone content — re-tested

Ordinary article passes live and single-target/rich-body fixtures pass. Actual standalone X candidate fails unknown extent. Independent social standalone/long-article success remains unverified; partly.

### Open findings (not dismissed)

- **F1 — important — eight mandatory entry points fail**, anchor `spec.md REQ-1 / report row 1`. Safe failure is not complete article delivery. Close only with permitted complete first-party membership/full captions and independent first/middle source comparisons for all remaining groups.
- **F3 — fatal — narrative membership unresolved**, anchor `internal/extract/threads.go:202` and `:245`, spec Design decision on resource-only replies, REQ-2/REQ-3, report rows 2/3/7/11. Current branch collection selects explicit same-author reply ancestry; count agreement establishes extent but does not establish that a separate resource recommendation belongs to the article. Fresh positive self-thread comparisons do not settle this broader branch contract, nor do failing live X sources. Close with independent source-grounded article-membership evidence, reject ambiguous resource-only author replies, and re-test both real entry points plus a resource-reply negative through production extraction. This witness did not repair the product or dismiss this finding.
- F2 provenance remains closed only for the independently verified positive outputs; no live X success is inferred. Standalone-social/long and real-provider summary limits remain disclosed.

### 1. All required first/middle articles — re-tested

All sixteen actual CLI attempts finished: eight successful exits across four Threads groups; eight failed exits, six X and two Threads AI colleague. Failed entries create zero body/summary files. Every URL is the exact public first/middle entry already listed in baseline section 1 and the confirmed spec.

| Group | First / middle exit | Actual result |
|---|---|---|
| X dashboard skills | 1 / 1 | headed retry: complete reply extent unknown |
| X Japan MCP tools | 1 / 1 | headed retry: complete reply extent unknown |
| X Google I/O | 1 / 1 | headed retry: complete reply extent unknown |
| Threads Open Design | 0 / 0 | complete independent source-caption comparison passes |
| Threads AI colleague | 1 / 1 | direct reply pagination incomplete in both attempts |
| Threads pstack | 0 / 0 | complete independent source-caption comparison passes |
| Threads Claude Code files | 0 / 0 | complete independent source-caption comparison passes |
| Threads AI adoption gap | 0 / 0 | complete independent source-caption comparison passes |

| Positive group | Members / links per body | Equal first/middle body SHA-256 | Fresh original HTML SHA-256 |
|---|---:|---|---|
| Open Design | 4 / 4 | aab776546687aad3d07ebe4f901dfd28a444d1f40a62fed1b1eabdbe63d1687b | 8d9321d5021a63529927ffbea9ecf9802c93a58f0a68a5f980a75f9bcb30e5ce |
| pstack | 6 / 6 | 840ae6fdfe38d8c1fd18086d24f954b8f47e668c51998183a4933b460b48e46e | e2520b115444889250aff3f752dd6f84d044552e6a920de9549aab73e4ddabc2 |
| Claude Code files | 5 / 5 | 8f46290a6b74d9df1c2aa5d5cc11d2067a47084ebed0763a1e23c632ed182085 | 511877b626d669785813fdf5f9d6dd1a419301ab2383dc07461808e229c62cfa |
| AI adoption gap | 5 / 5 | 8a37778517ffb465224a197856ea3126b0fd68f833bea1a976f153b870ba555e | 27eaf2858f65118a7eaee3cd0ad7cb30fd4c4e903010f69581fd0ba0f7d57fdf |

All source member codes, caption lengths and caption hashes were recomputed from the new original HTML and exactly equal the baseline per-member tables. Every positive body equals the full ordered caption sequence after only whitespace/provenance-marker removal. Source-link codes equal the complete member sequence once, and eight separate input identities remain intact. Raw comparison records are local `04-first/comparison.json`, `06-first/comparison.json`, `07-first/comparison.json`, `08-first/comparison.json` under the temporary live directory. A1 fails; four positive groups do not establish all eight.

### 2. Membership, ordering and exclusions — re-tested

All eight successful outputs have one complete ordered caption copy, exactly one source link per member, no additional body text, and exact first/middle equivalence. Source reader/quote/recommendation records were not used as expected article text. Focused fixtures pass reader/quoted-permalink exclusions, distinct-member preservation and conflicting identity rejection. These positive self-thread samples do not prove exclusion of every resource-only same-author reply admitted through the separate ancestry-based branch path. Google distinct seventh-section outputs remain unavailable. A2 is partly; F3 remains fatal and open.

### 5. First-post success — re-tested

All eight first URLs ran through the actual CLI: four complete-source positives with 4/6/5/5 members and four failures. No new switch, no partial-body success. Partly.

### 6. Middle-post success — re-tested

All eight middle URLs ran through the actual CLI: four exact body-hash matches to their first-entry output, each retaining its supplied URL identity; four failures. No other notes merged, renamed or deleted. Partly.

### 7. Correct scope and repetition — re-tested

Complete independent source comparisons settle positive bodies, ordering, once-only provenance and no extra text. Broader resource-only author-reply exclusion and distinct Google-section output remain unverified; partly rather than a universal membership pass.

### 18. Isolated verification — re-tested

All sixteen mandatory entry points were attempted with the new clean executable, separate guest profiles/configs/outputs and no credentials. All four positive groups have new independent original source HTML, terminal self-thread records and complete caption comparisons against both actual outputs. Failing groups are not claimed complete or source-verified. No production config/profile/account/notes/executable/process was touched. Only prescribed public URLs/IDs, counts, hashes and test results enter permanent evidence; original HTML/captions and account state remain outside Git. No new paid API, proxy, automatic login or external summary service was used. No required station was skipped, no important/fatal finding was dismissed, and acceptance remains blocked.

## Build investigation — authenticated source compatibility, 2026-10-29

This is orchestrator source research and focused implementation verification at `457fa8641def6b2830c150b18f298564109c32b5`, not a new independent acceptance report. The sixteen-entry witness above was not rerun at this revision. F1 and F3 remain open; no closing-review round, attestation, publication or merge is implied.

The user authorized reusing existing website sessions through an isolated copy. The original Chrome stayed open and unchanged. Only relevant website cookies and the minimal browser encryption configuration were copied locally; no credentials, request headers, account state or raw copyrighted captures enter Git. Native read-only request headers stayed in memory and were sent only to the original first-party endpoint.

| Source experiment | Observed result | What remains unproven |
|---|---|---|
| Guest Threads AI colleague | A login dialog covers the page, but embedded data retains 18 public posts including unrelated content; the parser recognizes the target. Programmatic internal scrolling fetches no further reply response. | An overlay does not establish an empty page, a complete article or access to the next page. |
| Authenticated Threads AI colleague | The native reply request uses the exact embedded `end_cursor` as `after`, targets post `3870860576310956047`, and returns the matching media-info owner with two new posts and `has_next_page: false`. | TOP ordering alone does not prove unfiltered extent; the product still ignores this response and narrative membership remains unresolved. |
| Authenticated X dashboard | Native probable-spam expansion adds three reader replies; observed distinct root children reach 11 of the reported 11. | Root reply extent is not complete-article membership or a product extraction pass. |
| Authenticated X Japan MCP | Native expansion yields 18 distinct direct children against the reported 18. A separate same-author resource recommendation is present. | Same author and ancestry do not justify including that recommendation in the article. |
| Authenticated X Google I/O | Initial response exposes 15 distinct direct children against the reported 68. Three native read-only cursor requests return HTTP 200 with no additional posts. | Empty subsequent pages do not prove that all author continuations were enumerated. |

The source decoder now handles modern nested-user records, raw GraphQL legacy records and the normalized `id_str`/`user.id_str` records actually retained by authenticated article props. Unknown ancestry stays unknown and conflicting identity, counts or parent observations fail. The actual authenticated dashboard initial snapshot recognizes its root with a complete 363-character body; scrolling still unmounts that target, which the current one-frame snapshot cannot recover on its own.

The actual Google article places its original body inside a nested `article[data-testid="twitterArticleReadView"]`. Generic nested-article quotation removal deleted that body before decoding. The fix retains the explicit native owned read-view while continuing to remove quoted cards and arbitrary nested articles. Live snapshot verification now returns a complete focal body of 8,487 characters including its title, SHA-256 `a797588cb7ebcbd60f49e7bca959613edbdab7047f82ab27ac26c5b786575c42`. Removing the title yields exactly the existing focal reader's 8,451-character result, SHA-256 `f39b90fe6c5739074f9c38238d2c825b37e448212f1bfb6d652f4903929a56f1`. This is a focal-body comparison with an existing reader, not independent complete-thread acceptance.

Verification: test-first RED/GREEN evidence for W9–W12 is retained in local `rlss-w9-*` through `rlss-w12-*` logs. The focused JavaScript suite passes all 19 tests. Trunk sync reports current origin/main already contained; the full `go test ./...` suite passes at the final functional revision (`/private/tmp/rlss-w12-package-suite.log`). The unchanged earned conflicting-root program passes (`/private/tmp/rlss-w12-adversarial.jsonl`). Final bundle generation is reproducible and the W12 complete diff whitespace check passes. No new independent acceptance success, standalone-social success, real summary-provider verification or resolution of F1/F3 is claimed.


## Build iteration: native Threads reply pagination (2026-10-29)

Scope: orchestrator integration verification at functional HEAD
`812242a486d9f5195b89faf432178f4e4fcef488`. This is not a replacement
independent acceptance report or a closing-review verdict. F1 and F3 remain
open; the previous sixteen-entry acceptance matrix was not rerun.

W13 joins supplemental native reply responses only when request post ID,
previous cursor and response media-info owner match the initial source.
W14 observes those first-party responses in bounded memory while scrolling
existing containers within the original attempt deadline. No requests are
replayed and no headers or credentials are retained. W15 preserves these
responses across the observed bare-to-www redirect, with port and other
response checks unchanged. W16 accepts absent per-post ancestry only within
a verified root-owned branch; explicit conflicting ancestry and standalone
identity remain failures. This proves source provenance, not narrative
membership.

Live production-entry runs used the previously authorized isolated Chrome
site-session copy, temporary configuration, temporary executable and output,
with summarization disabled. The original browser, configuration, notes,
executable and running production process were untouched.

- Before W15, the first entry failed with direct reply pagination incomplete.
- After W15, both first and middle entries reached the next guard and failed
  with unproven direct author branch; native source omitted redundant
  per-post root IDs despite a verified root-owned connection.
- After W16, both entries exit 0, Success 1, Failed 0. They produce identical
  bodies containing five distinct source URLs exactly once, in the same order.
- Group: Threads AI colleague, first `DW4Dy96FBAP`, middle `DW4D1bfFPUj`.
  Ordered members: `DW4Dy96FBAP`, `DW4D1bfFPUj`, `DW4D2SNlP5l`,
  `DW4D3LAFCCD`, `DW4D361FIv7`.
- Each caption hash equals the corresponding earlier research-source caption;
  character lengths are 321, 291, 349, 230 and 345. No trailing extra body
  remains. Combined-body SHA-256 is
  `86f73cd0e2198c40346625245817891d4ffe925c8ff59a1432c7ad8d85c3125c`.
- The reader-branch author response `DW5dImTlBe6` is absent. This exact
  comparison is research verification, not independent approval of generic
  narrative membership or exhaustive TOP reply-filter semantics.

Local evidence: `/private/tmp/rlss-w14-native-live/w16-comparison.json`,
`first/cli-w16.log`, `middle/cli-w16.log`, with corresponding isolated output
files. No article captions or session data are committed here.

Verification at W16:

- W13–W16 focused tests each observed RED before the minimal implementation
  and GREEN afterward. W15's real TLS fixture redirects bare host to www;
  W16 extends the existing first/middle branch fixture without its redundant
  member root ID. Existing wrong-owner and contradictory-source negatives pass.
- Trunk sync: HEAD already contains origin/main
  `8e37f33e252337dc5f2727631b35460eb7086066`; no sync commit.
- Complete `go test ./...`: PASS,
  `/private/tmp/rlss-w16-package-suite.log`.
- Direct `npm test`: PASS, `/private/tmp/rlss-w16-javascript-tests.log`;
  run directly because Go's cached package result does not track JS inputs.
- Unchanged adversarial program
  `go test -json internal/extract/probes/conflicting_root_test.go`: PASS,
  `/private/tmp/rlss-w16-adversarial.jsonl`. Probe SHA-256 remains
  `9d9389bfa8bac64f8c41829e9a2dcc02626f7ae7de2faf30bf813c380536489f`.
- Source diff whitespace checks pass. Complete cumulative diff check still
  reports two pre-existing third-party string whitespace locations in the
  regenerated bundle; no generated string was manually altered. Bundle SHA-256
  remains `7e315b2cd558cf885d5785f5f8afa2dad062f3fa3a72b00986bd17220a954eaf`,
  unchanged since the reproducible W13 generation.

Unresolved: X source pagination/virtualized snapshots, complete first/middle
source comparisons for all mandatory groups, and source-grounded distinction
between an article continuation and a separate same-author recommendation.
Native SelfThread display metadata also marks the observed Japan MCP resource
recommendation and cannot resolve that distinction by itself. No independent
acceptance pass, final review round, attestation, publication or merge is claimed.
