package source

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"google.golang.org/protobuf/encoding/protowire"
)

const readingListKeyPrefix = "reading_list-dt-"

// ChromeLevelDBSource reads Chrome Reading List directly from the Sync Data LevelDB.
// This approach does NOT require launching Chrome or any extensions.
// It reads the LevelDB files directly (read-only), bypassing SingletonLock issues.
type ChromeLevelDBSource struct {
	syncDataDir string // path to "Sync Data/LevelDB" directory
}

// NewChromeLevelDBSource creates a source that reads from the given profile's
// Sync Data LevelDB directory. profileDir is the Chrome profile folder path
// (e.g., "~/Library/Application Support/Google/Chrome/Default").
func NewChromeLevelDBSource(profileDir string) *ChromeLevelDBSource {
	return &ChromeLevelDBSource{
		syncDataDir: filepath.Join(profileDir, "Sync Data", "LevelDB"),
	}
}

func (c *ChromeLevelDBSource) Name() string { return "chrome" }

func (c *ChromeLevelDBSource) Fetch() ([]ReadingItem, error) {
	// Copy LevelDB to temp dir to avoid lock conflicts with running Chrome.
	tmpDir, err := copyLevelDB(c.syncDataDir)
	if err != nil {
		return nil, fmt.Errorf("copy LevelDB: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := leveldb.OpenFile(tmpDir, &opt.Options{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("open LevelDB %s: %w", tmpDir, err)
	}
	defer db.Close()

	iter := db.NewIterator(nil, nil)
	defer iter.Release()

	var items []ReadingItem
	for iter.Next() {
		key := string(iter.Key())
		if !strings.HasPrefix(key, readingListKeyPrefix) {
			continue
		}

		url := strings.TrimPrefix(key, readingListKeyPrefix)
		if url == "" {
			continue
		}

		entry, err := parseReadingListEntry(iter.Value())
		if err != nil {
			slog.Debug("skip malformed reading list entry", "url", url, "err", err)
			continue
		}

		// Prefer URL from protobuf if available, fall back to key
		itemURL := url
		if entry.url != "" {
			itemURL = entry.url
		}

		items = append(items, ReadingItem{
			Title:    entry.title,
			URL:      itemURL,
			DateAdded: entry.creationTime,
			IsUnread:  entry.status == 0, // 0=UNREAD, 1=READ, 2=UNSEEN
			Source:    "chrome",
		})
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("iterate LevelDB: %w", err)
	}

	slog.Info("read chrome reading list from LevelDB", "count", len(items), "dir", c.syncDataDir)
	return items, nil
}

// copyLevelDB copies the LevelDB directory to a temp location to avoid
// lock conflicts with a running Chrome instance. Only copies .ldb, .log,
// CURRENT, MANIFEST, and LOG files (skips LOCK file).
func copyLevelDB(srcDir string) (string, error) {
	tmpDir := filepath.Join(os.TempDir(), "rlss-leveldb-copy")
	os.RemoveAll(tmpDir) // clean previous copy
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", err
	}

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return "", fmt.Errorf("read dir %s: %w", srcDir, err)
	}

	for _, e := range entries {
		if e.IsDir() || e.Name() == "LOCK" {
			continue
		}
		src := filepath.Join(srcDir, e.Name())
		dst := filepath.Join(tmpDir, e.Name())
		data, err := os.ReadFile(src)
		if err != nil {
			slog.Debug("skip leveldb file", "name", e.Name(), "err", err)
			continue
		}
		if err := os.WriteFile(dst, data, 0644); err != nil {
			return tmpDir, err
		}
	}

	return tmpDir, nil
}

// readingListEntry holds parsed fields from ReadingListSpecifics protobuf.
type readingListEntry struct {
	entryID      string
	title        string
	url          string
	creationTime time.Time
	updateTime   time.Time
	status       int // 0=UNREAD, 1=READ, 2=UNSEEN
}

// parseReadingListEntry decodes a ReadingListSpecifics protobuf message.
//
// Wire format (from Chromium reading_list_specifics.proto):
//
//	field 1: entry_id (string)
//	field 2: title (string)
//	field 3: url (string)
//	field 4: creation_time_us (int64, microseconds since epoch)
//	field 5: update_time_us (int64, microseconds since epoch)
//	field 6: status (varint: 0=UNREAD, 1=READ, 2=UNSEEN)
//	field 7: first_read_time_us (int64)
//	field 8: update_title_time_us (int64)
//	field 9: estimated_read_time_seconds (int32)
func parseReadingListEntry(data []byte) (*readingListEntry, error) {
	entry := &readingListEntry{}

	for len(data) > 0 {
		num, typ, n := protowire.ConsumeTag(data)
		if n < 0 {
			return entry, nil // stop on invalid tag (may hit trailing bytes)
		}
		data = data[n:]

		switch typ {
		case protowire.VarintType:
			val, n := protowire.ConsumeVarint(data)
			if n < 0 {
				return entry, nil
			}
			data = data[n:]
			switch num {
			case 4: // creation_time_us
				entry.creationTime = microsToTime(int64(val))
			case 5: // update_time_us
				entry.updateTime = microsToTime(int64(val))
			case 6: // status
				entry.status = int(val)
			}

		case protowire.BytesType:
			val, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return entry, nil
			}
			data = data[n:]
			switch num {
			case 1: // entry_id
				entry.entryID = string(val)
			case 2: // title
				entry.title = string(val)
			case 3: // url
				entry.url = string(val)
			}

		case protowire.Fixed32Type:
			_, n := protowire.ConsumeFixed32(data)
			if n < 0 {
				return entry, nil
			}
			data = data[n:]

		case protowire.Fixed64Type:
			_, n := protowire.ConsumeFixed64(data)
			if n < 0 {
				return entry, nil
			}
			data = data[n:]

		default:
			return entry, nil // unknown wire type, stop
		}
	}

	return entry, nil
}

func microsToTime(us int64) time.Time {
	if us <= 0 {
		return time.Time{}
	}
	return time.UnixMicro(us)
}
