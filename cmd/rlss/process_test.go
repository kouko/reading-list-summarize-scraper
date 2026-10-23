package main

// concern: timestamp precision, stable undated placement, and limiting before date ordering can select the wrong reading items.

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/kouko/reading-list-summarize-scraper/internal/config"
	"github.com/kouko/reading-list-summarize-scraper/internal/source"
	"gopkg.in/yaml.v3"
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
				{URL: "new-safari", Source: "safari", DateAdded: base.Add(2 * time.Hour)},
				{URL: "rss", Source: "rss", DateAdded: base.Add(time.Hour)},
			},
			want: []string{"new-safari", "rss", "safari"},
		},
		{
			name: "first-url-occurrence",
			items: []source.ReadingItem{
				{URL: "duplicate", Title: "first", Source: "safari", DateAdded: base},
				{URL: "middle", Source: "rss", DateAdded: base.Add(time.Hour)},
				{URL: "duplicate", Title: "later", Source: "rss", DateAdded: base.Add(2 * time.Hour)},
			},
			want: []string{"middle", "duplicate"},
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
			dir := t.TempDir()
			var bookmarks, feed strings.Builder
			bookmarks.WriteString(`<plist version="1.0"><dict><key>Children</key><array><dict><key>Title</key><string>com.apple.ReadingList</string><key>Children</key><array>`)
			feed.WriteString(`<rss version="2.0"><channel><title>Fixture</title>`)
			for _, item := range tt.items {
				url := "https://example.test/" + item.URL
				if item.Source == "rss" {
					fmt.Fprintf(&feed, "<item><title>%s</title><link>%s</link><pubDate>%s</pubDate></item>", item.Title, url, item.DateAdded.Format(time.RFC1123Z))
					continue
				}
				fmt.Fprintf(&bookmarks, `<dict><key>WebBookmarkType</key><string>WebBookmarkTypeLeaf</string><key>URLString</key><string>%s</string><key>Title</key><string>%s</string><key>ReadingList</key><dict>`, url, item.Title)
				if !item.DateAdded.IsZero() {
					fmt.Fprintf(&bookmarks, "<key>DateAdded</key><date>%s</date>", item.DateAdded.Format(time.RFC3339Nano))
				}
				if !item.IsUnread {
					fmt.Fprintf(&bookmarks, "<key>DateLastViewed</key><date>%s</date>", base.Format(time.RFC3339))
				}
				bookmarks.WriteString(`</dict></dict>`)
			}
			bookmarks.WriteString(`</array></dict></array></dict></plist>`)
			feed.WriteString(`</channel></rss>`)
			plistPath := filepath.Join(dir, "Bookmarks.plist")
			if err := os.WriteFile(plistPath, []byte(bookmarks.String()), 0600); err != nil {
				t.Fatal(err)
			}
			cfg := config.DefaultConfig()
			cfg.Safari.Enabled, cfg.Chrome.Enabled = true, false
			cfg.Safari.PlistPath = plistPath
			cfg.Filter, cfg.OutputDir, cfg.Extract.UserDataDir = tt.filter, dir, dir
			if strings.Contains(feed.String(), "<item>") {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprint(w, feed.String())
				}))
				defer server.Close()
				cfg.RSS.Enabled, cfg.RSS.Feeds, cfg.RSS.Count = true, []string{server.URL}, len(tt.items)
			} else {
				cfg.RSS.Enabled = false
			}
			configData, err := yaml.Marshal(cfg)
			if err != nil {
				t.Fatal(err)
			}
			configPath := filepath.Join(dir, "config.yaml")
			if err := os.WriteFile(configPath, configData, 0600); err != nil {
				t.Fatal(err)
			}
			oldConfig, oldStdout := flagConfigPath, os.Stdout
			flagConfigPath = configPath
			stdout, err := os.CreateTemp(dir, "stdout")
			if err != nil {
				t.Fatal(err)
			}
			os.Stdout = stdout
			t.Cleanup(func() { flagConfigPath, os.Stdout = oldConfig, oldStdout; stdout.Close() })
			if err := runList(listCmd, nil); err != nil {
				t.Fatal(err)
			}
			out, err := os.ReadFile(stdout.Name())
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, line := range strings.Split(string(out), "\n") {
				fields := strings.Fields(line)
				if len(fields) > 0 && strings.HasPrefix(fields[0], "https://example.test/") {
					got = append(got, strings.TrimPrefix(fields[0], "https://example.test/"))
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("URLs = %v, want %v", got, tt.want)
			}
			if tt.name == "first-url-occurrence" && (!strings.Contains(string(out), "  first\n") || !strings.Contains(string(out), "(safari, 2026-10-23)")) {
				t.Fatalf("first URL metadata missing: %s", out)
			}
		})
	}
}
