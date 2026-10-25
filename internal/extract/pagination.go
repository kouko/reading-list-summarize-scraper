package extract

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// defaultMaxPages bounds how many pages a paginated article is followed for
// when a DomainRule sets no (or a non-positive) MaxPages.
const defaultMaxPages = 10

var numberedArticlePath = regexp.MustCompile(`^(.*)_([0-9]+)(\.[^/.]+)$`)

// Canonical URLs may omit query parameters but must not change article keys
// or contradict the actual document's position with an explicit page number.
func compatibleCanonical(identity, canonical string, position int) bool {
	canonicalIdentity, canonicalPage, _, err := articlePage(canonical)
	if err != nil {
		return false
	}
	a, _ := url.Parse(identity)
	b, _ := url.Parse(canonicalIdentity)
	actualCanonical, _ := url.Parse(canonical)
	_, explicitQueryPage := actualCanonical.Query()["page"]
	if (explicitQueryPage || numberedArticlePath.MatchString(actualCanonical.Path)) && canonicalPage != position {
		return false
	}
	if a.Scheme != b.Scheme || a.Host != b.Host || a.Path != b.Path {
		return false
	}
	for key, values := range b.Query() {
		if strings.Join(a.Query()[key], "\x00") != strings.Join(values, "\x00") {
			return false
		}
	}
	return true
}

// articlePage separates supported page positions from the stable article URL.
// Other query parameters remain part of the identity, not navigation guesses.
func articlePage(raw string) (identity string, number int, first string, err error) {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", 0, "", fmt.Errorf("article identity mismatch: invalid URL %q", raw)
	}
	u.Fragment = ""
	number = 1
	q := u.Query()
	values, hasPage := q["page"]
	if hasPage {
		if len(values) != 1 {
			return "", 0, "", fmt.Errorf("cannot establish article first page: ambiguous page number")
		}
		number, err = strconv.Atoi(values[0])
		if err != nil || number < 1 {
			return "", 0, "", fmt.Errorf("cannot establish article first page: invalid page number")
		}
		q.Del("page")
		u.RawQuery = q.Encode()
	}
	if match := numberedArticlePath.FindStringSubmatch(u.Path); match != nil {
		if number != 1 {
			return "", 0, "", fmt.Errorf("cannot establish article first page: conflicting page numbers")
		}
		number, err = strconv.Atoi(match[2])
		if err != nil || number < 1 {
			return "", 0, "", fmt.Errorf("cannot establish article first page: invalid path page number")
		}
		u.Path = match[1] + match[3]
	}
	u.RawQuery = q.Encode()
	identity = u.String()
	if hasPage {
		q.Set("page", "1")
		u.RawQuery = q.Encode()
	}
	return identity, number, u.String(), nil
}

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
