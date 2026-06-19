package summarize

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kouko/reading-list-summarize-scraper/internal/config"
)

func TestResolvePrompt_InlinePrompt(t *testing.T) {
	got, err := ResolvePrompt(config.SummaryConfig{Prompt: "Summarize this: {{title}}"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Summarize this: {{title}}" {
		t.Errorf("got %q, want the inline prompt verbatim", got)
	}
}

func TestResolvePrompt_FileTakesPrecedenceOverInline(t *testing.T) {
	dir := t.TempDir()
	pf := filepath.Join(dir, "prompt.md")
	if err := os.WriteFile(pf, []byte("FROM FILE {{content}}"), 0644); err != nil {
		t.Fatal(err)
	}
	// Both file and inline set → file wins (level 1 before level 2).
	got, err := ResolvePrompt(config.SummaryConfig{SummaryPromptFile: pf, Prompt: "inline"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "FROM FILE {{content}}" {
		t.Errorf("got %q, want the file contents (file precedence)", got)
	}
}

func TestResolvePrompt_FileMissing_Errors(t *testing.T) {
	if _, err := ResolvePrompt(config.SummaryConfig{SummaryPromptFile: "/no/such/prompt.md"}); err == nil {
		t.Error("missing prompt file should error")
	}
}

func TestResolvePrompt_BuiltinByLanguage(t *testing.T) {
	// Neither file nor inline → builtin template for the language (raw, with
	// placeholders still present since substitution happens later).
	got, err := ResolvePrompt(config.SummaryConfig{Language: "en"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(got) == "" {
		t.Error("builtin en summary prompt should be non-empty")
	}
	if !strings.Contains(got, "{{content}}") {
		t.Errorf("builtin summary template should contain the {{content}} placeholder, got %q", truncate(got))
	}
}

func TestReadPromptFile(t *testing.T) {
	dir := t.TempDir()
	pf := filepath.Join(dir, "p.md")
	if err := os.WriteFile(pf, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := readPromptFile(pf)
	if err != nil || got != "hello" {
		t.Errorf("readPromptFile = (%q, %v), want (hello, nil)", got, err)
	}
	if _, err := readPromptFile(filepath.Join(dir, "missing.md")); err == nil {
		t.Error("readPromptFile on a missing file should error")
	}
}

func TestBuiltinPromptPrefix(t *testing.T) {
	// "article" selects the article-style built-in; everything else (incl. the
	// default "", "outline", and typos) selects the plain outline prefix.
	if got := builtinPromptPrefix("article"); got != "summary-article" {
		t.Errorf("builtinPromptPrefix(article) = %q, want summary-article", got)
	}
	for _, style := range []string{"", "outline", "Article", "classic", "xxx"} {
		if got := builtinPromptPrefix(style); got != "summary" {
			t.Errorf("builtinPromptPrefix(%q) = %q, want summary (outline default)", style, got)
		}
	}
}

func TestLoadBuiltinPrompt_KnownAndFallback(t *testing.T) {
	// Known language (outline default style).
	if got, err := loadBuiltinPrompt("ja", ""); err != nil || strings.TrimSpace(got) == "" {
		t.Errorf("loadBuiltinPrompt(ja) = (%q, %v), want non-empty", truncate(got), err)
	}
	// Unknown language → falls back to en (non-empty, no error).
	got, err := loadBuiltinPrompt("zz-not-a-language", "")
	if err != nil {
		t.Fatalf("unknown language should fall back to en, got error: %v", err)
	}
	if strings.TrimSpace(got) == "" {
		t.Error("fallback (en) prompt should be non-empty")
	}
}

func TestKeywordPrompt(t *testing.T) {
	got, err := KeywordPrompt("MY SUMMARY TEXT", "en", 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "7") {
		t.Error("keyword prompt should substitute {{count}} → 7")
	}
	if !strings.Contains(got, "MY SUMMARY TEXT") {
		t.Error("keyword prompt should substitute {{summary}}")
	}
	if strings.Contains(got, "{{count}}") || strings.Contains(got, "{{summary}}") {
		t.Error("keyword prompt should leave no unsubstituted placeholders")
	}
	// Unknown language must not error (falls back to en).
	if _, err := KeywordPrompt("s", "zz", 3); err != nil {
		t.Errorf("unknown language should fall back, got error: %v", err)
	}
}

func TestMermaidPrompt(t *testing.T) {
	got, err := MermaidPrompt("MY SUMMARY TEXT", "en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "MY SUMMARY TEXT") {
		t.Error("mermaid prompt should substitute {{summary}}")
	}
	if strings.Contains(got, "{{summary}}") {
		t.Error("mermaid prompt should leave no unsubstituted {{summary}}")
	}
}

func TestResolveAndSubstitute_Inline(t *testing.T) {
	// Inline prompt with no {{content}} → content appended after the prompt.
	got, err := ResolveAndSubstitute(
		config.SummaryConfig{Prompt: "Title: {{title}}"},
		PromptVars{Title: "Hello", Content: "BODY TEXT"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "Title: Hello") {
		t.Error("should substitute {{title}}")
	}
	if !strings.Contains(got, "BODY TEXT") {
		t.Error("content should be appended when template has no {{content}}")
	}
}

func TestResolveAndSubstitute_BuiltinSubstitutesContent(t *testing.T) {
	got, err := ResolveAndSubstitute(
		config.SummaryConfig{Language: "en"},
		PromptVars{Title: "T", Content: "ARTICLE BODY", ContentLength: 12, Language: "en"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "ARTICLE BODY") {
		t.Error("builtin template's {{content}} should be substituted with the article body")
	}
	if strings.Contains(got, "{{content}}") {
		t.Error("no {{content}} placeholder should remain after substitution")
	}
}

func truncate(s string) string {
	if len(s) > 80 {
		return s[:80] + "…"
	}
	return s
}
