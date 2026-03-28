package builtin

import "embed"

// Prompts embeds all built-in prompt template files (*.md).
// Templates are named by convention: {prefix}-{language}.md
// e.g., summary-en.md, summary-zh-Hant.md, keywords-ja.md
//
//go:embed *.md
var Prompts embed.FS
