package extract

import (
	"log/slog"
	"net/url"
	"strings"
)

// defaultMaxPages bounds how many pages a paginated article is followed for
// when a DomainRule sets no (or a non-positive) MaxPages.
const defaultMaxPages = 10

// pageFetcher loads one page and returns its extracted content plus the raw
// href of its "next page" link ("" if none). It is the injectable seam that
// lets the paginate loop be tested without a real browser; the production
// implementation in ExtractPaginated drives chromedp.
type pageFetcher func(url string) (content, nextHref string, err error)

// paginate walks a paginated article: fetch a page, record it, resolve+follow
// its next link, repeat — stopping at no next link, an already-visited URL
// (loop guard), or maxPages (<=0 → defaultMaxPages). It is fail-soft: a page
// that errors mid-sequence stops the loop but keeps the pages already gathered
// (logged via slog.Warn); only a total failure (nothing gathered) returns the
// error. The browser-specific work lives behind the injected fetch.
func paginate(startURL string, maxPages int, fetch pageFetcher) (string, error) {
	if maxPages <= 0 {
		maxPages = defaultMaxPages
	}

	var pages []string
	visited := map[string]bool{}
	current := startURL
	var firstErr error

	for pageCount := 0; ; pageCount++ {
		visited[current] = true

		content, nextHref, err := fetch(current)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			// Fail-soft: keep the pages gathered so far, but surface the
			// failure so a truncated article isn't silently indistinguishable
			// from a genuinely short one.
			slog.Warn("paginated extract: page failed, keeping pages gathered so far",
				"url", current, "pages_gathered", len(pages), "err", err)
			break
		}
		pages = append(pages, content)

		next, ok := resolveNextURL(current, nextHref)
		if !ok || !shouldFollowNext(next, visited, pageCount+1, maxPages) {
			break
		}
		current = next
	}

	combined := joinPages(pages)
	if combined == "" && firstErr != nil {
		return "", firstErr
	}
	return combined, nil
}

// resolveNextURL resolves a "next page" href (possibly relative) against the
// current page's URL into an absolute URL. It returns ok=false for hrefs that
// are not real navigations: empty, javascript:/mailto: schemes, or a bare
// same-page fragment. The fragment is stripped from the result so that the
// visited-set comparison in shouldFollowNext is stable.
func resolveNextURL(baseURL, href string) (string, bool) {
	href = strings.TrimSpace(href)
	if href == "" || strings.HasPrefix(href, "#") {
		return "", false
	}
	switch {
	case strings.HasPrefix(strings.ToLower(href), "javascript:"),
		strings.HasPrefix(strings.ToLower(href), "mailto:"):
		return "", false
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", false
	}
	ref, err := url.Parse(href)
	if err != nil {
		return "", false
	}
	abs := base.ResolveReference(ref)
	abs.Fragment = ""
	return abs.String(), true
}

// shouldFollowNext reports whether the loop should fetch nextURL: it must be
// non-empty, not already visited (loop guard), and within the maxPages bound.
// maxPages <= 0 means "no cap" (callers pass the resolved default).
func shouldFollowNext(nextURL string, visited map[string]bool, pageCount, maxPages int) bool {
	if nextURL == "" {
		return false
	}
	if maxPages > 0 && pageCount >= maxPages {
		return false
	}
	return !visited[nextURL]
}

// joinPages concatenates per-page markdown bodies (in order) with a blank-line
// separator, skipping empty/whitespace-only pages.
func joinPages(pages []string) string {
	nonEmpty := make([]string, 0, len(pages))
	for _, p := range pages {
		if s := strings.TrimSpace(p); s != "" {
			nonEmpty = append(nonEmpty, s)
		}
	}
	return strings.Join(nonEmpty, "\n\n")
}
