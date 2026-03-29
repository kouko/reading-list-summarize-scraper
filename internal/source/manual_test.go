package source

import "testing"

func TestManualSource(t *testing.T) {
	url := "https://example.com/article"
	src := NewManualSource(url)

	t.Run("creates source with correct URL", func(t *testing.T) {
		if src.URL != url {
			t.Errorf("URL = %q, want %q", src.URL, url)
		}
	})

	t.Run("Name returns manual", func(t *testing.T) {
		if src.Name() != "manual" {
			t.Errorf("Name() = %q, want %q", src.Name(), "manual")
		}
	})

	t.Run("Fetch returns single item with correct fields", func(t *testing.T) {
		items, err := src.Fetch()
		if err != nil {
			t.Fatalf("Fetch() error: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("got %d items, want 1", len(items))
		}
		item := items[0]
		if item.URL != url {
			t.Errorf("URL = %q, want %q", item.URL, url)
		}
		if item.Source != "manual" {
			t.Errorf("Source = %q, want %q", item.Source, "manual")
		}
		if !item.IsUnread {
			t.Error("IsUnread should be true")
		}
		if item.DateAdded.IsZero() {
			t.Error("DateAdded should not be zero")
		}
	})
}
