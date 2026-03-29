package source

import (
	"os"
	"path/filepath"
	"testing"
)

const testPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Title</key>
    <string></string>
    <key>WebBookmarkType</key>
    <string>WebBookmarkTypeList</string>
    <key>Children</key>
    <array>
        <dict>
            <key>Title</key>
            <string>com.apple.ReadingList</string>
            <key>WebBookmarkType</key>
            <string>WebBookmarkTypeList</string>
            <key>Children</key>
            <array>
                <dict>
                    <key>WebBookmarkType</key>
                    <string>WebBookmarkTypeLeaf</string>
                    <key>URLString</key>
                    <string>https://example.com/article1</string>
                    <key>URIDictionary</key>
                    <dict>
                        <key>title</key>
                        <string>Test Article</string>
                    </dict>
                    <key>ReadingList</key>
                    <dict>
                        <key>DateAdded</key>
                        <date>2026-03-25T10:30:00Z</date>
                    </dict>
                </dict>
                <dict>
                    <key>WebBookmarkType</key>
                    <string>WebBookmarkTypeLeaf</string>
                    <key>URLString</key>
                    <string>https://example.com/article2</string>
                    <key>URIDictionary</key>
                    <dict>
                        <key>title</key>
                        <string>Read Article</string>
                    </dict>
                    <key>ReadingList</key>
                    <dict>
                        <key>DateAdded</key>
                        <date>2026-03-20T08:00:00Z</date>
                        <key>DateLastViewed</key>
                        <date>2026-03-22T14:00:00Z</date>
                    </dict>
                </dict>
            </array>
        </dict>
    </array>
</dict>
</plist>`

func TestSafariSource(t *testing.T) {
	dir := t.TempDir()
	plistPath := filepath.Join(dir, "Bookmarks.plist")
	os.WriteFile(plistPath, []byte(testPlist), 0644)

	src := NewSafariSource(plistPath)
	if src.Name() != "safari" {
		t.Errorf("Name() = %q, want safari", src.Name())
	}

	items, err := src.Fetch()
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}

	// First item: unread
	if items[0].Title != "Test Article" {
		t.Errorf("item[0].Title = %q", items[0].Title)
	}
	if items[0].URL != "https://example.com/article1" {
		t.Errorf("item[0].URL = %q", items[0].URL)
	}
	if !items[0].IsUnread {
		t.Error("item[0] should be unread")
	}
	if items[0].Source != "safari" {
		t.Errorf("item[0].Source = %q", items[0].Source)
	}

	// Second item: read
	if items[1].IsUnread {
		t.Error("item[1] should be read")
	}
}
