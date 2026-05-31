package summarize

import (
	"regexp"
	"strings"
)

// sourceInjectionRe matches （來源：…） tail-notes that the Antigravity CLI (agy)
// occasionally appends. Full-width parens （）, prefix 來源, non-greedy match to
// the closing ）. (?s) lets the group span newlines.
var sourceInjectionRe = regexp.MustCompile(`(?s)（來源.*?）`)

// stripSourceInjection removes every （來源…） note from s and trims surrounding
// whitespace. All other text is left intact.
func stripSourceInjection(s string) string {
	result := sourceInjectionRe.ReplaceAllString(s, "")
	return strings.TrimSpace(result)
}
