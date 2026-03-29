# Reading List Summarize Scraper — Design Spec

> **Tool name**: `reading-list-summarize-scraper` (binary: `rlss`)
> **Go module**: `github.com/kouko/reading-list-summarize-scraper`
> **Date**: 2026-03-28 (updated: 2026-03-29)
> **Status**: Implemented

---

## 1. Problem & Goal

macOS 的 Safari 和 Chrome Reading List 中累積大量「待讀」文章，但很少回頭閱讀。本工具自動化整個流程：取得 Reading List URL 清單 → chromedp + Defuddle JS 注入萃取網頁內容 → Agentic CLI 做摘要 → 保存為帶 YAML frontmatter 的 Obsidian Markdown 檔案。

### Design Constraints

| 約束 | 說明 |
|------|------|
| Language | Go — 靜態編譯、單一 Binary |
| Environment | macOS，利用本機 Chrome |
| Extraction | chromedp-undetected + esbuild 打包 Defuddle/full JS 注入（方案 A + 反偵測） |
| Summarization | Agentic CLI fallback chain（claude / gemini / qwen） |
| Output | Obsidian Flavored Markdown + YAML frontmatter |
| Naming | `YYYY-MM-DD__<sha8>__summary.md` / `YYYY-MM-DD__<sha8>__content.md` |

---

## 2. End-to-End Data Flow

```
┌─────────── 輸入層 ───────────┐
│ Safari Bookmarks.plist       │──┐
│ Chrome Sync Data LevelDB     │──┤── 統一 ReadingItem[]
│ 手動 --url                   │──┘
└──────────────────────────────┘
              │
              ▼
┌─────────── 篩選 + 去重 ──────┐
│ --unread / --since / --limit │
│ skip_existing (SHA256[:8])   │
└──────────────────────────────┘
              │
              ▼
┌─────────── 萃取層 ───────────┐
│ chromedp-undetected          │
│ + Defuddle/full JS 注入       │
│ (go:embed defuddle.min.js)   │
│ + anti-bot: headed_on_block  │
│ + min_content_length 檢查     │
└──────────────────────────────┘
              │
              ▼
┌─────────── 摘要層 ───────────┐
│ Agentic CLI fallback chain   │
│ + circuit breaker + cooldown │
│ Stage 1: Summary (blocking)  │
│ Stage 2: Keywords (optional) │
│ Stage 3: Mermaid (optional)  │
└──────────────────────────────┘
              │
              ▼
┌─────────── 輸出層 ───────────┐
│ {domain_dir}/                │
│   YYYY-MM-DD__<sha8>__*.md   │
│ + copy_to 模板複製            │
└──────────────────────────────┘
```

---

## 3. Chrome Reading List: LevelDB Direct Read

Chrome Reading List 透過**直接讀取 Sync Data LevelDB** 取得，不需要啟動 Chrome。

```
Chrome Profile/Sync Data/LevelDB/
  → 複製到 /tmp（跳過 LOCK 檔案，~1-2MB）
    → goleveldb ReadOnly 開啟
      → 掃描 key prefix "reading_list-dt-"
        → protowire 解碼 ReadingListSpecifics
          → URL, Title, Status, Timestamps
```

**優勢**（vs 原 Extension + CDP 方案）：
- 不需要啟動 Chrome（<1s vs ~15s）
- 無 SingletonLock 問題
- 無 Extension 載入問題
- 不需要 Google Account session

詳細技術筆記：[[2026-03-29 Chrome Reading List LevelDB 直讀技術實作筆記]]

### Chrome Instance Management (Extraction Layer)

萃取層仍使用 chromedp-undetected Lazy Pool（與 Reading List 取得無關）：

```
URL → 匹配 domain_rules → 得到 (headed, profile)
  → Pool 有此組合？
    ├─ 有 → 複用
    └─ 無 → 啟動新 Chrome 實例，加入 Pool
  → 結束時 CloseAll()
```

不匹配任何 rule → 使用預設 `(headless=true, extract.chrome_profile)`。

### Chrome Profile Resolution

支援兩種指定方式：

1. **Google Account email**（`chrome.google_account`）：掃描 `Local State` 的 `profile.info_cache` 中 `user_name` 欄位，找到匹配 email 的 profile
2. **Profile 名稱**（`chrome.profile`）：UI 顯示名稱或內部 folder name

解析優先級：
- email + profile 都有 → 兩者須匹配
- email 只有 → 自動找匹配的 profile
- profile 只有 → 按名稱查找
- 都沒有 → 使用 Default

---

## 4. Module Architecture

```
reading-list-summarize-scraper/
├── cmd/
│   └── rlss/
│       └── main.go                    ← CLI 入口（cobra）
├── internal/
│   ├── source/                        ← 輸入層
│   │   ├── types.go                   ← ReadingItem + Source interface + DeduplicateByURL
│   │   ├── safari.go                  ← Safari plist 解析（howett.net/plist）
│   │   ├── chrome_leveldb.go          ← Chrome Reading List（Sync Data LevelDB 直讀）
│   │   └── manual.go                  ← --url 手動輸入
│   │
│   ├── extract/                       ← 萃取層
│   │   ├── pool.go                    ← Chrome Lazy Pool (by headed+profile)
│   │   ├── browser.go                 ← 單一 chromedp 實例生命週期
│   │   ├── defuddle.go                ← JS 注入 + 結果回收
│   │   ├── domain.go                  ← domain_rules 匹配
│   │   └── profile.go                 ← Chrome profile UI name → folder
│   │
│   ├── summarize/                     ← 摘要層（與 ytss 相同架構）
│   │   ├── summarizer.go             ← interface + factory
│   │   ├── fallback.go                ← FallbackSummarizer
│   │   ├── circuit_breaker.go         ← CircuitBreaker 狀態機
│   │   ├── errors.go                  ← QuotaError + quotaPatterns
│   │   ├── prompt.go                  ← Prompt 解析 + 變數替換
│   │   ├── claude_code.go
│   │   ├── claude.go
│   │   ├── gemini.go
│   │   ├── qwen_code.go
│   │   ├── ollama.go
│   │   ├── llamacpp.go
│   │   └── openai_compat.go
│   │
│   ├── output/                        ← 輸出層
│   │   ├── obsidian.go                ← Frontmatter + Markdown 組裝 + Mermaid 插入
│   │   ├── filename.go                ← SHA8 + DomainDir + YYYY-MM-DD__<sha8>__type.md
│   │   ├── index.go                   ← 已處理 URL index（by SHA8）
│   │   └── copyto.go                  ← copy_to 模板展開 + 檔案複製
│   │
│   ├── pipeline/                      ← 流水線編排
│   │   ├── runner.go                  ← ProcessBatch + ProcessItem
│   │   ├── watch.go                   ← Watch 模式
│   │   └── stats.go                   ← 統計
│   │
│   └── config/                        ← 設定
│       ├── config.go                  ← 結構定義 + 載入
│       └── defaults.go                ← 預設值
│
├── embed/                             ← 嵌入資源
│   ├── embed.go                       ← //go:embed DefuddleJS
│   └── defuddle.min.js                ← esbuild IIFE（defuddle/full + turndown）
│
├── prompts/                           ← 內建 Prompt 模板
│   └── builtin/
│       ├── summary-en.md
│       ├── summary-ja.md
│       ├── summary-zh-Hant.md
│       ├── keywords-en.md
│       ├── keywords-ja.md
│       ├── keywords-zh-Hant.md
│       ├── mermaid-en.md
│       ├── mermaid-ja.md
│       └── mermaid-zh-Hant.md
│
├── inject.ts                          ← Defuddle 前端膠水碼
├── package.json                       ← esbuild + defuddle（build 用）
├── Makefile
├── config.example.yaml                ← 範例設定檔（含完整註解）
├── go.mod
└── go.sum
```

### Go Dependencies

| Package | Purpose |
|---------|---------|
| `howett.net/plist` | Safari Bookmarks.plist 解析 |
| `github.com/syndtr/goleveldb` | Chrome Sync Data LevelDB 讀取 |
| `google.golang.org/protobuf` | ReadingListSpecifics protowire 解碼 |
| `github.com/chromedp/chromedp` | Chrome CDP 控制（萃取層） |
| `github.com/chromedp/cdproto` | CDP 協定（runtime） |
| `github.com/Davincible/chromedp-undetected` | 反偵測 Chrome 啟動（萃取層） |
| `github.com/spf13/cobra` | CLI 框架 |
| `gopkg.in/yaml.v3` | Config YAML |
| `embed` | 嵌入 Defuddle JS（stdlib） |
| `os/exec` | 呼叫 Agentic CLI（stdlib） |
| `crypto/sha256` | URL hash（stdlib） |

---

## 5. Configuration

### Full config.yaml Structure

```yaml
# ./config.yaml (預設讀取 CWD 下的 config.yaml，與 ytss 一致)

# 輸出目錄
output_dir: ~/kouko-obsidian-vault/references/

# LLM 設定（與 ytss 完全相同結構）
llm:
  provider: "claude-code"              # 單一字串或 list ["claude-code", "gemini-cli"]
  provider_fallback_strategy:
    cooldown_seconds: 300
    failure_threshold: 1
  ollama:
    model: "llama3"
    endpoint: "http://localhost:11434"
    think: false
    timeout: 900
  llamacpp:
    endpoint: "http://localhost:8080"
  claude-api:
    api_key: "${CLAUDE_API_KEY}"
    model: "claude-sonnet-4-20250514"
  claude-code:
    model: "haiku"
    path: ""
    timeout: 900
  gemini-cli:
    model: "auto"
    path: ""
    timeout: 900
  qwen-code:
    model: "coder-model"
    path: ""
    timeout: 900
  openai-compat:
    endpoint: "http://127.0.0.1:8000/v1"
    model: ""
    api_key: ""
    timeout: 900

# 摘要設定（與 ytss 完全相同結構）
summary:
  language: "zh-Hant"                  # 內建 prompt 語言（en / zh-Hant / ja）
  prompt: ""                           # inline prompt（覆蓋內建）
  summary_prompt_file: ""              # 外部 prompt 檔案（最高優先）
  max_tokens: 10000
  keywords:
    enabled: true
    language: "en"
    count: 5
  mermaid:
    enabled: true

# Safari 設定
safari:
  enabled: true
  plist_path: ""                       # 空 = ~/Library/Safari/Bookmarks.plist

# Chrome Reading List 設定
chrome:
  enabled: true
  google_account: "user@gmail.com"     # Google Account email（自動找對應 profile）
  profile: ""                          # UI 顯示名稱（搭配 google_account 或單獨使用）
  user_data_dir: ""                    # 空 = 自動偵測系統 Chrome 路徑
  clone_profile: true                  # SingletonLock 時複製 profile（僅萃取層備用）

# 萃取設定
extract:
  headless: true                       # 預設 headless
  chrome_profile: "Default"            # 預設萃取用 profile（UI 名稱）
  google_account: ""                   # 萃取用 Google Account（per-domain 可覆蓋）
  user_data_dir: ""
  timeout: 30s
  headed_timeout: 60s                  # headed 模式超時（含等待反爬蟲自動通過）
  wait_after_load: 2s
  min_content_length: 100              # 內容低於此字數時跳過摘要（避免浪費 LLM token）
  headed_on_block: true                # 偵測到 Cloudflare 阻擋時自動切換 headed 模式重試
  domain_rules:
    - domains: ["medium.com"]
      headed: true
      google_account: "work@company.com"
      chrome_profile: ""
    - domains: ["*.substack.com"]
      headed: true
      chrome_profile: "Default"
    - domains: ["github.com"]
      headed: false
      chrome_profile: "Profile 2"

# 流水線設定
pipeline:
  skip_existing: true
  dry_run: false
  delay_min: 3
  delay_max: 8

# 篩選設定
filter:
  unread_only: false
  since: ""
  limit: 0

# Watch 模式
watch:
  enabled: false
  interval: 10

# copy_to 設定
copy_to:
  enabled: false
  path: "{output_dir}/by-source/{source}/{domain_dir}"
  filename: "{date}__{title}__{type}.md"
  files: ["summary", "content"]
  overwrite: false

# Obsidian 整合
obsidian:
  auto_tags: true
  wikilinks: true
```

### CLI Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--config` | | `./config.yaml` | Config path (CWD relative, like ytss) |
| `--safari` | `-s` | false | Safari Reading List |
| `--chrome` | `-c` | false | Chrome Reading List |
| `--all` | `-a` | false | All sources |
| `--unread` | | false | Unread only |
| `--since` | | "" | Date filter (YYYY-MM-DD) |
| `--limit` | `-n` | 0 | Max items |
| `--llm` | | config | Override primary provider |
| `--output` | `-o` | config | Output directory |
| `--profile` | | config | Chrome RL profile (UI name) |
| `--dry-run` | | false | List only |
| `--force` | | false | Reprocess existing |
| `--watch` | `-w` | false | Watch mode |
| `--interval` | | config | Watch interval (minutes) |
| `--verbose` | `-v` | false | Debug logging |

`rlss url` additional flags:
- `--headed`: Force headed mode for this URL
- `--extract-profile`: Override extract chrome_profile for this URL (distinct from `--profile` which sets Chrome RL fetch profile)

---

## 6. Input Layer

### Unified ReadingItem

```go
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
```

### Safari Source

- Parse `~/Library/Safari/Bookmarks.plist` using `howett.net/plist`
- Find `com.apple.ReadingList` folder in top-level Children
- Extract `URLString`, `URIDictionary["title"]` (fallback to `Title`), `ReadingList.DateAdded`, `ReadingList.DateLastViewed` (zero = unread)
- Requires Full Disk Access; on permission error:
  - Displays prominent ASCII banner with EN/JA/ZH-TW instructions
  - OSC 8 clickable hyperlink for supported terminals (iTerm2, Warp)
  - Automatically opens System Settings to Full Disk Access pane

### Chrome Source (LevelDB Direct Read)

Reading List 透過直接讀取 Chrome Sync Data LevelDB 取得，不啟動 Chrome：

1. 解析 profile（by `google_account` email 或 `profile` name）
2. 定位 `{userDataDir}/{profile}/Sync Data/LevelDB/`
3. 複製 LevelDB 檔案到 `/tmp/rlss-leveldb-copy/`（跳過 LOCK，~1-2MB）
4. `goleveldb` ReadOnly 開啟
5. 掃描 key prefix `reading_list-dt-` → URL 在 key 中
6. `protowire` 解碼 value（ReadingListSpecifics protobuf）→ title, status, timestamps
7. 回傳 `[]ReadingItem`

不需要：Chrome 執行、Extension、Google Account session、SingletonLock 處理

Chrome Extension source（`chrome.go`）已移除。LevelDB 直讀為唯一的 Chrome Reading List 取得方式。

### Manual Source (`rlss url`)

- Single URL → `ReadingItem{URL: url, Source: "manual", DateAdded: time.Now()}`
- Title resolved during extraction (from Defuddle result or `<title>` tag)

---

## 7. Extraction Layer

### Anti-Bot Stealth (chromedp-undetected)

Browser instances use `github.com/Davincible/chromedp-undetected` instead of raw chromedp:
- Hides `navigator.webdriver` property
- Realistic browser fingerprint
- `--disable-blink-features=AutomationControlled` flag
- On macOS, falls back to standard headless flag (Xvfb is Linux-only)

### Anti-Bot Retry: `headed_on_block`

When extraction detects Cloudflare/CAPTCHA content (pattern matching with ≥2 hits to avoid false positives), and `extract.headed_on_block: true`:
1. Automatically retries with headed Chrome (non-headless) using `headed_timeout` (60s)
2. Profile with existing cookies/history may auto-pass Cloudflare trust check
3. If still blocked after headed retry, records failure and continues

### Content Validation

After extraction, content is validated:
- **Empty content** → error, skip item
- **Anti-bot page detected** → trigger headed retry (if enabled) or error
- **Content below `min_content_length`** (default 100 chars) → save content.md but skip summarization (avoids wasting LLM tokens on login pages, error pages)

### Defuddle JS Injection (Option A)

Build pipeline:

```
inject.ts → esbuild (IIFE, minify) → defuddle.min.js → go:embed
```

**inject.ts** (uses `defuddle/full` for bundled Turndown markdown converter):

```typescript
import Defuddle from 'defuddle/full';

(window as any).extractArticle = async (): Promise<string> => {
    try {
        const df = new Defuddle(document, { markdown: true });
        const result = df.parse();
        return result?.content ?? "";
    } catch (e: any) {
        console.error("defuddle extraction error:", e);
        return "";
    }
};
```

Key implementation details:
- `defuddle/full` (not `defuddle`) — includes Turndown for HTML-to-Markdown conversion
- `{ markdown: true }` — outputs Markdown in `content` field (without this, content is HTML)
- `new Defuddle(document)` — takes DOM `document` object, not HTML string
- `parse()` is synchronous (not `parseAsync()`)

**chromedp extraction sequence**:

```go
chromedp.Navigate(url)
chromedp.WaitVisible(`body`, chromedp.ByQuery)
chromedp.Sleep(waitAfterLoad)
chromedp.Evaluate(defuddleJS, nil)             // inject
chromedp.Evaluate(`window.extractArticle()`,   // execute
    &content, chromedp.AwaitPromise)
```

### Lazy Pool

```go
type Pool struct {
    mu        sync.Mutex
    instances map[poolKey]*Browser  // poolKey = {headed bool, profile string}
    resolver  *ProfileResolver
    config    *config.ExtractConfig
}

func (p *Pool) GetBrowser(headed bool, profile string) (*Browser, error)
func (p *Pool) CloseAll()
```

### Domain Rules Matching

```go
func (p *Pool) ResolveForURL(rawURL string) (headed bool, profile string)
```

- Extract hostname from URL
- Match against `domain_rules[].domains` (exact match, wildcard `*.xxx.com`)
- `example.com` matches `example.com` and `*.example.com`
- No match → use defaults (`extract.headless`, `extract.chrome_profile`)

---

## 8. Summarization Layer

Identical architecture to youtube-summarize-scraper.

### Three-Stage Pipeline

| Stage | Blocking | Input | Output |
|-------|----------|-------|--------|
| 1. Main Summary | YES | Extracted content + prompt | Summary text |
| 2. Keywords | NO | Summary text | `[]string` tags |
| 3. Mermaid | NO | Summary text | `[]MermaidBlock{Title, Code}` |

### Prompt Resolution (4-level cascade)

1. Per-source prompt file (future: via domain_rules)
2. Global prompt file (`summary.summary_prompt_file`)
3. Inline prompt (`summary.prompt`)
4. Built-in by language (`summary.language` → `prompts/builtin/summary-{lang}.md`)

### Prompt Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `{{title}}` | Page title | "How to Build a CLI Tool in Go" |
| `{{domain}}` | Source domain | "example.com" |
| `{{date_added}}` | Date added to Reading List | "2026-03-25" |
| `{{source}}` | Source type | "safari" |
| `{{content_length}}` | Extracted content char count | 3200 |
| `{{content_tier}}` | Length tier | "中篇" |
| `{{content}}` | Extracted Markdown content | (full text) |

Content tier thresholds (CJK: zh-Hant, ja):

| Tier | CJK | English |
|------|-----|---------|
| Short | < 500 | < 1,000 |
| Medium | 500-3,000 | 1,000-5,000 |
| Long | 3,000-10,000 | 5,000-15,000 |
| Very Long | > 10,000 | > 15,000 |

### LLM Provider System (identical to ytss)

- **ProviderList**: YAML accepts scalar or list; `Primary()` / `Fallbacks()` methods
- **FallbackSummarizer**: Try providers in order; QuotaError triggers circuit breaker
- **CircuitBreaker**: Closed → Open → HalfOpen state machine with cooldown
- **QuotaError detection**: Pattern matching (`resource_exhausted`, `429`, `rate limit`, etc.)
- **Non-quota errors**: Try next provider without opening circuit
- **Per-provider configs**: Identical struct definitions to ytss (OllamaConfig, ClaudeCodeConfig, GeminiCLIConfig, QwenCodeConfig, ClaudeAPIConfig, LlamaCppConfig, OpenAICompatConfig)

### Summarizer Interface (identical to ytss)

```go
type Summarizer interface {
    Summarize(text string, opts SummarizeOptions) (SummarizeResult, error)
}

type SummarizeOptions struct {
    Prompt    string
    MaxTokens int
    Model     string
}

type SummarizeResult struct {
    Text     string
    Provider string
    Model    string
}
```

---

## 9. Output Layer

### File Structure

```
{output_dir}/
└── {domain_dir}/                              ← domain with . → _
    ├── YYYY-MM-DD__<sha8>__summary.md
    └── YYYY-MM-DD__<sha8>__content.md
```

Examples:

```
references/
├── example_com/
│   ├── 2026-03-28__a1b2c3d4__summary.md
│   └── 2026-03-28__a1b2c3d4__content.md
├── blog_example_com/
│   └── 2026-03-28__b2c3d4e5__summary.md
├── medium_com/
│   └── 2026-03-28__c3d4e5f6__summary.md
└── zenn_dev/
    └── 2026-03-28__d4e5f6g7__summary.md
```

### Summary File Frontmatter

```yaml
---
title: "How to Build a CLI Tool in Go"
type: reading-list-summary
date: 2026-03-28
url: "https://example.com/go-cli-guide"
domain: "example.com"
source: "safari"
date_added: 2026-03-25
llm_provider: "claude-code"
llm_model: "haiku"
content_length: 3200
content_tier: "中篇"
tags:
  - reading-list
  - auto-summary
  - golang
  - cli-tools
---
```

### Summary File Body

```markdown
> [!info] 來源資訊
> - **原始網址**：[example.com](https://example.com/go-cli-guide)
> - **加入日期**：2026-03-25
> - **來源**：Safari Reading List
> - **摘要工具**：claude-code (haiku)
> - **處理日期**：2026-03-28

---

### 概述

(Stage 1 summary text)

### 章節摘要

(with Mermaid diagrams inserted after matching section headings)

### 重點整理

(key takeaways)
```

No `# Title` heading (frontmatter already has title). No inline original text excerpt (separate content file).

### Content File

```yaml
---
title: "How to Build a CLI Tool in Go"
type: reading-list-content
date: 2026-03-28
url: "https://example.com/go-cli-guide"
domain: "example.com"
source: "safari"
date_added: 2026-03-25
content_length: 3200
extracted_by: "defuddle-js"
---

(Full Defuddle-extracted Markdown content)
```

### Index (Skip Detection)

```go
type FileIndex struct {
    entries map[string]fileInfo  // key = sha8
}

type fileInfo struct {
    Dir            string
    SummaryExists  bool
    ContentExists  bool
}
```

- Built at startup by scanning `output_dir/**/*__summary.md` and `*__content.md`
- Extract SHA8 from filename pattern `YYYY-MM-DD__<sha8>__type.md`
- O(1) lookup per URL
- `RebuildIndex()` called each watch iteration

### Resume Support

```
content exists + summary missing → read content, skip extraction, run summarization
both exist                      → skip (unless --force)
neither exists                  → full pipeline
```

### copy_to

```yaml
copy_to:
  enabled: false
  path: "{output_dir}/by-source/{source}/{domain_dir}"
  filename: "{date}__{title}__{type}.md"
  files: ["summary", "content"]
  overwrite: false
```

Template variables:

| Variable | Example | Description |
|----------|---------|-------------|
| `{output_dir}` | `~/vault/references` | Configured output_dir |
| `{date}` | `2026-03-28` | Processing date |
| `{date_added}` | `2026-03-25` | Date added to RL |
| `{title}` | `How to Build a CLI Tool in Go` | Page title (display-friendly sanitized) |
| `{sha8}` | `a1b2c3d4` | URL SHA256[:8] |
| `{source}` | `safari` | Source type |
| `{domain}` | `example.com` | Raw domain |
| `{domain_dir}` | `example_com` | Domain with . → _ |
| `{type}` | `summary` / `content` | File type |

Path length handling: progressive shortening (title 80→40→20) when path exceeds 255 bytes.

---

## 10. Pipeline

### ProcessItem (Single URL)

```
Step 1:   sha8 = sha256(url)[:8]
Step 2:   index.Has(sha8) && !force → skip
Step 3:   dry_run → print info, return
Step 4:   Resume detection (content exists, summary missing → skip to Step 9)
Step 5:   Domain rule matching → (headed, profile)
Step 6:   pool.ExtractURL() → Navigate → Defuddle JS inject → Markdown
Step 6a:  Empty content check → error
Step 6b:  Anti-bot detection → headed_on_block? → pool.ExtractURLHeaded() retry
Step 7:   Create output dir ({output_dir}/{domain_dir}/)
Step 8:   Write content file
Step 8a:  min_content_length check → skip summarization if too short
Step 9:   Summarization (Stage 1 blocking, Stage 2+3 non-blocking)
Step 10:  Assemble + write summary file
Step 11:  ExecuteCopyTo()
```

### ProcessBatch

```
1. Fetch sources
   ├─ --safari → SafariSource.Fetch() (plist, <100ms)
   ├─ --chrome → ChromeLevelDBSource.Fetch() (LevelDB 直讀，<1s)
   └─ Merge + URL dedup
2. Filter (--unread / --since / --limit)
3. Build/rebuild Index
4. Process items in Reading List order
   for item in items:
     if p.stopped() → break
     ProcessItem(item)
     random delay (delay_min ~ delay_max)
     update stats
5. pool.CloseAll()
```

### Graceful Shutdown

- SIGINT/SIGTERM during processing → `p.cancel()` → finish current URL, stop loop
- SIGINT/SIGTERM during sleep → immediate exit
- Chrome Pool guaranteed `CloseAll()` on exit

### Error Handling

| Stage | On Failure | Logging |
|-------|-----------|---------|
| Safari plist | ASCII banner (EN/JA/ZH) + auto-open System Settings + skip Safari | stderr banner |
| Chrome LevelDB | Log error, skip Chrome source | slog.Error |
| Extraction (empty) | Record error, skip item | stats.Errors + slog.Error |
| Extraction (blocked) | Auto-retry with headed mode if headed_on_block, else skip | slog.Warn → slog.Error |
| Content too short | Save content.md, skip summarization | slog.Warn |
| Summary Stage 1 | Fallback chain exhausted → skip (content preserved) | stats.Errors + slog.Error |
| Keywords Stage 2 | Warn, continue without tags | slog.Warn |
| Mermaid Stage 3 | Warn, continue without diagrams | slog.Warn |
| File write | Immediate error | slog.Error + abort |
| copy_to | Warn, continue | slog.Warn |

Sentinel error pattern (identical to ytss):

```go
var errSkipped = fmt.Errorf("skipped")
func IsSkipped(err error) bool { return err == errSkipped }
```

### Stats

```go
type Stats struct {
    Success int
    Skipped int
    Failed  int
    Errors  []ItemError  // {URL, Title, Err}
    Start   time.Time
    End     time.Time
}
```

---

## 11. Watch Mode

```yaml
watch:
  enabled: false
  interval: 10    # minutes
```

### Watch Loop

```
Loop (until SIGINT/SIGTERM):
  1. ResetContext()    ← fresh context for graceful shutdown
  2. ReloadConfig()   ← partial reload (filter, watch, llm, summary)
  3. RebuildIndex()   ← rescan processed files
  4. ProcessBatch()   ← in goroutine
  5. Print iteration stats
  6. Sleep interval minutes
```

### Config Partial Reload (per watch iteration)

Reloaded: `filter`, `watch.interval`, `llm`, `summary`

Not reloaded (requires restart): `output_dir`, `safari`, `chrome`, `extract`

---

## 12. CLI Interface

### Commands

```
rlss [flags]                   ← default: process
rlss process [flags]           ← batch process Reading Lists
rlss list [flags]              ← list items only (= --dry-run)
rlss url <URL> [flags]         ← process single URL
rlss config                    ← show current config
```

### Terminal Output

```
$ rlss --all --unread --llm claude-code

📚 Reading List Summarize Scraper v0.1.0
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📖 來源掃描
  Safari Reading List: 8 篇未讀
  Chrome Reading List: 5 篇未讀
  去重後: 12 篇（已處理 3 篇，待處理 9 篇）

🔄 處理進度 [3/9]
  ✅ How to Build a CLI Tool in Go
     → example_com/2026-03-28__a1b2c3d4__summary.md
  ✅ 日本經濟展望 2026Q2
     → nikkei_com/2026-03-28__b2c3d4e5__summary.md
  🔄 Understanding WebAssembly...
     萃取中... (chromedp headless)
  ⏳ The Future of AI Agents
  ... (5 more)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 處理報告
  ✅ 成功: 8
  ⏭️ 跳過: 3 (已處理)
  ❌ 失敗: 1 (Cloudflare 阻擋)
  ⏱️ 耗時: 6m 12s
  📁 輸出: ~/kouko-obsidian-vault/references/
```

---

## 13. Build Pipeline

### Overview

```
npm install (defuddle + esbuild)
  → esbuild: inject.ts → embed/defuddle.min.js (IIFE, minified)
    → go:embed embeds defuddle.min.js into Go binary
      → go build → rlss binary (~15MB)
```

`embed/defuddle.min.js` is **committed to git** for convenience (users can `go build` without Node.js). However, CI and `make build` always regenerate it from the latest Defuddle source to ensure freshness.

### Project Files

```
reading-list-summarize-scraper/
├── inject.ts              ← Defuddle 前端膠水碼（import + window.extractArticle）
├── package.json           ← defuddle + esbuild as devDependencies
├── package-lock.json      ← 鎖定版本（committed）
├── embed/
│   └── defuddle.min.js    ← build artifact（committed for go-build-only users）
├── embed.go               ← //go:generate + //go:embed 指令
└── Makefile               ← 自動化入口
```

### Step 1: package.json

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

### Step 2: inject.ts

```typescript
import Defuddle from 'defuddle/full';

(window as any).extractArticle = async (): Promise<string> => {
    try {
        const df = new Defuddle(document, { markdown: true });
        const result = df.parse();
        return result?.content ?? "";
    } catch (e: any) {
        console.error("defuddle extraction error:", e);
        return "";
    }
};
```

Note: Uses `defuddle/full` (not `defuddle`) to include Turndown HTML-to-Markdown converter (~568KB bundle vs ~168KB).

### Step 3: embed package

`embed/embed.go` provides `//go:embed` for defuddle.min.js.
Go's `//go:embed` only allows relative paths, so the embed package lives alongside the assets.

### Step 4: Makefile

```makefile
.PHONY: build js js-quick clean generate

# 完整 build：強制重裝依賴 → 打包 JS → 編譯 Go
build: js
	go build -ldflags="-s -w" -o rlss ./cmd/rlss/

# 強制重新安裝 defuddle + 打包（預設）
js:
	rm -rf node_modules
	npm install
	npm run build:defuddle
	@echo "Bundled defuddle $$(npm list defuddle --depth=0 | grep defuddle)"

# 快速打包（不重裝，用於開發）
js-quick: node_modules
	npm run build:defuddle

node_modules: package.json package-lock.json
	npm install
	@touch node_modules

# go generate
generate: node_modules
	go generate ./...

# 清理
clean:
	rm -f rlss embed/defuddle.min.js
	rm -rf node_modules
```

### CI/CD: GitHub Actions

```yaml
name: Build rlss

on:
  push:
    branches: [main]
  workflow_dispatch:

jobs:
  build:
    runs-on: macos-latest              # macOS（本工具主要平台）
    steps:
      - uses: actions/checkout@v4

      # 1. Node.js 環境（esbuild + defuddle 打包用）
      - uses: actions/setup-node@v4
        with:
          node-version: '22'
          cache: 'npm'

      # 2. 安裝 JS 依賴 + 打包 defuddle.min.js
      - name: Install JS dependencies
        run: npm ci

      - name: Bundle defuddle.min.js
        run: npm run build:defuddle

      # 3. 檢查打包結果是否與 committed 版本不同（提醒更新）
      - name: Check defuddle.min.js freshness
        run: |
          if git diff --name-only | grep -q 'embed/defuddle.min.js'; then
            echo "::warning::embed/defuddle.min.js is outdated. Please run 'make update-defuddle' and commit."
          fi

      # 4. Go 環境
      - uses: actions/setup-go@v5
        with:
          go-version-file: 'go.mod'

      # 5. 編譯 Go Binary（defuddle.min.js 自動被 go:embed 嵌入）
      - name: Build Go binary
        run: go build -ldflags="-s -w" -o rlss ./cmd/rlss/

      # 6. 上傳產出物
      - uses: actions/upload-artifact@v4
        with:
          name: rlss-${{ runner.os }}-${{ runner.arch }}
          path: rlss
```

### Build Scenarios

| Scenario | Command | Node.js Required | Result |
|----------|---------|------------------|--------|
| Quick build（use committed JS） | `go build ./cmd/rlss/` | No | Uses existing `embed/defuddle.min.js` |
| Full build（force reinstall + regenerate JS） | `make` / `make build` | Yes | rm node_modules → npm install → esbuild → go build |
| Dev build（skip reinstall） | `make js-quick && go build ./cmd/rlss/` | Yes | Use existing node_modules, rebundle only |
| CI build | GitHub Actions | Yes (in CI) | Always regenerate from latest locked version |
| go generate | `go generate ./...` | Yes | Triggers `npm run build:defuddle` |

### Defuddle Version Management

- `package.json` 中 `defuddle: "latest"` 確保 `npm update` 取得最新版
- `package-lock.json` committed，確保可重現 build
- `make update-defuddle` = 更新 + 重新打包（開發者手動觸發）
- CI 使用 `npm ci`（from lock file），確保一致性
- CI 額外檢查 committed 的 `defuddle.min.js` 是否過時，發出 warning

---

## 14. Reference Research Documents

- [[2026-03-28 Reading List 自動摘要工具規劃 — Go chromedp + Defuddle + Agentic CLI]] — Primary planning document
- [[2026-03-28 macOS Reading List 自動化技術研究 — Safari 與 Chrome 的 Go 實作方案]] — Safari plist + Chrome Extension technical details
- [[2026-03-28 Go 網頁內容萃取與 Agentic CLI 摘要自動化研究]] — Extraction + CLI summarization architecture
- [[2026-03-29 Chrome Reading List LevelDB 直讀技術實作筆記]] — Chrome LevelDB 直讀方案的技術實作與業界參考
- youtube-summarize-scraper — Reference implementation for LLM provider system, fallback chain, config structure, pipeline patterns, watch mode, copy_to, index/skip detection
