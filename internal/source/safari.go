package source

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"howett.net/plist"
)

// FullDiskAccessError indicates Safari plist cannot be read due to macOS permissions.
type FullDiskAccessError struct {
	Path string
}

func (e *FullDiskAccessError) Error() string {
	return fmt.Sprintf("cannot read %s: permission denied (Full Disk Access required)", e.Path)
}

// FormatFullDiskAccessBanner returns a prominent multi-language banner for FDA errors.
func FormatFullDiskAccessBanner(path string) string {
	return "\n" +
		"╔══════════════════════════════════════════════════════════════════╗\n" +
		"║  ⚠  Full Disk Access Required                                  ║\n" +
		"╠══════════════════════════════════════════════════════════════════╣\n" +
		"║                                                                ║\n" +
		"║  Safari's Bookmarks.plist is protected by macOS.               ║\n" +
		"║  Your terminal app needs Full Disk Access to read it.          ║\n" +
		"║                                                                ║\n" +
		"║  → System Settings > Privacy & Security > Full Disk Access     ║\n" +
		"║    Add: Terminal / iTerm2 / Warp / VS Code / your terminal     ║\n" +
		"║                                                                ║\n" +
		"║  Or run this command to open the settings directly:            ║\n" +
		"║  open \"x-apple.systempreferences:com.apple.preference.        ║\n" +
		"║        security?Privacy_AllFiles\"                              ║\n" +
		"║                                                                ║\n" +
		"╠══════════════════════════════════════════════════════════════════╣\n" +
		"║  ⚠  フルディスクアクセスが必要です                             ║\n" +
		"╠══════════════════════════════════════════════════════════════════╣\n" +
		"║                                                                ║\n" +
		"║  Safari の Bookmarks.plist は macOS により保護されています。    ║\n" +
		"║  ターミナルアプリにフルディスクアクセスを許可してください。      ║\n" +
		"║                                                                ║\n" +
		"║  → システム設定 > プライバシーとセキュリティ >                  ║\n" +
		"║    フルディスクアクセス → ターミナルアプリを追加                ║\n" +
		"║                                                                ║\n" +
		"╠══════════════════════════════════════════════════════════════════╣\n" +
		"║  ⚠  需要「完整磁碟取用權限」                                   ║\n" +
		"╠══════════════════════════════════════════════════════════════════╣\n" +
		"║                                                                ║\n" +
		"║  Safari 的 Bookmarks.plist 受 macOS 保護。                     ║\n" +
		"║  請授予終端機 App「完整磁碟取用權限」。                         ║\n" +
		"║                                                                ║\n" +
		"║  → 系統設定 > 隱私與安全性 > 完整磁碟取用權限                  ║\n" +
		"║    加入你的終端機（Terminal / iTerm2 / Warp 等）                ║\n" +
		"║                                                                ║\n" +
		"╚══════════════════════════════════════════════════════════════════╝\n"
}

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
			return nil, &FullDiskAccessError{Path: s.plistPath}
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
