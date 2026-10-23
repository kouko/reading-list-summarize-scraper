# Reading list date order — spec
intent: 2026-10-23-reading-list-date-order@232216c
confirmed-behavior: 2026-10-23 @d4cae13
pre-build-review: not-required — bounded ordering change in the shared CLI collection path; no new flags, external API, authentication, or irreversible data operation

## Requirements
REQ-1 — Newest-first order
  WHEN reading items are listed or processed in a batch, the CLI shall order them by their recorded addition timestamps from newest to oldest; equivalent timestamps retain their previous relative order, and absent timestamps appear last → Acceptance #1
REQ-2 — Combined-source ordering
  WHEN multiple reading sources are enabled, the CLI shall use one timestamp order across the combined deduplicated result rather than group items by source; existing first-occurrence URL deduplication remains unchanged → Acceptance #2
REQ-3 — Ordered limit
  WHEN a maximum item count is configured, the CLI shall select the first qualifying items in newest-first order after the existing unread/date filters, including each watch cycle → Acceptance #3

## Design decision
- agent-decided — stably sort the filtered result in applyFilters before its existing limit, so list, process, and watch share the same ordering without new configuration or helper abstractions.
- agent-decided — compare full recorded time.Time values rather than displayed calendar dates, preserving correct same-day and timezone ordering.
- agent-decided — keep equivalent timestamps stable and place zero timestamps after dated items, avoiding invented dates or additional prompts.
- agent-decided — preserve URL deduplication before sorting; its existing first-occurrence metadata choice remains unchanged, preventing an unrelated source-precedence change.
- agent-decided — completed-item detection stays in the pipeline after selection; sorting does not introduce backfilling from beyond the configured limit.
- user-decided — combine sources and use newest-first ordering before selecting the configured maximum.

## Alternatives considered
- Sort only Chrome source: rejected because the confirmed behavior includes unified ordering across sources.
- Sort each source independently: rejected because concatenation would still give source precedence rather than a shared date order.
- Add a sort direction setting or CLI flag: rejected because only newest-first behavior was requested.
- Move completed-item exclusion before the limit: rejected as a separate behavior change outside the confirmed scope.

## Current state evidence
- Forward: cmd/rlss/list.go:runList and cmd/rlss/process.go:runProcess both consume fetchAndFilter.
- Reverse: cmd/rlss/process.go:fetchAndFilter concatenates Safari, Chrome and optional RSS results, then calls DeduplicateByURL and applyFilters; watch uses the same fetch function.
- Error: cmd/rlss/process.go:applyFilters logs an invalid since date and ignores that filter; source fetch failures retain the existing handling. Ordering adds no new error path.
- Data: internal/source/types.go:ReadingItem.DateAdded is time.Time; Safari and Chrome map recorded addition times into it. DeduplicateByURL retains the first occurrence and its metadata.
- Boundary: internal/pipeline/runner.go:ProcessItem performs skip detection after collection; output file formats, credentials, extraction, summarization, and running processes are unchanged.

## UI flows
| Case | What the user does | What they see |
|---|---|---|
| List success | Lists one or more enabled reading sources | Items appear from newest added to oldest, across sources; displayed dates keep their existing format. |
| Batch success | Processes a batch or lets the next watch cycle run | Processing starts with the newest qualifying items, up to the configured maximum. |
| Equal timestamps | Lists or processes items recorded at the same instant | Those items keep their previous relative order. |
| Missing timestamp | Lists or processes an item without a recorded addition time | The undated item appears after dated items; no date is invented. |
| Existing filters | Uses unread-only or a since date | The same items qualify, but their resulting order is newest first before the maximum is applied. |
| Duplicate URL | Enables multiple sources containing the same URL | The URL appears only once, retaining the original source-selection and timestamp behavior. |
| No qualifying items | Lists or processes an empty result | Existing no-items output remains unchanged; no browser or AI request is added. |
| In progress | Starts a reading list operation | Existing fetch and processing logs remain; sorting adds no new progress interface. |
| Existing error | Uses an invalid since date or encounters a source read failure | Existing warnings and recovery behavior remain; the user corrects the date or source access and reruns. |
| Already completed items | Processes a limited batch whose selected items already have summaries | Those items remain skipped; the batch does not automatically select older replacements beyond the limit. |

The shared result is reused by list, batch, and each watch cycle. The direct single-URL operation remains unaffected. Rerunning a command or the next watch scan recomputes order from currently read data; no persistent queue or file migration is introduced. Currently running binaries require a rebuild and restart to use the new behavior; this change does not stop or replace them automatically.
