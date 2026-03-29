package output

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kouko/reading-list-summarize-scraper/internal/config"
)

func TestExpandTemplate(t *testing.T) {
	vars := CopyToVars{
		OutputDir: "/vault/references",
		Date:      "2026-03-28",
		DateAdded: "2026-03-25",
		Title:     "Test Article Title",
		SHA8:      "a1b2c3d4",
		Source:    "safari",
		Domain:    "example.com",
		DomainDir: "example_com",
		Type:      "summary",
	}

	got := ExpandTemplate("{output_dir}/by-source/{source}/{domain_dir}", vars)
	want := "/vault/references/by-source/safari/example_com"
	if got != want {
		t.Errorf("ExpandTemplate path = %q, want %q", got, want)
	}

	got = ExpandTemplate("{date}__{title}__{type}.md", vars)
	want = "2026-03-28__Test Article Title__summary.md"
	if got != want {
		t.Errorf("ExpandTemplate filename = %q, want %q", got, want)
	}
}

func TestSanitizeTitleForDisplay(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello/World: Test", "Hello World Test"},
		{"日本語テスト", "日本語テスト"},
		{`A "quoted" <title>`, "A quoted title"},
	}
	for _, tt := range tests {
		got := SanitizeTitleForDisplay(tt.input)
		if got != tt.want {
			t.Errorf("SanitizeTitleForDisplay(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestExecuteCopyTo_Enabled(t *testing.T) {
	// Set up source directory with a matching file
	srcDir := t.TempDir()
	targetBase := t.TempDir()

	sha8 := "a1b2c3d4"
	srcContent := "# Summary\nTest content"
	srcFile := filepath.Join(srcDir, "2026-03-28__a1b2c3d4__summary.md")
	if err := os.WriteFile(srcFile, []byte(srcContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := config.CopyToConfig{
		Enabled:   true,
		Path:      filepath.Join(targetBase, "output"),
		Filename:  "{sha8}__{type}.md",
		Files:     []string{"summary"},
		Overwrite: true,
	}

	vars := CopyToVars{SHA8: sha8}
	if err := ExecuteCopyTo(cfg, srcDir, sha8, vars); err != nil {
		t.Fatalf("ExecuteCopyTo() error: %v", err)
	}

	targetPath := filepath.Join(targetBase, "output", sha8+"__summary.md")
	data, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("target file not created: %v", err)
	}
	if string(data) != srcContent {
		t.Errorf("content = %q, want %q", string(data), srcContent)
	}
}

func TestExecuteCopyTo_Disabled(t *testing.T) {
	cfg := config.CopyToConfig{Enabled: false}
	err := ExecuteCopyTo(cfg, "/nonexistent", "abc12345", CopyToVars{})
	if err != nil {
		t.Errorf("disabled should be no-op, got error: %v", err)
	}
}

func TestExecuteCopyTo_OverwriteFalse(t *testing.T) {
	srcDir := t.TempDir()
	targetBase := t.TempDir()
	sha8 := "a1b2c3d4"

	// Create source file
	srcFile := filepath.Join(srcDir, "2026-03-28__a1b2c3d4__summary.md")
	os.WriteFile(srcFile, []byte("new content"), 0644)

	// Pre-create target directory and file
	targetDir := filepath.Join(targetBase, "out")
	os.MkdirAll(targetDir, 0755)
	targetPath := filepath.Join(targetDir, sha8+"__summary.md")
	os.WriteFile(targetPath, []byte("old content"), 0644)

	cfg := config.CopyToConfig{
		Enabled:   true,
		Path:      targetDir,
		Filename:  "{sha8}__{type}.md",
		Files:     []string{"summary"},
		Overwrite: false,
	}

	vars := CopyToVars{SHA8: sha8}
	if err := ExecuteCopyTo(cfg, srcDir, sha8, vars); err != nil {
		t.Fatalf("ExecuteCopyTo() error: %v", err)
	}

	data, _ := os.ReadFile(targetPath)
	if string(data) != "old content" {
		t.Errorf("overwrite=false should skip; got %q, want %q", string(data), "old content")
	}
}

func TestExecuteCopyTo_OverwriteTrue(t *testing.T) {
	srcDir := t.TempDir()
	targetBase := t.TempDir()
	sha8 := "a1b2c3d4"

	srcFile := filepath.Join(srcDir, "2026-03-28__a1b2c3d4__summary.md")
	os.WriteFile(srcFile, []byte("new content"), 0644)

	targetDir := filepath.Join(targetBase, "out")
	os.MkdirAll(targetDir, 0755)
	targetPath := filepath.Join(targetDir, sha8+"__summary.md")
	os.WriteFile(targetPath, []byte("old content"), 0644)

	cfg := config.CopyToConfig{
		Enabled:   true,
		Path:      targetDir,
		Filename:  "{sha8}__{type}.md",
		Files:     []string{"summary"},
		Overwrite: true,
	}

	vars := CopyToVars{SHA8: sha8}
	if err := ExecuteCopyTo(cfg, srcDir, sha8, vars); err != nil {
		t.Fatalf("ExecuteCopyTo() error: %v", err)
	}

	data, _ := os.ReadFile(targetPath)
	if string(data) != "new content" {
		t.Errorf("overwrite=true should replace; got %q, want %q", string(data), "new content")
	}
}
