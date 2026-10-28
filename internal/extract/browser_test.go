package extract

// concern: First-page recovery must reject unrelated, missing, or incomplete article chains without returning partial content.

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	cu "github.com/Davincible/chromedp-undetected"
	"github.com/chromedp/chromedp"

	assets "github.com/kouko/reading-list-summarize-scraper/embed"
	"github.com/kouko/reading-list-summarize-scraper/internal/config"
)

func newTestBrowser(t *testing.T, flags ...chromedp.ExecAllocatorOption) *Browser {
	t.Helper()
	// Allocator flags apply to every tab, including those opened by Extract.
	ctx, cancel, err := cu.New(cu.NewConfig(cu.WithUserDataDir(t.TempDir()),
		cu.WithChromeFlags(append(flags, chromedp.Flag("headless", true))...)))
	if err != nil {
		t.Fatal(err)
	}
	b := &Browser{ctx: ctx, cancel: cancel}
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

func TestBrowserExtractSocialCompleteOnly(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<body><article>ordinary article</article></body>`)
	}))
	defer server.Close()
	// Keep the real social hostname for production dispatch while every tab
	// resolves it only to this fixture; other hostnames cannot resolve.
	b := newTestBrowser(t,
		chromedp.Flag("host-resolver-rules", "MAP x.com 127.0.0.1, MAP * ~NOTFOUND, EXCLUDE 127.0.0.1"),
		chromedp.Flag("no-proxy-server", true),
		chromedp.Flag("ignore-certificate-errors", true))
	socialURL := strings.Replace(server.URL, "127.0.0.1", "x.com", 1) + "/author/status/1"
	js := `window.extractArticle=()=>document.querySelector('article').textContent;
window.socialSnapshot=()=>({version:1,platform:'x',targetID:'1',posts:[{id:'1',url:location.href,authorID:'a',body:'verified article',bodyComplete:true,parentID:null,replyCount:0}]})`
	body, err := b.Extract(socialURL, js, 20*time.Second, 0)
	if err != nil || body != "verified article\n\n[Source]("+socialURL+")" {
		t.Fatalf("social: %q, %v", body, err)
	}
	body, err = b.Extract(socialURL, strings.Replace(js, "bodyComplete:true", "bodyComplete:false", 1), 20*time.Second, 0)
	if err == nil || body != "" {
		t.Fatalf("preview: %q, %v", body, err)
	}
	body, err = b.Extract(socialURL, `window.socialSnapshot=()=>{throw Error('missing target')};window.extractArticle=()=>{throw Error('Cloudflare security challenge')}`, 20*time.Second, 0)
	if err == nil || body != "" || !strings.Contains(err.Error(), "Cloudflare") {
		t.Fatalf("challenge identity: %q, %v", body, err)
	}
	body, err = b.Extract(server.URL, js, 20*time.Second, 0)
	if err != nil || body != "ordinary article" {
		t.Fatalf("ordinary: %q, %v", body, err)
	}
}

func TestBrowserExtractPaginated(t *testing.T) {
	b := newTestBrowser(t)
	// Check extraction behavior within the production budget, not a two-second
	// performance target. The pending resource remains blocked until canceled.
	timeout := config.DefaultConfig().Extract.Timeout
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
			case "missing":
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, `<html><body><h1>Error response</h1><p>Error code: 404</p><p>Message: Not Found.</p><p>Error code explanation: 404 - Nothing matches the given URI.</p></body></html>`)
			case "challenge":
				fmt.Fprint(w, `<body><article><p>Cloudflare security challenge. CAPTCHA.</p></article></body>`)
			case "challenge-heading":
				fmt.Fprint(w, `<html><body><h1>Cloudflare CAPTCHA</h1><p>Verify you are human</p></body></html>`)
			case "challenge-title":
				fmt.Fprint(w, `<html><head><title>Cloudflare CAPTCHA</title></head><body><p>Verify you are human</p></body></html>`)
			case "article":
				fmt.Fprint(w, `<body><article><h1>Cloudflare CAPTCHA</h1><p>Please sign in to read is a phrase commonly used by publishers.</p><p>The researchers explain how publishers fund reporting through subscriptions and describe the effects on their readers.</p></article></body>`)
			default:
				fmt.Fprint(w, `<body><article><p>Please sign in to read the full article.</p></article></body>`)
			}
		case "/1":
			fmt.Fprint(w, `<body><article>first</article><a class="next" href="/2">next</a><img src="/pending"><img src="/missing-image"></body>`)
		case "/missing-image":
			w.WriteHeader(http.StatusNotFound)
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
		case "/query-mode":
			switch r.URL.Query().Get("page") {
			case "1":
				fmt.Fprint(w, `<body><article>first</article><a class="next" href="?page=2&amp;id=article">next</a></body>`)
			case "2":
				fmt.Fprint(w, `<body><article>second</article></body>`)
			default:
				fmt.Fprint(w, `<body><article>different presentation without a forward chain</article></body>`)
			}
		case "/story.html":
			fmt.Fprint(w, `<body><article>first</article><a class="next" href="/story_2.html">next</a></body>`)
		case "/story_2.html":
			fmt.Fprint(w, `<link rel="prev" href="/story.html"><body><article>second</article></body>`)
		case "/recover":
			kind := r.URL.Query().Get("kind")
			if r.URL.Query().Get("page") == "2" {
				prev := "?kind=" + kind
				if kind == "foreign" {
					prev = "/another-article"
				}
				fmt.Fprintf(w, `<link rel="prev" href="%s"><body><article>second</article></body>`, prev)
			} else if kind == "unavailable" {
				fmt.Fprint(w, `<body><article> </article></body>`)
			} else if kind == "canonical" {
				fmt.Fprint(w, `<link rel="canonical" href="?page=2&amp;kind=canonical"><body><article>second</article><a class="next" href="?page=2&amp;kind=canonical">next</a></body>`)
			} else {
				fmt.Fprint(w, `<body><article>first</article></body>`)
			}
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
				got, err := b.ExtractPaginated(server.URL+"/incomplete-start?kind="+tc.kind, js, ".next", tc.maxPages, timeout, 0)
				if err == nil || got != "" || !strings.Contains(err.Error(), "page "+failedPage) || !strings.Contains(err.Error(), tc.reason) {
					t.Fatalf("got %q, %v; want no body and failed page %s with %s", got, err, failedPage, tc.reason)
				}
			})
		}
	})
	t.Run("real extractor validates later page", func(t *testing.T) {
		for _, tc := range []struct{ kind, reason string }{{"login", "login required"}, {"challenge", "Cloudflare/CAPTCHA"}, {"challenge-heading", "Cloudflare/CAPTCHA"}, {"challenge-title", "Cloudflare/CAPTCHA"}, {"missing", "HTTP 404"}, {"article", ""}} {
			got, err := b.ExtractPaginated(server.URL+"/wall-start?kind="+tc.kind, assets.DefuddleJS, ".next", 10, timeout, 0)
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
		for _, path := range []string{"/query-base-prev", "/query-no-prev"} {
			got, err := b.ExtractPaginated(server.URL+path+"?page=2&id=article", js, ".next", 10, timeout, 0)
			if err == nil || got != "" {
				t.Errorf("%s later query entry: got %q, %v", path, got, err)
			}
		}
		for _, path := range []string{"/query?page=2&id=article", "/story_2.html", "/query-mode?page=2&id=article"} {
			got, err := b.ExtractPaginated(server.URL+path, js, ".next", 2, timeout, 0)
			if err != nil || got != "first\n\nsecond" {
				t.Errorf("later entry %s: got %q, %v", path, got, err)
			}
		}
		for _, tc := range []struct {
			kind, reason string
			maxPages     int
		}{{"foreign", "article identity mismatch", 2}, {"missing-entry", "omits requested page", 2}, {"unavailable", "empty content", 2}, {"limit", "recovery exceeds maxPages", 1}, {"canonical", "article identity mismatch: canonical", 2}} {
			got, err := b.ExtractPaginated(server.URL+"/recover?page=2&kind="+tc.kind, js, ".next", tc.maxPages, timeout, 0)
			if err == nil || got != "" || !strings.Contains(err.Error(), tc.reason) {
				t.Errorf("%s recovery: got %q, %v; want %s", tc.kind, got, err, tc.reason)
			}
		}
		got, err := b.ExtractPaginated(server.URL+"/query?page=1&id=article", js, ".next", 10, timeout, 0)
		if err != nil || got != "first\n\nsecond" {
			t.Fatalf("first query entry: got %q, %v", got, err)
		}
	})
	t.Run("ready body pending subresource", func(t *testing.T) {
		got, err := b.ExtractPaginated(server.URL+"/1", js, ".next", 10, timeout, 0)
		if err != nil || got != "first\n\nsecond" {
			t.Fatalf("got %q, %v", got, err)
		}
	})
	t.Run("DOMContentLoaded before extraction", func(t *testing.T) {
		got, err := b.ExtractPaginated(server.URL+"/deferred", js, ".next", 10, timeout, 0)
		if err != nil || got != "ready" {
			t.Fatalf("got %q, %v", got, err)
		}
	})
	t.Run("unmatched selector with standard hint", func(t *testing.T) {
		for _, path := range []string{"/hint", "/empty-next"} {
			got, err := b.ExtractPaginated(server.URL+path, js, ".next", 10, timeout, 0)
			if err == nil || got != "" {
				t.Fatalf("%s: got %q, %v", path, got, err)
			}
		}
	})
	t.Run("navigation error", func(t *testing.T) {
		got, err := b.ExtractPaginated("http://127.0.0.1:1/", js, ".next", 10, timeout, 0)
		if err == nil || got != "" || !strings.Contains(err.Error(), "net::ERR_") {
			t.Fatalf("got %q, %v", got, err)
		}
	})
}
