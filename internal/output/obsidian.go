package output

import (
	"fmt"
	"strings"
	"time"
)

type SummaryParams struct {
	Title         string
	URL           string
	Domain        string
	Source        string
	DateAdded     time.Time
	ProcessedDate time.Time
	LLMProvider   string
	LLMModel      string
	ContentLength int
	ContentTier   string
	SummaryText   string
	Keywords      []string
	MermaidBlocks []MermaidBlock
}

type MermaidBlock struct {
	Title string
	Code  string
}

type ContentParams struct {
	Title         string
	URL           string
	Domain        string
	Source        string
	DateAdded     time.Time
	ProcessedDate time.Time
	ContentLength int
	ExtractedBy   string
	Content       string
}

func AssembleSummary(p SummaryParams) string {
	var b strings.Builder

	// Frontmatter
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("title: %q\n", p.Title))
	b.WriteString("type: reading-list-summary\n")
	b.WriteString(fmt.Sprintf("date: %s\n", p.ProcessedDate.Format("2006-01-02")))
	b.WriteString(fmt.Sprintf("url: %q\n", p.URL))
	b.WriteString(fmt.Sprintf("domain: %q\n", p.Domain))
	b.WriteString(fmt.Sprintf("source: %q\n", p.Source))
	b.WriteString(fmt.Sprintf("date_added: %s\n", p.DateAdded.Format("2006-01-02")))
	b.WriteString(fmt.Sprintf("llm_provider: %q\n", p.LLMProvider))
	b.WriteString(fmt.Sprintf("llm_model: %q\n", p.LLMModel))
	b.WriteString(fmt.Sprintf("content_length: %d\n", p.ContentLength))
	b.WriteString(fmt.Sprintf("content_tier: %q\n", p.ContentTier))
	b.WriteString("tags:\n")
	b.WriteString("  - reading-list\n")
	b.WriteString("  - auto-summary\n")
	for _, kw := range p.Keywords {
		b.WriteString(fmt.Sprintf("  - %s\n", kw))
	}
	b.WriteString("---\n\n")

	// Info callout
	b.WriteString("> [!info] 來源資訊\n")
	b.WriteString(fmt.Sprintf("> - **原始網址**：[%s](%s)\n", p.Domain, p.URL))
	b.WriteString(fmt.Sprintf("> - **加入日期**：%s\n", p.DateAdded.Format("2006-01-02")))
	b.WriteString(fmt.Sprintf("> - **來源**：%s Reading List\n", sourceDisplayName(p.Source)))
	b.WriteString(fmt.Sprintf("> - **摘要工具**：%s (%s)\n", p.LLMProvider, p.LLMModel))
	b.WriteString(fmt.Sprintf("> - **處理日期**：%s\n", p.ProcessedDate.Format("2006-01-02")))
	b.WriteString("\n---\n\n")

	// Summary body with Mermaid insertion
	body := p.SummaryText
	if len(p.MermaidBlocks) > 0 {
		body = insertMermaidBlocks(body, p.MermaidBlocks)
	}
	b.WriteString(body)
	b.WriteString("\n")

	return b.String()
}

func AssembleContent(p ContentParams) string {
	var b strings.Builder

	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("title: %q\n", p.Title))
	b.WriteString("type: reading-list-content\n")
	b.WriteString(fmt.Sprintf("date: %s\n", p.ProcessedDate.Format("2006-01-02")))
	b.WriteString(fmt.Sprintf("url: %q\n", p.URL))
	b.WriteString(fmt.Sprintf("domain: %q\n", p.Domain))
	b.WriteString(fmt.Sprintf("source: %q\n", p.Source))
	b.WriteString(fmt.Sprintf("date_added: %s\n", p.DateAdded.Format("2006-01-02")))
	b.WriteString(fmt.Sprintf("content_length: %d\n", p.ContentLength))
	b.WriteString(fmt.Sprintf("extracted_by: %q\n", p.ExtractedBy))
	b.WriteString("---\n\n")
	b.WriteString(p.Content)
	b.WriteString("\n")

	return b.String()
}

func sourceDisplayName(source string) string {
	switch source {
	case "safari":
		return "Safari"
	case "chrome":
		return "Chrome"
	default:
		return "Manual"
	}
}

// insertMermaidBlocks inserts Mermaid diagrams after matching section headings.
// Matches block.Title against ### headings in body.
// Unmatched blocks inserted after the first section (before second ### heading).
func insertMermaidBlocks(body string, blocks []MermaidBlock) string {
	lines := strings.Split(body, "\n")

	for _, block := range blocks {
		normalizedTitle := strings.TrimSpace(block.Title)
		normalizedTitle = strings.TrimPrefix(normalizedTitle, "#### ")
		normalizedTitle = strings.TrimPrefix(normalizedTitle, "### ")

		matched := false
		for i := range lines {
			lineTitle := strings.TrimSpace(lines[i])
			lineTitle = strings.TrimPrefix(lineTitle, "#### ")
			lineTitle = strings.TrimPrefix(lineTitle, "### ")

			if lineTitle == normalizedTitle {
				// Find next heading
				insertIdx := len(lines)
				for j := i + 1; j < len(lines); j++ {
					trimmed := strings.TrimSpace(lines[j])
					if strings.HasPrefix(trimmed, "###") {
						insertIdx = j
						break
					}
				}
				mermaid := fmt.Sprintf("\n```mermaid\n%s\n```\n", block.Code)
				newLines := make([]string, 0, len(lines)+4)
				newLines = append(newLines, lines[:insertIdx]...)
				newLines = append(newLines, mermaid)
				newLines = append(newLines, lines[insertIdx:]...)
				lines = newLines
				matched = true
				break
			}
		}

		if !matched {
			// Insert after first section
			headingCount := 0
			insertIdx := len(lines)
			for i, line := range lines {
				if strings.HasPrefix(strings.TrimSpace(line), "###") {
					headingCount++
					if headingCount == 2 {
						insertIdx = i
						break
					}
				}
			}
			mermaid := fmt.Sprintf("\n```mermaid\n%s\n```\n", block.Code)
			newLines := make([]string, 0, len(lines)+4)
			newLines = append(newLines, lines[:insertIdx]...)
			newLines = append(newLines, mermaid)
			newLines = append(newLines, lines[insertIdx:]...)
			lines = newLines
		}
	}

	return strings.Join(lines, "\n")
}
