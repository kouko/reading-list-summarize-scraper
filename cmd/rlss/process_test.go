package main

// concern: timestamp precision, stable undated placement, and limiting before date ordering can select the wrong reading items.

import (
	"reflect"
	"testing"
	"time"

	"github.com/kouko/reading-list-summarize-scraper/internal/config"
	"github.com/kouko/reading-list-summarize-scraper/internal/source"
)

func TestApplyFiltersDateOrder(t *testing.T) {
	base := time.Date(2026, 10, 23, 8, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		items  []source.ReadingItem
		filter config.FilterConfig
		want   []string
	}{
		{
			name: "precise-newest-first",
			items: []source.ReadingItem{
				{URL: "older", DateAdded: base},
				{URL: "newest", DateAdded: base.Add(time.Nanosecond)},
				{URL: "earlier-local-date", DateAdded: time.Date(2026, 10, 23, 1, 0, 0, 0, time.FixedZone("west", -8*60*60))},
			},
			want: []string{"earlier-local-date", "newest", "older"},
		},
		{
			name: "stable-ties-undated",
			items: []source.ReadingItem{
				{URL: "undated-first"},
				{URL: "tie-first", DateAdded: base},
				{URL: "undated-second"},
				{URL: "tie-second", DateAdded: base.In(time.FixedZone("east", 8*60*60))},
			},
			want: []string{"tie-first", "tie-second", "undated-first", "undated-second"},
		},
		{
			name: "mixed-sources",
			items: []source.ReadingItem{
				{URL: "safari", Source: "safari", DateAdded: base},
				{URL: "chrome", Source: "chrome", DateAdded: base.Add(2 * time.Hour)},
				{URL: "rss", Source: "rss", DateAdded: base.Add(time.Hour)},
			},
			want: []string{"chrome", "rss", "safari"},
		},
		{
			name: "filtered-top-n",
			items: []source.ReadingItem{
				{URL: "old", DateAdded: base.AddDate(0, 0, -1), IsUnread: true},
				{URL: "older-qualifying", DateAdded: base, IsUnread: true},
				{URL: "read", DateAdded: base.Add(4 * time.Hour)},
				{URL: "newest", DateAdded: base.Add(3 * time.Hour), IsUnread: true},
				{URL: "second", DateAdded: base.Add(2 * time.Hour), IsUnread: true},
			},
			filter: config.FilterConfig{UnreadOnly: true, Since: "2026-10-23", Limit: 2},
			want:   []string{"newest", "second"},
		},
		{name: "empty", filter: config.FilterConfig{Limit: 1}},
		{
			name: "unlimited-invalid-since",
			items: []source.ReadingItem{
				{URL: "older", DateAdded: base},
				{URL: "newer", DateAdded: base.Add(time.Hour)},
			},
			filter: config.FilterConfig{Since: "invalid"},
			want:   []string{"newer", "older"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			for _, item := range applyFilters(tt.items, tt.filter) {
				got = append(got, item.URL)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("URLs = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDateOrderKeepsFirstURLMetadata(t *testing.T) {
	base := time.Date(2026, 10, 23, 8, 0, 0, 0, time.UTC)
	first := source.ReadingItem{URL: "duplicate", Title: "first", Source: "safari", DateAdded: base}
	middle := source.ReadingItem{URL: "middle", Source: "rss", DateAdded: base.Add(time.Hour)}
	items := []source.ReadingItem{
		first, middle,
		{URL: "duplicate", Title: "later", Source: "chrome", DateAdded: base.Add(2 * time.Hour)},
	}
	got := applyFilters(source.DeduplicateByURL(items), config.FilterConfig{})
	want := []source.ReadingItem{middle, first}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("items = %+v, want %+v", got, want)
	}
}
