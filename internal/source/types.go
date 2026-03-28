package source

import "time"

type ReadingItem struct {
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	DateAdded   time.Time `json:"date_added"`
	IsUnread    bool      `json:"is_unread"`
	PreviewText string    `json:"preview_text,omitempty"`
	Source      string    `json:"source"` // "safari" | "chrome" | "manual"
}

type Source interface {
	Name() string
	Fetch() ([]ReadingItem, error)
}

// DeduplicateByURL removes duplicate items keeping the first occurrence.
func DeduplicateByURL(items []ReadingItem) []ReadingItem {
	seen := make(map[string]bool)
	var result []ReadingItem
	for _, item := range items {
		if !seen[item.URL] {
			seen[item.URL] = true
			result = append(result, item)
		}
	}
	return result
}
