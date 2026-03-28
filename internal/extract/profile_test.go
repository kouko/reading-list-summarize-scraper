package extract

import (
	"os"
	"path/filepath"
	"testing"
)

const testLocalState = `{
    "profile": {
        "info_cache": {
            "Default": {"name": "Personal"},
            "Profile 1": {"name": "Work Account"},
            "Profile 5": {"name": "ReadingList-Auto"}
        }
    }
}`

func TestProfileResolver(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "Local State"), []byte(testLocalState), 0644)

	r, err := NewProfileResolver(dir)
	if err != nil {
		t.Fatalf("NewProfileResolver error: %v", err)
	}

	tests := []struct {
		input string
		want  string
	}{
		{"ReadingList-Auto", "Profile 5"},
		{"Personal", "Default"},
		{"Work Account", "Profile 1"},
		{"Default", "Default"},
		{"Profile 1", "Profile 1"},
	}

	for _, tt := range tests {
		got, err := r.Resolve(tt.input)
		if err != nil {
			t.Errorf("Resolve(%q) error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("Resolve(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}

	_, err = r.Resolve("Nonexistent")
	if err == nil {
		t.Error("Resolve(Nonexistent) should error")
	}
}
