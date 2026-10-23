package extract

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"

	assets "github.com/kouko/reading-list-summarize-scraper/embed"
)

func newTestBrowser(t *testing.T) *Browser {
	t.Helper()
	b, err := NewBrowser(false, "", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	// Register after TempDir so Chrome finishes writing before profile removal.
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(b.ctx, 10*time.Second)
		defer cancel()
		defer b.Close()
		if err := chromedp.Cancel(ctx); err != nil {
			t.Errorf("close test browser: %v", err)
		}
	})
	// Allocate on the root context so child navigation deadlines exclude
	// process startup and every extraction reuses the same browser.
	if err := chromedp.Run(b.ctx); err != nil {
		t.Fatalf("start test browser: %v", err)
	}
	return b
}

func TestBrowserExtractPaginated(t *testing.T) {
	b := newTestBrowser(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/pending" {
			<-r.Context().Done()
			return
		}
		w.Header().Set("Content-Type", "text/html")
		switch r.URL.Path {
		case "/incomplete-start":
			next := "/incomplete-end?kind=" + r.URL.Query().Get("kind")
			switch r.URL.Query().Get("kind") {
			case "unfollowable":
				next = "javascript:void(0)"
			case "later-failure":
				next = "http://127.0.0.1:1/"
			}
			fmt.Fprintf(w, `<body><article>first</article><a class="next" href="%s">next</a></body>`, next)
		case "/incomplete-end":
			if r.URL.Query().Get("kind") == "loop" {
				fmt.Fprint(w, `<body><article>second</article><a class="next" href="/incomplete-start?kind=loop">next</a></body>`)
			} else {
				fmt.Fprint(w, `<body><article>  </article></body>`)
			}
		case "/wall-start":
			fmt.Fprintf(w, `<body><article><p>The investigation follows a complete account of the events and explains the evidence collected by the reporters.</p></article><a class="next" href="/wall-end?kind=%s">next</a></body>`, r.URL.Query().Get("kind"))
		case "/wall-end":
			switch r.URL.Query().Get("kind") {
			case "challenge":
				fmt.Fprint(w, `<body><article><p>Cloudflare security challenge. CAPTCHA.</p></article></body>`)
			case "article":
				fmt.Fprint(w, `<body><article><p>Please sign in to read is a phrase commonly used by publishers.</p><p>The researchers explain how publishers fund reporting through subscriptions and describe the effects on their readers.</p></article></body>`)
			default:
				fmt.Fprint(w, `<body><article><p>Please sign in to read the full article.</p></article></body>`)
			}
		case "/1":
			fmt.Fprint(w, `<body><article>first</article><a class="next" href="/2">next</a><img src="/pending"></body>`)
		case "/2":
			fmt.Fprint(w, `<link rel="prev" href="/1"><body><article>second</article><img src="/pending"></body>`)
		case "/query":
			if r.URL.Query().Get("page") == "2" {
				fmt.Fprint(w, `<body><article>second</article><a href="?page=1&amp;id=article">1</a></body>`)
			} else {
				fmt.Fprint(w, `<link rel="first" href="?page=1&amp;id=article"><body><article>first</article><a href="?page=0&amp;id=other">other</a><a class="next" href="?page=2&amp;id=article">next</a></body>`)
			}
		case "/query-base-prev":
			fmt.Fprint(w, `<body><article>second</article><a class="prev" href="?id=article">previous</a></body>`)
		case "/query-no-prev":
			fmt.Fprint(w, `<body><article>second</article></body>`)
		case "/hint":
			fmt.Fprint(w, `<link rel="next" href="/2"><body><article>first</article></body>`)
		case "/empty-next":
			fmt.Fprint(w, `<link rel="next" href="/2"><body><article>first</article><a class="next" href="">next</a></body>`)
		case "/deferred":
			fmt.Fprint(w, `<body><article>early</article><script defer src="/script"></script></body>`)
		case "/script":
			time.Sleep(100 * time.Millisecond)
			fmt.Fprint(w, `document.querySelector('article').textContent='ready'`)
		}
	}))
	defer server.Close()
	js := `window.extractArticle=()=>document.querySelector('article').textContent`
	t.Run("incomplete sequences", func(t *testing.T) {
		for _, tc := range []struct {
			kind, page, reason string
			maxPages           int
		}{
			{"cap", "/incomplete-start?kind=cap", "beyond maxPages 1", 1},
			{"loop", "/incomplete-end?kind=loop", "repeats continuation", 10},
			{"later-failure", "http://127.0.0.1:1/", "net::ERR_UNSAFE_PORT", 10},
			{"unfollowable", "/incomplete-start?kind=unfollowable", "unfollowable continuation", 10},
			{"empty", "/incomplete-end?kind=empty", "empty content", 10},
		} {
			t.Run(tc.kind, func(t *testing.T) {
				failedPage := tc.page
				if strings.HasPrefix(failedPage, "/") {
					failedPage = server.URL + failedPage
				}
				got, err := b.ExtractPaginated(server.URL+"/incomplete-start?kind="+tc.kind, js, ".next", tc.maxPages, 2*time.Second, 0)
				if err == nil || got != "" || !strings.Contains(err.Error(), "page "+failedPage) || !strings.Contains(err.Error(), tc.reason) {
					t.Fatalf("got %q, %v; want no body and failed page %s with %s", got, err, failedPage, tc.reason)
				}
			})
		}
	})
	t.Run("real extractor validates later page", func(t *testing.T) {
		for _, tc := range []struct{ kind, reason string }{{"login", "login required"}, {"challenge", "Cloudflare/CAPTCHA"}, {"article", ""}} {
			got, err := b.ExtractPaginated(server.URL+"/wall-start?kind="+tc.kind, assets.DefuddleJS, ".next", 10, 2*time.Second, 0)
			if tc.reason == "" {
				if err != nil || !strings.Contains(got, "The researchers explain") {
					t.Errorf("%s: got %q, %v", tc.kind, got, err)
				}
			} else if err == nil || got != "" || !strings.Contains(err.Error(), "/wall-end") || !strings.Contains(err.Error(), tc.reason) {
				t.Errorf("%s: got %q, %v", tc.kind, got, err)
			}
		}
	})
	t.Run("initial earlier page evidence", func(t *testing.T) {
		for _, path := range []string{"/query", "/query-base-prev", "/query-no-prev"} {
			got, err := b.ExtractPaginated(server.URL+path+"?page=2&id=article", js, ".next", 10, 2*time.Second, 0)
			if err == nil || got != "" {
				t.Errorf("%s later query entry: got %q, %v", path, got, err)
			}
		}
		got, err := b.ExtractPaginated(server.URL+"/query?page=1&id=article", js, ".next", 10, 2*time.Second, 0)
		if err != nil || got != "first\n\nsecond" {
			t.Fatalf("first query entry: got %q, %v", got, err)
		}
	})
	t.Run("ready body pending subresource", func(t *testing.T) {
		got, err := b.ExtractPaginated(server.URL+"/1", js, ".next", 10, 2*time.Second, 0)
		if err != nil || got != "first\n\nsecond" {
			t.Fatalf("got %q, %v", got, err)
		}
	})
	t.Run("DOMContentLoaded before extraction", func(t *testing.T) {
		got, err := b.ExtractPaginated(server.URL+"/deferred", js, ".next", 10, 2*time.Second, 0)
		if err != nil || got != "ready" {
			t.Fatalf("got %q, %v", got, err)
		}
	})
	t.Run("unmatched selector with standard hint", func(t *testing.T) {
		for _, path := range []string{"/hint", "/empty-next"} {
			got, err := b.ExtractPaginated(server.URL+path, js, ".next", 10, 2*time.Second, 0)
			if err == nil || got != "" {
				t.Fatalf("%s: got %q, %v", path, got, err)
			}
		}
	})
	t.Run("navigation error", func(t *testing.T) {
		got, err := b.ExtractPaginated("http://127.0.0.1:1/", js, ".next", 10, 2*time.Second, 0)
		if err == nil || got != "" || !strings.Contains(err.Error(), "net::ERR_") {
			t.Fatalf("got %q, %v", got, err)
		}
	})
}
