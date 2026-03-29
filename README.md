# Reading List Summarize Scraper (rlss)

A Go CLI tool that fetches URLs from macOS Safari/Chrome Reading Lists, extracts web content via chromedp + Defuddle, summarizes with Agentic CLI (Claude/Gemini/Qwen), and outputs Obsidian Markdown files.

```
Safari Bookmarks.plist ──┐
Chrome Sync Data LevelDB ┤── Extract (chromedp + Defuddle) → Summarize (LLM) → Obsidian .md
Manual --url ────────────┘
```

## Features

- **Safari Reading List**: Direct plist parsing, Full Disk Access guidance
- **Chrome Reading List**: LevelDB direct read — no Chrome launch needed, <1s
- **Content Extraction**: chromedp-undetected (anti-bot stealth) + Defuddle Markdown output
- **LLM Summarization**: 7 providers with fallback chain + circuit breaker
- **Obsidian Output**: YAML frontmatter, keywords, Mermaid diagrams
- **Watch Mode**: Periodic re-scan for new items

## Requirements

- macOS
- Google Chrome (for content extraction)
- One of: `claude`, `gemini`, `qwen`, or local LLM (for summarization)
- Go 1.23+ (for building from source)
- Node.js 22+ (for rebuilding Defuddle JS bundle)

## Installation

### From GitHub Release

Download the latest binary from [Releases](https://github.com/kouko/reading-list-summarize-scraper/releases).

### Build from source

```bash
git clone https://github.com/kouko/reading-list-summarize-scraper.git
cd reading-list-summarize-scraper

# Full build (reinstalls npm deps + rebuilds Defuddle JS + compiles Go)
make

# Or quick build (uses pre-committed defuddle.min.js, no Node.js needed)
go build -o rlss ./cmd/rlss/
```

## Quick Start

```bash
# 1. Copy example config
cp config.example.yaml config.yaml

# 2. Edit config.yaml — set output_dir and LLM provider

# 3. List Safari Reading List items (dry run)
./rlss --safari --dry-run

# 4. Process the 3 most recent items
./rlss --safari --limit 3

# 5. Process a single URL
./rlss url "https://example.com/article"

# 6. Extract only (skip summarization)
./rlss --safari --limit 3 --no-summary
```

## Usage

```
rlss [flags]
rlss process [flags]       # Batch process Reading Lists (default command)
rlss list [flags]          # List items without processing
rlss url <URL> [flags]     # Process a single URL
rlss config                # Show current configuration
```

### Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--safari` | `-s` | Process Safari Reading List |
| `--chrome` | `-c` | Process Chrome Reading List |
| `--all` | `-a` | Process all sources |
| `--unread-only` | | Only unread items |
| `--since` | | Items added after date (YYYY-MM-DD) |
| `--limit` | `-n` | Max items to process |
| `--provider` | | Override LLM provider |
| `--output` | `-o` | Override output directory |
| `--force` | | Reprocess existing items |
| `--dry-run` | | List items without processing |
| `--no-summary` | | Extract only, skip summarization |
| `--watch` | | Enable watch mode |
| `--verbose` | `-v` | Debug logging |
| `--config` | | Config file path (default: ./config.yaml) |

### `rlss url` additional flags

| Flag | Description |
|------|-------------|
| `--headed` | Force headed (non-headless) Chrome mode |
| `--extract-profile` | Chrome profile for extraction |

## Configuration

Copy `config.example.yaml` to `config.yaml` and edit. See the example file for detailed comments on every option.

### Key settings

```yaml
output_dir: ./rlss-output

llm:
  provider: "claude-code"              # or ["claude-code", "gemini-cli"]

safari:
  enabled: true

chrome:
  enabled: true
  google_account: "you@gmail.com"      # Auto-find matching Chrome profile

extract:
  headed_on_block: true                # Auto-retry with headed mode on Cloudflare
  domain_rules:
    - domains: ["medium.com"]
      headed: true
      chrome_profile: "Work"

summary:
  language: "zh-Hant"                  # en, zh-Hant, ja
  keywords:
    enabled: true
  mermaid:
    enabled: true
```

## Output Structure

```
{output_dir}/
└── {domain_dir}/                              # domain with . → _
    ├── YYYY-MM-DD__{sha8}__summary.md         # LLM summary with frontmatter
    └── YYYY-MM-DD__{sha8}__content.md         # Extracted Markdown content
```

### Summary frontmatter

```yaml
---
title: "Article Title"
type: reading-list-summary
date: 2026-03-29
url: "https://example.com/article"
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
  - keyword1
---
```

## macOS Permissions

### Safari Reading List

Requires **Full Disk Access** for the terminal app. On permission error, rlss automatically opens System Settings to the correct pane.

### Chrome Reading List

Reads directly from Chrome's Sync Data LevelDB files. **No special permissions needed** — works while Chrome is running.

## Build

| Command | Description |
|---------|-------------|
| `make` | Full build: reinstall npm deps → rebuild Defuddle JS → compile Go |
| `make js-quick` | Rebuild JS only (skip npm reinstall) |
| `go build -o rlss ./cmd/rlss/` | Quick Go build using committed defuddle.min.js |
| `make clean` | Remove binary, JS bundle, node_modules |
| `go test ./...` | Run all tests (108 tests) |

## Release

Releases are automated via GitHub Actions. To create a new release:

```bash
# 1. Ensure main branch is up to date
git checkout main
git pull

# 2. Create a version tag (semantic versioning)
git tag v0.1.0

# 3. Push the tag — triggers the Release workflow
git push origin v0.1.0
```

The Release workflow will:
1. Run all tests
2. Build the binary with the version embedded (`./rlss --version` → `v0.1.0`)
3. Create a GitHub Release with the binary attached
4. Auto-generate release notes from commits

### Versioning

| Tag format | Release type |
|------------|-------------|
| `v1.0.0` | Stable release |
| `v1.0.0-beta` | Pre-release (tags containing `-`) |
| `v1.0.0-rc.1` | Pre-release |

### CI/CD Workflows

| Workflow | Trigger | Action |
|----------|---------|--------|
| **CI** (`build.yml`) | Push/PR to `main` | Test + compile verification |
| **Release** (`release.yml`) | Push `v*` tag | Test + build + GitHub Release |

## License

MIT
