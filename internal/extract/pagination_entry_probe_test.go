package extract

// concern: Starting on a later article page must not save a known incomplete suffix as a complete article.

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestPagination_laterentry_rejectsincomplete exercises explicit first/previous-page evidence.
func TestPagination_laterentry_rejectsincomplete(t *testing.T) {
	b, err := NewBrowser(false, "", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		switch r.URL.Path {
		case "/1":
			fmt.Fprint(w, `<body><article>first</article><a class="next" href="/2">next</a></body>`)
		case "/2":
			fmt.Fprint(w, `<link rel="first" href="/1"><link rel="prev" href="/1"><body><article>second</article><a class="next" href="/3">next</a></body>`)
		case "/3":
			fmt.Fprint(w, `<link rel="first" href="/1"><link rel="prev" href="/2"><body><article>third</article></body>`)
		}
	}))
	defer s.Close()
	for _, entry := range []string{"/2", "/3"} {
		t.Run(entry, func(t *testing.T) {
			got, err := b.ExtractPaginated(s.URL+entry, `window.extractArticle=()=>document.querySelector('article').textContent`, ".next", 10, 3*time.Second, 0)
			if err == nil && got != "first\n\nsecond\n\nthird" {
				t.Fatalf("known later-page entry returned incomplete success: %q", got)
			}
			if err != nil && got != "" {
				t.Fatalf("failure leaked partial body: %q, %v", got, err)
			}
		})
	}
}
