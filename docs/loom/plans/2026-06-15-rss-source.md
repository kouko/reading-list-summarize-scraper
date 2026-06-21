# Plan: RSS subscription as a new article source

Source brief: docs/loom/specs/2026-06-15-rss-source.md
Total tasks: 4
Critical-path depth: 2 (≤5)  ← longest Dependencies chain: (Task 1 or 2) → Task 3
Execution order: parallel-where-possible (Tasks 1 & 2 independent; 3 & 4 follow)
Plan-document-reviewer verdict: PASS (2026-06-15, 14/14; amended post-PASS per Notes — schema-safe)

## Task 1 — RSS config struct
- Description: Add `RSSConfig{Enabled bool; Feeds []string; Count int}` and an `RSS RSSConfig` field (`yaml:"rss"`) on `Config`. Default `Count` to 5 when 0 is applied at the consumer (source constructor), not in DefaultConfig — matches rlss's existing "defaults applied at use site" convention (cf. summarizer.go cooldown defaults). No DefaultConfig change required.
- Module: internal/config
- Files touched: internal/config/config.go, internal/config/config_test.go
- Context paths:
  - internal/config/config.go  (FallbackStrategyConfig / SafariConfig pattern for struct + yaml tags)
  - internal/config/config_test.go  (TestLoad_FallbackStrategy as the parse-test template)
- Acceptance:
  - RED: TestLoad_RSSConfig — load YAML with `rss: {enabled: true, count: 5, feeds: [https://a/feed]}`; fails to compile/parse (fields undefined).
  - GREEN: parsed `cfg.LLM`-sibling `cfg.RSS.Enabled==true`, `cfg.RSS.Count==5`, `len(cfg.RSS.Feeds)==1`.
- Dependencies: none
- Independent: true
- Brief item covered: "config.go: `RSS RSSConfig` with `Enabled bool` + `Feeds []string` (+ a backfill bound)"

## Task 2 — RSSSource (parse + map + newest-N)
- Description: Add `internal/source/rss.go`: `RSSSource{feeds []string; count int}` with `NewRSSSource(feeds []string, count int)`, `Name() "rss"`, and `Fetch()`. Factor the entry→ReadingItem mapping + per-feed newest-N (sort by PublishedParsed desc, cap to count; count<=0 → default 5) into a pure helper `feedToItems(feed *gofeed.Feed, count int) []ReadingItem` (mapping: Title, URL=item.Link, DateAdded=item.PublishedParsed, PreviewText=item.Description, Source="rss") so it's testable offline via `gofeed.NewParser().ParseString(sampleXML)`. `Fetch()` = ParseURL per feed → feedToItems → aggregate; per-feed parse error logs + continues (mirrors fetchAndFilter tolerance). Add `github.com/mmcdole/gofeed` dependency.
- Module: internal/source
- Files touched: internal/source/rss.go, internal/source/rss_test.go, go.mod, go.sum
- Context paths:
  - internal/source/types.go  (Source interface + ReadingItem fields + DeduplicateByURL)
  - internal/source/manual.go  (smallest existing Source impl as shape reference)
- Acceptance:
  - RED: TestFeedToItems_MapsAndCapsNewestN — parse a sample RSS XML (5 entries) via ParseString, feedToItems(feed, 3) → 3 newest items, mapped fields correct (Title, URL=Link, DateAdded=PublishedParsed, Source="rss"); plus TestFeedToItems_CountZeroDefaults (0 → default cap). Fails: rss.go absent.
  - GREEN: both pass; `go build ./...` resolves gofeed.
- External surfaces: github.com/mmcdole/gofeed (third-party RSS/Atom/JSON-Feed parser — no stdlib equivalent for lenient multi-format feed parsing); net/http (gofeed ParseURL default client).
- Dependencies: none
- Independent: true
- Brief item covered: "internal/source/rss.go: RSSSource ... Fetch() parses each feed with gofeed, maps each entry → ReadingItem"; "Resolved Decisions: per-feed newest-N cap (default N=5)"

## Task 3 — Wire RSSSource into fetchAndFilter
- Description: In `cmd/rlss/process.go fetchAndFilter`, add `if cfg.RSS.Enabled { sources = append(sources, source.NewRSSSource(cfg.RSS.Feeds, cfg.RSS.Count)) }`, mirroring the existing Safari/Chrome `Enabled`-gated append. RSS items then flow through the existing per-source fetch/tolerate-error loop, DeduplicateByURL, applyFilters, ProcessBatch unchanged.
- Module: cmd/rlss
- Files touched: cmd/rlss/process.go
- Context paths:
  - cmd/rlss/process.go:151-225  (fetchAndFilter source-list assembly + per-source fetch loop)
- Acceptance:
  - RED: `go build ./...` is the diagnostic — referencing cfg.RSS / source.NewRSSSource before Tasks 1-2 land fails to compile; pre-wiring, RSS feeds are never fetched.
  - GREEN: `go build ./...` succeeds; with `rss.enabled: true` + feeds configured, a run fetches RSS items (logged "fetched items" source=rss). cmd/ has no test harness (no test files) — verified by build + the tested Fetch unit from Task 2; wiring is config-gated delegation mirroring the existing Safari/Chrome pattern (trivial-delegation TDD exemption).
- Dependencies: Tasks 1, 2 complete first
- Independent: false
- Brief item covered: "cmd/rlss/process.go fetchAndFilter: if cfg.RSS.Enabled { sources = append(..., source.NewRSSSource(...)) }"

## Task 4 — Document rss config in config.example.yaml
- Description: Add a commented `rss:` block to config.example.yaml (enabled / feeds list / count) documenting the new source and the newest-N backfill semantics, mirroring the existing safari/chrome example blocks.
- Module: docs (config.example.yaml)
- Files touched: config.example.yaml
- Context paths:
  - config.example.yaml  (existing safari:/chrome: example blocks for style)
- Acceptance:
  - RED: grep for `rss:` / `feeds:` in config.example.yaml returns nothing (doc absent).
  - GREEN: `rss:` block present with enabled/feeds/count keys and a one-line newest-N explanation; field names match Task 1's yaml tags.
- Dependencies: Task 1 completes first
- Independent: false  # doc-mirrors-code: field names must match Task 1's yaml tags (see Notes)
- Brief item covered: Smallest End State "config list of feed URLs" — user-facing documentation of the new config surface.

## Notes
- Execution: depth-2, 4 tasks. Tasks 1 & 2 are independent leaves (disjoint files: internal/config vs internal/source+go.mod, no shared symbol). Tasks 3 (cmd wiring) and 4 (doc) both depend on earlier tasks and are sequential.
- Out of scope (from brief): conditional GET / ETag, seen-GUID store, feed-content reuse, OPML / auto-discovery / feed-management subcommands, per-feed overrides, new scheduler.
