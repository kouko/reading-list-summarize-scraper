package source

import "testing"

func TestDeduplicateByURL(t *testing.T) {
	t.Run("removes duplicates keeps first", func(t *testing.T) {
		items := []ReadingItem{
			{URL: "https://a.com", Title: "First A"},
			{URL: "https://b.com", Title: "B"},
			{URL: "https://a.com", Title: "Second A"},
		}
		got := DeduplicateByURL(items)
		if len(got) != 2 {
			t.Fatalf("got %d items, want 2", len(got))
		}
		if got[0].Title != "First A" {
			t.Errorf("first occurrence not kept: got %q", got[0].Title)
		}
		if got[1].Title != "B" {
			t.Errorf("second item: got %q, want %q", got[1].Title, "B")
		}
	})

	t.Run("empty input", func(t *testing.T) {
		got := DeduplicateByURL(nil)
		if len(got) != 0 {
			t.Errorf("got %d items, want 0", len(got))
		}
	})

	t.Run("no duplicates", func(t *testing.T) {
		items := []ReadingItem{
			{URL: "https://a.com"},
			{URL: "https://b.com"},
			{URL: "https://c.com"},
		}
		got := DeduplicateByURL(items)
		if len(got) != 3 {
			t.Errorf("got %d items, want 3", len(got))
		}
	})

	t.Run("multiple duplicates of same URL", func(t *testing.T) {
		items := []ReadingItem{
			{URL: "https://a.com", Title: "1st"},
			{URL: "https://a.com", Title: "2nd"},
			{URL: "https://a.com", Title: "3rd"},
			{URL: "https://b.com", Title: "B"},
		}
		got := DeduplicateByURL(items)
		if len(got) != 2 {
			t.Fatalf("got %d items, want 2", len(got))
		}
		if got[0].Title != "1st" {
			t.Errorf("first occurrence not kept: got %q", got[0].Title)
		}
	})
}
