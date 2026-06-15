package source

import (
	"testing"

	"github.com/mmcdole/gofeed"
)

// sampleRSS has 5 items with ascending pubDates (item 5 is newest).
const sampleRSS = `<?xml version="1.0"?>
<rss version="2.0">
  <channel>
    <title>Example Feed</title>
    <item><title>Post 1</title><link>https://example.com/1</link><description>desc 1</description><pubDate>Mon, 01 Jun 2026 10:00:00 GMT</pubDate></item>
    <item><title>Post 2</title><link>https://example.com/2</link><description>desc 2</description><pubDate>Tue, 02 Jun 2026 10:00:00 GMT</pubDate></item>
    <item><title>Post 3</title><link>https://example.com/3</link><description>desc 3</description><pubDate>Wed, 03 Jun 2026 10:00:00 GMT</pubDate></item>
    <item><title>Post 4</title><link>https://example.com/4</link><description>desc 4</description><pubDate>Thu, 04 Jun 2026 10:00:00 GMT</pubDate></item>
    <item><title>Post 5</title><link>https://example.com/5</link><description>desc 5</description><pubDate>Fri, 05 Jun 2026 10:00:00 GMT</pubDate></item>
  </channel>
</rss>`

func parseSample(t *testing.T, xml string) *gofeed.Feed {
	t.Helper()
	feed, err := gofeed.NewParser().ParseString(xml)
	if err != nil {
		t.Fatalf("ParseString: %v", err)
	}
	return feed
}

func TestFeedToItems_MapsAndCapsNewestN(t *testing.T) {
	feed := parseSample(t, sampleRSS)
	items := feedToItems(feed, 3)

	if len(items) != 3 {
		t.Fatalf("got %d items, want 3 (newest-N cap)", len(items))
	}
	// Newest first: Post 5, 4, 3.
	wantTitles := []string{"Post 5", "Post 4", "Post 3"}
	for i, want := range wantTitles {
		if items[i].Title != want {
			t.Errorf("items[%d].Title = %q, want %q", i, items[i].Title, want)
		}
	}
	// Field mapping on the newest item.
	top := items[0]
	if top.URL != "https://example.com/5" {
		t.Errorf("URL = %q, want the item link", top.URL)
	}
	if top.PreviewText != "desc 5" {
		t.Errorf("PreviewText = %q, want the item description", top.PreviewText)
	}
	if top.Source != "rss" {
		t.Errorf("Source = %q, want rss", top.Source)
	}
	if top.DateAdded.IsZero() {
		t.Error("DateAdded should be set from the item's published date")
	}
}

func TestFeedToItems_CountZeroDefaults(t *testing.T) {
	feed := parseSample(t, sampleRSS)
	// count 0 → default cap (5); the sample has exactly 5 so all come through.
	items := feedToItems(feed, 0)
	if len(items) != 5 {
		t.Errorf("count=0 should apply the default cap; got %d items, want 5", len(items))
	}
}

// sampleAtomMixedDates exercises the date-fallback ladder: entry A has both
// published+updated (published must win), B has only updated (fallback), C has
// neither (zero time → sorts last). gofeed is lenient about the missing dates.
const sampleAtomMixedDates = `<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Atom Example</title>
  <entry><title>A</title><link href="https://x/a"/><id>a</id><published>2026-06-10T00:00:00Z</published><updated>2026-06-12T00:00:00Z</updated></entry>
  <entry><title>B</title><link href="https://x/b"/><id>b</id><updated>2026-06-11T00:00:00Z</updated></entry>
  <entry><title>C</title><link href="https://x/c"/><id>c</id></entry>
</feed>`

func TestFeedToItems_UndatedAndUpdatedFallback(t *testing.T) {
	feed := parseSample(t, sampleAtomMixedDates)
	items := feedToItems(feed, 10)
	if len(items) != 3 {
		t.Fatalf("got %d items, want 3", len(items))
	}

	// Order newest-first by effective time: B(updated 06-11) > A(published 06-10)
	// > C(no date → zero, last).
	wantOrder := []string{"B", "A", "C"}
	for i, want := range wantOrder {
		if items[i].Title != want {
			t.Errorf("items[%d].Title = %q, want %q (order: %v)", i, items[i].Title, want, wantOrder)
		}
	}

	byTitle := map[string]ReadingItem{}
	for _, it := range items {
		byTitle[it.Title] = it
	}
	// A: published wins over updated → 06-10, not 06-12.
	if d := byTitle["A"].DateAdded.UTC().Format("2006-01-02"); d != "2026-06-10" {
		t.Errorf("A DateAdded = %s, want 2026-06-10 (published precedence)", d)
	}
	// B: falls back to updated → 06-11.
	if d := byTitle["B"].DateAdded.UTC().Format("2006-01-02"); d != "2026-06-11" {
		t.Errorf("B DateAdded = %s, want 2026-06-11 (updated fallback)", d)
	}
	// C: no date → zero time.
	if !byTitle["C"].DateAdded.IsZero() {
		t.Errorf("C DateAdded = %v, want zero (no date in feed)", byTitle["C"].DateAdded)
	}
}

func TestFeedToItems_FewerThanCap(t *testing.T) {
	feed := parseSample(t, sampleRSS)
	// Asking for more than available returns all, no panic.
	items := feedToItems(feed, 50)
	if len(items) != 5 {
		t.Errorf("got %d, want all 5 when cap exceeds available", len(items))
	}
}
