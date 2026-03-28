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
