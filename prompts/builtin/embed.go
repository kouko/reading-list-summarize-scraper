package builtin

import "embed"

//go:embed summary-en.md summary-ja.md summary-zh-Hant.md
//go:embed summary-article-en.md summary-article-ja.md summary-article-zh-Hant.md
//go:embed keywords-en.md keywords-ja.md keywords-zh-Hant.md
//go:embed mermaid-en.md mermaid-ja.md mermaid-zh-Hant.md
var FS embed.FS
