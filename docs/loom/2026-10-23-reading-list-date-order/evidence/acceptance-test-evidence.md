# Reading list date order — acceptance test evidence

Tried on 2026-10-23 in a clean detached worktree at 9157f23bd76119990dd1b772d9314305213dce6c.
Runner profile: ordinary inherited; overrides: null; unobservable-main-profile.

## Setup and scope

Commands (all exited 0):

```text
git worktree add --detach /private/tmp/rlss-date-acceptance 9157f23bd76119990dd1b772d9314305213dce6c
GOCACHE=/private/tmp/rlss-go-build go build -o rlss ./cmd/rlss/
GOCACHE=/private/tmp/rlss-go-build go test ./cmd/rlss -run 'TestApplyFiltersDateOrder|TestDateOrderKeepsFirstURLMetadata' -count=1 -v
```

The build follows README's quick source-build procedure using the committed JS bundle. It emitted a nonfatal module stat-cache permission warning; binary execution succeeded. No dependencies installed. Full build with Node was not required by the documented quick-build route.

Focused test output:

```text
--- PASS: TestApplyFiltersDateOrder
    --- PASS: TestApplyFiltersDateOrder/precise-newest-first
    --- PASS: TestApplyFiltersDateOrder/stable-ties-undated
    --- PASS: TestApplyFiltersDateOrder/mixed-sources
    --- PASS: TestApplyFiltersDateOrder/filtered-top-n
    --- PASS: TestApplyFiltersDateOrder/empty
    --- PASS: TestApplyFiltersDateOrder/unlimited-invalid-since
--- PASS: TestDateOrderKeepsFirstURLMetadata
PASS
ok github.com/kouko/reading-list-summarize-scraper/cmd/rlss 0.424s
```

This witness did not run `go test ./...`. Package-tests are not skipped; finalize-review will run the full suite and block acceptance on failure. No important-or-worse dismissed findings were supplied.

The explicit local synthetic Safari fixture is not the user's reading list. Python plistlib generated a root Children list containing a com.apple.ReadingList folder; each leaf has WebBookmarkTypeLeaf, Title, URLString, and ReadingList metadata. URL prefix: https://example.invalid/. Insertion order:

| Title / URL suffix | DateAdded (UTC) | Read |
|---|---|---|
| old | 2026-10-21 08:00 | no |
| tie-first | 2026-10-23 08:00 | no |
| undated | absent | no |
| newest-read | 2026-10-23 11:00 | yes |
| newest | 2026-10-23 10:00 | no |
| tie-second | 2026-10-23 08:00 | no |
| old (duplicate) | 2026-10-24 08:00 | no |

Read means DateLastViewed=2026-10-23 12:00; otherwise absent. Config acceptance.yaml: output_dir=/private/tmp/rlss-date-acceptance/output; safari.enabled=true; safari.plist_path=/private/tmp/rlss-date-acceptance/fixture.plist; chrome.enabled=false; extract.user_data_dir=/private/tmp/rlss-date-acceptance/no-profiles. RSS disabled by default. All fixture/output/config writes were in the clean temporary worktree. No real Safari plist, browser process, root-workspace binary or vault was modified. Batch's existing resolver may read local profile metadata; only synthetic Safari items were fetched. No extraction, paid service or AI call ran.

## 1. A1 — list and batch newest first

Probe: test_cli_synthetic_list_and_batch_newest_first.

```text
./rlss list --config acceptance.yaml
fetched items source=safari count=7
1. [new]   newest-read (safari, 2026-10-23)
2. [new] * newest (safari, 2026-10-23)
3. [new] * tie-first (safari, 2026-10-23)
4. [new] * tie-second (safari, 2026-10-23)
5. [new] * old (safari, 2026-10-21)
6. [new] * undated (safari, 0001-01-01)
Total: 6 items

./rlss process --config acceptance.yaml --no-summary --dry-run --unread-only --since 2026-10-23 --limit 2
processing item index=1 total=2 url=https://example.invalid/newest title=newest
dry-run: would process url=https://example.invalid/newest sha8=30975863
processing item index=2 total=2 url=https://example.invalid/tie-first title=tie-first
dry-run: would process url=https://example.invalid/tie-first sha8=d8cf3e6f
Total: 2; Success: 0; Skipped: 2; Partial: 0; Failed: 0
```

Both exit 0. List initially warns that isolated no-profiles/Local State is absent; this does not prevent Safari collection. Focused precise-newest-first test additionally checks nanoseconds and differing timezones.

## 2. A2 — combined-source order

Probe: test_filters_mixed_sources_global_order. Executed TestApplyFiltersDateOrder/mixed-sources: PASS; inputs Safari at 08:00, Chrome at 10:00, RSS at 09:00; asserted Chrome, RSS, Safari order. TestDateOrderKeepsFirstURLMetadata: PASS; later duplicate metadata did not replace the first item. This is a focused test result, not a live Chrome/RSS integration run. Actual collection deduplicates before filtering at cmd/rlss/process.go:223 and :226. list calls the shared collector at cmd/rlss/list.go:31; batch at cmd/rlss/process.go:74.

## 3. A3 — filter, sort, then limit

Probe: test_cli_filtered_limit_selects_newest.

```text
./rlss list --config acceptance.yaml --unread-only --since 2026-10-23 --limit 2
1. [new] * newest (safari, 2026-10-23)
2. [new] * tie-first (safari, 2026-10-23)
Total: 2 items
```

Exit 0. Same batch output in section 1 and focused filtered-top-n PASS corroborate selection. Watch cycle evidence is in section 5.

## 4. UI — List success

Probe: test_cli_list_dated_items_descend. Section 1 actual output has newest-read, newest, ties, old, undated; displayed calendar date format unchanged. Mixed-source part covered by section 2's executed test, not a fabricated live source.

## 5. UI — Batch success and watch

Probe: test_cli_watch_cycle_uses_ordered_limit. Section 1 batch dry-run ran before synthetic completed records were added. A separate watch.yaml used the same Safari fixture and output, filter.unread_only=true, filter.since="2026-10-23", filter.limit=2, pipeline.dry_run=true, pipeline.skip_existing=true. These settings were in config to survive the existing watch reload. A Python subprocess launched the command, read until `watch: sleeping`, sent SIGINT only to that child, then collected output with a five-second timeout:

```text
./rlss process --config watch.yaml --no-summary --watch
fetched items source=safari count=7
processing item index=1 total=2 url=https://example.invalid/newest title=newest
processing item index=2 total=2 url=https://example.invalid/tie-first title=tie-first
watch: cycle complete report="Pipeline completed in 0s\n Total: 2\n Success: 0\n Skipped: 2\n Partial: 0\n Failed: 0\n"
watch: sleeping minutes=10
watch: signal during sleep, exiting signal=interrupt
exit: 0
```

This was one actual isolated watch cycle, not a long-running real-data session. The two selected items had synthetic completion records by this point and skipped. No extraction or AI occurred. Watch callback calls fetchAndFilter at cmd/rlss/process.go:67; each cycle invokes it at internal/pipeline/watch.go:51 and passes its items to ProcessBatch at :58.

## 6. UI — Equal timestamps

Probe: test_cli_equal_timestamps_keep_relative_order. Section 1 shows tie-first before tie-second; stable-ties-undated focused test passed, including equal instants represented with different timezones.

## 7. UI — Missing timestamp

Probe: test_cli_missing_timestamp_sorts_last. Section 1 places undated last. Existing display prints zero time as 0001-01-01; the change does not invent an addition time or change that display. Focused stable-ties-undated passed for multiple missing timestamps.

## 8. UI — Existing filters

Probe: test_cli_existing_filters_keep_qualifying_items. Section 3 excludes newest-read, old and undated, selects newest then tie-first; executed filtered-top-n test passed.

## 9. UI — Duplicate URL

Probe: test_filters_duplicate_url_keeps_first_metadata. Section 1 fetched seven leaves and displayed six URLs; old retains 2026-10-21 despite its later duplicate dated 2026-10-24. Executed TestDateOrderKeepsFirstURLMetadata verifies a cross-source duplicate and entire retained item metadata.

## 10. UI — No qualifying items

Probe: test_cli_future_since_returns_empty.

```text
./rlss list --config acceptance.yaml --since 2099-01-01
fetched items source=safari count=7
No items found.
./rlss process --config acceptance.yaml --no-summary --since 2099-01-01
fetched items source=safari count=7
no items to process
```

Both exit 0; no browser or AI requests. Focused empty test also passed.

## 11. UI — In progress

Probe: test_cli_operation_keeps_existing_progress. Sections 1 and 5 captured fetched-items, processing-item, dry-run and watch completion/sleep messages. No additional sorting interface appeared.

## 12. UI — Existing error

Probe: test_cli_invalid_date_and_missing_source_keep_recovery.

```text
./rlss list --config acceptance.yaml --since invalid --limit 1
WARN invalid --since date, ignoring value=invalid
1. [new] newest-read (safari, 2026-10-23)
Total: 1 items

./rlss list --config missing.yaml
ERROR source fetch failed source=safari err="open /private/tmp/rlss-date-acceptance/missing.plist: ... no such file or directory"
No items found.
```

Both exit 0. missing.yaml enabled Safari at the nonexistent temporary path and disabled Chrome. Did not provoke real filesystem permission errors or open System Settings. Correct-path reruns already succeeded in sections 1 and 3. Focused unlimited-invalid-since passed.

## 13. UI — Already completed items

Probe: test_cli_completed_selection_does_not_backfill. Added two synthetic summary records in temporary output/example_invalid, named 2026-10-23__30975863__summary.md and 2026-10-23__d8cf3e6f__summary.md, each containing a fixture-only sentence. Then ran an actual non-dry batch with --no-summary; both selected records skip before extraction:

```text
./rlss process --config acceptance.yaml --no-summary --unread-only --since 2026-10-23 --limit 2 --verbose
processing item index=1 total=2 url=https://example.invalid/newest title=newest
DEBUG skipping existing url=https://example.invalid/newest sha8=30975863
DEBUG delay between items seconds=7
processing item index=2 total=2 url=https://example.invalid/tie-first title=tie-first
DEBUG skipping existing url=https://example.invalid/tie-first sha8=d8cf3e6f
Pipeline completed in 7s
Total: 2; Success: 0; Skipped: 2; Partial: 0; Failed: 0
```

Exit 0. tie-second remained beyond the limit, not backfilled. Real skip check is internal/pipeline/runner.go:187, after shared selection. No real summaries were read or modified.
