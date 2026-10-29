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
- Test: A1 positive: exact-direct-child-count; boundary: unknown-parent. A3 positive: exact-direct-child-count; negative: missing-direct-child. A4 positive: bounded-complete-X-source; negative: invalid-reply-count.
- Risk: agent-decided — REQ-1,3–4: require exact nonnegative source counts and unique explicit direct children within one snapshot. Unknown or mismatched counts, display-group endings and timeline termination remain insufficient. Existing incomplete live samples remain failures.

### Wave 7 — Consistent reply-count evidence

**W7-01 Reject contradictory reply counts across collection** after: W6-01 acceptance: 2, 3
- Files: internal/extract/threads.go (collectSocial), internal/extract/threads_test.go
- Test: A3 positive: consistent-count-observations; negative: contradictory-count-observations. A2 positive: reader-branch-excluded; negative: unrelated-author-branch.
- Risk: agent-decided — REQ-2–3: reject changing explicit counts and counts smaller than known distinct direct children before accepting terminal evidence. Preserve unknown fields and source membership rules; correct the existing zero-count reader fixture without changing its exclusion assertion.

### Wave 8 — Production social dispatch coverage

**W8-01 Exercise social browser behavior through Browser.Extract** after: W7-01 acceptance: 3, 4
- Files: internal/extract/browser_test.go (TestBrowserExtractSocialCompleteOnly and existing browser helper)
- Test: A3 positive: complete-social-entry; negative: incomplete-social-entry. A4 positive: production-social-dispatch; negative: incomplete-social-entry.
- Risk: agent-decided — REQ-3–4: move existing social assertions through the production entry using an isolated intercepted social hostname. Preserve ordinary extraction and remove superseded helper calls; use no production profile, external page requests or new production configuration.

### Wave 9 — Authenticated X source compatibility

**W9-01 Decode both observed public X post schemas without inventing ancestry** after: W8-01 acceptance: 1, 3
- Files: extraction.ts (socialSnapshot X source decoder), extraction.test.mjs, embed/defuddle.min.js
- Test: A1 positive: authenticated-legacy-source-identity; boundary: modern-source-preserved. A3 positive: explicit-source-parent-or-root; negative: unknown-or-conflicting-source-relationship.
- Risk: agent-decided — REQ-1,3: widen extraction.test.mjs for authenticated legacy fields; preserve modern fixtures. Matching conversation_id_str establishes the root only without contradictory ancestry; absent fields remain unknown. Pagination and narrative membership remain unresolved.

### Wave 10 — Rendered authenticated X identity

**W10-01 Read the authenticated post shape actually retained by article props** after: W9-01 acceptance: 1, 3
- Files: extraction.ts (socialSnapshot X source decoder), extraction.test.mjs, embed/defuddle.min.js
- Test: A1 positive: normalized-rendered-source-identity; boundary: raw-and-modern-source-preserved. A3 positive: explicit-normalized-root-and-parent; negative: unknown-or-conflicting-normalized-relationship.
- Risk: agent-decided — REQ-1,3: extend existing source-decoder fixtures for observed article props containing id_str, user.id_str and top-level legacy fields. Preserve W9 guards and bounded article scope; never interpret analytics IDs as posts or absent parent as root.

### Wave 11 — Authenticated rich article body compatibility

**W11-01 Preserve the owned classic X rich article body in source snapshots** after: W10-01 acceptance: 1, 3
- Files: extraction.ts (modernXBody used by socialSnapshot), extraction.test.mjs, embed/defuddle.min.js
- Test: A1 positive: owned-classic-rich-body; boundary: modern-rich-body-preserved. A3 positive: verified-owned-rich-container; negative: quoted-or-ambiguous-rich-body.
- Risk: agent-decided — REQ-1,3: widen existing rich-body fixtures for the observed twitterArticleReadView/twitterArticleRichTextView layout and upstream title selector. Keep own-article identity, quotation removal and incomplete-body rejection; a readable root body does not prove full thread extent.

### Wave 12 — Owned rich body versus quoted article

**W12-01 Preserve the native owned rich read-view during quotation removal** after: W11-01 acceptance: 1, 2, 3
- Files: extraction.ts (ownTweet used by socialSnapshot), extraction.test.mjs, embed/defuddle.min.js
- Test: A1 positive: nested-native-rich-body; boundary: modern-rich-preserved. A2 positive: owned-read-view-retained; negative: quoted-article-excluded. A3 positive: complete-owned-body; negative: ambiguous-rich-body.
- Risk: agent-decided — REQ-1–3: widen existing classic rich fixture to the observed nested article read-view. Preserve known quotation/card exclusions and reject duplicate bodies; nested native ownership differs from an arbitrary quoted article.

### Wave 13 — Verified Threads response pages

**W13-01 Merge root-owned native reply pages into the existing source snapshot** after: W12-01 acceptance: 1, 3
- Files: extraction.ts (socialSnapshot), extraction.test.mjs, inject.ts (window.socialSnapshot), embed/defuddle.min.js
- Test: A1 positive: native-reply-page-continuity; boundary: static-self-thread-preserved. A3 positive: matching-request-and-response-owner; negative: wrong-owner-or-cursor.
- Risk: agent-decided — REQ-1,3: extend existing fixtures and snapshot entry with bounded supplemental pages. Match request post ID, previous cursor and response media-info owner before joining; retain source conflicts and unknown pagination. No article-membership relaxation.

### Wave 14 — Browser reply-page ingestion

**W14-01 Collect native Threads reply responses while loading the article** after: W13-01 acceptance: 1, 3, 4
- Files: internal/extract/browser.go (Browser.Extract), internal/extract/browser_test.go
- Test: A1 positive: native-replies-through-browser-entry; boundary: first-middle-source-equivalence. A3 positive: verified-page-delivered; negative: wrong-owner-or-unresolved-page. A4 positive: existing-social-and-ordinary-dispatch; boundary: shared-attempt-deadline.
- Risk: agent-decided — REQ-1,3–4: widen browser_test.go using isolated intercepted hosts. Observe only first-party reply requests; keep headers out of captures and all data in bounded memory. Scroll existing containers within the attempt deadline; preserve X, retry and challenge behavior.

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
5. agent-decided — Acceptance F1 remains open: authenticated source progress does not repair product pagination or establish complete narrative membership. Preserve failed acceptance until all groups have permitted complete-source evidence and product verification.
6. agent-decided — X selects partial conversation branches (https://help.x.com/en/resources/recommender-systems/conversations-recommendations); counts measure direct replies (https://help.x.com/en/using-x/x-conversations). Guest dashboard exposes two of eleven direct children. Termination and module tip IDs do not prove author completeness.
7. agent-decided — Closing review also leaves source-grounded narrative membership unresolved: author identity plus reply ancestry may include a separate resource recommendation. Exact reply counts prove extent, not article membership. Retain this blocker alongside F1; do not weaken confirmed guarantees.
8. user-decided — Existing Chrome website sessions may be reused through an isolated read-only copy for source experiments. The original browser stays open and unchanged; no credentials or account data enter committed evidence. Product profile selection is unchanged.
9. agent-decided — Authenticated experiments obtained all eleven dashboard and eighteen Japan MCP root replies through probable-spam expansion; Google remains partial. Threads followed its exact embedded cursor to two new root-owned posts, but TOP ordering does not prove unfiltered completeness.
10. agent-decided — Guest Threads retains eighteen parseable public posts, including unrelated content, behind its login dialog; internal scrolling fetches no further replies. Read loaded source before classifying failure; a login overlay proves neither emptiness nor completeness.
11. agent-decided — W9 fixes raw response compatibility but fails live authenticated identity: article props normalize rest_id/core/legacy into id_str/user/top-level fields. The bounded owned-article probe locates the target at ancestor props.tweet; use that observed shape rather than widening into account state.
12. agent-decided — W10 recognizes authenticated targets and passes package/probe checks. Google yields an empty snapshot body despite its owned classic rich container; the existing focal reader returns 8,451 characters. Repair that layout while retaining unresolved pagination and membership findings.
13. agent-decided — W11 passes fixtures but live cloning still removes the body. The target tweet contains a nested article explicitly marked twitterArticleReadView; generic nested-article removal deletes it before rich decoding. Retain this native owned view without admitting quoted cards.
14. agent-decided — Implement proven Threads cursor ingestion before further X inference. Initial scripts remain authoritative for ownership; supplemental direct-reply pages must connect without gaps. Native TOP terminal pagination establishes that connection only; F1/F3 still require independent article verification.
15. agent-decided — Observed X tweetDisplayType SelfThread also labels the Japan MCP resource recommendation. Neither that flag nor VerticalConversation modules prove narrative membership; do not replace the unresolved membership requirement with this display hint.
