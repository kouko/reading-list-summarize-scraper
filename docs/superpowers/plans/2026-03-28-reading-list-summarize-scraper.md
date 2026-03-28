# Reading List Summarize Scraper Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go CLI tool (`rlss`) that fetches URLs from macOS Safari/Chrome Reading Lists, extracts web content via chromedp + Defuddle JS injection, summarizes with Agentic CLI fallback chain, and outputs Obsidian Markdown files.

**Architecture:** Modular pipeline — source layer (Safari plist / Chrome Extension CDP / manual URL) feeds unified ReadingItem structs into an extraction layer (chromedp lazy pool + Defuddle JS injection), then a summarization layer (LLM fallback chain ported from youtube-summarize-scraper), and finally an output layer (Obsidian Markdown with frontmatter + copy_to). Config via YAML, CLI via Cobra.

**Tech Stack:** Go 1.23+, chromedp, howett.net/plist, Cobra, esbuild, Defuddle (TypeScript), Agentic CLIs (claude/gemini/qwen)

**Reference project:** `/Users/kouko/VisualStudioCodeProject/youtube-summarize-scraper` (ytss) — the summarize layer, config patterns, pipeline patterns, and prompt templates are ported from this project.

**Design spec:** `docs/superpowers/specs/2026-03-28-reading-list-summarize-scraper-design.md`

---

## File Map

### Root Files
| File | Responsibility |
|------|---------------|
| `go.mod` | Go module definition |
| `Makefile` | Build automation (js, build, clean, update-defuddle) |
| `package.json` | Node.js deps for esbuild + defuddle |
| `inject.ts` | Defuddle browser glue code |
| `embed.go` | `//go:generate` + `//go:embed` for defuddle.min.js |
| `config.yaml` | Example config |

### embed/
| File | Responsibility |
|------|---------------|
| `embed/defuddle.min.js` | esbuild-bundled Defuddle IIFE (committed + regenerated in CI) |
| `embed/extension/manifest.json` | Chrome Extension manifest for Reading List API |
| `embed/extension/background.js` | Minimal service worker |

### cmd/rlss/
| File | Responsibility |
|------|---------------|
| `cmd/rlss/main.go` | CLI entry point, version, logging setup |
| `cmd/rlss/root.go` | Root command, persistent flags, config loading |
| `cmd/rlss/process.go` | `process` subcommand (default) |
| `cmd/rlss/list.go` | `list` subcommand (dry-run) |
| `cmd/rlss/url.go` | `url` subcommand (single URL) |
| `cmd/rlss/config_cmd.go` | `config` subcommand (show config) |

### internal/config/
| File | Responsibility |
|------|---------------|
| `internal/config/config.go` | All config structs + Load() + YAML parsing |
| `internal/config/defaults.go` | DefaultConfig() with sensible defaults |

### internal/source/
| File | Responsibility |
|------|---------------|
| `internal/source/types.go` | ReadingItem struct + Source interface |
| `internal/source/safari.go` | Safari Bookmarks.plist parsing |
| `internal/source/safari_test.go` | Tests with embedded test plist |
| `internal/source/chrome.go` | Chrome CDP + Extension + Service Worker |
| `internal/source/manual.go` | Manual --url source |

### internal/extract/
| File | Responsibility |
|------|---------------|
| `internal/extract/browser.go` | Single chromedp instance lifecycle |
| `internal/extract/defuddle.go` | JS injection + result extraction |
| `internal/extract/pool.go` | Lazy pool keyed by (headed, profile) |
| `internal/extract/domain.go` | domain_rules matching logic |
| `internal/extract/domain_test.go` | Domain matching tests |
| `internal/extract/profile.go` | Chrome Local State → UI name → folder name |
| `internal/extract/profile_test.go` | Profile resolution tests |

### internal/summarize/ (ported from ytss)
| File | Responsibility |
|------|---------------|
| `internal/summarize/summarizer.go` | Interface + factory + NewSummarizer |
| `internal/summarize/fallback.go` | FallbackSummarizer |
| `internal/summarize/circuit_breaker.go` | CircuitBreaker state machine |
| `internal/summarize/circuit_breaker_test.go` | CB tests |
| `internal/summarize/errors.go` | QuotaError + quotaPatterns |
| `internal/summarize/errors_test.go` | Error detection tests |
| `internal/summarize/prompt.go` | Prompt resolution + variable substitution |
| `internal/summarize/prompt_test.go` | Prompt tests |
| `internal/summarize/keywords.go` | Keyword extraction prompt + parsing |
| `internal/summarize/mermaid.go` | Mermaid prompt + validation |
| `internal/summarize/claude_code.go` | ClaudeCodeSummarizer |
| `internal/summarize/claude.go` | ClaudeSummarizer (API) |
| `internal/summarize/gemini.go` | GeminiCLISummarizer |
| `internal/summarize/qwen_code.go` | QwenCodeSummarizer |
| `internal/summarize/ollama.go` | OllamaSummarizer |
| `internal/summarize/llamacpp.go` | LlamaCppSummarizer |
| `internal/summarize/openai_compat.go` | OpenAICompatSummarizer |
| `internal/summarize/fallback_test.go` | Fallback chain tests |
| `internal/summarize/summarizer_test.go` | Factory tests |

### internal/output/
| File | Responsibility |
|------|---------------|
| `internal/output/filename.go` | SHA8 computation + filename generation |
| `internal/output/filename_test.go` | Filename tests |
| `internal/output/index.go` | FileIndex scan + lookup |
| `internal/output/index_test.go` | Index tests |
| `internal/output/obsidian.go` | Frontmatter + body assembly + Mermaid insertion |
| `internal/output/obsidian_test.go` | Assembly tests |
| `internal/output/copyto.go` | copy_to template expansion + file copy |
| `internal/output/copyto_test.go` | copy_to tests |

### internal/pipeline/
| File | Responsibility |
|------|---------------|
| `internal/pipeline/runner.go` | Pipeline struct + ProcessBatch + ProcessItem |
| `internal/pipeline/watch.go` | Watch mode loop |
| `internal/pipeline/stats.go` | Stats struct + reporting |

### prompts/builtin/
| File | Responsibility |
|------|---------------|
| `prompts/builtin/embed.go` | `//go:embed` for all prompt files |
| `prompts/builtin/summary-{en,ja,zh-Hant}.md` | Summary prompts (3 langs) |
| `prompts/builtin/keywords-{en,ja,zh-Hant}.md` | Keyword prompts (3 langs) |
| `prompts/builtin/mermaid-{en,ja,zh-Hant}.md` | Mermaid prompts (3 langs) |

---

## Task 1: Project Scaffold + Build Pipeline

**Files:**
- Create: `go.mod`, `Makefile`, `package.json`, `inject.ts`, `embed.go`, `.gitignore`
- Create: `embed/extension/manifest.json`, `embed/extension/background.js`

- [ ] **Step 1: Initialize Go module**

```bash
cd /Users/kouko/VisualStudioCodeProject/reading-list-summarize-scraper
go mod init github.com/kouko/reading-list-summarize-scraper
```

- [ ] **Step 2: Create .gitignore**

Create `.gitignore`:

```gitignore
# Binary
rlss

# Node
node_modules/

# OS
.DS_Store

# IDE
.idea/
.vscode/
```

Note: `embed/defuddle.min.js` and `package-lock.json` are intentionally NOT gitignored — they are committed.

- [ ] **Step 3: Create package.json**

Create `package.json`:

```json
{
  "private": true,
  "scripts": {
    "build:defuddle": "esbuild inject.ts --bundle --format=iife --global-name=DefuddleExtractor --platform=browser --target=es2020 --minify --outfile=embed/defuddle.min.js",
    "update:defuddle": "npm update defuddle && npm run build:defuddle"
  },
  "devDependencies": {
    "defuddle": "latest",
    "esbuild": "^0.25.0"
  }
}
```

- [ ] **Step 4: Create inject.ts**

Create `inject.ts`:

```typescript
import { Defuddle } from 'defuddle';

(window as any).extractArticle = async (): Promise<string> => {
    try {
        const html = document.documentElement.outerHTML;
        const df = new Defuddle(html);
        const result = await df.parse();
        return result?.content ?? "";
    } catch (e: any) {
        return "";
    }
};
```

- [ ] **Step 5: Run npm install and build defuddle.min.js**

```bash
npm install
npm run build:defuddle
```

Expected: `embed/defuddle.min.js` created (~150-300KB minified IIFE).

- [ ] **Step 6: Create embed.go**

Create `embed.go` at project root:

```go
package rlss

import _ "embed"

//go:generate npm run build:defuddle

//go:embed embed/defuddle.min.js
var DefuddleJS string
```

- [ ] **Step 7: Create Chrome Extension files**

Create `embed/extension/manifest.json`:

```json
{
  "manifest_version": 3,
  "name": "Reading List Exporter",
  "version": "1.0",
  "permissions": ["readingList"],
  "background": {
    "service_worker": "background.js"
  }
}
```

Create `embed/extension/background.js`:

```javascript
console.log("Reading List Exporter loaded");
```

- [ ] **Step 8: Create Makefile**

Create `Makefile`:

```makefile
.PHONY: build js clean update-defuddle generate

build: js
	go build -ldflags="-s -w" -o rlss ./cmd/rlss/

js: node_modules
	npm run build:defuddle

node_modules: package.json package-lock.json
	npm install
	@touch node_modules

update-defuddle:
	npm update defuddle
	npm run build:defuddle
	@echo "Updated defuddle to $$(npm list defuddle --depth=0 | grep defuddle)"

generate: node_modules
	go generate ./...

clean:
	rm -f rlss embed/defuddle.min.js
	rm -rf node_modules
```

- [ ] **Step 9: Commit**

```bash
git add .gitignore go.mod package.json package-lock.json inject.ts embed.go Makefile
git add embed/defuddle.min.js embed/extension/
git commit -m "feat: project scaffold with esbuild + defuddle build pipeline"
```

---

## Task 2: Config System

**Files:**
- Create: `internal/config/config.go`, `internal/config/defaults.go`, `config.yaml`

- [ ] **Step 1: Create config structs**

Create `internal/config/config.go`. This file defines all config structs. The LLM-related structs (ProviderList, FallbackStrategyConfig, LLMConfig, and all provider configs) must be copied from ytss `config/config.go` with the same field names, YAML tags, and ProviderList methods (UnmarshalYAML, Primary, Fallbacks, SetPrimary, Contains, MarshalYAML, Equal, String).

The rlss-specific structs are:

```go
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	OutputDir string         `yaml:"output_dir"`
	LLM       LLMConfig      `yaml:"llm"`
	Summary   SummaryConfig  `yaml:"summary"`
	Safari    SafariConfig   `yaml:"safari"`
	Chrome    ChromeConfig   `yaml:"chrome"`
	Extract   ExtractConfig  `yaml:"extract"`
	Pipeline  PipelineConfig `yaml:"pipeline"`
	Filter    FilterConfig   `yaml:"filter"`
	Watch     WatchConfig    `yaml:"watch"`
	CopyTo    CopyToConfig   `yaml:"copy_to"`
	Obsidian  ObsidianConfig `yaml:"obsidian"`
}

type SummaryConfig struct {
	Language          string         `yaml:"language"`
	Prompt            string         `yaml:"prompt"`
	SummaryPromptFile string         `yaml:"summary_prompt_file"`
	MaxTokens         int            `yaml:"max_tokens"`
	Keywords          KeywordsConfig `yaml:"keywords"`
	Mermaid           MermaidConfig  `yaml:"mermaid"`
}

type KeywordsConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Language string `yaml:"language"`
	Count    int    `yaml:"count"`
}

type MermaidConfig struct {
	Enabled bool `yaml:"enabled"`
}

type SafariConfig struct {
	Enabled   bool   `yaml:"enabled"`
	PlistPath string `yaml:"plist_path"`
}

type ChromeConfig struct {
	Enabled     bool   `yaml:"enabled"`
	Profile     string `yaml:"profile"`
	UserDataDir string `yaml:"user_data_dir"`
}

type DomainRule struct {
	Domains       []string `yaml:"domains"`
	Headed        bool     `yaml:"headed"`
	ChromeProfile string   `yaml:"chrome_profile"`
}

type ExtractConfig struct {
	Headless      bool          `yaml:"headless"`
	ChromeProfile string        `yaml:"chrome_profile"`
	UserDataDir   string        `yaml:"user_data_dir"`
	Timeout       time.Duration `yaml:"timeout"`
	WaitAfterLoad time.Duration `yaml:"wait_after_load"`
	DomainRules   []DomainRule  `yaml:"domain_rules"`
}

type PipelineConfig struct {
	SkipExisting bool `yaml:"skip_existing"`
	DryRun       bool `yaml:"dry_run"`
	DelayMin     int  `yaml:"delay_min"`
	DelayMax     int  `yaml:"delay_max"`
}

type FilterConfig struct {
	UnreadOnly bool   `yaml:"unread_only"`
	Since      string `yaml:"since"`
	Limit      int    `yaml:"limit"`
}

type WatchConfig struct {
	Enabled  bool `yaml:"enabled"`
	Interval int  `yaml:"interval"`
}

type CopyToConfig struct {
	Enabled   bool     `yaml:"enabled"`
	Path      string   `yaml:"path"`
	Filename  string   `yaml:"filename"`
	Files     []string `yaml:"files"`
	Overwrite bool     `yaml:"overwrite"`
}

type ObsidianConfig struct {
	AutoTags  bool `yaml:"auto_tags"`
	Wikilinks bool `yaml:"wikilinks"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			return &cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	cfg.OutputDir = expandPath(cfg.OutputDir)
	if cfg.Safari.PlistPath != "" {
		cfg.Safari.PlistPath = expandPath(cfg.Safari.PlistPath)
	}

	return &cfg, nil
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
```

Copy the following from ytss `config/config.go` into the same file, adapting the package name to `config`:
- `ProviderList` type and all its methods
- `FallbackStrategyConfig` struct
- `LLMConfig` struct
- All provider config structs: `OllamaConfig`, `LlamaCppConfig`, `ClaudeAPIConfig`, `ClaudeCodeConfig`, `GeminiCLIConfig`, `QwenCodeConfig`, `OpenAICompatConfig`

Reference: `/Users/kouko/VisualStudioCodeProject/youtube-summarize-scraper/config/config.go`

- [ ] **Step 2: Create defaults**

Create `internal/config/defaults.go`:

```go
package config

import "time"

func DefaultConfig() Config {
	return Config{
		OutputDir: "~/reading-list-summaries",
		LLM: LLMConfig{
			Provider: ProviderList{"claude-code"},
			ProviderFallbackStrategy: FallbackStrategyConfig{
				CooldownSeconds:  300,
				FailureThreshold: 1,
			},
			Ollama: OllamaConfig{
				Model:    "llama3",
				Endpoint: "http://localhost:11434",
				Think:    ptrBool(false),
				Timeout:  900,
			},
			LlamaCpp: LlamaCppConfig{
				Endpoint: "http://localhost:8080",
			},
			ClaudeCode: ClaudeCodeConfig{
				Model:   "haiku",
				Timeout: 900,
			},
			GeminiCLI: GeminiCLIConfig{
				Model:   "auto",
				Timeout: 900,
			},
			QwenCode: QwenCodeConfig{
				Model:   "coder-model",
				Timeout: 900,
			},
			OpenAICompat: OpenAICompatConfig{
				Timeout: 900,
			},
		},
		Summary: SummaryConfig{
			Language:  "zh-Hant",
			MaxTokens: 10000,
			Keywords: KeywordsConfig{
				Enabled:  true,
				Language: "en",
				Count:    5,
			},
			Mermaid: MermaidConfig{
				Enabled: true,
			},
		},
		Safari: SafariConfig{
			Enabled: true,
		},
		Chrome: ChromeConfig{
			Enabled: true,
			Profile: "Default",
		},
		Extract: ExtractConfig{
			Headless:      true,
			ChromeProfile: "Default",
			Timeout:       30 * time.Second,
			WaitAfterLoad: 2 * time.Second,
		},
		Pipeline: PipelineConfig{
			SkipExisting: true,
			DelayMin:     3,
			DelayMax:     8,
		},
		Filter: FilterConfig{},
		Watch: WatchConfig{
			Interval: 10,
		},
		CopyTo: CopyToConfig{
			Files: []string{"summary", "content"},
		},
		Obsidian: ObsidianConfig{
			AutoTags:  true,
			Wikilinks: true,
		},
	}
}

func ptrBool(b bool) *bool {
	return &b
}
```

- [ ] **Step 3: Create example config.yaml**

Create `config.yaml` at project root — copy the full config YAML from the design spec Section 5.

- [ ] **Step 4: Verify config loads**

Create a temporary test in `internal/config/` to verify Load works:

```bash
go get gopkg.in/yaml.v3
go vet ./internal/config/
```

- [ ] **Step 5: Commit**

```bash
git add internal/config/ config.yaml go.mod go.sum
git commit -m "feat: config system with all structs, defaults, and YAML loading"
```

---

## Task 3: Source Types + Manual Source

**Files:**
- Create: `internal/source/types.go`, `internal/source/manual.go`

- [ ] **Step 1: Create ReadingItem and Source interface**

Create `internal/source/types.go`:

```go
package source

import "time"

type ReadingItem struct {
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	DateAdded   time.Time `json:"date_added"`
	IsUnread    bool      `json:"is_unread"`
	PreviewText string    `json:"preview_text,omitempty"`
	Source      string    `json:"source"` // "safari" | "chrome" | "manual"
}

type Source interface {
	Name() string
	Fetch() ([]ReadingItem, error)
}

// DeduplicateByURL removes duplicate items keeping the first occurrence.
func DeduplicateByURL(items []ReadingItem) []ReadingItem {
	seen := make(map[string]bool)
	var result []ReadingItem
	for _, item := range items {
		if !seen[item.URL] {
			seen[item.URL] = true
			result = append(result, item)
		}
	}
	return result
}
```

- [ ] **Step 2: Create ManualSource**

Create `internal/source/manual.go`:

```go
package source

import "time"

type ManualSource struct {
	URL string
}

func NewManualSource(url string) *ManualSource {
	return &ManualSource{URL: url}
}

func (m *ManualSource) Name() string { return "manual" }

func (m *ManualSource) Fetch() ([]ReadingItem, error) {
	return []ReadingItem{
		{
			URL:       m.URL,
			Source:    "manual",
			DateAdded: time.Now(),
			IsUnread:  true,
		},
	}, nil
}
```

- [ ] **Step 3: Verify compilation**

```bash
go vet ./internal/source/
```

- [ ] **Step 4: Commit**

```bash
git add internal/source/
git commit -m "feat: ReadingItem types, Source interface, ManualSource"
```

---

## Task 4: Output — Filename + SHA8 + Index

**Files:**
- Create: `internal/output/filename.go`, `internal/output/filename_test.go`, `internal/output/index.go`, `internal/output/index_test.go`

- [ ] **Step 1: Write filename tests**

Create `internal/output/filename_test.go`:

```go
package output

import (
	"testing"
	"time"
)

func TestSHA8(t *testing.T) {
	got := SHA8("https://example.com/article")
	if len(got) != 8 {
		t.Errorf("SHA8 length = %d, want 8", len(got))
	}
	// Same URL = same hash
	if SHA8("https://example.com/article") != got {
		t.Error("SHA8 not deterministic")
	}
	// Different URL = different hash
	if SHA8("https://other.com") == got {
		t.Error("SHA8 collision")
	}
}

func TestDomainDir(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://example.com/path", "example_com"},
		{"https://blog.example.com/path", "blog_example_com"},
		{"https://www.nikkei.com/article", "www_nikkei_com"},
	}
	for _, tt := range tests {
		got := DomainDir(tt.url)
		if got != tt.want {
			t.Errorf("DomainDir(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}

func TestFilename(t *testing.T) {
	date := time.Date(2026, 3, 28, 0, 0, 0, 0, time.UTC)
	sha := "a1b2c3d4"

	got := SummaryFilename(date, sha)
	if got != "2026-03-28__a1b2c3d4__summary.md" {
		t.Errorf("SummaryFilename = %q", got)
	}

	got = ContentFilename(date, sha)
	if got != "2026-03-28__a1b2c3d4__content.md" {
		t.Errorf("ContentFilename = %q", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/output/ -v
```

Expected: FAIL — functions not defined.

- [ ] **Step 3: Implement filename.go**

Create `internal/output/filename.go`:

```go
package output

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// SHA8 returns the first 8 hex characters of the SHA256 hash of s.
func SHA8(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h[:4])
}

// DomainDir extracts the hostname from a URL and replaces dots with underscores.
func DomainDir(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return "unknown"
	}
	host := u.Hostname() // strips port
	return strings.ReplaceAll(host, ".", "_")
}

// SummaryFilename returns "YYYY-MM-DD__<sha8>__summary.md".
func SummaryFilename(date time.Time, sha8 string) string {
	return fmt.Sprintf("%s__%s__summary.md", date.Format("2006-01-02"), sha8)
}

// ContentFilename returns "YYYY-MM-DD__<sha8>__content.md".
func ContentFilename(date time.Time, sha8 string) string {
	return fmt.Sprintf("%s__%s__content.md", date.Format("2006-01-02"), sha8)
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/output/ -v
```

Expected: PASS.

- [ ] **Step 5: Write index tests**

Create `internal/output/index_test.go`:

```go
package output

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileIndex(t *testing.T) {
	// Create temp dir with test files
	dir := t.TempDir()
	domainDir := filepath.Join(dir, "example_com")
	os.MkdirAll(domainDir, 0755)
	os.WriteFile(filepath.Join(domainDir, "2026-03-28__a1b2c3d4__summary.md"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(domainDir, "2026-03-28__a1b2c3d4__content.md"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(domainDir, "2026-03-28__b2c3d4e5__content.md"), []byte("test"), 0644)

	idx := NewFileIndex()
	idx.Build(dir)

	// a1b2c3d4: both files exist
	if !idx.Has("a1b2c3d4") {
		t.Error("index should have a1b2c3d4")
	}
	info := idx.Get("a1b2c3d4")
	if !info.SummaryExists || !info.ContentExists {
		t.Error("a1b2c3d4 should have both files")
	}

	// b2c3d4e5: only content
	info = idx.Get("b2c3d4e5")
	if info.SummaryExists {
		t.Error("b2c3d4e5 should not have summary")
	}
	if !info.ContentExists {
		t.Error("b2c3d4e5 should have content")
	}

	// unknown: not found
	if idx.Has("ffffffff") {
		t.Error("should not have ffffffff")
	}
}
```

- [ ] **Step 6: Implement index.go**

Create `internal/output/index.go`:

```go
package output

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var filePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}__([a-f0-9]{8})__(summary|content)\.md$`)

type FileInfo struct {
	Dir            string
	SummaryExists  bool
	ContentExists  bool
}

type FileIndex struct {
	entries map[string]*FileInfo
}

func NewFileIndex() *FileIndex {
	return &FileIndex{entries: make(map[string]*FileInfo)}
}

func (idx *FileIndex) Build(outputDir string) {
	idx.entries = make(map[string]*FileInfo)

	dirs, _ := os.ReadDir(outputDir)
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		domainDir := filepath.Join(outputDir, d.Name())
		files, _ := os.ReadDir(domainDir)
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			m := filePattern.FindStringSubmatch(f.Name())
			if m == nil {
				continue
			}
			sha8 := m[1]
			fileType := m[2]

			info, ok := idx.entries[sha8]
			if !ok {
				info = &FileInfo{Dir: domainDir}
				idx.entries[sha8] = info
			}
			switch fileType {
			case "summary":
				info.SummaryExists = true
			case "content":
				info.ContentExists = true
			}
		}
	}
}

func (idx *FileIndex) Has(sha8 string) bool {
	_, ok := idx.entries[sha8]
	return ok
}

func (idx *FileIndex) Get(sha8 string) FileInfo {
	if info, ok := idx.entries[sha8]; ok {
		return *info
	}
	return FileInfo{}
}

// ContentPath returns the full path to the content file for a sha8, or empty if not found.
func (idx *FileIndex) ContentPath(sha8 string) string {
	info, ok := idx.entries[sha8]
	if !ok || !info.ContentExists {
		return ""
	}
	files, _ := os.ReadDir(info.Dir)
	for _, f := range files {
		if strings.Contains(f.Name(), sha8) && strings.HasSuffix(f.Name(), "__content.md") {
			return filepath.Join(info.Dir, f.Name())
		}
	}
	return ""
}
```

- [ ] **Step 7: Run all tests**

```bash
go test ./internal/output/ -v
```

Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add internal/output/
git commit -m "feat: filename generation (SHA8, domain_dir) and file index for skip detection"
```

---

## Task 5: Safari Source

**Files:**
- Create: `internal/source/safari.go`, `internal/source/safari_test.go`

- [ ] **Step 1: Write Safari test with embedded test data**

Create `internal/source/safari_test.go`:

```go
package source

import (
	"os"
	"path/filepath"
	"testing"
)

// Minimal XML plist for testing
const testPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Title</key>
	<string></string>
	<key>WebBookmarkType</key>
	<string>WebBookmarkTypeList</string>
	<key>Children</key>
	<array>
		<dict>
			<key>Title</key>
			<string>com.apple.ReadingList</string>
			<key>WebBookmarkType</key>
			<string>WebBookmarkTypeList</string>
			<key>Children</key>
			<array>
				<dict>
					<key>WebBookmarkType</key>
					<string>WebBookmarkTypeLeaf</string>
					<key>URLString</key>
					<string>https://example.com/article1</string>
					<key>URIDictionary</key>
					<dict>
						<key>title</key>
						<string>Test Article</string>
					</dict>
					<key>ReadingList</key>
					<dict>
						<key>DateAdded</key>
						<date>2026-03-25T10:30:00Z</date>
					</dict>
				</dict>
				<dict>
					<key>WebBookmarkType</key>
					<string>WebBookmarkTypeLeaf</string>
					<key>URLString</key>
					<string>https://example.com/article2</string>
					<key>URIDictionary</key>
					<dict>
						<key>title</key>
						<string>Read Article</string>
					</dict>
					<key>ReadingList</key>
					<dict>
						<key>DateAdded</key>
						<date>2026-03-20T08:00:00Z</date>
						<key>DateLastViewed</key>
						<date>2026-03-22T14:00:00Z</date>
					</dict>
				</dict>
			</array>
		</dict>
	</array>
</dict>
</plist>`

func TestSafariSource(t *testing.T) {
	// Write test plist to temp file
	dir := t.TempDir()
	plistPath := filepath.Join(dir, "Bookmarks.plist")
	os.WriteFile(plistPath, []byte(testPlist), 0644)

	src := NewSafariSource(plistPath)
	if src.Name() != "safari" {
		t.Errorf("Name() = %q, want safari", src.Name())
	}

	items, err := src.Fetch()
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}

	// First item: unread (no DateLastViewed)
	if items[0].Title != "Test Article" {
		t.Errorf("item[0].Title = %q", items[0].Title)
	}
	if items[0].URL != "https://example.com/article1" {
		t.Errorf("item[0].URL = %q", items[0].URL)
	}
	if !items[0].IsUnread {
		t.Error("item[0] should be unread")
	}
	if items[0].Source != "safari" {
		t.Errorf("item[0].Source = %q", items[0].Source)
	}

	// Second item: read (has DateLastViewed)
	if items[1].IsUnread {
		t.Error("item[1] should be read")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go get howett.net/plist
go test ./internal/source/ -v -run TestSafari
```

Expected: FAIL — NewSafariSource not defined.

- [ ] **Step 3: Implement safari.go**

Create `internal/source/safari.go`:

```go
package source

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"howett.net/plist"
)

const (
	typeLeaf  = "WebBookmarkTypeLeaf"
	typeList  = "WebBookmarkTypeList"
	readingListFolder = "com.apple.ReadingList"
)

type bookmarksPlist struct {
	Title           string           `plist:"Title"`
	WebBookmarkType string           `plist:"WebBookmarkType"`
	Children        []bookmarkEntry  `plist:"Children"`
}

type bookmarkEntry struct {
	Title              string            `plist:"Title"`
	WebBookmarkType    string            `plist:"WebBookmarkType"`
	URLString          string            `plist:"URLString"`
	URIDictionary      map[string]string `plist:"URIDictionary"`
	Children           []bookmarkEntry   `plist:"Children"`
	ReadingList        *readingListMeta  `plist:"ReadingList"`
}

type readingListMeta struct {
	DateAdded      time.Time `plist:"DateAdded"`
	DateLastViewed time.Time `plist:"DateLastViewed"`
	PreviewText    string    `plist:"PreviewText"`
}

type SafariSource struct {
	plistPath string
}

func NewSafariSource(plistPath string) *SafariSource {
	if plistPath == "" {
		home, _ := os.UserHomeDir()
		plistPath = filepath.Join(home, "Library", "Safari", "Bookmarks.plist")
	}
	return &SafariSource{plistPath: plistPath}
}

func (s *SafariSource) Name() string { return "safari" }

func (s *SafariSource) Fetch() ([]ReadingItem, error) {
	file, err := os.Open(s.plistPath)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w (may need Full Disk Access)", s.plistPath, err)
	}
	defer file.Close()

	var root bookmarksPlist
	if err := plist.NewDecoder(file).Decode(&root); err != nil {
		return nil, fmt.Errorf("decode plist: %w", err)
	}

	for _, child := range root.Children {
		if child.Title == readingListFolder {
			return extractItems(child.Children), nil
		}
	}

	return nil, fmt.Errorf("Reading List folder (%s) not found", readingListFolder)
}

func extractItems(entries []bookmarkEntry) []ReadingItem {
	var items []ReadingItem
	for _, e := range entries {
		if e.WebBookmarkType != typeLeaf || e.URLString == "" {
			continue
		}

		item := ReadingItem{
			Title:  itemTitle(e),
			URL:    e.URLString,
			Source: "safari",
		}

		if e.ReadingList != nil {
			item.DateAdded = e.ReadingList.DateAdded
			item.IsUnread = e.ReadingList.DateLastViewed.IsZero()
			item.PreviewText = e.ReadingList.PreviewText
		}

		items = append(items, item)
	}
	return items
}

func itemTitle(e bookmarkEntry) string {
	if t, ok := e.URIDictionary["title"]; ok && t != "" {
		return t
	}
	return e.Title
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/source/ -v -run TestSafari
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/source/ go.mod go.sum
git commit -m "feat: Safari plist Reading List parser with tests"
```

---

## Task 6: Chrome Profile Resolver

**Files:**
- Create: `internal/extract/profile.go`, `internal/extract/profile_test.go`

- [ ] **Step 1: Write profile resolver tests**

Create `internal/extract/profile_test.go`:

```go
package extract

import (
	"os"
	"path/filepath"
	"testing"
)

const testLocalState = `{
	"profile": {
		"info_cache": {
			"Default": {
				"name": "Personal"
			},
			"Profile 1": {
				"name": "Work Account"
			},
			"Profile 5": {
				"name": "ReadingList-Auto"
			}
		}
	}
}`

func TestProfileResolver(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "Local State"), []byte(testLocalState), 0644)

	r, err := NewProfileResolver(dir)
	if err != nil {
		t.Fatalf("NewProfileResolver error: %v", err)
	}

	tests := []struct {
		input string
		want  string
	}{
		{"ReadingList-Auto", "Profile 5"},
		{"Personal", "Default"},
		{"Work Account", "Profile 1"},
		// Direct folder name also works
		{"Default", "Default"},
		{"Profile 1", "Profile 1"},
	}

	for _, tt := range tests {
		got, err := r.Resolve(tt.input)
		if err != nil {
			t.Errorf("Resolve(%q) error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("Resolve(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}

	// Unknown name
	_, err = r.Resolve("Nonexistent")
	if err == nil {
		t.Error("Resolve(Nonexistent) should error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/extract/ -v -run TestProfileResolver
```

- [ ] **Step 3: Implement profile.go**

Create `internal/extract/profile.go`:

```go
package extract

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ProfileResolver struct {
	// uiNameToFolder maps UI display name → folder name
	uiNameToFolder map[string]string
	// folderNames is the set of valid folder names
	folderNames map[string]bool
}

func NewProfileResolver(chromeUserDataDir string) (*ProfileResolver, error) {
	if chromeUserDataDir == "" {
		home, _ := os.UserHomeDir()
		chromeUserDataDir = filepath.Join(home, "Library", "Application Support", "Google", "Chrome")
	}

	localStatePath := filepath.Join(chromeUserDataDir, "Local State")
	data, err := os.ReadFile(localStatePath)
	if err != nil {
		return nil, fmt.Errorf("read Local State: %w", err)
	}

	var state struct {
		Profile struct {
			InfoCache map[string]struct {
				Name string `json:"name"`
			} `json:"info_cache"`
		} `json:"profile"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parse Local State: %w", err)
	}

	r := &ProfileResolver{
		uiNameToFolder: make(map[string]string),
		folderNames:    make(map[string]bool),
	}

	for folder, info := range state.Profile.InfoCache {
		r.uiNameToFolder[info.Name] = folder
		r.folderNames[folder] = true
	}

	return r, nil
}

// Resolve converts a profile name (UI display name or folder name) to a folder name.
func (r *ProfileResolver) Resolve(name string) (string, error) {
	// Try as folder name first
	if r.folderNames[name] {
		return name, nil
	}
	// Try as UI display name
	if folder, ok := r.uiNameToFolder[name]; ok {
		return folder, nil
	}
	return "", fmt.Errorf("Chrome profile %q not found. Available: %s", name, r.availableNames())
}

func (r *ProfileResolver) availableNames() string {
	var names []string
	for uiName, folder := range r.uiNameToFolder {
		names = append(names, fmt.Sprintf("%q (%s)", uiName, folder))
	}
	return strings.Join(names, ", ")
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/extract/ -v -run TestProfileResolver
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/extract/
git commit -m "feat: Chrome profile resolver (UI display name → folder name)"
```

---

## Task 7: Extract — Domain Rules Matching

**Files:**
- Create: `internal/extract/domain.go`, `internal/extract/domain_test.go`

- [ ] **Step 1: Write domain matching tests**

Create `internal/extract/domain_test.go`:

```go
package extract

import (
	"testing"

	"github.com/kouko/reading-list-summarize-scraper/internal/config"
)

func TestMatchDomainRules(t *testing.T) {
	rules := []config.DomainRule{
		{Domains: []string{"medium.com"}, Headed: true, ChromeProfile: "Work"},
		{Domains: []string{"*.substack.com"}, Headed: true, ChromeProfile: "Default"},
		{Domains: []string{"github.com"}, Headed: false, ChromeProfile: "Dev"},
	}

	tests := []struct {
		url         string
		wantHeaded  bool
		wantProfile string
		wantMatch   bool
	}{
		{"https://medium.com/article", true, "Work", true},
		{"https://www.medium.com/article", true, "Work", true},
		{"https://foo.substack.com/post", true, "Default", true},
		{"https://github.com/repo", false, "Dev", true},
		{"https://example.com/page", false, "", false},
	}

	for _, tt := range tests {
		headed, profile, matched := MatchDomainRules(tt.url, rules)
		if matched != tt.wantMatch {
			t.Errorf("MatchDomainRules(%q) matched=%v, want %v", tt.url, matched, tt.wantMatch)
			continue
		}
		if matched {
			if headed != tt.wantHeaded {
				t.Errorf("MatchDomainRules(%q) headed=%v, want %v", tt.url, headed, tt.wantHeaded)
			}
			if profile != tt.wantProfile {
				t.Errorf("MatchDomainRules(%q) profile=%q, want %q", tt.url, profile, tt.wantProfile)
			}
		}
	}
}
```

- [ ] **Step 2: Implement domain.go**

Create `internal/extract/domain.go`:

```go
package extract

import (
	"net/url"
	"strings"

	"github.com/kouko/reading-list-summarize-scraper/internal/config"
)

// MatchDomainRules checks if a URL matches any domain rule.
// Returns (headed, chromeProfile, matched).
func MatchDomainRules(rawURL string, rules []config.DomainRule) (bool, string, bool) {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return false, "", false
	}
	host := strings.ToLower(u.Hostname())

	for _, rule := range rules {
		for _, pattern := range rule.Domains {
			pattern = strings.ToLower(pattern)
			if matchDomain(host, pattern) {
				return rule.Headed, rule.ChromeProfile, true
			}
		}
	}
	return false, "", false
}

func matchDomain(host, pattern string) bool {
	if strings.HasPrefix(pattern, "*.") {
		// Wildcard: *.example.com matches foo.example.com, bar.example.com
		suffix := pattern[1:] // ".example.com"
		return strings.HasSuffix(host, suffix)
	}
	// Exact match: example.com matches example.com and *.example.com
	return host == pattern || strings.HasSuffix(host, "."+pattern)
}
```

- [ ] **Step 3: Run tests**

```bash
go test ./internal/extract/ -v -run TestMatchDomainRules
```

Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/extract/domain.go internal/extract/domain_test.go
git commit -m "feat: domain rules matching with wildcard support"
```

---

## Task 8: Extract — Browser + Defuddle + Pool

**Files:**
- Create: `internal/extract/browser.go`, `internal/extract/defuddle.go`, `internal/extract/pool.go`

- [ ] **Step 1: Get chromedp dependency**

```bash
go get github.com/chromedp/chromedp
go get github.com/chromedp/cdproto
```

- [ ] **Step 2: Create browser.go**

Create `internal/extract/browser.go`:

```go
package extract

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

type Browser struct {
	allocCtx context.Context
	cancel   context.CancelFunc
	headed   bool
	profile  string
}

func NewBrowser(headed bool, profileDir string, userDataDir string) (*Browser, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", !headed),
		chromedp.Flag("disable-gpu", true),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"),
	)

	if profileDir != "" {
		opts = append(opts, chromedp.Flag("profile-directory", profileDir))
	}
	if userDataDir != "" {
		opts = append(opts, chromedp.UserDataDir(userDataDir))
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	return &Browser{
		allocCtx: allocCtx,
		cancel:   cancel,
		headed:   headed,
		profile:  profileDir,
	}, nil
}

// Extract navigates to URL, injects Defuddle JS, and returns extracted Markdown content.
func (b *Browser) Extract(url string, defuddleJS string, timeout time.Duration, waitAfterLoad time.Duration) (string, error) {
	ctx, cancel := chromedp.NewContext(b.allocCtx)
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	var content string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(waitAfterLoad),
		chromedp.Evaluate(defuddleJS, nil),
		chromedp.Evaluate(`window.extractArticle()`, &content, chromedp.AwaitPromise),
	)
	return content, err
}

func (b *Browser) Close() {
	b.cancel()
}
```

- [ ] **Step 3: Create defuddle.go**

Create `internal/extract/defuddle.go`:

```go
package extract

import (
	_ "embed"
)

//go:embed ../../embed/defuddle.min.js
var defuddleJS string

// DefuddleJS returns the embedded Defuddle JavaScript code.
func DefuddleJS() string {
	return defuddleJS
}
```

Note: The embed path is relative to the file's location. If the Go compiler requires the embed to be in the same package or a different arrangement, adjust the embed.go at root to export the variable and import it here. The important thing is that defuddleJS is accessible from the extract package.

If the relative embed path doesn't work (Go only allows embedding from same directory or subdirectories), move the `//go:embed` to the root `embed.go` and pass the JS string into the Pool/Browser at construction time instead.

- [ ] **Step 4: Create pool.go**

Create `internal/extract/pool.go`:

```go
package extract

import (
	"fmt"
	"sync"
	"time"

	"github.com/kouko/reading-list-summarize-scraper/internal/config"
)

type poolKey struct {
	headed  bool
	profile string // resolved folder name
}

type Pool struct {
	mu          sync.Mutex
	instances   map[poolKey]*Browser
	resolver    *ProfileResolver
	cfg         *config.ExtractConfig
	defuddleJS string
}

func NewPool(cfg *config.ExtractConfig, resolver *ProfileResolver, defuddleJS string) *Pool {
	return &Pool{
		instances:   make(map[poolKey]*Browser),
		resolver:    resolver,
		cfg:         cfg,
		defuddleJS: defuddleJS,
	}
}

// ExtractURL extracts content from a URL using the appropriate browser instance.
func (p *Pool) ExtractURL(rawURL string) (string, error) {
	headed, profile := p.resolveForURL(rawURL)

	browser, err := p.getBrowser(headed, profile)
	if err != nil {
		return "", fmt.Errorf("get browser (headed=%v, profile=%q): %w", headed, profile, err)
	}

	return browser.Extract(rawURL, p.defuddleJS, p.cfg.Timeout, p.cfg.WaitAfterLoad)
}

func (p *Pool) resolveForURL(rawURL string) (bool, string) {
	headed, profileName, matched := MatchDomainRules(rawURL, p.cfg.DomainRules)
	if !matched {
		headed = !p.cfg.Headless
		profileName = p.cfg.ChromeProfile
	}

	// Resolve UI name to folder name
	if p.resolver != nil && profileName != "" {
		if folder, err := p.resolver.Resolve(profileName); err == nil {
			profileName = folder
		}
	}
	return headed, profileName
}

func (p *Pool) getBrowser(headed bool, profile string) (*Browser, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := poolKey{headed: headed, profile: profile}
	if b, ok := p.instances[key]; ok {
		return b, nil
	}

	b, err := NewBrowser(headed, profile, p.cfg.UserDataDir)
	if err != nil {
		return nil, err
	}
	p.instances[key] = b
	return b, nil
}

func (p *Pool) CloseAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, b := range p.instances {
		b.Close()
	}
	p.instances = make(map[poolKey]*Browser)
}
```

- [ ] **Step 5: Verify compilation**

```bash
go vet ./internal/extract/
```

Note: If the `//go:embed` relative path in defuddle.go fails, create an alternative approach: remove the embed from defuddle.go, and instead have the Pool receive `defuddleJS string` as a constructor parameter (which it already does). The caller will read it from the root-level embed.go. Update defuddle.go to just be a placeholder or remove it.

- [ ] **Step 6: Commit**

```bash
git add internal/extract/ go.mod go.sum
git commit -m "feat: chromedp browser, Defuddle JS injection, and lazy browser pool"
```

---

## Task 9: Summarize Layer (Port from ytss)

**Files:**
- Create all files in `internal/summarize/`

This task ports the entire summarize layer from youtube-summarize-scraper. The code is functionally identical — only the package path and prompt variable names change.

- [ ] **Step 1: Port core files from ytss**

Copy and adapt the following files from `/Users/kouko/VisualStudioCodeProject/youtube-summarize-scraper/summarizer/` to `internal/summarize/`:

| ytss file | rlss file | Changes needed |
|-----------|-----------|---------------|
| `summarizer.go` | `internal/summarize/summarizer.go` | Change package to `summarize`, update config import path |
| `fallback.go` | `internal/summarize/fallback.go` | Change package |
| `circuit_breaker.go` | `internal/summarize/circuit_breaker.go` | Change package |
| `errors.go` | `internal/summarize/errors.go` | Change package |
| `keywords.go` | `internal/summarize/keywords.go` | Change package |
| `mermaid.go` | `internal/summarize/mermaid.go` | Change package |
| `claude_code.go` | `internal/summarize/claude_code.go` | Change package, config import |
| `claude.go` | `internal/summarize/claude.go` | Change package, config import |
| `gemini.go` | `internal/summarize/gemini.go` | Change package, config import |
| `qwen_code.go` | `internal/summarize/qwen_code.go` | Change package, config import |
| `ollama.go` | `internal/summarize/ollama.go` | Change package, config import |
| `llamacpp.go` | `internal/summarize/llamacpp.go` | Change package, config import |
| `openai_compat.go` | `internal/summarize/openai_compat.go` | Change package, config import |

For each file:
1. Change `package summarizer` → `package summarize`
2. Change config import from `github.com/kouko/youtube-summarize-scraper/config` → `github.com/kouko/reading-list-summarize-scraper/internal/config`
3. Keep all logic, types, and method signatures identical

- [ ] **Step 2: Port and adapt prompt.go**

Copy ytss `summarizer/prompt.go` to `internal/summarize/prompt.go`. Then modify:

1. Change package name
2. Update `PromptVars` struct to match rlss variables:

```go
type PromptVars struct {
	Title         string
	Domain        string
	DateAdded     string
	Source        string
	Content       string
	ContentLength int
}
```

3. Update `SubstituteVars` to replace rlss-specific placeholders:
   - `{{title}}`, `{{domain}}`, `{{date_added}}`, `{{source}}`, `{{content_length}}`, `{{content_tier}}`, `{{content}}`

4. Update `CalculateTier` to use the rlss tier names and thresholds from the spec (same thresholds as ytss but for `content` instead of `transcription`).

5. Keep the `ResolvePrompt` 4-level cascade logic identical, but without the per-channel level (simplify to 3 levels for now: global file → inline → built-in).

- [ ] **Step 3: Port tests**

Copy and adapt from ytss:
- `circuit_breaker_test.go` → `internal/summarize/circuit_breaker_test.go`
- `errors_test.go` → `internal/summarize/errors_test.go`
- `fallback_test.go` → `internal/summarize/fallback_test.go`
- `prompt_test.go` → `internal/summarize/prompt_test.go`
- `summarizer_test.go` → `internal/summarize/summarizer_test.go`

Change package names and config imports in each.

- [ ] **Step 4: Run tests**

```bash
go test ./internal/summarize/ -v
```

Expected: All tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/summarize/
git commit -m "feat: summarize layer ported from ytss (fallback chain, circuit breaker, all providers)"
```

---

## Task 10: Built-in Prompt Templates

**Files:**
- Create all files in `prompts/builtin/`

- [ ] **Step 1: Create embed.go for prompts**

Create `prompts/builtin/embed.go`:

```go
package builtin

import "embed"

//go:embed summary-en.md summary-ja.md summary-zh-Hant.md
//go:embed keywords-en.md keywords-ja.md keywords-zh-Hant.md
//go:embed mermaid-en.md mermaid-ja.md mermaid-zh-Hant.md
var FS embed.FS
```

- [ ] **Step 2: Create summary prompts (3 languages)**

Create `prompts/builtin/summary-zh-Hant.md`, `summary-en.md`, `summary-ja.md`.

These follow the same structure as ytss summary prompts but adapted for web articles:
- Replace `{{channel_name}}` → `{{domain}}`
- Replace `{{upload_date}}` → `{{date_added}}`
- Remove `{{duration}}`, `{{tags}}`
- Replace `{{transcription_length}}` → `{{content_length}}`
- Replace `{{transcription_tier}}` → `{{content_tier}}`
- Replace `{{transcript}}` → `{{content}}`
- Adjust context from "video transcript" to "web article extracted content"

Copy the structure, output scale guide, format rules, and section headings from the corresponding ytss files in `/Users/kouko/VisualStudioCodeProject/youtube-summarize-scraper/prompts/builtin/`.

- [ ] **Step 3: Create keyword prompts (3 languages)**

Copy directly from ytss — these are identical (they operate on summary text, not source content):
- `keywords-en.md`, `keywords-ja.md`, `keywords-zh-Hant.md`

- [ ] **Step 4: Create mermaid prompts (3 languages)**

Copy directly from ytss — these are identical:
- `mermaid-en.md`, `mermaid-ja.md`, `mermaid-zh-Hant.md`

- [ ] **Step 5: Verify embed works**

```bash
go vet ./prompts/builtin/
```

- [ ] **Step 6: Commit**

```bash
git add prompts/
git commit -m "feat: built-in prompt templates (summary, keywords, mermaid) in 3 languages"
```

---

## Task 11: Output — Obsidian Markdown Assembly

**Files:**
- Create: `internal/output/obsidian.go`, `internal/output/obsidian_test.go`

- [ ] **Step 1: Write assembly test**

Create `internal/output/obsidian_test.go`:

```go
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

	// Check frontmatter
	if !strings.Contains(result, `title: "Test Article"`) {
		t.Error("missing title in frontmatter")
	}
	if !strings.Contains(result, `url: "https://example.com/test"`) {
		t.Error("missing url in frontmatter")
	}
	if !strings.Contains(result, "- golang") {
		t.Error("missing keyword tag")
	}

	// Check body — no # Title heading
	if strings.Contains(result, "# Test Article") {
		t.Error("should not have # Title heading")
	}

	// Check info callout
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/output/ -v -run TestAssemble
```

- [ ] **Step 3: Implement obsidian.go**

Create `internal/output/obsidian.go`:

```go
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
// Port the logic from ytss pipeline.go insertMermaidBlocksAfterFirstHeading.
func insertMermaidBlocks(body string, blocks []MermaidBlock) string {
	// For each block, try to match its Title to a ### heading in the body.
	// If matched, insert after that section (before the next ### heading).
	// Unmatched blocks are inserted after the first ### section.
	// This logic should be ported from ytss.
	lines := strings.Split(body, "\n")
	var result []string
	var unmatchedBlocks []MermaidBlock

	for _, block := range blocks {
		matched := false
		normalizedTitle := strings.TrimPrefix(strings.TrimPrefix(block.Title, "#### "), "### ")

		for i, line := range lines {
			lineTitle := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(line), "#### "), "### ")
			if lineTitle == normalizedTitle {
				// Find next heading or end
				insertIdx := len(lines)
				for j := i + 1; j < len(lines); j++ {
					if strings.HasPrefix(strings.TrimSpace(lines[j]), "###") {
						insertIdx = j
						break
					}
				}
				// Insert before the next heading
				mermaidText := fmt.Sprintf("\n```mermaid\n%s\n```\n", block.Code)
				newLines := make([]string, 0, len(lines)+4)
				newLines = append(newLines, lines[:insertIdx]...)
				newLines = append(newLines, mermaidText)
				newLines = append(newLines, lines[insertIdx:]...)
				lines = newLines
				matched = true
				break
			}
		}
		if !matched {
			unmatchedBlocks = append(unmatchedBlocks, block)
		}
	}

	// Insert unmatched blocks after overview section (second ### heading)
	if len(unmatchedBlocks) > 0 {
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
		var mermaidText string
		for _, block := range unmatchedBlocks {
			mermaidText += fmt.Sprintf("\n```mermaid\n%s\n```\n", block.Code)
		}
		newLines := make([]string, 0, len(lines)+len(unmatchedBlocks)*4)
		newLines = append(newLines, lines[:insertIdx]...)
		newLines = append(newLines, mermaidText)
		newLines = append(newLines, lines[insertIdx:]...)
		lines = newLines
	}

	result = lines
	return strings.Join(result, "\n")
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/output/ -v -run TestAssemble
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/output/obsidian.go internal/output/obsidian_test.go
git commit -m "feat: Obsidian Markdown assembly (summary + content with frontmatter)"
```

---

## Task 12: Output — copy_to

**Files:**
- Create: `internal/output/copyto.go`, `internal/output/copyto_test.go`

- [ ] **Step 1: Write copy_to tests**

Create `internal/output/copyto_test.go`:

```go
package output

import (
	"testing"

	"github.com/kouko/reading-list-summarize-scraper/internal/config"
)

func TestExpandTemplate(t *testing.T) {
	vars := CopyToVars{
		OutputDir: "/vault/references",
		Date:      "2026-03-28",
		DateAdded: "2026-03-25",
		Title:     "Test Article Title",
		SHA8:      "a1b2c3d4",
		Source:    "safari",
		Domain:    "example.com",
		DomainDir: "example_com",
		Type:      "summary",
	}

	got := ExpandTemplate("{output_dir}/by-source/{source}/{domain_dir}", vars)
	want := "/vault/references/by-source/safari/example_com"
	if got != want {
		t.Errorf("ExpandTemplate path = %q, want %q", got, want)
	}

	got = ExpandTemplate("{date}__{title}__{type}.md", vars)
	want = "2026-03-28__Test Article Title__summary.md"
	if got != want {
		t.Errorf("ExpandTemplate filename = %q, want %q", got, want)
	}
}

func TestSanitizeTitleForDisplay(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello/World: Test", "Hello World Test"},
		{"日本語テスト", "日本語テスト"},
		{`A "quoted" <title>`, "A quoted title"},
	}
	for _, tt := range tests {
		got := SanitizeTitleForDisplay(tt.input)
		if got != tt.want {
			t.Errorf("SanitizeTitleForDisplay(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
```

- [ ] **Step 2: Implement copyto.go**

Create `internal/output/copyto.go`:

```go
package output

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/kouko/reading-list-summarize-scraper/internal/config"
)

type CopyToVars struct {
	OutputDir string
	Date      string
	DateAdded string
	Title     string
	SHA8      string
	Source    string
	Domain    string
	DomainDir string
	Type      string
}

func ExpandTemplate(tmpl string, vars CopyToVars) string {
	r := strings.NewReplacer(
		"{output_dir}", vars.OutputDir,
		"{date}", vars.Date,
		"{date_added}", vars.DateAdded,
		"{title}", vars.Title,
		"{sha8}", vars.SHA8,
		"{source}", vars.Source,
		"{domain}", vars.Domain,
		"{domain_dir}", vars.DomainDir,
		"{type}", vars.Type,
	)
	return r.Replace(tmpl)
}

var unsafeChars = regexp.MustCompile(`[\\/:*?"<>|]`)

func SanitizeTitleForDisplay(title string) string {
	clean := unsafeChars.ReplaceAllString(title, "")
	clean = strings.TrimSpace(clean)
	if utf8.RuneCountInString(clean) > 80 {
		runes := []rune(clean)
		clean = string(runes[:80])
	}
	return clean
}

func ExecuteCopyTo(cfg config.CopyToConfig, sourceDir string, sha8 string, vars CopyToVars) error {
	if !cfg.Enabled {
		return nil
	}

	for _, fileType := range cfg.Files {
		vars.Type = fileType
		srcPattern := fmt.Sprintf("*__%s__%s.md", sha8, fileType)

		matches, _ := filepath.Glob(filepath.Join(sourceDir, srcPattern))
		if len(matches) == 0 {
			slog.Warn("copy_to: source file not found", "pattern", srcPattern)
			continue
		}
		srcPath := matches[0]

		targetDir := ExpandTemplate(cfg.Path, vars)
		targetFile := ExpandTemplate(cfg.Filename, vars)
		targetPath := filepath.Join(targetDir, targetFile)

		// Check overwrite
		if !cfg.Overwrite {
			if _, err := os.Stat(targetPath); err == nil {
				slog.Debug("copy_to: skip existing", "path", targetPath)
				continue
			}
		}

		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("create dir %s: %w", targetDir, err)
		}

		if err := copyFile(srcPath, targetPath); err != nil {
			return fmt.Errorf("copy %s → %s: %w", srcPath, targetPath, err)
		}
		slog.Info("copy_to", "src", srcPath, "dst", targetPath)
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
```

- [ ] **Step 3: Run tests**

```bash
go test ./internal/output/ -v -run "TestExpandTemplate|TestSanitize"
```

Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/output/copyto.go internal/output/copyto_test.go
git commit -m "feat: copy_to with template expansion and file copying"
```

---

## Task 13: Pipeline — Stats + Runner

**Files:**
- Create: `internal/pipeline/stats.go`, `internal/pipeline/runner.go`

- [ ] **Step 1: Create stats.go**

Create `internal/pipeline/stats.go`:

```go
package pipeline

import (
	"fmt"
	"strings"
	"time"
)

type ItemError struct {
	URL   string
	Title string
	Err   error
}

type Stats struct {
	Success int
	Skipped int
	Failed  int
	Errors  []ItemError
	Start   time.Time
	End     time.Time
}

func (s *Stats) Duration() time.Duration {
	return s.End.Sub(s.Start)
}

func (s *Stats) Report() string {
	var b strings.Builder
	b.WriteString("📊 處理報告\n")
	b.WriteString(fmt.Sprintf("  ✅ 成功: %d\n", s.Success))
	b.WriteString(fmt.Sprintf("  ⏭️ 跳過: %d\n", s.Skipped))
	b.WriteString(fmt.Sprintf("  ❌ 失敗: %d\n", s.Failed))
	b.WriteString(fmt.Sprintf("  ⏱️ 耗時: %s\n", s.Duration().Truncate(time.Second)))
	if len(s.Errors) > 0 {
		b.WriteString("  失敗項目:\n")
		for _, e := range s.Errors {
			b.WriteString(fmt.Sprintf("    - %s: %v\n", e.Title, e.Err))
		}
	}
	return b.String()
}
```

- [ ] **Step 2: Create runner.go**

Create `internal/pipeline/runner.go`:

```go
package pipeline

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/kouko/reading-list-summarize-scraper/internal/config"
	"github.com/kouko/reading-list-summarize-scraper/internal/extract"
	"github.com/kouko/reading-list-summarize-scraper/internal/output"
	"github.com/kouko/reading-list-summarize-scraper/internal/source"
	"github.com/kouko/reading-list-summarize-scraper/internal/summarize"
)

var errSkipped = fmt.Errorf("skipped")

func IsSkipped(err error) bool { return err == errSkipped }

type Pipeline struct {
	config     *config.Config
	pool       *extract.Pool
	summarizer summarize.Summarizer
	index      *output.FileIndex
	ctx        context.Context
	cancel     context.CancelFunc
	force      bool
	dryRun     bool
}

func New(cfg *config.Config, pool *extract.Pool, sum summarize.Summarizer) *Pipeline {
	ctx, cancel := context.WithCancel(context.Background())
	return &Pipeline{
		config:     cfg,
		pool:       pool,
		summarizer: sum,
		index:      output.NewFileIndex(),
		ctx:        ctx,
		cancel:     cancel,
		force:      !cfg.Pipeline.SkipExisting,
		dryRun:     cfg.Pipeline.DryRun,
	}
}

func (p *Pipeline) Shutdown()          { p.cancel() }
func (p *Pipeline) stopped() bool      { return p.ctx.Err() != nil }

func (p *Pipeline) ResetContext() {
	p.ctx, p.cancel = context.WithCancel(context.Background())
}

func (p *Pipeline) RebuildIndex() {
	p.index.Build(p.config.OutputDir)
}

func (p *Pipeline) ProcessBatch(items []source.ReadingItem) Stats {
	stats := Stats{Start: time.Now()}
	p.index.Build(p.config.OutputDir)

	for i, item := range items {
		if p.stopped() {
			break
		}

		slog.Info("processing", "index", fmt.Sprintf("%d/%d", i+1, len(items)), "title", item.Title, "url", item.URL)

		err := p.ProcessItem(item)
		switch {
		case err == nil:
			stats.Success++
		case IsSkipped(err):
			stats.Skipped++
		default:
			stats.Failed++
			stats.Errors = append(stats.Errors, ItemError{
				URL:   item.URL,
				Title: item.Title,
				Err:   err,
			})
			slog.Error("failed", "url", item.URL, "err", err)
		}

		// Random delay between items
		if i < len(items)-1 && !p.dryRun {
			delay := p.config.Pipeline.DelayMin
			if p.config.Pipeline.DelayMax > delay {
				delay += rand.IntN(p.config.Pipeline.DelayMax - delay)
			}
			time.Sleep(time.Duration(delay) * time.Second)
		}
	}

	stats.End = time.Now()
	return stats
}

func (p *Pipeline) ProcessItem(item source.ReadingItem) error {
	// Step 1: SHA8
	sha8 := output.SHA8(item.URL)
	domainDir := output.DomainDir(item.URL)
	now := time.Now()

	// Step 2: Skip detection
	if !p.force && p.index.Has(sha8) {
		info := p.index.Get(sha8)
		if info.SummaryExists {
			slog.Debug("skip: already processed", "sha8", sha8)
			return errSkipped
		}
	}

	// Step 3: Dry run
	if p.dryRun {
		slog.Info("dry-run", "title", item.Title, "url", item.URL, "sha8", sha8)
		return errSkipped
	}

	// Step 4: Resume detection
	var extractedContent string
	if p.index.Has(sha8) {
		info := p.index.Get(sha8)
		if info.ContentExists && !info.SummaryExists {
			contentPath := p.index.ContentPath(sha8)
			if contentPath != "" {
				data, err := os.ReadFile(contentPath)
				if err == nil {
					// Extract content after frontmatter
					extractedContent = stripFrontmatter(string(data))
					slog.Info("resume: using existing content", "sha8", sha8)
				}
			}
		}
	}

	// Step 5+6: Extract (if not resuming)
	if extractedContent == "" {
		content, err := p.pool.ExtractURL(item.URL)
		if err != nil {
			return fmt.Errorf("extract: %w", err)
		}
		if content == "" {
			return fmt.Errorf("extract: empty content")
		}
		extractedContent = content

		// Resolve title from content if manual source
		if item.Title == "" {
			item.Title = extractTitle(extractedContent, item.URL)
		}
	}

	contentLength := utf8.RuneCountInString(extractedContent)
	domain := extractDomain(item.URL)

	// Step 7: Create output dir
	outDir := filepath.Join(p.config.OutputDir, domainDir)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	// Step 8: Write content file
	contentParams := output.ContentParams{
		Title:         item.Title,
		URL:           item.URL,
		Domain:        domain,
		Source:        item.Source,
		DateAdded:     item.DateAdded,
		ProcessedDate: now,
		ContentLength: contentLength,
		ExtractedBy:   "defuddle-js",
		Content:       extractedContent,
	}
	contentFilename := output.ContentFilename(now, sha8)
	contentPath := filepath.Join(outDir, contentFilename)
	if err := os.WriteFile(contentPath, []byte(output.AssembleContent(contentParams)), 0644); err != nil {
		return fmt.Errorf("write content: %w", err)
	}
	slog.Info("wrote content", "path", contentPath)

	// Step 9: Summarization
	if p.summarizer == nil {
		slog.Warn("no summarizer configured, skipping summary")
		return nil
	}

	summaryResult, keywords, mermaidBlocks, err := p.runSummarization(extractedContent, item, domain, contentLength)
	if err != nil {
		return fmt.Errorf("summarize: %w", err)
	}

	// Step 10: Write summary file
	tier := summarize.CalculateTier(contentLength, p.config.Summary.Language)
	summaryParams := output.SummaryParams{
		Title:         item.Title,
		URL:           item.URL,
		Domain:        domain,
		Source:        item.Source,
		DateAdded:     item.DateAdded,
		ProcessedDate: now,
		LLMProvider:   summaryResult.Provider,
		LLMModel:      summaryResult.Model,
		ContentLength: contentLength,
		ContentTier:   tier,
		SummaryText:   summaryResult.Text,
		Keywords:      keywords,
		MermaidBlocks: mermaidBlocks,
	}
	summaryFilename := output.SummaryFilename(now, sha8)
	summaryPath := filepath.Join(outDir, summaryFilename)
	if err := os.WriteFile(summaryPath, []byte(output.AssembleSummary(summaryParams)), 0644); err != nil {
		return fmt.Errorf("write summary: %w", err)
	}
	slog.Info("wrote summary", "path", summaryPath)

	// Step 11: copy_to
	if p.config.CopyTo.Enabled {
		vars := output.CopyToVars{
			OutputDir: p.config.OutputDir,
			Date:      now.Format("2006-01-02"),
			DateAdded: item.DateAdded.Format("2006-01-02"),
			Title:     output.SanitizeTitleForDisplay(item.Title),
			SHA8:      sha8,
			Source:    item.Source,
			Domain:    domain,
			DomainDir: domainDir,
		}
		if err := output.ExecuteCopyTo(p.config.CopyTo, outDir, sha8, vars); err != nil {
			slog.Warn("copy_to failed", "err", err)
		}
	}

	return nil
}

func (p *Pipeline) runSummarization(content string, item source.ReadingItem, domain string, contentLength int) (summarize.SummarizeResult, []string, []output.MermaidBlock, error) {
	// Stage 1: Main summary (blocking)
	prompt, err := summarize.ResolveAndSubstitute(p.config.Summary, summarize.PromptVars{
		Title:         item.Title,
		Domain:        domain,
		DateAdded:     item.DateAdded.Format("2006-01-02"),
		Source:        item.Source,
		Content:       content,
		ContentLength: contentLength,
	})
	if err != nil {
		return summarize.SummarizeResult{}, nil, nil, fmt.Errorf("resolve prompt: %w", err)
	}

	opts := summarize.SummarizeOptions{
		Prompt:    prompt,
		MaxTokens: p.config.Summary.MaxTokens,
	}
	result, err := p.summarizer.Summarize(content, opts)
	if err != nil {
		return summarize.SummarizeResult{}, nil, nil, err
	}

	// Stage 2: Keywords (non-blocking)
	var keywords []string
	if p.config.Summary.Keywords.Enabled {
		kw, kwErr := p.extractKeywords(result.Text)
		if kwErr != nil {
			slog.Warn("keywords extraction failed", "err", kwErr)
		} else {
			keywords = kw
		}
	}

	// Stage 3: Mermaid (non-blocking)
	var mermaidBlocks []output.MermaidBlock
	if p.config.Summary.Mermaid.Enabled {
		blocks, mErr := p.generateMermaid(result.Text)
		if mErr != nil {
			slog.Warn("mermaid generation failed", "err", mErr)
		} else {
			mermaidBlocks = blocks
		}
	}

	return result, keywords, mermaidBlocks, nil
}

func (p *Pipeline) extractKeywords(summaryText string) ([]string, error) {
	prompt := summarize.KeywordPrompt(summaryText, p.config.Summary.Keywords)
	opts := summarize.SummarizeOptions{Prompt: prompt, MaxTokens: 500}
	result, err := p.summarizer.Summarize(summaryText, opts)
	if err != nil {
		return nil, err
	}
	return summarize.ParseKeywords(result.Text), nil
}

func (p *Pipeline) generateMermaid(summaryText string) ([]output.MermaidBlock, error) {
	prompt := summarize.MermaidPrompt(summaryText, p.config.Summary.Language)
	opts := summarize.SummarizeOptions{Prompt: prompt, MaxTokens: 4000}
	result, err := p.summarizer.Summarize(summaryText, opts)
	if err != nil {
		return nil, err
	}
	blocks := summarize.ValidateMermaidBlocks(result.Text)
	var out []output.MermaidBlock
	for _, b := range blocks {
		out = append(out, output.MermaidBlock{Title: b.Title, Code: b.Code})
	}
	return out, nil
}

func stripFrontmatter(content string) string {
	if !strings.HasPrefix(content, "---\n") {
		return content
	}
	end := strings.Index(content[4:], "\n---\n")
	if end == -1 {
		return content
	}
	return strings.TrimSpace(content[end+8:])
}

func extractDomain(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "unknown"
	}
	return u.Hostname()
}

func extractTitle(content string, fallbackURL string) string {
	// Try to find first # heading
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return fallbackURL
}
```

- [ ] **Step 3: Verify compilation**

```bash
go vet ./internal/pipeline/
```

- [ ] **Step 4: Commit**

```bash
git add internal/pipeline/stats.go internal/pipeline/runner.go
git commit -m "feat: pipeline runner with ProcessBatch, ProcessItem, 3-stage summarization"
```

---

## Task 14: Pipeline — Watch Mode

**Files:**
- Create: `internal/pipeline/watch.go`

- [ ] **Step 1: Implement watch.go**

Create `internal/pipeline/watch.go`:

```go
package pipeline

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kouko/reading-list-summarize-scraper/internal/config"
	"github.com/kouko/reading-list-summarize-scraper/internal/source"
)

// RunWatch executes the watch loop until interrupted.
func (p *Pipeline) RunWatch(fetchSources func() ([]source.ReadingItem, error), configPath string) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		// Reset context for this iteration
		p.ResetContext()

		// Partial config reload
		if err := p.reloadConfig(configPath); err != nil {
			slog.Warn("config reload failed, using previous config", "err", err)
		}

		// Rebuild index
		p.RebuildIndex()

		// Fetch and filter items
		items, err := fetchSources()
		if err != nil {
			slog.Error("fetch sources failed", "err", err)
		} else {
			// Run batch in this goroutine (signal handling below)
			doneCh := make(chan Stats, 1)
			go func() {
				stats := p.ProcessBatch(items)
				doneCh <- stats
			}()

			// Wait for batch completion or signal
			select {
			case stats := <-doneCh:
				slog.Info("iteration complete", "success", stats.Success, "skipped", stats.Skipped, "failed", stats.Failed)
			case sig := <-sigCh:
				slog.Info("received signal during processing, shutting down", "signal", sig)
				p.Shutdown()
				<-doneCh // wait for current item to finish
				slog.Info("shutdown complete")
				return
			}
		}

		// Sleep with signal awareness
		interval := time.Duration(p.config.Watch.Interval) * time.Minute
		slog.Info("sleeping until next iteration", "interval", interval)

		select {
		case <-time.After(interval):
			// Continue to next iteration
		case sig := <-sigCh:
			slog.Info("received signal during sleep, exiting", "signal", sig)
			return
		}
	}
}

func (p *Pipeline) reloadConfig(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}
	// Partial reload: only runtime-safe fields
	p.config.Filter = cfg.Filter
	p.config.Watch = cfg.Watch
	p.config.LLM = cfg.LLM
	p.config.Summary = cfg.Summary
	p.config.Pipeline = cfg.Pipeline
	return nil
}
```

- [ ] **Step 2: Verify compilation**

```bash
go vet ./internal/pipeline/
```

- [ ] **Step 3: Commit**

```bash
git add internal/pipeline/watch.go
git commit -m "feat: watch mode with signal handling and partial config reload"
```

---

## Task 15: Chrome Source

**Files:**
- Create: `internal/source/chrome.go`

- [ ] **Step 1: Implement Chrome source**

Create `internal/source/chrome.go`:

```go
package source

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)

//go:embed ../../embed/extension/manifest.json
var extensionManifest []byte

//go:embed ../../embed/extension/background.js
var extensionBackground []byte

type ChromeSource struct {
	profileDir  string // resolved folder name (e.g., "Profile 5")
	userDataDir string
}

func NewChromeSource(profileDir string, userDataDir string) *ChromeSource {
	if userDataDir == "" {
		home, _ := os.UserHomeDir()
		userDataDir = filepath.Join(home, "Library", "Application Support", "Google", "Chrome")
	}
	return &ChromeSource{profileDir: profileDir, userDataDir: userDataDir}
}

func (c *ChromeSource) Name() string { return "chrome" }

func (c *ChromeSource) Fetch() ([]ReadingItem, error) {
	// Extract embedded extension to temp dir
	extDir, err := extractExtension()
	if err != nil {
		return nil, fmt.Errorf("extract extension: %w", err)
	}
	defer os.RemoveAll(extDir)

	// Launch headed Chrome with extension + profile
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-extensions", false),
		chromedp.Flag("disable-extensions-except", extDir),
		chromedp.Flag("load-extension", extDir),
		chromedp.Flag("profile-directory", c.profileDir),
		chromedp.UserDataDir(c.userDataDir),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	ctx, ctxCancel := chromedp.NewContext(allocCtx)
	defer ctxCancel()

	ctx, timeoutCancel := context.WithTimeout(ctx, 30*time.Second)
	defer timeoutCancel()

	// Navigate to trigger browser startup
	if err := chromedp.Run(ctx, chromedp.Navigate("about:blank")); err != nil {
		return nil, fmt.Errorf("launch chrome: %w", err)
	}

	// Wait for extension to load
	time.Sleep(2 * time.Second)

	// Find extension service worker target
	targets, err := target.GetTargets().Do(chromedp.FromContext(ctx).WithExecutor(ctx, chromedp.FromContext(ctx).Browser))
	if err != nil {
		// Fallback: try using the cdp package directly
		slog.Warn("GetTargets via helper failed, trying direct", "err", err)
	}

	var swTargetID target.ID
	for _, t := range targets {
		if t.Type == "service_worker" {
			swTargetID = t.TargetID
			break
		}
	}

	if swTargetID == "" {
		return nil, fmt.Errorf("extension service worker not found")
	}

	// Attach to service worker
	sessionID, err := target.AttachToTarget(swTargetID).WithFlatten(true).Do(
		chromedp.FromContext(ctx).WithExecutor(ctx, chromedp.FromContext(ctx).Browser),
	)
	if err != nil {
		return nil, fmt.Errorf("attach to SW: %w", err)
	}

	// Execute chrome.readingList.query({}) in SW context
	evalResult, exceptionDetails, err := runtime.Evaluate(`
		(async () => {
			const entries = await chrome.readingList.query({});
			return JSON.stringify(entries);
		})()
	`).WithAwaitPromise(true).Do(
		// Route to the correct session
		chromedp.FromContext(ctx).WithExecutor(ctx, chromedp.FromContext(ctx).Target),
	)
	_ = sessionID
	if err != nil {
		return nil, fmt.Errorf("evaluate in SW: %w", err)
	}
	if exceptionDetails != nil {
		return nil, fmt.Errorf("JS exception: %s", exceptionDetails.Text)
	}

	// Parse JSON result
	var jsonStr string
	if err := json.Unmarshal(evalResult.Value, &jsonStr); err != nil {
		return nil, fmt.Errorf("unmarshal result: %w", err)
	}

	var entries []chromeReadingListEntry
	if err := json.Unmarshal([]byte(jsonStr), &entries); err != nil {
		return nil, fmt.Errorf("parse entries: %w", err)
	}

	return chromeEntriesToItems(entries), nil
}

type chromeReadingListEntry struct {
	URL            string  `json:"url"`
	Title          string  `json:"title"`
	HasBeenRead    bool    `json:"hasBeenRead"`
	CreationTime   float64 `json:"creationTime"`
	LastUpdateTime float64 `json:"lastUpdateTime"`
}

func chromeEntriesToItems(entries []chromeReadingListEntry) []ReadingItem {
	items := make([]ReadingItem, 0, len(entries))
	for _, e := range entries {
		items = append(items, ReadingItem{
			Title:    e.Title,
			URL:      e.URL,
			DateAdded: time.UnixMilli(int64(e.CreationTime)),
			IsUnread:  !e.HasBeenRead,
			Source:    "chrome",
		})
	}
	return items
}

func extractExtension() (string, error) {
	dir, err := os.MkdirTemp("", "rlss-ext-*")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), extensionManifest, 0644); err != nil {
		return dir, err
	}
	if err := os.WriteFile(filepath.Join(dir, "background.js"), extensionBackground, 0644); err != nil {
		return dir, err
	}
	return dir, nil
}
```

Note: The CDP target/session routing for the service worker is complex and the exact API may need adjustment during implementation. The chromedp library's internal API for accessing the browser-level executor may differ from what's shown. Consult the chromedp examples and cdproto documentation. The core approach is correct: GetTargets → find SW → AttachToTarget → Evaluate in that session.

- [ ] **Step 2: Verify compilation**

```bash
go vet ./internal/source/
```

If the `//go:embed` with `../../embed/` relative path fails, restructure: have the extension bytes passed in as constructor parameters to ChromeSource, with the embed happening at the cmd/rlss level.

- [ ] **Step 3: Commit**

```bash
git add internal/source/chrome.go
git commit -m "feat: Chrome Reading List source via CDP Extension + Service Worker"
```

---

## Task 16: CLI — Cobra Commands

**Files:**
- Create: `cmd/rlss/main.go`, `cmd/rlss/root.go`, `cmd/rlss/process.go`, `cmd/rlss/list.go`, `cmd/rlss/url.go`, `cmd/rlss/config_cmd.go`

- [ ] **Step 1: Get cobra dependency**

```bash
go get github.com/spf13/cobra
```

- [ ] **Step 2: Create main.go**

Create `cmd/rlss/main.go`:

```go
package main

import (
	"log/slog"
	"os"
)

var version = "dev"

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func setupLogging(verbose bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})))
}
```

- [ ] **Step 3: Create root.go**

Create `cmd/rlss/root.go`:

```go
package main

import (
	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	outputDir  string
	llmFlag    string
	profileFlag string
	verbose    bool
	force      bool
	dryRun     bool
	watchFlag  bool
	intervalFlag int
	safariFlag bool
	chromeFlag bool
	allFlag    bool
	unreadFlag bool
	sinceFlag  string
	limitFlag  int
)

var rootCmd = &cobra.Command{
	Use:   "rlss",
	Short: "Reading List Summarize Scraper",
	Long:  "Fetch URLs from macOS Safari/Chrome Reading Lists, extract content, and generate Obsidian summaries.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return processCmd.RunE(cmd, args)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "", "", "config file (default: ~/.config/rlss/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&outputDir, "output", "o", "", "output directory")
	rootCmd.PersistentFlags().StringVar(&llmFlag, "llm", "", "override primary LLM provider")
	rootCmd.PersistentFlags().StringVar(&profileFlag, "profile", "", "Chrome Reading List profile (UI name)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging")
	rootCmd.PersistentFlags().BoolVar(&force, "force", false, "force reprocess existing")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "list items without processing")
	rootCmd.PersistentFlags().BoolVarP(&watchFlag, "watch", "w", false, "watch mode")
	rootCmd.PersistentFlags().IntVar(&intervalFlag, "interval", 0, "watch interval in minutes")
	rootCmd.PersistentFlags().BoolVarP(&safariFlag, "safari", "s", false, "process Safari Reading List")
	rootCmd.PersistentFlags().BoolVarP(&chromeFlag, "chrome", "c", false, "process Chrome Reading List")
	rootCmd.PersistentFlags().BoolVarP(&allFlag, "all", "a", false, "process all sources")
	rootCmd.PersistentFlags().BoolVar(&unreadFlag, "unread", false, "unread items only")
	rootCmd.PersistentFlags().StringVar(&sinceFlag, "since", "", "only items added after date (YYYY-MM-DD)")
	rootCmd.PersistentFlags().IntVarP(&limitFlag, "limit", "n", 0, "max items to process")

	rootCmd.AddCommand(processCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(urlCmd)
	rootCmd.AddCommand(configShowCmd)
}
```

- [ ] **Step 4: Create process.go**

Create `cmd/rlss/process.go`. This command loads config, applies CLI overrides, initializes the pipeline, fetches sources, filters, and runs ProcessBatch or watch mode.

```go
package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/kouko/reading-list-summarize-scraper/internal/config"
	"github.com/kouko/reading-list-summarize-scraper/internal/extract"
	"github.com/kouko/reading-list-summarize-scraper/internal/output"
	"github.com/kouko/reading-list-summarize-scraper/internal/pipeline"
	"github.com/kouko/reading-list-summarize-scraper/internal/source"
	"github.com/kouko/reading-list-summarize-scraper/internal/summarize"
)

var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Batch process Reading Lists",
	RunE:  runProcess,
}

func runProcess(cmd *cobra.Command, args []string) error {
	setupLogging(verbose)

	cfg, cfgPath, err := loadAndOverrideConfig()
	if err != nil {
		return err
	}

	// Initialize pool
	var resolver *extract.ProfileResolver
	resolver, _ = extract.NewProfileResolver("") // ignore error if Chrome not installed

	defuddleJS := getDefuddleJS() // from embed.go or root package
	pool := extract.NewPool(&cfg.Extract, resolver, defuddleJS)
	defer pool.CloseAll()

	// Initialize summarizer
	sum, err := summarize.NewSummarizer(cfg.LLM)
	if err != nil {
		slog.Warn("summarizer init failed, summaries will be skipped", "err", err)
	}

	// Create pipeline
	p := pipeline.New(cfg, pool, sum)

	fetchSources := func() ([]source.ReadingItem, error) {
		return fetchAndFilter(cfg, resolver)
	}

	if cfg.Watch.Enabled {
		p.RunWatch(fetchSources, cfgPath)
		return nil
	}

	// Single run
	items, err := fetchSources()
	if err != nil {
		return err
	}

	if len(items) == 0 {
		fmt.Println("No items to process.")
		return nil
	}

	fmt.Printf("📚 Reading List Summarize Scraper %s\n", version)
	fmt.Printf("   %d items to process\n\n", len(items))

	stats := p.ProcessBatch(items)
	pool.CloseAll()

	fmt.Println()
	fmt.Print(stats.Report())
	fmt.Printf("  📁 輸出: %s\n", cfg.OutputDir)

	return nil
}

func loadAndOverrideConfig() (*config.Config, string, error) {
	cfgPath := cfgFile
	if cfgPath == "" {
		home, _ := os.UserHomeDir()
		cfgPath = filepath.Join(home, ".config", "rlss", "config.yaml")
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, cfgPath, fmt.Errorf("load config: %w", err)
	}

	// Apply CLI overrides
	if outputDir != "" {
		cfg.OutputDir = config.ExpandPath(outputDir)
	}
	if llmFlag != "" {
		cfg.LLM.Provider.SetPrimary(llmFlag)
	}
	if profileFlag != "" {
		cfg.Chrome.Profile = profileFlag
	}
	if force {
		cfg.Pipeline.SkipExisting = false
	}
	if dryRun {
		cfg.Pipeline.DryRun = true
	}
	if watchFlag {
		cfg.Watch.Enabled = true
	}
	if intervalFlag > 0 {
		cfg.Watch.Interval = intervalFlag
	}
	if allFlag {
		cfg.Safari.Enabled = true
		cfg.Chrome.Enabled = true
	}
	if safariFlag {
		cfg.Safari.Enabled = true
	}
	if chromeFlag {
		cfg.Chrome.Enabled = true
	}
	if unreadFlag {
		cfg.Filter.UnreadOnly = true
	}
	if sinceFlag != "" {
		cfg.Filter.Since = sinceFlag
	}
	if limitFlag > 0 {
		cfg.Filter.Limit = limitFlag
	}

	return cfg, cfgPath, nil
}

func fetchAndFilter(cfg *config.Config, resolver *extract.ProfileResolver) ([]source.ReadingItem, error) {
	var allItems []source.ReadingItem

	if cfg.Safari.Enabled {
		src := source.NewSafariSource(cfg.Safari.PlistPath)
		items, err := src.Fetch()
		if err != nil {
			slog.Warn("Safari fetch failed", "err", err)
		} else {
			slog.Info("Safari", "count", len(items))
			allItems = append(allItems, items...)
		}
	}

	if cfg.Chrome.Enabled {
		profileDir := cfg.Chrome.Profile
		if resolver != nil {
			if resolved, err := resolver.Resolve(cfg.Chrome.Profile); err == nil {
				profileDir = resolved
			}
		}
		src := source.NewChromeSource(profileDir, cfg.Chrome.UserDataDir)
		items, err := src.Fetch()
		if err != nil {
			slog.Warn("Chrome fetch failed", "err", err)
		} else {
			slog.Info("Chrome", "count", len(items))
			allItems = append(allItems, items...)
		}
	}

	// Dedup
	allItems = source.DeduplicateByURL(allItems)

	// Filter
	allItems = applyFilters(allItems, cfg.Filter)

	return allItems, nil
}

func applyFilters(items []source.ReadingItem, f config.FilterConfig) []source.ReadingItem {
	var result []source.ReadingItem

	for _, item := range items {
		if f.UnreadOnly && !item.IsUnread {
			continue
		}
		if f.Since != "" {
			sinceDate, err := time.Parse("2006-01-02", f.Since)
			if err == nil && item.DateAdded.Before(sinceDate) {
				continue
			}
		}
		result = append(result, item)
	}

	if f.Limit > 0 && len(result) > f.Limit {
		result = result[:f.Limit]
	}
	return result
}

// getDefuddleJS returns the embedded Defuddle JavaScript.
// This function should read from the root embed.go's exported variable.
func getDefuddleJS() string {
	// Import from root package or use extract.DefuddleJS()
	return extract.DefuddleJS()
}
```

- [ ] **Step 5: Create list.go**

Create `cmd/rlss/list.go`:

```go
package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/kouko/reading-list-summarize-scraper/internal/extract"
	"github.com/kouko/reading-list-summarize-scraper/internal/output"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List Reading List items without processing",
	RunE: func(cmd *cobra.Command, args []string) error {
		setupLogging(verbose)

		cfg, _, err := loadAndOverrideConfig()
		if err != nil {
			return err
		}

		resolver, _ := extract.NewProfileResolver("")
		items, err := fetchAndFilter(cfg, resolver)
		if err != nil {
			return err
		}

		// Build index for skip info
		idx := output.NewFileIndex()
		idx.Build(cfg.OutputDir)

		fmt.Printf("📖 Reading List (%d items)\n\n", len(items))
		for i, item := range items {
			sha8 := output.SHA8(item.URL)
			status := "⬚"
			if idx.Has(sha8) && idx.Get(sha8).SummaryExists {
				status = "✅"
			}
			unread := ""
			if item.IsUnread {
				unread = " [未讀]"
			}
			fmt.Printf("  %s %d. %s%s\n", status, i+1, item.Title, unread)
			fmt.Printf("     %s · %s · %s\n", output.DomainDir(item.URL), item.DateAdded.Format("2006-01-02"), sha8)
		}

		return nil
	},
}
```

- [ ] **Step 6: Create url.go**

Create `cmd/rlss/url.go`:

```go
package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kouko/reading-list-summarize-scraper/internal/extract"
	"github.com/kouko/reading-list-summarize-scraper/internal/pipeline"
	"github.com/kouko/reading-list-summarize-scraper/internal/source"
	"github.com/kouko/reading-list-summarize-scraper/internal/summarize"
)

var (
	headedFlag        bool
	extractProfileFlag string
)

var urlCmd = &cobra.Command{
	Use:   "url <URL>",
	Short: "Process a single URL",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		setupLogging(verbose)

		cfg, _, err := loadAndOverrideConfig()
		if err != nil {
			return err
		}

		// Apply url-specific overrides
		if headedFlag {
			cfg.Extract.Headless = false
		}
		if extractProfileFlag != "" {
			cfg.Extract.ChromeProfile = extractProfileFlag
		}

		resolver, _ := extract.NewProfileResolver("")
		defuddleJS := extract.DefuddleJS()
		pool := extract.NewPool(&cfg.Extract, resolver, defuddleJS)
		defer pool.CloseAll()

		sum, _ := summarize.NewSummarizer(cfg.LLM)

		p := pipeline.New(cfg, pool, sum)

		src := source.NewManualSource(args[0])
		items, _ := src.Fetch()

		stats := p.ProcessBatch(items)
		pool.CloseAll()

		fmt.Println()
		fmt.Print(stats.Report())
		return nil
	},
}

func init() {
	urlCmd.Flags().BoolVar(&headedFlag, "headed", false, "force headed mode")
	urlCmd.Flags().StringVar(&extractProfileFlag, "extract-profile", "", "Chrome profile for extraction (UI name)")
}
```

- [ ] **Step 7: Create config_cmd.go**

Create `cmd/rlss/config_cmd.go`:

```go
package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configShowCmd = &cobra.Command{
	Use:   "config",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, cfgPath, err := loadAndOverrideConfig()
		if err != nil {
			return err
		}

		fmt.Printf("# Config loaded from: %s\n\n", cfgPath)
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return err
		}
		fmt.Print(string(data))
		return nil
	},
}
```

- [ ] **Step 8: Verify compilation and build**

```bash
go vet ./cmd/rlss/
go build -o rlss ./cmd/rlss/
```

Note: There will likely be import issues and missing exported functions (e.g., `config.ExpandPath`, `extract.DefuddleJS`). Fix these by:
1. Exporting `expandPath` → `ExpandPath` in config.go
2. Ensuring DefuddleJS() is accessible from the extract package (fix embed path or pass it through)
3. Adjusting any missing function signatures

- [ ] **Step 9: Commit**

```bash
git add cmd/rlss/
git commit -m "feat: CLI commands (process, list, url, config) with Cobra"
```

---

## Task 17: Integration — Build + Smoke Test

**Files:**
- Modify: various files for integration fixes

- [ ] **Step 1: Fix all compilation errors**

```bash
go build ./cmd/rlss/ 2>&1
```

Fix any import path issues, missing exports, or type mismatches. Common issues:
- `//go:embed` relative paths from nested packages (may need restructuring)
- Missing `ExpandPath` export in config package
- Summarize package function signatures not matching pipeline expectations

- [ ] **Step 2: Run all tests**

```bash
go test ./... -v
```

Fix any test failures.

- [ ] **Step 3: Full build via Makefile**

```bash
make clean
make build
```

Expected: `rlss` binary created.

- [ ] **Step 4: Smoke test — help**

```bash
./rlss --help
./rlss process --help
./rlss url --help
./rlss list --help
./rlss config
```

- [ ] **Step 5: Smoke test — dry run with Safari**

```bash
./rlss list --safari --verbose
```

Expected: Lists Safari Reading List items (requires Full Disk Access for Terminal).

- [ ] **Step 6: Smoke test — single URL**

```bash
./rlss url "https://go.dev/doc/effective_go" --dry-run --verbose
```

Expected: Shows dry-run info for the URL.

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "fix: integration fixes for full build and smoke tests"
```

---

## Task 18: CI Workflow

**Files:**
- Create: `.github/workflows/build.yml`

- [ ] **Step 1: Create GitHub Actions workflow**

Create `.github/workflows/build.yml`:

```yaml
name: Build rlss

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  workflow_dispatch:

jobs:
  build:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: '22'
          cache: 'npm'

      - name: Install JS dependencies
        run: npm ci

      - name: Bundle defuddle.min.js
        run: npm run build:defuddle

      - name: Check defuddle.min.js freshness
        run: |
          if git diff --name-only | grep -q 'embed/defuddle.min.js'; then
            echo "::warning::embed/defuddle.min.js is outdated. Run 'make update-defuddle' and commit."
          fi

      - uses: actions/setup-go@v5
        with:
          go-version-file: 'go.mod'

      - name: Run tests
        run: go test ./... -v

      - name: Build
        run: go build -ldflags="-s -w" -o rlss ./cmd/rlss/

      - uses: actions/upload-artifact@v4
        with:
          name: rlss-${{ runner.os }}-${{ runner.arch }}
          path: rlss
```

- [ ] **Step 2: Commit**

```bash
git add .github/
git commit -m "ci: GitHub Actions build workflow (Node.js + Go)"
```

---

## Summary

| Task | Component | Key Files |
|------|-----------|-----------|
| 1 | Project scaffold + build pipeline | go.mod, Makefile, package.json, inject.ts, embed/ |
| 2 | Config system | internal/config/ |
| 3 | Source types + Manual | internal/source/types.go, manual.go |
| 4 | Filename + SHA8 + Index | internal/output/filename.go, index.go |
| 5 | Safari source | internal/source/safari.go |
| 6 | Chrome profile resolver | internal/extract/profile.go |
| 7 | Domain rules matching | internal/extract/domain.go |
| 8 | Browser + Defuddle + Pool | internal/extract/browser.go, pool.go |
| 9 | Summarize layer (port ytss) | internal/summarize/ (all files) |
| 10 | Built-in prompt templates | prompts/builtin/ |
| 11 | Obsidian Markdown assembly | internal/output/obsidian.go |
| 12 | copy_to | internal/output/copyto.go |
| 13 | Pipeline runner + stats | internal/pipeline/runner.go, stats.go |
| 14 | Watch mode | internal/pipeline/watch.go |
| 15 | Chrome source | internal/source/chrome.go |
| 16 | CLI commands | cmd/rlss/ |
| 17 | Integration + smoke tests | cross-cutting fixes |
| 18 | CI workflow | .github/workflows/ |
