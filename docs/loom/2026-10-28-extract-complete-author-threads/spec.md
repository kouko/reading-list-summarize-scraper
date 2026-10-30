# X specified-post extraction — staged release spec
intent: 2026-10-28-extract-complete-author-threads@183b82f
confirmed-behavior: 2026-10-30 — user accepted the explicit X specified-post stage; original full-thread behavior remains deferred
pre-build-review: required — Source identity, complete focal bodies, scope labels and replacement of saved content affect the article contract; experimental thread assembly must not be reachable in this stage.

## Staged release — confirmed 2026-10-30

The user accepted releasing X focal-post/long-article extraction first, with an
explicit notice that subsequent author replies are excluded. Full X and Threads
thread assembly is deferred, not an automatic fallback or a successful complete
article. This section and REQ-1–4 below govern this release. The original sample
audit and design below remain historical/deferred requirements; their complete
thread UI flows are not shipped or represented as accepted.

Agent-decided — Make this release's X behavior fixed and explicit in its saved
content and summary, without adding an experimental full-thread configuration
switch. Unverified complete-thread collection must be unreachable from normal,
headed and paginated production entry points. Retain its isolated regression
coverage for later work. Threads post processing fails clearly as deferred;
completed-note skips remain unchanged. Do not silently downgrade a failed thread.

## Requirements
REQ-1 — Complete specified X post or long article
  WHEN an X/Twitter status URL is processed, the scraper shall save only the verified complete body belonging to that exact target, including source line breaks, authored reference URLs and owned rich article content. First and middle URLs intentionally produce their respective posts, not equivalent whole-thread bodies. Unknown conversation extent must not block a verified focal body → Acceptance #1

REQ-2 — Explicit delivered scope and exclusions
  WHEN a focal X post succeeds, both its content note and any generated summary shall visibly state that only the specified post was extracted and author follow-up replies are not included; retain its target source link, exclude other posts, reader replies, quoted bodies and recommendations, and do not claim full-thread success. Full X/Threads thread collection is unavailable in production entry points for this stage → Acceptance #2

REQ-3 — Honest failures and preserved files
  IF target identity or complete focal body cannot be established, or a Threads post is processed in this deferred stage, THEN report extraction failure, do not use legacy social cache or generic page extraction to bypass the scope checks, start no summary and preserve existing body and summary files. Login/verification pages, previews, mismatched targets and empty bodies are never successful articles → Acceptance #3

REQ-4 — Existing extraction and retries
  WHEN extraction runs, retain headless-first and headed retry behavior, headed_on_block effects, standalone/long article and ordinary/YouTube extraction, completed-item skipping and explicit forced refresh. Verification uses isolated settings, profile, executable and outputs; installing this stage does not bulk rewrite existing notes → Acceptance #4

## Active staged UI flows

| Case | What the user does | What they see |
|---|---|---|
| X first, middle or standalone post | Run the existing URL command with summary disabled | Only the specified complete post is saved, with its source and a visible notice excluding author follow-up replies. First and middle posts remain distinct. Success 1, Failed 0, exit 0. No new switch is required. |
| X long article | Process a status containing an owned rich article | Its full owned body, title, formatting and image references are retained, with the same single-post notice. Quoted articles and other posts are not expanded. |
| Threads deferred | Process a Threads post through normal, headed or paginated extraction | Clear unsupported/deferred extraction failure, no content or summary replacement, no summary request. Full-thread collection cannot be enabled through existing domain rules. |
| Missing or unverified focal body | Process an unavailable, wrong-target, preview-only, login or verification page | Failure with no saved replacement or summary request; a generic readable page or old social cache cannot turn it into success. |
| Retry and verification restriction | Use headless or explicit headed settings, or encounter a recognized challenge | Existing single headed retry and explicit headed behavior remain. headed_on_block false prevents headed retry for recognized challenges; ordinary failures keep the existing retry policy. |
| Summary enabled | Process a verified focal post with summary configured | The summary receives only verified focal content and a scope instruction; the saved summary visibly excludes follow-up replies independently of generated prose. Summary-service failure retains existing Partial behavior. |
| Existing notes and force | Run normal batch skip, explicit URL/forced refresh, or resume content without summary | Completed items skip as before; requested social refresh validates live scope before writing. Failure preserves both old files. No bulk rewrite, rename or migration. |
| Batch, watch, dry run and empty input | Run existing processing entry points | Existing result buckets, continuation after an individual failure, retries on later watch cycles, no-item and no-write dry-run behavior remain. |
| Other extraction | Process ordinary articles or YouTube | Existing extraction, original-language subtitle validation and pagination/retry behavior remain. |
| Isolated verification | Check the staged behavior | Fresh build and actual CLI/source checks use temporary outputs, executable and isolated browser session; production config, notes and process are unchanged. Original eight-group full-thread failures remain recorded as deferred. |

## Design decisions

- user-decided — Ship X specified-post/long-article extraction first; exclude
  author follow-up replies explicitly. Defer full X/Threads thread delivery.
- agent-decided — Use existing verified `socialSnapshot` target ownership and
  complete-body checks, not the older generic single-post fallback. Source
  reply counts and thread extent are irrelevant to the smaller delivered scope.
- agent-decided — Return a target-bound, versioned single-post result and validate
  it at the pipeline save boundary. Visible content and saved summary scope
  labels do not depend on the summary model following instructions.
- agent-decided — Keep the experimental complete-thread code inaccessible from
  production entry points, including domain pagination rules and headed retry.
  Preserve its regression tests for later work without claiming its original
  acceptance conditions pass.
- agent-decided — Reject incomplete focal bodies, unavailable/mismatched targets,
  login and verification pages; preserve existing body/summary files and start
  no summary on extraction failure. Cached social text cannot bypass validation.
- agent-decided — Preserve input URL storage identity, completed-item skipping,
  explicit/forced refresh, summary Partial handling and the existing retries.
  No automatic bulk rewrite, file migration, new account, paid API or login.
- agent-decided — Keep development/acceptance browser sessions and outputs
  isolated. Permanent evidence contains sanitized fixtures and public identities,
  URLs, counts and hashes; no credentials or full captured source articles.

## Retained sample groups

These groups came from the earlier complete-thread audit. For this stage, X
first/middle URLs test distinct focal bodies; Threads URLs test explicit deferred
failure. They no longer imply complete-thread acceptance. Original full-thread
requirements and failed results remain in the earlier committed spec/report and
retained evidence; the staged release does not resolve or mark those failures
as passing.

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

- Keep full-thread processing enabled experimentally: rejected for this release
  because unresolved narrative membership can admit unrelated author replies.
- Treat a failed complete thread as a single-post success: rejected because the
  delivered article would silently change scope. This stage selects focal scope
  before extraction and marks it explicitly.
- Defer all X support: unnecessary; isolated source experiments prove focal plain
  and rich bodies can be read in authenticated headless and headed Chrome.
- Add a full-thread opt-in or a second crawler/API: deferred; neither is needed
  to deliver the accepted smaller scope, and neither resolves narrative proof.
- Automatically rewrite or merge earlier notes: outside scope. Normal explicit
  reprocessing retains existing per-URL replacement behavior after validation.

## Verification boundaries

Independent staged acceptance must build from a clean committed copy, compare
actual X focal output to original target source across the retained plain/rich
and first/middle cases, verify excluded posts and durable scope notices, exercise
failure preservation and headed/pagination guards, and run the active UI flows.
Existing complete-thread research is context, not a staged acceptance pass.
All production config, browser data, notes, executable and running processes stay
unchanged. Authenticated source availability does not guarantee logged-out or
universal site support; unavailable or unverified cases remain honest failures.
