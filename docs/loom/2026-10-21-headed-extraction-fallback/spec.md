# Headed extraction fallback — spec
intent: 2026-10-21-headed-extraction-fallback@ca2e7cd
confirmed-behavior: 2026-10-22 @9592aa0
pre-build-review: not-required — bounded local extraction recovery; no new authentication, data migration, or external service

## Requirements
REQ-1 — All-site recovery
  WHEN an initially headless extraction fails on any domain, the pipeline shall retry once with a visible browser, except that detected anti-bot verification pages shall honor the existing headed_on_block switch → Acceptance #1
REQ-2 — Single-run recovery
  WHEN the visible-browser retry obtains the target article body, the pipeline shall save that body during the same invocation without a manual rerun → Acceptance #2
REQ-3 — Explicit terminal failure
  IF the final extraction attempt fails THEN the pipeline shall report an extraction failure and produce no new content or summary for that attempt → Acceptance #3
REQ-4 — Correct article content
  IF a response contains only a login, verification, maintenance, or subscription wall instead of the target body THEN the pipeline shall reject it rather than report article success or summarize it; login prompts alongside a real body shall not alone cause rejection → Acceptance #4

## Design decision
- agent-decided — consolidate validation and retry orchestration so empty results, navigation failures, and recognized wall pages follow the same bounded recovery path.
- agent-decided — preserve explicit headed configuration and --headed: already-visible attempts are not retried in the same mode.
- user-decided — preserve headed_on_block and its original effect: false prevents headed retry for detected anti-bot verification pages, true permits it. Navigation errors and empty content remain eligible for generic recovery; this switch does not govern those failures.
- agent-decided — use existing headed_timeout for recovery and preserve pagination settings and the selected browser profile; close the headless browser before reusing its data directory.
- agent-decided — strengthen verification/login recognition using inspected page evidence and fixtures; do not rely on bare Log in text, which also appears beside accessible X articles.
- agent-decided — apply correctness validation to resumable cached content before summarization; invalid cached content must not bypass the checks. Do not delete existing notes.
- user-decided — prioritize correctness; an unavailable target must fail explicitly rather than create a misleading summary. Absolute availability is not guaranteed.

## Alternatives considered
- X-specific headed rule: rejected because the confirmed recovery applies to every domain.
- Always-headed browsing: rejected because successful headless extraction should not open a window.
- Repeated headed retries: rejected because repeated identical attempts add latency without a new recovery strategy.
- Reject every page containing login wording: rejected because observed accessible X pages include login prompts alongside genuine article content.

## Current state evidence
- Forward: cmd/rlss/url.go:runURL and cmd/rlss/process.go:runProcess construct the extraction pool and pipeline.
- Reverse: internal/pipeline/runner.go:ProcessItem calls the pool; internal/extract/pool.go:ExtractURL applies domain rules and pagination.
- Error: the uncommitted internal/pipeline/extraction.go:extractWithFallback handles navigation failure, but existing internal/pipeline/runner.go:isBlockedPage misses the observed `to, err:= human()` challenge.
- Data: internal/extract/browser.go:Extract returns Defuddle Markdown; internal/output/obsidian.go assembles article and summary documents.
- Boundary: internal/pipeline/runner.go:ProcessItem resume path can reuse saved content; existing successful notes are not rewritten or removed by this change. Authentication and AI provider selection are unchanged.

## UI flows
| Case | What the user does | What they see |
|---|---|---|
| Initial success | Processes an article with normal configuration | The article is saved without opening a browser window for retry. |
| In progress / first attempt failed | Processes a page that fails to load, returns no content, or yields a recognized wall page | A warning explains the first failure; a Chrome window opens for one retry, using the configured headed timeout. |
| Retry success | Waits for the retry | The target body is saved in the same run; normal configured summarization can continue. |
| Final error / empty result | Waits for a failed retry | The item is reported failed; no new content or summary is produced for it. The single-URL command exits nonzero; batch output counts the failed item and continues other items. The user can resolve network/access issues and rerun. |
| Login or verification only | Processes a target that returns a wall rather than its body | The response is rejected. Anti-bot verification retries only when headed_on_block is true; other failures follow generic recovery. The user can resolve access in the configured browser and rerun; automatic login is outside this change. |
| Login prompt plus article | Processes an accessible public X article | The genuine article remains eligible for extraction despite incidental login prompts. |
| Already visible | Uses --headed or a headed domain rule | One visible attempt runs; its failure is reported without opening another retry window. |
| Anti-bot switch disabled | Uses headed_on_block: false and encounters a detected verification page | The item fails without a visible retry; no wall content or summary is saved. Navigation errors and empty content can still retry visibly. |
| Anti-bot switch enabled | Uses headed_on_block: true and encounters a detected verification page | One visible retry runs; a remaining wall is reported as failure without saving it as article content. |
| Resume | Processes an item whose saved original lacks a summary | Saved content is validated before use; a recognized cached wall cannot be summarized as the article. Existing files are not deleted. |

Successful recovery returns to the ordinary save/summary path. Terminal failure returns to the caller; batch processing can proceed to the next item. No new interactive login prompt, new CLI flag, or automatic authentication is introduced. This change adds no AI request before valid content exists.
