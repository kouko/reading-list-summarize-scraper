package output

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var filePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}__([a-f0-9]{8})__(summary|content)\.md$`)

type FileInfo struct {
	Dir           string
	SummaryExists bool
	ContentExists bool
}

type FileIndex struct {
	entries map[string]*FileInfo
}

func NewFileIndex() *FileIndex {
	return &FileIndex{entries: make(map[string]*FileInfo)}
}

func (idx *FileIndex) Build(outputDir string) {
	idx.entries = make(map[string]*FileInfo)
	dirs, _ := os.ReadDir(outputDir)
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		domainDir := filepath.Join(outputDir, d.Name())
		files, _ := os.ReadDir(domainDir)
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			m := filePattern.FindStringSubmatch(f.Name())
			if m == nil {
				continue
			}
			sha8 := m[1]
			fileType := m[2]
			info, ok := idx.entries[sha8]
			if !ok {
				info = &FileInfo{Dir: domainDir}
				idx.entries[sha8] = info
			}
			switch fileType {
			case "summary":
				info.SummaryExists = true
			case "content":
				info.ContentExists = true
			}
		}
	}
}

func (idx *FileIndex) Has(sha8 string) bool {
	_, ok := idx.entries[sha8]
	return ok
}

func (idx *FileIndex) Get(sha8 string) FileInfo {
	if info, ok := idx.entries[sha8]; ok {
		return *info
	}
	return FileInfo{}
}

func (idx *FileIndex) ContentPath(sha8 string) string {
	info, ok := idx.entries[sha8]
	if !ok || !info.ContentExists {
		return ""
	}
	files, _ := os.ReadDir(info.Dir)
	for _, f := range files {
		if strings.Contains(f.Name(), sha8) && strings.HasSuffix(f.Name(), "__content.md") {
			return filepath.Join(info.Dir, f.Name())
		}
	}
	return ""
}
