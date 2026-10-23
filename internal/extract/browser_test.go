package extract

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBrowserExtractPaginated(t *testing.T) {
	b, err := NewBrowser(false, "", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/pending" {
			<-r.Context().Done()
			return
		}
		w.Header().Set("Content-Type", "text/html")
		switch r.URL.Path {
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
	t.Run("initial earlier page evidence", func(t *testing.T) {
		got, err := b.ExtractPaginated(server.URL+"/query?page=2&id=article", js, ".next", 10, 2*time.Second, 0)
		if err == nil || got != "" {
			t.Fatalf("later query entry: got %q, %v", got, err)
		}
		got, err = b.ExtractPaginated(server.URL+"/query?page=1&id=article", js, ".next", 10, 2*time.Second, 0)
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
