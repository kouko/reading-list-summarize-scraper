package output

import (
	"testing"
	"time"
)

func TestSHA8(t *testing.T) {
	got := SHA8("https://example.com/article")
	if len(got) != 8 {
		t.Errorf("SHA8 length = %d, want 8", len(got))
	}
	if SHA8("https://example.com/article") != got {
		t.Error("SHA8 not deterministic")
	}
	if SHA8("https://other.com") == got {
		t.Error("SHA8 collision")
	}
}

func TestDomainDir(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://example.com/path", "example_com"},
		{"https://blog.example.com/path", "blog_example_com"},
		{"https://www.nikkei.com/article", "www_nikkei_com"},
	}
	for _, tt := range tests {
		got := DomainDir(tt.url)
		if got != tt.want {
			t.Errorf("DomainDir(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}

func TestFilename(t *testing.T) {
	date := time.Date(2026, 3, 28, 0, 0, 0, 0, time.UTC)
	sha := "a1b2c3d4"

	got := SummaryFilename(date, sha)
	if got != "2026-03-28__a1b2c3d4__summary.md" {
		t.Errorf("SummaryFilename = %q", got)
	}

	got = ContentFilename(date, sha)
	if got != "2026-03-28__a1b2c3d4__content.md" {
		t.Errorf("ContentFilename = %q", got)
	}
}
