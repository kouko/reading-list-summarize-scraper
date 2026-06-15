package extract

import "testing"

func TestResolveNextURL(t *testing.T) {
	base := "https://www.itmedia.co.jp/pcuser/articles/2606/08/news096.html"
	cases := []struct {
		name, href, want string
		ok               bool
	}{
		{"relative suffix", "news096_2.html", "https://www.itmedia.co.jp/pcuser/articles/2606/08/news096_2.html", true},
		{"query param", "?page=2", "https://www.itmedia.co.jp/pcuser/articles/2606/08/news096.html?page=2", true},
		{"absolute", "https://x.example/2", "https://x.example/2", true},
		{"empty", "", "", false},
		{"javascript", "javascript:void(0)", "", false},
		{"fragment only", "#top", "", false},
	}
	for _, c := range cases {
		got, ok := resolveNextURL(base, c.href)
		if ok != c.ok {
			t.Errorf("%s: ok=%v, want %v", c.name, ok, c.ok)
			continue
		}
		if ok && got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestShouldFollowNext(t *testing.T) {
	visited := map[string]bool{"https://x/seen": true}
	if shouldFollowNext("", visited, 0, 10) {
		t.Error("empty nextURL should stop")
	}
	if shouldFollowNext("https://x/seen", visited, 1, 10) {
		t.Error("already-visited nextURL should stop (loop guard)")
	}
	if shouldFollowNext("https://x/new", visited, 10, 10) {
		t.Error("reaching maxPages should stop")
	}
	if !shouldFollowNext("https://x/new", visited, 1, 10) {
		t.Error("fresh nextURL under maxPages should continue")
	}
}

func TestJoinPages(t *testing.T) {
	if got := joinPages([]string{"a", "b"}); got != "a\n\nb" {
		t.Errorf("got %q, want %q", got, "a\n\nb")
	}
	if got := joinPages([]string{"a", "   ", "", "b"}); got != "a\n\nb" {
		t.Errorf("empties should be skipped: got %q", got)
	}
	if got := joinPages(nil); got != "" {
		t.Errorf("empty input should join to empty, got %q", got)
	}
}

// scriptedFetcher returns a pageFetcher backed by a fixed url→(content,next,err)
// map, plus a pointer to the fetch-call count for cap assertions.
func scriptedFetcher(pages map[string]struct {
	content, next string
	err           error
}) (pageFetcher, *int) {
	calls := 0
	f := func(u string) (string, string, error) {
		calls++
		p, ok := pages[u]
		if !ok {
			return "", "", errTestUnexpectedURL
		}
		return p.content, p.next, p.err
	}
	return f, &calls
}

var errTestUnexpectedURL = errTest("unexpected url")

type errTest string

func (e errTest) Error() string { return string(e) }

func TestPaginate_FollowsToEnd(t *testing.T) {
	fetch, calls := scriptedFetcher(map[string]struct {
		content, next string
		err           error
	}{
		"https://s/1": {content: "p1", next: "https://s/2"},
		"https://s/2": {content: "p2", next: "https://s/3"},
		"https://s/3": {content: "p3", next: ""}, // no next → stop
	})
	got, err := paginate("https://s/1", 10, fetch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "p1\n\np2\n\np3" {
		t.Errorf("got %q, want p1\\n\\np2\\n\\np3", got)
	}
	if *calls != 3 {
		t.Errorf("fetched %d pages, want 3", *calls)
	}
}

func TestPaginate_MaxPagesCap(t *testing.T) {
	// Every page links to a next page; the cap must stop fetching at maxPages.
	fetch, calls := scriptedFetcher(map[string]struct {
		content, next string
		err           error
	}{
		"https://s/1": {content: "p1", next: "https://s/2"},
		"https://s/2": {content: "p2", next: "https://s/3"},
		"https://s/3": {content: "p3", next: "https://s/4"},
	})
	got, err := paginate("https://s/1", 2, fetch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "p1\n\np2" {
		t.Errorf("got %q, want only first 2 pages", got)
	}
	if *calls != 2 {
		t.Errorf("fetched %d pages, want 2 (maxPages cap)", *calls)
	}
}

func TestPaginate_VisitedLoopGuard(t *testing.T) {
	// page2 links back to page1: must stop, not loop forever.
	fetch, calls := scriptedFetcher(map[string]struct {
		content, next string
		err           error
	}{
		"https://s/1": {content: "p1", next: "https://s/2"},
		"https://s/2": {content: "p2", next: "https://s/1"}, // back-link
	})
	got, err := paginate("https://s/1", 100, fetch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "p1\n\np2" {
		t.Errorf("got %q, want p1\\n\\np2 (stopped at revisit)", got)
	}
	if *calls != 2 {
		t.Errorf("fetched %d pages, want 2 (loop guard)", *calls)
	}
}

func TestPaginate_FailSoftKeepsPriorPages(t *testing.T) {
	// page1 ok, page2 errors → keep page1, no error returned (fail-soft).
	fetch, _ := scriptedFetcher(map[string]struct {
		content, next string
		err           error
	}{
		"https://s/1": {content: "p1", next: "https://s/2"},
		"https://s/2": {err: errTest("boom")},
	})
	got, err := paginate("https://s/1", 10, fetch)
	if err != nil {
		t.Fatalf("partial failure should be fail-soft (nil err), got %v", err)
	}
	if got != "p1" {
		t.Errorf("got %q, want p1 (prior page kept despite later failure)", got)
	}
}

func TestPaginate_AllFail_ReturnsError(t *testing.T) {
	// First page errors → nothing gathered → return the error.
	fetch, _ := scriptedFetcher(map[string]struct {
		content, next string
		err           error
	}{
		"https://s/1": {err: errTest("boom")},
	})
	if _, err := paginate("https://s/1", 10, fetch); err == nil {
		t.Fatal("total failure should return an error")
	}
}

func TestPaginate_StopsOnUnfollowableNext(t *testing.T) {
	// A non-empty next href that resolveNextURL rejects (javascript:) must stop
	// the loop after the current page, not error or loop.
	fetch, calls := scriptedFetcher(map[string]struct {
		content, next string
		err           error
	}{
		"https://s/1": {content: "p1", next: "javascript:void(0)"},
	})
	got, err := paginate("https://s/1", 10, fetch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "p1" {
		t.Errorf("got %q, want p1 (unfollowable next → stop)", got)
	}
	if *calls != 1 {
		t.Errorf("fetched %d pages, want 1", *calls)
	}
}
