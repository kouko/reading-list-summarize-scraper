package extract

import (
	"fmt"
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

// paginate returns a body only after every page reaches a normal end without
// a continuation. Known incomplete sequences return no body and an error.
func paginate(startURL string, maxPages int, fetch pageFetcher) (string, error) {
	if maxPages <= 0 {
		maxPages = defaultMaxPages
	}

	var pages []string
	visited := map[string]bool{}
	current := startURL

	for pageCount := 0; ; pageCount++ {
		visited[current] = true

		content, nextHref, err := fetch(current)
		if err != nil {
			return "", fmt.Errorf("paginated extract: page %s: %w", current, err)
		}
		if strings.TrimSpace(content) == "" {
			return "", fmt.Errorf("paginated extract: page %s has empty content", current)
		}
		pages = append(pages, content)
		if strings.TrimSpace(nextHref) == "" {
			return joinPages(pages), nil
		}
		next, ok := resolveNextURL(current, nextHref)
		if !ok {
			return "", fmt.Errorf("paginated extract: page %s has unfollowable continuation %q", current, nextHref)
		}
		if visited[next] {
			return "", fmt.Errorf("paginated extract: page %s repeats continuation %s", current, next)
		}
		if pageCount+1 >= maxPages {
			return "", fmt.Errorf("paginated extract: page %s has continuation %s beyond maxPages %d", current, next, maxPages)
		}
		current = next
	}
}

// resolveNextURL resolves a "next page" href (possibly relative) against the
// current page's URL into an absolute URL. It returns ok=false for hrefs that
// are not real navigations: empty, javascript:/mailto: schemes, or a bare
// same-page fragment. The fragment is stripped from the result so that the
// visited-set comparison in paginate is stable.
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
	if (abs.Scheme != "http" && abs.Scheme != "https") || abs.Host == "" {
		return "", false
	}
	abs.Fragment = ""
	return abs.String(), true
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
