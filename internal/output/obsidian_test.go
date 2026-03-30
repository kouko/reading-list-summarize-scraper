package output

import (
	"strings"
	"testing"
	"time"
)

func TestAssembleSummary(t *testing.T) {
	params := SummaryParams{
		Title:         "Test Article",
		URL:           "https://example.com/test",
		Domain:        "example.com",
		Source:        "safari",
		DateAdded:     time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC),
		ProcessedDate: time.Date(2026, 3, 28, 0, 0, 0, 0, time.UTC),
		LLMProvider:   "claude-code",
		LLMModel:      "haiku",
		ContentLength: 3200,
		ContentTier:   "中篇",
		SummaryText:   "### 概述\n\nThis is a test summary.",
		Keywords:      []string{"golang", "testing"},
	}

	result := AssembleSummary(params)

	if !strings.Contains(result, `title: "Test Article"`) {
		t.Error("missing title in frontmatter")
	}
	if !strings.Contains(result, `url: "https://example.com/test"`) {
		t.Error("missing url in frontmatter")
	}
	if !strings.Contains(result, "- golang") {
		t.Error("missing keyword tag")
	}
	if strings.Contains(result, "# Test Article\n") {
		t.Error("should not have # Title heading")
	}
	if !strings.Contains(result, "[!info]") {
		t.Error("missing info callout")
	}
}

func TestSourceDisplayName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"safari", "Safari"},
		{"chrome", "Chrome"},
		{"manual", "Manual"},
		{"unknown", "Manual"},
		{"", "Manual"},
	}
	for _, tt := range tests {
		got := sourceDisplayName(tt.input)
		if got != tt.want {
			t.Errorf("sourceDisplayName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestInsertMermaidBlocks(t *testing.T) {
	t.Run("matched heading", func(t *testing.T) {
		body := "### Overview\n\nSome text.\n\n### Details\n\nMore text."
		blocks := []MermaidBlock{{Title: "### Overview", Code: "graph LR\nA-->B"}}
		result := insertMermaidBlocks(body, blocks)
		if !strings.Contains(result, "```mermaid\ngraph LR\nA-->B\n```") {
			t.Errorf("mermaid block not inserted, got:\n%s", result)
		}
		// Mermaid should appear between Overview content and Details heading
		overviewIdx := strings.Index(result, "### Overview")
		mermaidIdx := strings.Index(result, "```mermaid")
		detailsIdx := strings.Index(result, "### Details")
		if mermaidIdx < overviewIdx || mermaidIdx > detailsIdx {
			t.Errorf("mermaid block not in correct position: overview=%d mermaid=%d details=%d", overviewIdx, mermaidIdx, detailsIdx)
		}
	})

	t.Run("unmatched falls back to after first section", func(t *testing.T) {
		body := "### First\n\nContent.\n\n### Second\n\nMore."
		blocks := []MermaidBlock{{Title: "### Nonexistent", Code: "graph LR\nX-->Y"}}
		result := insertMermaidBlocks(body, blocks)
		if !strings.Contains(result, "```mermaid") {
			t.Error("mermaid block should still be inserted")
		}
		mermaidIdx := strings.Index(result, "```mermaid")
		secondIdx := strings.Index(result, "### Second")
		if mermaidIdx > secondIdx {
			t.Errorf("unmatched block should be before second heading: mermaid=%d second=%d", mermaidIdx, secondIdx)
		}
	})

	t.Run("no headings", func(t *testing.T) {
		body := "Just some text without headings."
		blocks := []MermaidBlock{{Title: "### Missing", Code: "graph LR\nA-->B"}}
		result := insertMermaidBlocks(body, blocks)
		if !strings.Contains(result, "```mermaid") {
			t.Error("mermaid block should be appended")
		}
	})

	t.Run("empty blocks", func(t *testing.T) {
		body := "### Heading\n\nContent."
		result := insertMermaidBlocks(body, nil)
		if result != body {
			t.Errorf("empty blocks should return body unchanged, got:\n%s", result)
		}
	})
}

func TestAssembleSummary_EmbedContent(t *testing.T) {
	base := SummaryParams{
		Title:         "Test Article",
		URL:           "https://example.com/test",
		Domain:        "example.com",
		Source:        "safari",
		DateAdded:     time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC),
		ProcessedDate: time.Date(2026, 3, 28, 0, 0, 0, 0, time.UTC),
		LLMProvider:   "claude-code",
		LLMModel:      "haiku",
		ContentLength: 3200,
		ContentTier:   "中篇",
		SummaryText:   "### 概述\n\nSummary text here.",
	}

	t.Run("no embed when empty", func(t *testing.T) {
		p := base
		p.EmbedContent = ""
		result := AssembleSummary(p)
		if strings.Contains(result, "原文內容") || strings.Contains(result, "Original Content") {
			t.Error("should not contain embed heading when EmbedContent is empty")
		}
		if !strings.Contains(result, "embed_content: false") {
			t.Error("frontmatter should have embed_content: false")
		}
	})

	t.Run("embed zh-Hant", func(t *testing.T) {
		p := base
		p.EmbedContent = "# 原始文章\n\n這是原文。"
		p.Language = "zh-Hant"
		result := AssembleSummary(p)
		if !strings.Contains(result, "## 原文內容") {
			t.Error("missing zh-Hant embed heading")
		}
		if !strings.Contains(result, "這是原文。") {
			t.Error("missing embedded content")
		}
		if !strings.Contains(result, "embed_content: true") {
			t.Error("frontmatter should have embed_content: true")
		}
	})

	t.Run("embed ja", func(t *testing.T) {
		p := base
		p.EmbedContent = "# 元の記事\n\nこれは原文です。"
		p.Language = "ja"
		result := AssembleSummary(p)
		if !strings.Contains(result, "## 原文") {
			t.Error("missing ja embed heading")
		}
		// Ensure it's exactly "## 原文" not "## 原文內容"
		if strings.Contains(result, "## 原文內容") {
			t.Error("should use ja heading, not zh-Hant")
		}
	})

	t.Run("embed en", func(t *testing.T) {
		p := base
		p.EmbedContent = "# Original Article\n\nThis is the content."
		p.Language = "en"
		result := AssembleSummary(p)
		if !strings.Contains(result, "## Original Content") {
			t.Error("missing en embed heading")
		}
	})

	t.Run("separator structure", func(t *testing.T) {
		p := base
		p.EmbedContent = "Article content."
		p.Language = "en"
		result := AssembleSummary(p)
		// Verify: two blank lines before first ---, heading, then --- before content
		expected := "\n\n\n---\n\n## Original Content\n\n---\n\nArticle content.\n"
		if !strings.Contains(result, expected) {
			t.Errorf("unexpected separator structure, got:\n%s", result)
		}
	})
}

func TestEmbedContentHeading(t *testing.T) {
	tests := []struct {
		lang string
		want string
	}{
		{"zh-Hant", "## 原文內容"},
		{"ja", "## 原文"},
		{"en", "## Original Content"},
		{"", "## Original Content"},
		{"fr", "## Original Content"},
	}
	for _, tt := range tests {
		got := embedContentHeading(tt.lang)
		if got != tt.want {
			t.Errorf("embedContentHeading(%q) = %q, want %q", tt.lang, got, tt.want)
		}
	}
}

func TestAssembleContent(t *testing.T) {
	params := ContentParams{
		Title:         "Test Article",
		URL:           "https://example.com/test",
		Domain:        "example.com",
		Source:        "safari",
		DateAdded:     time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC),
		ProcessedDate: time.Date(2026, 3, 28, 0, 0, 0, 0, time.UTC),
		ContentLength: 3200,
		ExtractedBy:   "defuddle-js",
		Content:       "# Article\n\nFull extracted content here.",
	}

	result := AssembleContent(params)

	if !strings.Contains(result, `type: reading-list-content`) {
		t.Error("missing type in frontmatter")
	}
	if !strings.Contains(result, "Full extracted content here.") {
		t.Error("missing content body")
	}
}
