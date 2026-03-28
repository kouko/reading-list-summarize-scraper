package output

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileIndex(t *testing.T) {
	dir := t.TempDir()
	domainDir := filepath.Join(dir, "example_com")
	os.MkdirAll(domainDir, 0755)
	os.WriteFile(filepath.Join(domainDir, "2026-03-28__a1b2c3d4__summary.md"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(domainDir, "2026-03-28__a1b2c3d4__content.md"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(domainDir, "2026-03-28__b2c3d4e5__content.md"), []byte("test"), 0644)

	idx := NewFileIndex()
	idx.Build(dir)

	if !idx.Has("a1b2c3d4") {
		t.Error("index should have a1b2c3d4")
	}
	info := idx.Get("a1b2c3d4")
	if !info.SummaryExists || !info.ContentExists {
		t.Error("a1b2c3d4 should have both files")
	}

	info = idx.Get("b2c3d4e5")
	if info.SummaryExists {
		t.Error("b2c3d4e5 should not have summary")
	}
	if !info.ContentExists {
		t.Error("b2c3d4e5 should have content")
	}

	if idx.Has("ffffffff") {
		t.Error("should not have ffffffff")
	}
}
