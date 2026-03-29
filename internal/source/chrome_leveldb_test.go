package source

import (
	"testing"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"google.golang.org/protobuf/encoding/protowire"
)

// encodeReadingListEntry builds a ReadingListSpecifics protobuf message
// using raw protowire encoding (no generated proto code needed).
func encodeReadingListEntry(entryID, title, url string, creationTimeUs int64, status int) []byte {
	var b []byte
	b = protowire.AppendTag(b, 1, protowire.BytesType)
	b = protowire.AppendString(b, entryID)
	b = protowire.AppendTag(b, 2, protowire.BytesType)
	b = protowire.AppendString(b, title)
	b = protowire.AppendTag(b, 3, protowire.BytesType)
	b = protowire.AppendString(b, url)
	b = protowire.AppendTag(b, 4, protowire.VarintType)
	b = protowire.AppendVarint(b, uint64(creationTimeUs))
	b = protowire.AppendTag(b, 6, protowire.VarintType)
	b = protowire.AppendVarint(b, uint64(status))
	return b
}

func TestParseReadingListEntry(t *testing.T) {
	now := time.Now()
	creationUs := now.UnixMicro()

	data := encodeReadingListEntry("id-1", "Test Article", "https://example.com/article", creationUs, 0)
	entry, err := parseReadingListEntry(data)
	if err != nil {
		t.Fatalf("parseReadingListEntry returned error: %v", err)
	}

	if entry.entryID != "id-1" {
		t.Errorf("entryID = %q, want %q", entry.entryID, "id-1")
	}
	if entry.title != "Test Article" {
		t.Errorf("title = %q, want %q", entry.title, "Test Article")
	}
	if entry.url != "https://example.com/article" {
		t.Errorf("url = %q, want %q", entry.url, "https://example.com/article")
	}
	if entry.status != 0 {
		t.Errorf("status = %d, want 0 (UNREAD)", entry.status)
	}
	if entry.creationTime.UnixMicro() != creationUs {
		t.Errorf("creationTime = %v, want %v", entry.creationTime, now)
	}
}

func TestParseReadingListEntry_ReadStatus(t *testing.T) {
	data := encodeReadingListEntry("id-2", "Read Article", "https://example.com/read", 1000000, 1)
	entry, err := parseReadingListEntry(data)
	if err != nil {
		t.Fatalf("parseReadingListEntry returned error: %v", err)
	}
	if entry.status != 1 {
		t.Errorf("status = %d, want 1 (READ)", entry.status)
	}
}

func TestParseReadingListEntry_Malformed(t *testing.T) {
	// Completely garbage bytes -- should not panic, returns partial entry.
	entry, err := parseReadingListEntry([]byte{0xff, 0xfe, 0xfd})
	if err != nil {
		t.Fatalf("unexpected error on malformed data: %v", err)
	}
	// Should return an empty/partial entry without crashing.
	if entry == nil {
		t.Fatal("expected non-nil entry even for malformed data")
	}
}

func TestParseReadingListEntry_Empty(t *testing.T) {
	entry, err := parseReadingListEntry([]byte{})
	if err != nil {
		t.Fatalf("unexpected error on empty data: %v", err)
	}
	if entry == nil {
		t.Fatal("expected non-nil entry for empty data")
	}
	if entry.entryID != "" || entry.title != "" || entry.url != "" {
		t.Errorf("expected empty fields for empty data, got %+v", entry)
	}
}

func TestChromeLevelDBSource_FetchFromTempDB(t *testing.T) {
	// Create a temporary LevelDB with fake reading list entries.
	tmpDir := t.TempDir()

	db, err := leveldb.OpenFile(tmpDir, nil)
	if err != nil {
		t.Fatalf("open temp LevelDB: %v", err)
	}

	creationUs := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC).UnixMicro()

	// UNREAD entry
	unreadVal := encodeReadingListEntry("entry-1", "Unread Article", "https://example.com/unread", creationUs, 0)
	if err := db.Put([]byte("reading_list-dt-https://example.com/unread"), unreadVal, nil); err != nil {
		t.Fatal(err)
	}

	// READ entry
	readVal := encodeReadingListEntry("entry-2", "Read Article", "https://example.com/read", creationUs, 1)
	if err := db.Put([]byte("reading_list-dt-https://example.com/read"), readVal, nil); err != nil {
		t.Fatal(err)
	}

	// Non-reading-list entry (should be ignored)
	if err := db.Put([]byte("some-other-key"), []byte("irrelevant"), nil); err != nil {
		t.Fatal(err)
	}

	// Malformed reading list entry (should be skipped gracefully)
	if err := db.Put([]byte("reading_list-dt-https://example.com/broken"), []byte{0xff, 0xfe}, nil); err != nil {
		t.Fatal(err)
	}

	db.Close()

	// The ChromeLevelDBSource expects profileDir, and appends "Sync Data/LevelDB".
	// We need to set up that directory structure, or test the internal parsing directly.
	// Since Fetch() copies from syncDataDir, we can test by pointing syncDataDir at our temp DB.
	src := &ChromeLevelDBSource{syncDataDir: tmpDir}
	items, err := src.Fetch()
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	// The malformed entry ({0xff, 0xfe}) does not cause a parse error --
	// parseReadingListEntry returns a partial (empty) entry, and the URL is
	// extracted from the key. So we get 3 items total.
	if len(items) != 3 {
		t.Fatalf("got %d items, want 3", len(items))
	}

	// Build a URL -> item map for easier assertions.
	byURL := make(map[string]ReadingItem)
	for _, item := range items {
		byURL[item.URL] = item
	}

	unread, ok := byURL["https://example.com/unread"]
	if !ok {
		t.Fatal("missing unread item")
	}
	if unread.Title != "Unread Article" {
		t.Errorf("unread title = %q, want %q", unread.Title, "Unread Article")
	}
	if !unread.IsUnread {
		t.Error("expected unread item to have IsUnread=true")
	}
	if unread.Source != "chrome" {
		t.Errorf("source = %q, want %q", unread.Source, "chrome")
	}

	readItem, ok := byURL["https://example.com/read"]
	if !ok {
		t.Fatal("missing read item")
	}
	if readItem.Title != "Read Article" {
		t.Errorf("read title = %q, want %q", readItem.Title, "Read Article")
	}
	if readItem.IsUnread {
		t.Error("expected read item to have IsUnread=false")
	}
}

func TestChromeLevelDBSource_EmptyDB(t *testing.T) {
	tmpDir := t.TempDir()

	db, err := leveldb.OpenFile(tmpDir, nil)
	if err != nil {
		t.Fatalf("open temp LevelDB: %v", err)
	}
	// Add some non-reading-list keys.
	if err := db.Put([]byte("other-data"), []byte("value"), nil); err != nil {
		t.Fatal(err)
	}
	db.Close()

	src := &ChromeLevelDBSource{syncDataDir: tmpDir}
	items, err := src.Fetch()
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("got %d items from empty DB, want 0", len(items))
	}
}

func TestNewChromeLevelDBSource(t *testing.T) {
	src := NewChromeLevelDBSource("/fake/profile")
	if src.Name() != "chrome" {
		t.Errorf("Name() = %q, want %q", src.Name(), "chrome")
	}
	// Verify the syncDataDir is constructed correctly.
	want := "/fake/profile/Sync Data/LevelDB"
	if src.syncDataDir != want {
		t.Errorf("syncDataDir = %q, want %q", src.syncDataDir, want)
	}
}

func TestMicrosToTime(t *testing.T) {
	// Zero should return zero time.
	if !microsToTime(0).IsZero() {
		t.Error("microsToTime(0) should return zero time")
	}
	if !microsToTime(-1).IsZero() {
		t.Error("microsToTime(-1) should return zero time")
	}

	us := int64(1705312800000000) // 2024-01-15T10:00:00Z
	got := microsToTime(us)
	if got.UnixMicro() != us {
		t.Errorf("microsToTime(%d) = %v, want UnixMicro=%d", us, got, us)
	}
}
