# Full-content extraction of multi-page (paginated) articles

Feasibility brief — 2026-06-15. Status: discovery (no code yet).

## Problem
(Axis 1 — JTBD) Today rlss extracts **one URL = one page**. Many news/long-form
sites split an article across N pages; rlss currently captures only page 1, so the
summary is built from a fraction of the article. The job: *when an article is
paginated, fetch and stitch together all its pages so the summary reflects the whole
piece, not just the first screen.* Confirmed by three real examples the user hits
(itmedia, Yahoo!ニュース, 日経中文網) — all paginate, all currently lose pages 2..N.

## Users
(Axis 2) kouko, reading JP/CN/EN news that commonly paginates. Job story: *When I save
a multi-page article (e.g. an itmedia 5-page feature), I want rlss to summarize the
complete article, so the summary isn't truncated to page 1.* Runs the existing
extract→summarize flow; the gap is invisible today (a page-1-only summary still "looks"
complete), so silent truncation is the real harm.

## Smallest End State
(Axis 3) Per-domain opt-in "follow the next-page link and concatenate" extraction.
- Extend `DomainRule` with `NextPageSelector string` (CSS selector for the next-page
  anchor on that site) + optional `MaxPages int` (safety cap, default e.g. 10).
- New `Browser.ExtractPaginated(startURL, js, nextSelector, maxPages, …)`: reuse ONE
  chromedp context; loop = navigate → `window.extractArticle()` (existing Defuddle JS) →
  read `document.querySelector(nextSelector)?.href` → if present, not yet visited, and
  page count < maxPages, navigate to it and repeat; else stop. Concatenate the per-page
  markdown bodies with a separator. Returns combined markdown.
- `pool.ExtractURL`: when the matched `DomainRule` has a `NextPageSelector`, route to
  `ExtractPaginated`; otherwise the existing single-page `Extract` (zero behavior change
  for every site without a rule).
- v1 is **explicit per-domain selector only** — pagination activates solely for a domain
  whose `DomainRule` sets a `NextPageSelector`. itmedia is covered by configuring its
  selector value explicitly as `a[rel=next]` (it emits `rel=next`); no implicit auto-detect.
  (Implicit `rel=next` auto-default is deferred — see Out of Scope.)

The 3 example sites map cleanly:
| Site | next-page mechanism | selector / note |
|---|---|---|
| itmedia | `news096.html → _2…_5.html` | `link[rel=next]` — itmedia marks next as `<link rel="next">` in `<head>`, NOT `<a rel=next>` (verified by integration smoke: single-page 9801 → paginated 46655 chars across 5 pages) |
| Yahoo!ニュース | `?page=2,3` + 「次へ」 | per-domain selector for the 「次へ」 pager anchor |
| 日経中文網 | `?start=0,1,…` + `.article_pagination_next` | `.article_pagination_next a` (also has a `?print=1` all-in-one view — see Out of Scope) |

## Current State Evidence
- **Forward** (extract path): `internal/pipeline/runner.go:222` `p.pool.ExtractURL(item.URL)`
  → `internal/extract/pool.go:32` `ExtractURL` → `internal/extract/browser.go:61`
  `Extract` = `chromedp.Navigate(url)` → inject Defuddle JS → `window.extractArticle()`
  returns the **current page's** markdown only.
- **Reverse** (per-domain config / SSOT): `internal/config/config.go:252` `DomainRule{Domains,
  Headed, GoogleAccount, ChromeProfile}`; matched by `internal/extract/domain.go:10`
  `MatchDomainRules`; held in `ExtractConfig.DomainRules` (`config.go:268`). This is the
  established per-domain extension point — pagination config is a peer field here.
- **Error**: `ExtractURL` returns `(markdown, err)`; `runner.go:230` already does a headed
  retry on anti-bot blocks. Paginated fetch inherits this for page 1; pages 2..N should
  fail soft (a failed later page → keep what we have + log, don't lose the whole article).
- **Data**: Defuddle (`internal/extract/defuddle.go`, `GetDefuddleJS`) operates on ONE
  DOM; `window.extractArticle()` (browser.go:73) returns markdown for the loaded page. It
  has **no** multi-page awareness — the loop must live in rlss.
- **Boundary**: `Browser.Extract` creates a fresh chromedp context per call
  (`browser.go:62`). Pagination wants ONE context reused across sequential navigations
  (cheaper; lets `querySelector` read the next-link from the just-rendered DOM) — a new
  method rather than looping `Extract`.

## Decision
Build per-domain, next-link-following pagination (route **B** — the experiment-confirmed
workhorse): a new `ExtractPaginated` browser method driven by a `NextPageSelector` (+
`MaxPages`) field on the existing `DomainRule`. Concatenate page markdowns. Single-page
sites and any site without a rule are completely unchanged. Do NOT build generic
Mercury-style next-link heuristic scoring, per-site print-view shortcuts (route C), or
URL-template enumeration in v1.

**Refined scope (post-experiment, 2026-06-15):**
- Route C ("prefer single-page/print/view-all") covers only 1/3 of the user's real sites
  (nikkei `?print=1`); **itmedia + Yahoo have NO view-all** (`?page=all` → 404, canonical
  self) and REQUIRE next-following. So B is the mandatory core; C is deferred.
- `rel="next"` is a **weak signal** — Google deprecated it (2019) and adoption is spotty —
  so v1 does NOT auto-detect it. Per-domain explicit selectors are the only mechanism;
  itmedia simply uses `a[rel=next]` as its configured selector *value*. (Auto-defaulting
  to `rel=next` when no rule is set is deferred to Out of Scope — it adds mis-fire risk
  for marginal benefit since the one site that has it is trivially configured explicitly.)
- nikkei is **JS client-side pagination** (curl returns identical HTML for `start=0/1/2`);
  rlss uses chromedp (runs JS) + Defuddle, so nikkei likely already extracts fully with NO
  pagination handling — verify with chromedp during implementation before adding a nikkei rule.
- **Testability**: the extract package's chromedp/browser path is NOT unit-tested today
  (only pure logic is — `defuddle_test`, `domain_test`, `profile_test`). Follow that: extract
  the pagination *decision* logic into pure, unit-tested helpers (next-URL resolve / absolute-ize,
  visited-set + maxPages stop condition, page-markdown join); the `ExtractPaginated` chromedp
  loop itself is build-verified + integration/manual, like the existing `Browser.Extract`.

## Alternatives Considered
(Axis 4 — WebSearch EN+JA, 2026-06)

1. **Next-link following** (Postlight/Mercury Parser `next_page_url` model) — load page,
   extract, find the next-page link, repeat; stop when none. EN: Mercury Parser exposes
   `next_page_url` + `total_pages`/`rendered_pages` and a `fetchAllPages` option. JA:
   scraping guides describe the same "「次へ」 を辿る" loop. → **chosen** (per-domain
   selector + `rel=next` default). Fits rlss's browser model + DomainRule pattern; all 3
   sites have a discoverable next link.
2. **Per-domain URL-template enumeration** (`_{n}.html` / `?page={n}` / `?start={n}`, stop
   on 404/duplicate) — more stable than CSS when a site has a clean numeric scheme, but
   needs a per-site template AND a robust last-page stop condition, and still per-domain
   config. Folded in as a possible future selector-alternative; not v1.
3. **Generic heuristic next-link detection** (Mercury's link-scoring: anchor text "next"/
   「次」/page numbers near content) — zero per-site config, but complex to port and
   mis-fires (next-article vs next-page). Rejected for v1: too much machinery for a 3-site
   need; per-domain selectors are deterministic.
4. **Per-site print / single-page view** (日経 `?tmpl=component&print=1`, some sites
   `?page=all`) — one fetch gets everything, no loop. Great when it exists, but
   site-specific and not universal (itmedia/Yahoo lack a clean one). Deferred (could be a
   `SinglePageURL` template per domain later).

Sources: Postlight/Mercury Parser README (`next_page_url`, `fetchAllPages`, `total_pages`)
(EN); Octoparse「ページ番号でページネーションを処理する方法」/ BeautifulSoup 複数ページ
スクレイピング guides (JA) — both split pagination into "next-button follow" vs
"page-number enumerate", matching alternatives 1 & 2.

### Deeper industry research (2026-06, EN+JA) — refines the above

Five real-world approaches and two trend signals that change the recommendation:

- **A. Heuristic next-link scoring** — Arc90 Readability (2010) `nextPageLink` scored
  candidate links by URL pattern + anchor text ("next"/"›"/page numbers) + link density +
  content proximity. Mercury/Postlight inherited it (`fetchAllPages`). BUT Mozilla
  Readability.js (Firefox Reader View) **dropped multi-page joining** — it's single-doc now.
  Zero per-site config; mis-fires (next-article vs next-page); complex to port.
- **B. Crowd-sourced per-site selector DB** — **AutoPagerize / uAutoPagerize SITEINFO**
  (JP origin, 1000+ sites incl. JP news): community XPath per site for next-link +
  content element. JP-strong → likely already covers itmedia/Yahoo. Accurate for covered
  sites with no per-user config; depends on DB freshness; format is browser-tool-oriented.
- **C. Single-page / view-all / print via canonical** — **trafilatura's explicit advice:
  "for multi-page articles, follow canonical links or force the print-view URL."** SEO best
  practice canonicalizes a paginated series to a view-all page. One fetch, no stitching, no
  boilerplate-dedup, no mis-fire — *most robust when it exists* (日経 has `?print=1`; many
  sites have `?page=all` or `rel=canonical` → single page). Not universal.
- **D. Browser-native reader concat** — Safari Reader historically auto-stitched multi-page
  (now degraded/removed per user reports); Firefox Reader View never did. Standalone
  extensions PageZipper / AutoPagerize / Page One do it at read-time. Not a server/library route.
- **E. Per-domain URL-template enumeration** — alternative 2 above.

Trend signals (temper the design):
- **`rel=next`/`rel=prev` is a weak signal now**: Google **deprecated it for indexing in
  March 2019** (low correct adoption); only Bing still treats it as a hint. Still valid HTML
  (itmedia emits it) but spottily present/maintained — do NOT lean on `rel=next` as the
  primary detector.
- **Pagination is in decline** (reader backlash; shift to single-page + infinite scroll), so
  this is increasingly a *legacy-site* problem — though JP news (itmedia etc.) still does it.

Refined take: the extraction community's most robust route is **C (prefer a detectable
single-page/print/canonical version)** first, falling back to **A/B (follow the next link)**.
For the user's JP/CN sites specifically, **B (AutoPagerize SITEINFO)** is the proven precedent
and may already cover them. This argues for: try single-page/canonical → else per-domain
next-selector (optionally seeded from AutoPagerize SITEINFO) → `rel=next` only as a last hint.

Research sources (each labeled by language):
- (EN) Mozilla/Arc90 Readability heuristics overview — joyboseroy.medium.com; arc90-readability (GitHub masukomi)
- (EN) trafilatura docs + "for multi-page, follow canonical / force print view" — trafilatura.readthedocs.io; htdocs.dev crawler comparison
- (EN) Google rel=next/prev deprecation (2019-03) — developers.google.com/search/blog; yoast.com
- (EN) View-All + canonical best practice — amsive.com; lumar.io; en.wikipedia.org/wiki/Canonical_link_element
- (EN) Safari Reader multi-page concat + PageZipper — howtogeek.com
- (JA) AutoPagerize / uAutoPagerize SITEINFO (1000+ JP sites) — weekly.ascii.jp; untic.blog; toyger.co.jp
- (JA) 「複数ページにまたがる記事を1ページにまとめる」— lifehacker.jp; itmedia.co.jp/bizid

## What Becomes Obsolete
(Axis 5) Nothing removed — additive (single-page extraction stays the default path). The
additive-ness is bounded: it only activates for domains the user explicitly configures (or
that expose `rel=next`), so it can't regress existing single-page behavior.

## Out of Scope (v1)
- **Implicit `rel=next` auto-default** (paginating sites with no `DomainRule`): deferred —
  weak/deprecated signal + mis-fire risk; the one example site (itmedia) is covered by an
  explicit `next_page_selector: "a[rel=next]"`. Revisit if zero-config paging is wanted.
- Generic/heuristic next-link detection (no per-site config) — alternative 3.
- Per-site print/all-in-one-view shortcut (`?print=1`, `?page=all`) — alternative 4.
- URL-template page enumeration — alternative 2.
- De-duplicating repeated headers/footers/boilerplate across concatenated pages beyond
  what Defuddle already strips per page.
- Infinite-scroll / "load more" (JS-appended) articles — different mechanism than discrete
  paginated URLs.

## Open Questions
1. ~~Detection strategy~~ — RESOLVED (user chose route B): explicit per-domain
   `NextPageSelector` only; implicit `rel=next` auto-default and URL-template enumeration
   are deferred (Out of Scope).
2. Page-join separator + whether to re-strip duplicated nav/boilerplate across pages.
3. `MaxPages` default value (proposed 10) and behavior on a mid-sequence page failure
   (proposed: keep pages fetched so far, log, stop).
