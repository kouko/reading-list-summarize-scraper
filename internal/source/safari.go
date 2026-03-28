package source

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"howett.net/plist"
)

const (
	typeLeaf          = "WebBookmarkTypeLeaf"
	typeList          = "WebBookmarkTypeList"
	readingListFolder = "com.apple.ReadingList"
)

type bookmarksPlist struct {
	Title           string          `plist:"Title"`
	WebBookmarkType string          `plist:"WebBookmarkType"`
	Children        []bookmarkEntry `plist:"Children"`
}

type bookmarkEntry struct {
	Title           string            `plist:"Title"`
	WebBookmarkType string            `plist:"WebBookmarkType"`
	URLString       string            `plist:"URLString"`
	URIDictionary   map[string]string `plist:"URIDictionary"`
	Children        []bookmarkEntry   `plist:"Children"`
	ReadingList     *readingListMeta  `plist:"ReadingList"`
}

type readingListMeta struct {
	DateAdded      time.Time `plist:"DateAdded"`
	DateLastViewed time.Time `plist:"DateLastViewed"`
	PreviewText    string    `plist:"PreviewText"`
}

type SafariSource struct {
	plistPath string
}

func NewSafariSource(plistPath string) *SafariSource {
	if plistPath == "" {
		home, _ := os.UserHomeDir()
		plistPath = filepath.Join(home, "Library", "Safari", "Bookmarks.plist")
	}
	return &SafariSource{plistPath: plistPath}
}

func (s *SafariSource) Name() string { return "safari" }

func (s *SafariSource) Fetch() ([]ReadingItem, error) {
	file, err := os.Open(s.plistPath)
	if err != nil {
		if os.IsPermission(err) {
			return nil, fmt.Errorf(
				"cannot read %s: permission denied\n"+
					"  Safari's Bookmarks.plist requires Full Disk Access.\n"+
					"  To fix: System Settings > Privacy & Security > Full Disk Access\n"+
					"  Add your terminal app (Terminal, iTerm2, Warp, etc.)\n"+
					"  Or run: open \"x-apple.systempreferences:com.apple.preference.security?Privacy_AllFiles\"",
				s.plistPath,
			)
		}
		return nil, fmt.Errorf("open %s: %w", s.plistPath, err)
	}
	defer file.Close()

	var root bookmarksPlist
	if err := plist.NewDecoder(file).Decode(&root); err != nil {
		return nil, fmt.Errorf("decode plist: %w", err)
	}

	for _, child := range root.Children {
		if child.Title == readingListFolder {
			return extractItems(child.Children), nil
		}
	}

	return nil, fmt.Errorf("Reading List folder (%s) not found", readingListFolder)
}

func extractItems(entries []bookmarkEntry) []ReadingItem {
	var items []ReadingItem
	for _, e := range entries {
		if e.WebBookmarkType != typeLeaf || e.URLString == "" {
			continue
		}
		item := ReadingItem{
			Title:  itemTitle(e),
			URL:    e.URLString,
			Source: "safari",
		}
		if e.ReadingList != nil {
			item.DateAdded = e.ReadingList.DateAdded
			item.IsUnread = e.ReadingList.DateLastViewed.IsZero()
			item.PreviewText = e.ReadingList.PreviewText
		}
		items = append(items, item)
	}
	return items
}

func itemTitle(e bookmarkEntry) string {
	if t, ok := e.URIDictionary["title"]; ok && t != "" {
		return t
	}
	return e.Title
}
