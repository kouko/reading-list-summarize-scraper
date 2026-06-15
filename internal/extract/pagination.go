package extract

import (
	"net/url"
	"strings"
)

// defaultMaxPages bounds how many pages a paginated article is followed for
// when a DomainRule sets no (or a non-positive) MaxPages.
const defaultMaxPages = 10

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
