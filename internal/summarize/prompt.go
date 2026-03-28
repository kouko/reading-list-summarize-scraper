package summarize

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kouko/reading-list-summarize-scraper/internal/config"
)

// PromptVars holds all variables available for prompt template substitution.
type PromptVars struct {
	Title         string
	Domain        string
	DateAdded     string
	Source        string
	Content       string
	ContentLength int
}

// ResolvePrompt resolves the prompt template using a 3-level cascade:
//  1. summaryConfig.SummaryPromptFile (if set)
//  2. summaryConfig.Prompt (inline, if set)
//  3. Built-in prompt for summaryConfig.Language (default)
func ResolvePrompt(summaryConfig config.SummaryConfig) (string, error) {
	// Level 1: global prompt file
	if summaryConfig.SummaryPromptFile != "" {
		return readPromptFile(summaryConfig.SummaryPromptFile)
	}

	// Level 2: inline prompt
	if summaryConfig.Prompt != "" {
		return summaryConfig.Prompt, nil
	}

	// Level 3: built-in prompt by language
	return loadBuiltinPrompt(summaryConfig.Language)
}

// readPromptFile reads a prompt template from the given file path.
func readPromptFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading prompt file %q: %w", path, err)
	}
	return string(data), nil
}

// loadBuiltinPrompt loads a built-in summary prompt template for the given language.
func loadBuiltinPrompt(language string) (string, error) {
	return loadBuiltinPromptByPrefix("summary", language)
}

// SubstituteVars replaces {{variable}} placeholders in a template with values from vars.
// For inline prompts (those without {{content}}), the content is appended after the prompt.
func SubstituteVars(template string, vars PromptVars) string {
	tier := CalculateTier(vars.ContentLength, "en") // default to en for tier calculation
	lengthStr := strconv.Itoa(vars.ContentLength)

	replacer := strings.NewReplacer(
		"{{title}}", vars.Title,
		"{{domain}}", vars.Domain,
		"{{date_added}}", vars.DateAdded,
		"{{source}}", vars.Source,
		"{{content}}", vars.Content,
		"{{content_length}}", lengthStr,
		"{{content_tier}}", tier,
	)

	result := replacer.Replace(template)

	// If the template had no {{content}} placeholder, append content
	if !strings.Contains(template, "{{content}}") {
		result = result + "\n\n" + vars.Content
	}

	return result
}

// ResolveAndSubstitute is a convenience function that combines ResolvePrompt + SubstituteVars.
// It resolves the prompt template from config, then substitutes all variables.
func ResolveAndSubstitute(summaryConfig config.SummaryConfig, vars PromptVars) (string, error) {
	template, err := ResolvePrompt(summaryConfig)
	if err != nil {
		return "", err
	}
	return SubstituteVars(template, vars), nil
}

// CalculateTier returns a tier label based on character count and language.
// CJK languages (zh-Hant, ja) use lower thresholds and language-specific units.
// English and other languages use higher thresholds with "chars" unit.
func CalculateTier(charCount int, language string) string {
	switch language {
	case "zh-Hant":
		return calculateCJKTier(charCount, "字")
	case "ja":
		return calculateCJKTier(charCount, "文字")
	default:
		return calculateEnTier(charCount)
	}
}

func calculateCJKTier(charCount int, unit string) string {
	switch {
	case charCount < 500:
		return fmt.Sprintf("< 500 %s", unit)
	case charCount <= 3000:
		return fmt.Sprintf("500-3,000 %s", unit)
	case charCount <= 10000:
		return fmt.Sprintf("3,000-10,000 %s", unit)
	default:
		return fmt.Sprintf("> 10,000 %s", unit)
	}
}

func calculateEnTier(charCount int) string {
	switch {
	case charCount < 1000:
		return "< 1,000 chars"
	case charCount <= 5000:
		return "1,000-5,000 chars"
	case charCount <= 15000:
		return "5,000-15,000 chars"
	default:
		return "> 15,000 chars"
	}
}
