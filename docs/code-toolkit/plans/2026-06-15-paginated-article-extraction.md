# Plan: Paginated article extraction (route B — per-domain next-link following)

Source brief: docs/code-toolkit/specs/2026-06-15-paginated-article-extraction.md
Total tasks: 6
Critical-path depth: 3 (≤5)  ← longest chains: T1→T3→T5 and T2→T4→T5
Execution order: parallel-where-possible (T1 & T2 independent leaves; rest follow)
Plan-document-reviewer verdict: PASS (2026-06-15, round 2, 14/14)

## Task 1 — DomainRule pagination fields
- Description: Add `NextPageSelector string` (`yaml:"next_page_selector"`) and `MaxPages int` (`yaml:"max_pages"`) to `DomainRule`. No default in struct; MaxPages default (e.g. 10) applied at the consumer (ExtractPaginated), per rlss's defaults-at-use-site convention.
- Module: internal/config
- Files touched: internal/config/config.go, internal/config/config_test.go
- Context paths:
  - internal/config/config.go  (DomainRule struct ~:252; TestLoad_FallbackStrategy as parse-test template)
- Acceptance:
  - RED: TestLoad_DomainRulePagination — load YAML `extract.domain_rules` with `next_page_selector: "a[rel=next]"` + `max_pages: 5`; fails to compile (fields undefined).
  - GREEN: parsed `cfg.Extract.DomainRules[0].NextPageSelector == "a[rel=next]"` and `.MaxPages == 5`.
- Dependencies: none
- Independent: true
- Brief item covered: "Extend `DomainRule` with `NextPageSelector string` (+ `MaxPages int`)"

## Task 2 — Pure pagination-control helpers
- Description: New `internal/extract/pagination.go` with three pure (no-I/O) helpers: `resolveNextURL(baseURL, href string) (string, bool)` (absolute-ize href against base; reject empty / `javascript:` / pure `#fragment` → ok=false); `shouldFollowNext(nextURL string, visited map[string]bool, pageCount, maxPages int) bool` (false when nextURL empty, already in visited, or pageCount >= maxPages); `joinPages(pages []string) string` (concatenate page markdowns with a `\n\n` separator, skipping empties).
- Module: internal/extract
- Files touched: internal/extract/pagination.go, internal/extract/pagination_test.go
- Context paths:
  - internal/extract/domain.go  (existing pure-helper style in this package)
- Acceptance:
  - RED: pagination_test.go — TestResolveNextURL (relative→absolute, reject javascript:/empty/fragment), TestShouldFollowNext (stop on empty / visited / maxPages reached; continue otherwise), TestJoinPages (joins with separator, skips empties). Fails: funcs undefined.
  - GREEN: all three test funcs pass; `go test ./internal/extract/` green.
- External surfaces: net/url (stdlib, for resolveNextURL absolute-ization).
- Dependencies: none
- Independent: true
- Brief item covered: "Testability: extract the pagination decision logic into pure, unit-tested helpers (next-URL resolve / absolute-ize, visited-set + maxPages stop condition, page-markdown join)"

## Task 3 — MatchDomainRules returns pagination config
- Description: Extend `MatchDomainRules` to also surface the matched rule's pagination config (return `nextPageSelector string, maxPages int`, or return the matched `*config.DomainRule`). Update existing callers. Keeps the existing headed/profile/googleAccount returns.
- Module: internal/extract
- Files touched: internal/extract/domain.go, internal/extract/domain_test.go
- Context paths:
  - internal/extract/domain.go:10  (MatchDomainRules current signature + callers)
  - internal/extract/domain_test.go  (TestMatchDomainRules)
- Acceptance:
  - RED: TestMatchDomainRules_ReturnsPaginationConfig — a rule with `NextPageSelector`/`MaxPages` set; assert MatchDomainRules returns them for a matching URL and zero-values for a non-match. Fails: signature lacks the new returns.
  - GREEN: test passes; existing TestMatchDomainRules still green.
- Dependencies: Task 1 completes first
- Independent: false
- Brief item covered: "matched by `MatchDomainRules`; pagination config is a peer field here" — surfacing the new DomainRule fields to the extract layer.

## Task 4 — Browser.ExtractPaginated (chromedp loop)
- Description: Add `Browser.ExtractPaginated(startURL, jsCode string, nextSelector string, maxPages int, timeout, waitAfterLoad time.Duration) (string, error)`: reuse ONE chromedp context; loop = Navigate → WaitVisible body → Sleep → inject Defuddle JS → `window.extractArticle()` (page markdown) → read next href via `document.querySelector(<nextSelector>)?.href` → use `resolveNextURL` + `shouldFollowNext` (Task 2) to decide; on continue, navigate next and repeat; on stop, `joinPages` the collected markdowns. A mid-sequence page error keeps pages gathered so far + returns them (fail-soft). maxPages<=0 → default 10.
- Module: internal/extract
- Files touched: internal/extract/browser.go
- Context paths:
  - internal/extract/browser.go:60  (existing single-page Extract — same chromedp.Run/Evaluate pattern to mirror)
  - internal/extract/pagination.go  (helpers from Task 2)
- Acceptance:
  - RED: `go build ./internal/extract/` — references to `ExtractPaginated` (and its use of the Task-2 helpers) fail to compile until defined. (Per this package's convention the chromedp/browser path is NOT unit-tested — only the pure helpers it calls are, in Task 2; cf. existing `Browser.Extract` which has no unit test.)
  - GREEN: `go build ./...` succeeds; ExtractPaginated compiles and calls the Task-2 helpers; pure stop/join/resolve logic is green via Task 2; manual/integration verification against a real paginated site (itmedia) deferred to verification step.
- External surfaces: github.com/chromedp/chromedp (browser automation — same surface the existing Extract already uses; no new dependency).
- Dependencies: Task 2 completes first
- Independent: false
- Brief item covered: "a new `Browser.ExtractPaginated(startURL, js, nextSelector, maxPages, …)`: reuse ONE chromedp context; loop = navigate → extractArticle → read next href → … → concatenate the per-page markdown"

## Task 5 — Route ExtractURL to ExtractPaginated when configured
- Description: In `pool.ExtractURL`, use the pagination config from `MatchDomainRules` (Task 3): when the matched rule has a non-empty `NextPageSelector`, call `ExtractPaginated(url, js, selector, maxPages, …)`; otherwise the existing single-page `Extract` (unchanged default for every site without a rule).
- Module: internal/extract
- Files touched: internal/extract/pool.go
- Context paths:
  - internal/extract/pool.go:32  (ExtractURL — current MatchDomainRules use + Extract call)
- Acceptance:
  - RED: `go build ./internal/extract/` — ExtractURL referencing the new ExtractPaginated + the new MatchDomainRules returns fails to compile until wired. (pool's browser path is integration-only, like Task 4; routing is a thin config-gated branch mirroring the existing headed/profile gating.)
  - GREEN: `go build ./...` succeeds; a configured domain routes to ExtractPaginated, an unconfigured one to Extract (behavior unchanged for the default path); full `go test ./...` green.
- Dependencies: Tasks 3, 4 complete first
- Independent: false
- Brief item covered: "`pool.ExtractURL`: when the matched `DomainRule` has a `NextPageSelector`, route to `ExtractPaginated`; otherwise the existing single-page `Extract`"

## Task 6 — Document pagination in config.example.yaml
- Description: Add `next_page_selector` + `max_pages` keys (commented, with the itmedia `a[rel=next]` example and a 1-line explanation) to the `domain_rules` block in config.example.yaml, mirroring the existing domain_rules example style.
- Module: docs (config.example.yaml)
- Files touched: config.example.yaml
- Context paths:
  - config.example.yaml  (existing extract.domain_rules example block)
- Acceptance:
  - RED: grep for `next_page_selector` in config.example.yaml returns nothing.
  - GREEN: `next_page_selector` + `max_pages` documented under a domain_rules example; key names match Task 1's yaml tags.
- Dependencies: Task 1 completes first  (doc-mirrors-code: key names must match the struct yaml tags — see Notes)
- Independent: false
- Brief item covered: Smallest End State — the per-domain pagination config surface; user-facing documentation of `next_page_selector`/`max_pages`.

## Notes
- Round-1 review (Check 8) flagged an uncovered brief item: an *implicit* `a[rel=next]`
  auto-default. Resolved by tightening the brief — v1 is explicit-per-domain-selector only;
  the implicit auto-default moved to brief §Out of Scope. The plan is unchanged (T5 routes
  only when an explicit `NextPageSelector` is set; T6 documents `a[rel=next]` as the example
  *value* for itmedia's explicit selector) and now fully covers the brief.
- Depth-3, 6 tasks. T1 (config) & T2 (pure helpers) are independent leaves: disjoint files (internal/config vs internal/extract/pagination.go), no shared symbol. T3/T4/T5 are the sequential extract-layer chain; T6 (doc) is sequential after T1 (doc-mirrors-code, kept `Independent: false` though files are disjoint).
- Testability convention (from brief): internal/extract's chromedp/browser path has no unit tests today (defuddle/domain/profile pure-logic only). T2 (pure helpers) + T3 (MatchDomainRules) carry real unit tests; T4 (ExtractPaginated chromedp loop) + T5 (pool routing) are build-verified + integration/manual, matching the existing untested `Browser.Extract`. Live verification against itmedia (5-page, has rel=next) is the integration check after the chain lands.
- Out of scope (brief): route C single-page/print shortcut, heuristic next-link scoring, URL-template enumeration, cross-page boilerplate dedup, infinite-scroll.
