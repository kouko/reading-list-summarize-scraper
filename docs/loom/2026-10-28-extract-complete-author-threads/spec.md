# Complete author threads — spec
intent: 2026-10-28-extract-complete-author-threads@183b82f
confirmed-behavior: 2026-10-28 @6c5aee3
pre-build-review: required — Thread membership, source completeness and replacement of saved content affect the article contract; unrelated replies and cached partial bodies must not become successful articles.

## Requirements
REQ-1 — Complete articles from first and middle posts
  WHEN an article from any of the three confirmed X groups or five confirmed Threads groups is processed, the scraper shall deliver its complete ordered author body from either a first-post or middle-post URL, with equivalent body content across those entry points and independent source comparisons for all eight groups → Acceptance #1

REQ-2 — Correct membership, ordering and deduplication
  WHEN the scraper combines an author thread, the scraper shall retain the original author-body order and source links, include each member post once, and exclude reader comments, author conversations with readers, quoted-post bodies and recommended articles → Acceptance #2

REQ-3 — Honest incomplete results and preserved existing files
  IF target identity, continuation membership or complete body cannot be established THEN the scraper shall report extraction failure without delivering missing sections, previews, login or verification content as a complete successful article or using unverified cached social content to bypass that failure, while preserving existing body and summary files → Acceptance #3

REQ-4 — Existing extraction and retry behavior
  WHEN supported standalone posts, long articles or ordinary articles are processed, the scraper shall retain their successful extraction and existing headless-first, headed-retry and verification-setting behavior, with development verification isolated from production configuration, browser data, notes, executable and running processes → Acceptance #4

## Design decision
- agent-decided — Extend the existing browser extraction and save boundary rather than introduce a second crawler, account index or conversation archive; the requested result is one complete author article for the supplied URL.
- agent-decided — Treat author identity and explicit source conversation relationships as necessary membership evidence. Same-author spelling, adjacent display, timestamps, numbering or matching text alone cannot prove continuation; a reader branch must not be merged back into the article.
- agent-decided — Discover the first member and exhaust the verified continuation sequence before returning content. Expand or open member previews through first-party source pages when necessary; stop with failure on missing members, unresolved pagination, inaccessible bodies, loops or exhausted existing request limits rather than silently truncate.
- agent-decided — A page with no observed continuation is not automatically proven to be a standalone post. Require adequate source evidence for the delivered article's extent; the audit's single-scroll snapshots establish positive samples, not proof of absence or complete membership.
- agent-decided — Deduplicate repeated observations by stable post identity. Preserve distinct authored posts instead of silently deleting them merely because their wording is similar; the Google sample has two distinct URLs for its seventh section and requires source-grounded resolution, not a text-similarity shortcut.
- agent-decided — Retain author-written references and links inside member bodies, but do not expand quoted social cards, recommended content or linked articles into the combined body. Resource-only author replies are not automatically article members merely because the author matches.
- agent-decided — Preserve the supplied URL's existing output identity and storage conventions; equivalent first/middle bodies do not authorize merging, renaming or deleting existing notes across URLs.
- agent-decided — Explicit URL processing and forced social reprocessing must revalidate the live complete article even if a content-only cache exists. Normal completed-item skipping remains; no bulk backfill or migration is introduced. Unrelated resume behavior stays intact.
- agent-decided — Retain existing result categories and summary behavior: successful extraction is followed by the configured summary step, and summary failure remains Partial. No summary starts from a partial or unverified social article.
- agent-decided — Use existing browser configuration and first-party access without a new paid API, account requirement, proxy service, file-format migration or automatic login. These are reversible, zero-obligation choices; no new existing-data operation beyond explicitly requested normal reprocessing is authorized.
- agent-decided — Keep PR #24 and its unmerged work untouched. Planning must account for its reviewed modern X layout fixes without claiming it has merged or modifying that PR without authorization.
- agent-decided — Use sanitized permanent fixtures and counts, identities and source URLs in evidence. Full copyrighted source captures, account state and the personal reading-list inventory remain outside version control.

### Confirmed sample groups

The audit confirmed eight groups at nine previously processed URLs. Seven additional entry URLs below are derived from the same groups to verify both first-post and middle-post behavior; they require live verification, not an assumed pass. Additional source members must be discovered and independently checked during implementation and acceptance.

| Group | First-post URL | Middle-post URL | Audit evidence and limits |
|---|---|---|---|
| X dashboard skills | https://x.com/SuguruKun_ai/status/2039372810053967953 | https://x.com/SuguruKun_ai/status/2039372826311098624 | Both URLs were already processed; each saved body lacks surrounding members. Same conversation was also opened independently from the second post. Other author resource replies require separate membership checks. |
| X Japan MCP tools | https://x.com/SuguruKun_ai/status/2038872366743502989 | https://x.com/SuguruKun_ai/status/2038872379066327076 | Supplied URL is the middle post; preceding tool list is visible in source context but missing from the saved body. The derived first-post URL needs its own live check. |
| X Google I/O | https://x.com/Khazix0918/status/2056894400320708671 | https://x.com/Khazix0918/status/2056894667413958907 | Later author posts continue sections six and seven; some are previews. Seventh-section URLs 2056894891205296229 and 2056895055617724832 are distinct and must not be silently collapsed. Full source sequence remains to be established. |
| Threads Open Design | https://www.threads.com/@aiposthub/post/DYliXKSkRBP | https://www.threads.com/@aiposthub/post/DYliX_qkbFu | Source self-thread lists three continuations; their observed text is already present in the saved body. |
| Threads AI colleague | https://www.threads.com/@ci.fullstack/post/DW4Dy96FBAP | https://www.threads.com/@ci.fullstack/post/DW4D1bfFPUj | One self-thread member and three substantive direct author replies were observed; two of four continuation bodies are incomplete in the saved output. Direct-reply membership requires verification. |
| Threads pstack | https://www.threads.com/@ci.fullstack/post/De13wZUiRHk | https://www.threads.com/@ci.fullstack/post/De138E9CYfv | Source self-thread has numbered continuations 1/ through 5/; none is present in the saved body, and all were published before that file was written. |
| Threads Claude Code files | https://www.threads.com/@gakikun_claudecode/post/DeJwYehANl- | https://www.threads.com/@gakikun_claudecode/post/DeJwZRmAIRk | Four source self-thread members were observed; related same-author recommendations are excluded from membership. |
| Threads AI adoption gap | https://www.threads.com/@lcj.hankliu/post/DXB4kzEFLdn | https://www.threads.com/@lcj.hankliu/post/DXB4lahlFpz | Four source self-thread members were observed; an author response to another reader is excluded. |

Audit provenance: 101 previously processed unique URLs were opened in a fresh logged-out browser on 2026-10-28, with observations before and after one bottom scroll. Nine inputs correspond to eight confirmed groups; thirteen other inputs have author replies or supplements without confirmed narrative membership, seventy-three showed no author continuation in the observed page, and six targets were unavailable. None of those negative observations proves no hidden continuation exists. Local research evidence is retained at `/private/tmp/rlss-thread-audit-20261028/`; raw captures and the personal inventory are not committed.

## Alternatives considered
- Merge all same-author posts on screen: rejected because recommendations, quotations, earlier articles, resource replies and reader conversations can share the author.
- Stop at the supplied post or the currently visible replies: rejected because the samples include middle entry points, omitted earlier members and truncated continuation previews.
- Use numbering or a language model alone to infer membership and the end: rejected because inference does not establish source identity or completeness and introduces unsupported success claims.
- Trust any cached nonempty body: rejected because existing saved social content demonstrably lacks continuations and the current content-only resume path can bypass fresh extraction.
- Backfill all existing notes or canonicalize their storage automatically: not part of the requested capability; normal explicit reprocessing is sufficient and preserves existing-data boundaries.

## Current state evidence
- Forward: `cmd/rlss/url.go`, `runURL` starts explicit URL processing and sets SkipExisting false; `cmd/rlss/process.go`, `runProcess` and force handling supply batch entry points.
- Reverse: `internal/pipeline/runner.go`, `ProcessItem` calls `extractWithFallback`, then writes body and invokes summarization; content-only resume currently bypasses extraction.
- Error: `extraction.ts`, `social` rejects unavailable targets but selects a matching single post; `internal/pipeline/extraction.go`, `extractWithFallback` validates returned text and controls headed retry.
- Data: `extraction.ts`, `sameTarget`, `ownTweet`, `bodyMarkdown` and the Threads post/container/JSON-LD branches handle source identity and Markdown; saved files currently do not establish whole-thread membership or completeness.
- Boundary: `internal/pipeline/extraction.go`, headedOnBlock behavior and `internal/output/index.go` existing URL/file indexing remain; the production configuration, notes and running process are outside verification writes.

## UI flows

| Case | What the user does | What they see |
|---|---|---|
| First-post success | Run the existing single-URL command for a complete first-post sample with summary disabled | One article body contains the entire verified author thread in order, with original sections and source links; the report shows Success 1 and Failed 0, exit 0. Thread detection is automatic and requires no new switch. |
| Middle-post success | Run the same command for a middle member of that sample | The article includes verified earlier and later members and has the same body as first-post processing. The supplied URL retains its existing output identity; existing notes for other URLs are not merged or deleted. |
| Correct scope and repetition | Process a thread containing comments, quoted posts, recommendation cards or repeated observations | Only verified author article members appear. A member seen more than once is included once. Author-written reference links remain; quoted-card bodies, reader conversations and recommendation bodies do not. Distinct authored posts are not silently deleted just because their wording is similar. |
| In progress and ordinary retry | Start extraction with the existing default configuration | Existing progress remains visible; the scraper may take longer while loading article members. Failed headless extraction follows the existing headed retry. Explicit headed execution retains its existing behavior. |
| Verification restriction | Encounter a recognized verification page | With headed_on_block false, the blocked attempt does not open a headed retry; with it enabled, the existing headed retry is allowed. Normal navigation or missing-body failures retain their existing retry behavior. |
| Missing target or body | Process an unavailable, mismatched, login-only or empty target | The existing Failed count increases; the single-URL command exits nonzero with an Inaccessible social target error explaining the reason. No article body or summary is replaced. Retry after the source becomes accessible. |
| Incomplete or ambiguous thread | Process a thread with missing members, unreadable previews, unresolved earlier/later content or ambiguous membership/order | The command reports Failed 1 and exits nonzero; the Inaccessible social target error explains that the complete thread could not be established. Existing body and summary remain unchanged. Retry after source availability is restored; the next attempt revalidates the whole article. |
| Old content without summary | Explicitly process a social URL whose old body exists but summary is missing | The scraper first revalidates the complete article instead of immediately summarizing the old body. Success replaces that URL's body through normal processing; extraction failure preserves the old body and summary and starts no summary from partial content. |
| Summary enabled | Process a verified complete article using the existing summary configuration | The summary step receives the complete combined body. Existing summary success and Partial behavior remain; an extraction failure starts no summary. External summary-service availability is not changed by this feature. |
| Completed notes and explicit reprocessing | Run batch processing with existing skip settings, or explicitly request URL/forced reprocessing | Completed notes retain existing Skipped behavior. Explicit reprocessing verifies a fresh complete social article and may replace that URL's saved body after extraction succeeds. Installing this change does not automatically batch-rewrite old notes. |
| Batch and watch | Continue the existing batch or watch operation | Each selected article follows the same complete-content checks. An individual Failed item does not stop later items; existing Success, Failed, Skipped and Partial reporting and batch exit behavior remain. Watch may retry a failed item on a later cycle. |
| Empty or dry run | Run with no selected items or use the existing dry-run option | Existing no-items or dry-run reporting remains, with no new extraction or body/summary writes. |
| Existing standalone content | Process an ordinary article, standalone post or single long article | Existing body extraction and output conventions remain. A known incomplete social thread does not silently fall back to single-post success. |
| Isolated verification | Verify the eight groups during development and acceptance | Evidence compares complete source members, order, body sections and exclusions for each group and both entry points. All settings, browser data, executables and outputs are isolated; production notes and running processes remain unchanged. An unavailable or incompletely verified sample is not reported as passing. |

No bulk rewrite, deletion, new paid service, new required account or data-format migration is introduced. Normal explicit reprocessing has the existing consequence of replacing that URL's content after successful extraction; failure preserves its previous files. Every sample group must pass independent actual-source verification before the feature is described as completed.
