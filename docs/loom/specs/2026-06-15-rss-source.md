# RSS subscription as a new article source

Feasibility brief — 2026-06-15. Status: discovery (no code yet).

## Problem
(Axis 1 — JTBD) Today rlss only ingests articles the user **manually saved** to a
Safari/Chrome reading list (pull model). The job behind "add RSS": *automatically
discover new articles from sources I follow (blogs / news / newsletters) and run them
through the existing extract→summarize→Obsidian pipeline, without manually bookmarking
each one* — i.e. add a **push/subscription** input channel alongside the existing
manual one. Confident read from session context; stated as committed interpretation.

## Users
(Axis 2) Single operator (kouko), macOS, runs the `rlss` CLI, output lands in Obsidian.
Job story: *When a blog I follow publishes a new post, I want it summarized into my vault
automatically on the next run, so I don't have to manually save every article.* Runs
periodically (existing `--watch` mode or an external cron). Constraint surfaced below:
must not drown in history when first subscribing to a feed.

## Smallest End State
(Axis 3) One new `RSSSource` implementing the existing `Source` interface, driven by a
config list of feed URLs, registered in `fetchAndFilter` next to Safari/Chrome, run under
the **existing** watch loop. No new pipeline, no new dedup state.

- `internal/source/rss.go`: `RSSSource{feeds []string}` → `Fetch()` parses each feed
  with `gofeed`, maps each entry → `ReadingItem{Title, URL: item.Link,
  DateAdded: item.PublishedParsed, PreviewText: item.Description, Source: "rss"}`.
- `config.go`: `RSS RSSConfig` with `Enabled bool` + `Feeds []string` (+ a backfill
  bound — see Open Questions).
- `cmd/rlss/process.go fetchAndFilter`: `if cfg.RSS.Enabled { sources = append(..., source.NewRSSSource(cfg.RSS.Feeds)) }`.
- Ongoing "only new" is **free**: `ProcessItem` step 1 skips any URL whose summary
  already exists (via `FileIndex`, before extraction), so re-polling the whole feed each
  cycle never re-summarizes. `DeduplicateByURL` handles within-batch dupes.

## Current State Evidence
- **Forward** (entry → pipeline): `cmd/rlss/process.go:151` `fetchAndFilter` builds
  `[]source.Source` (gated by `cfg.X.Enabled`), fetches each (`:201`), dedupes
  (`:219`), filters (`:222`) → `ProcessBatch` (`process.go:83`).
- **Reverse** (source contract / SSOT): `internal/source/types.go:14` `Source` interface =
  `Name() string` + `Fetch() ([]ReadingItem, error)`; `ReadingItem` fields at `types.go:5`.
  Three impls today: `safari.go:118`, `chrome_leveldb.go:36`, `manual.go:15`. RSS = a 4th
  peer; no existing code changes, only an addition + one `append` site.
- **Error**: `fetchAndFilter:204` already tolerates per-source failures (logs, `continue`).
  A dead/4xx feed degrades gracefully without breaking other sources — RSS inherits this.
- **Data**: dedup/skip is keyed on `output.SHA8(item.URL)`; `ProcessItem:170` skip-if-
  `SummaryExists` runs **before** extract (`runner.go:170,188`). So feed re-polling is cheap.
- **Boundary**: `pipeline.Watch` (`watch.go:22`) already re-fetches via a `FetchFunc`
  every `Watch.Interval` minutes, rebuilding the index each cycle — RSS subscription is a
  polling model that maps 1:1 onto this; no scheduler work needed.

## Decision
Build `RSSSource` as a 4th `Source` impl parsed by `gofeed`, config-driven, wired into the
existing fetch→dedup→filter→batch flow and run under existing watch mode. Bound first-
subscribe backfill with a **stateless max-age** (default: only entries published in the
last N days). Always extract the article URL via Defuddle as today (do NOT trust feed-
provided `content:encoded`) for consistent output quality. Do NOT add conditional-GET,
a seen-GUID store, or a feeds-management UI in v1.

## Alternatives Considered
(Axis 4 — WebSearch EN+JA, 2026-06)

**RSS parsing library**
1. **gofeed (mmcdole)** — de-facto standard. RSS 0.9x–2.0 + Atom 0.3/1.0 + JSON Feed
   1.x, one unified model, lenient parsing of broken XML/dates. Used widely; JA tutorials
   converge on it. → **chosen.**
2. ungerik/go-rss / hand-roll `encoding/xml` — fewer deps but RSS-2.0/Atom only, no JSON
   Feed, no broken-feed tolerance. Rejected: re-implements what gofeed solved.

**"Only new" / dedup**
1. **Reuse `FileIndex` skip** (summary-exists ⇒ skip before extract) — zero new state,
   already battle-tested. → **chosen** for ongoing dedup.
2. Conditional GET (ETag/If-None-Match, Last-Modified/If-Modified-Since) — saves
   bandwidth (304 Not Modified) on unchanged feeds. Real best practice, but an
   optimization, not correctness. → deferred (can pass a custom client to gofeed later).
3. Persistent seen-GUID store — needed only for "process strictly entries published after
   subscribe". Heavier; the max-age bound covers the same first-run concern statelessly.

**Backfill bound** (the one genuine fork — see Open Questions)
- max-age (last N days) — stateless, simple, default. ‖ per-feed newest-N cap — reuses the
  existing per-source count idea. ‖ mark-all-seen on first fetch — cleanest "from now on"
  but needs the seen-store from above.

## What Becomes Obsolete
(Axis 5) Nothing removed — purely additive (a new input channel, not a replacement for
manual/reading-list sources). Additive is justified here (genuinely new capability), but
it means the change must stay small (Smallest End State above) to avoid scope creep.

## Out of Scope (v1)
- Conditional GET / ETag bandwidth optimization.
- Persistent per-feed seen-GUID state.
- Using feed-embedded full content instead of Defuddle extraction.
- OPML import, feed auto-discovery, add/remove-feed CLI subcommands, per-feed filters.
- Any new scheduler (reuse `--watch` / external cron).

## Resolved Decisions
1. **Backfill bound** → per-feed **newest-N** cap (default N = 5), applied after sorting a
   feed's entries by published date desc. Stateless; bounds first-subscribe history and
   every subsequent poll (already-summarized ones skip cheaply via FileIndex, so steady
   state only ever processes genuinely new entries within the newest-N window).
3. **Dependency** `github.com/mmcdole/gofeed` accepted.

## Open Questions
2. Per-feed overrides (own count / filter) — global `count` only in v1; per-feed override
   deferred unless needed.
