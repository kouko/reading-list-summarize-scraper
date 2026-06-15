package source

import (
	"log/slog"
	"sort"
	"time"

	"github.com/mmcdole/gofeed"
)

// defaultRSSCount caps how many newest entries per feed are taken when the
// configured count is unset (<= 0). Bounds first-subscribe backfill.
const defaultRSSCount = 5

// RSSSource fetches articles from a list of RSS/Atom/JSON feed URLs. Each poll
// returns the newest `count` entries per feed; the pipeline's file-index skip
// dedupes already-summarized URLs, so steady state only processes new entries.
type RSSSource struct {
	feeds  []string
	count  int
	parser *gofeed.Parser
}

// NewRSSSource creates an RSSSource. count <= 0 applies defaultRSSCount.
func NewRSSSource(feeds []string, count int) *RSSSource {
	return &RSSSource{feeds: feeds, count: count, parser: gofeed.NewParser()}
}

func (r *RSSSource) Name() string { return "rss" }

// Fetch parses every configured feed and returns the newest entries as
// ReadingItems. A feed that fails to fetch/parse is logged and skipped so one
// dead feed does not sink the rest (mirrors the source-loop tolerance in cmd).
func (r *RSSSource) Fetch() ([]ReadingItem, error) {
	var items []ReadingItem
	for _, url := range r.feeds {
		feed, err := r.parser.ParseURL(url)
		if err != nil {
			slog.Warn("rss: feed fetch/parse failed, skipping", "feed", url, "err", err)
			continue
		}
		items = append(items, feedToItems(feed, r.count)...)
	}
	return items, nil
}

// feedToItems maps a parsed feed's entries to ReadingItems, newest first,
// capped to count (count <= 0 → defaultRSSCount). Pure (no I/O) for testing.
func feedToItems(feed *gofeed.Feed, count int) []ReadingItem {
	if count <= 0 {
		count = defaultRSSCount
	}
	if feed == nil {
		return nil
	}

	entries := make([]*gofeed.Item, len(feed.Items))
	copy(entries, feed.Items)

	// Newest first. Entries without a parsed date sort to the end (treated as
	// oldest) so undated items don't crowd out genuinely-recent ones.
	sort.SliceStable(entries, func(i, j int) bool {
		return entryTime(entries[i]).After(entryTime(entries[j]))
	})

	if len(entries) > count {
		entries = entries[:count]
	}

	items := make([]ReadingItem, 0, len(entries))
	for _, e := range entries {
		items = append(items, ReadingItem{
			Title:       e.Title,
			URL:         e.Link,
			DateAdded:   entryTime(e),
			PreviewText: e.Description,
			Source:      "rss",
		})
	}
	return items
}

// entryTime returns the entry's published (or updated, as fallback) time, or
// the zero time when neither is parseable.
func entryTime(e *gofeed.Item) time.Time {
	if e.PublishedParsed != nil {
		return *e.PublishedParsed
	}
	if e.UpdatedParsed != nil {
		return *e.UpdatedParsed
	}
	return time.Time{}
}
