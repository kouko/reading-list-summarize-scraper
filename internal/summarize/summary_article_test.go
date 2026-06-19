package summarize

import (
	"strings"
	"testing"
)

// TestArticlePrompts_LoadAndShape pins the structural contract of the built-in
// article-style summary prompts: they must load for each supported language,
// use rlss's content placeholders, lead with the heading-safe "### TL;DR", and
// must NOT carry ytss-only transcript placeholders (guards against copy-paste
// from the sibling project). Wording quality is tuned manually via the
// RLSS_LIVE_PROMPT harness, not asserted here.
func TestArticlePrompts_LoadAndShape(t *testing.T) {
	for _, lang := range []string{"en", "ja", "zh-Hant"} {
		got, err := loadBuiltinPrompt(lang, "article")
		if err != nil {
			t.Errorf("loadBuiltinPrompt(%q, article) error: %v", lang, err)
			continue
		}
		if strings.TrimSpace(got) == "" {
			t.Errorf("%s article prompt is empty", lang)
			continue
		}
		if !strings.Contains(got, "{{content}}") {
			t.Errorf("%s article prompt must contain the {{content}} placeholder", lang)
		}
		if !strings.Contains(got, "### TL;DR") {
			t.Errorf("%s article prompt must lead with the '### TL;DR' heading (preamble-stripper safe)", lang)
		}
		for _, bad := range []string{"{{transcript}}", "{{transcription_tier}}"} {
			if strings.Contains(got, bad) {
				t.Errorf("%s article prompt contains ytss-only placeholder %q (must use rlss content vars)", lang, bad)
			}
		}
	}
}
