package summarize

import (
	"strings"
	"testing"
)

// TestOutlinePrompts_EnforceOutputLanguage pins an emphatic output-language
// directive in each outline (default) summary prompt. Dogfooding surfaced that
// the weak one-line "use Traditional Chinese" guideline let a model drift to
// Simplified Chinese; each prompt must carry a strong "always output in <lang>,
// even if the source is another language" instruction (and zh-Hant must call
// out Simplified→Traditional conversion explicitly).
func TestOutlinePrompts_EnforceOutputLanguage(t *testing.T) {
	cases := map[string][]string{
		"zh-Hant": {"繁體", "簡體"}, // Traditional output + Simplified→Traditional conversion
		"ja":      {"必ず日本語"},    // emphatic Japanese-output directive
		"en":      {"Always write the output in English"},
	}
	for lang, needles := range cases {
		got, err := loadBuiltinPrompt(lang)
		if err != nil {
			t.Errorf("loadBuiltinPrompt(%q): %v", lang, err)
			continue
		}
		for _, n := range needles {
			if !strings.Contains(got, n) {
				t.Errorf("outline %s prompt must contain an output-language directive %q", lang, n)
			}
		}
	}
}
