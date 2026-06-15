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
