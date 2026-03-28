package pipeline

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/kouko/reading-list-summarize-scraper/internal/config"
	"github.com/kouko/reading-list-summarize-scraper/internal/extract"
	"github.com/kouko/reading-list-summarize-scraper/internal/output"
	"github.com/kouko/reading-list-summarize-scraper/internal/source"
	"github.com/kouko/reading-list-summarize-scraper/internal/summarize"
)

var errSkipped = fmt.Errorf("skipped")

// IsSkipped reports whether the error indicates the item was skipped.
func IsSkipped(err error) bool { return err == errSkipped }

// Pipeline orchestrates extraction, summarization, and output for reading list items.
type Pipeline struct {
	config     *config.Config
	pool       *extract.Pool
	summarizer summarize.Summarizer // may be nil
	index      *output.FileIndex
	ctx        context.Context
	cancel     context.CancelFunc
	force      bool
	dryRun     bool
}

// New creates a new Pipeline. The summarizer may be nil (extraction-only mode).
func New(cfg *config.Config, pool *extract.Pool, sum summarize.Summarizer) *Pipeline {
	ctx, cancel := context.WithCancel(context.Background())
	idx := output.NewFileIndex()
	idx.Build(cfg.OutputDir)

	return &Pipeline{
		config:     cfg,
		pool:       pool,
		summarizer: sum,
		index:      idx,
		ctx:        ctx,
		cancel:     cancel,
		force:      !cfg.Pipeline.SkipExisting,
		dryRun:     cfg.Pipeline.DryRun,
	}
}

// Shutdown cancels the pipeline context, signalling ProcessBatch to stop
// after the current item completes.
func (p *Pipeline) Shutdown() {
	p.cancel()
}

// ResetContext replaces the pipeline context with a fresh one.
// Called at the start of each watch cycle.
func (p *Pipeline) ResetContext() {
	p.ctx, p.cancel = context.WithCancel(context.Background())
}

// RebuildIndex rescans the output directory to refresh the file index.
func (p *Pipeline) RebuildIndex() {
	p.index.Build(p.config.OutputDir)
}

// ProcessBatch processes a slice of reading items sequentially, collecting stats.
// It inserts a random delay between items and respects context cancellation.
func (p *Pipeline) ProcessBatch(items []source.ReadingItem) Stats {
	stats := Stats{Start: time.Now()}
	defer func() { stats.End = time.Now() }()

	for i, item := range items {
		// Check for shutdown.
		select {
		case <-p.ctx.Done():
			slog.Info("pipeline: shutdown requested, stopping batch")
			return stats
		default:
		}

		slog.Info("processing item",
			"index", i+1,
			"total", len(items),
			"url", item.URL,
			"title", item.Title,
		)

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
			slog.Error("item failed", "url", item.URL, "err", err)
		}

		// Random delay between items (not after the last one, skip in dry-run).
		if i < len(items)-1 && !p.dryRun {
			p.delayBetweenItems()
		}
	}

	return stats
}

// ProcessItem runs the full pipeline for a single reading item.
func (p *Pipeline) ProcessItem(item source.ReadingItem) error {
	sha8 := output.SHA8(item.URL)
	domainDir := output.DomainDir(item.URL)
	now := time.Now()

	// ── 1. Skip detection ──
	if !p.force && p.index.Has(sha8) {
		info := p.index.Get(sha8)
		if info.SummaryExists {
			slog.Debug("skipping existing", "url", item.URL, "sha8", sha8)
			return errSkipped
		}
	}

	// ── 2. Dry run ──
	if p.dryRun {
		slog.Info("dry-run: would process", "url", item.URL, "sha8", sha8)
		return errSkipped
	}

	// ── 3. Resume detection ──
	// If content exists but summary is missing, read existing content and skip to summarization.
	var markdown string
	info := p.index.Get(sha8)
	if info.ContentExists && !info.SummaryExists {
		contentPath := p.index.ContentPath(sha8)
		if contentPath != "" {
			data, err := os.ReadFile(contentPath)
			if err == nil {
				slog.Info("resuming: content exists, skipping extraction", "sha8", sha8)
				markdown = stripFrontmatter(string(data))
			}
		}
	}

	outDir := filepath.Join(p.config.OutputDir, domainDir)

	// ── 4. Extract ──
	if markdown == "" {
		// Domain rule matching is handled inside pool.ExtractURL.
		var err error
		markdown, err = p.pool.ExtractURL(item.URL)
		if err != nil {
			return fmt.Errorf("extract %s: %w", item.URL, err)
		}
		if strings.TrimSpace(markdown) == "" {
			return fmt.Errorf("extract %s: empty content (Defuddle returned nothing)", item.URL)
		}
		if isBlockedPage(markdown) {
			if p.config.Extract.HeadedOnBlock {
				slog.Warn("blocked by anti-bot, retrying with headed mode", "url", item.URL)
				markdown, err = p.pool.ExtractURLHeaded(item.URL)
				if err != nil {
					return fmt.Errorf("extract (headed retry) %s: %w", item.URL, err)
				}
				if strings.TrimSpace(markdown) == "" || isBlockedPage(markdown) {
					return fmt.Errorf("extract %s: still blocked after headed retry", item.URL)
				}
			} else {
				return fmt.Errorf("extract %s: blocked by anti-bot protection (Cloudflare/CAPTCHA). Try setting extract.headed_on_block: true or adding this domain to extract.domain_rules with headed: true", item.URL)
			}
		}

		// Create output directory.
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("create dir %s: %w", outDir, err)
		}

		// Write content file.
		contentFile := output.ContentFilename(now, sha8)
		contentPath := filepath.Join(outDir, contentFile)
		domain := extractDomain(item.URL)

		contentDoc := output.AssembleContent(output.ContentParams{
			Title:         item.Title,
			URL:           item.URL,
			Domain:        domain,
			Source:        item.Source,
			DateAdded:     item.DateAdded,
			ProcessedDate: now,
			ContentLength: len(markdown),
			ExtractedBy:   "defuddle",
			Content:       markdown,
		})
		if err := os.WriteFile(contentPath, []byte(contentDoc), 0644); err != nil {
			return fmt.Errorf("write content: %w", err)
		}
		slog.Info("wrote content", "path", contentPath)
	}

	// ── 5. Summarization (3-stage) ──
	if p.summarizer == nil {
		slog.Info("no summarizer configured, skipping summary", "sha8", sha8)
		return nil
	}

	domain := extractDomain(item.URL)

	// Stage 1: Main summary (blocking).
	summaryPrompt, err := summarize.ResolveAndSubstitute(p.config.Summary, summarize.PromptVars{
		Title:         item.Title,
		Domain:        domain,
		DateAdded:     item.DateAdded.Format("2006-01-02"),
		Source:        item.Source,
		Content:       markdown,
		ContentLength: len(markdown),
		Language:      p.config.Summary.Language,
	})
	if err != nil {
		return fmt.Errorf("resolve prompt: %w", err)
	}

	summaryResult, err := p.summarizer.Summarize(markdown, summarize.SummarizeOptions{
		Prompt:    summaryPrompt,
		MaxTokens: p.config.Summary.MaxTokens,
	})
	if err != nil {
		return fmt.Errorf("summarize: %w", err)
	}
	summaryText := summarize.StripThinkingTags(summaryResult.Text)

	// Stage 2 & 3: Keywords and Mermaid (concurrent, non-blocking).
	var (
		keywords      []string
		mermaidBlocks []output.MermaidBlock
		wg            sync.WaitGroup
		kwErr         error
		mermaidErr    error
	)

	if p.config.Summary.Keywords.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			kwCfg := p.config.Summary.Keywords
			prompt, err := summarize.KeywordPrompt(summaryText, kwCfg.Language, kwCfg.Count)
			if err != nil {
				kwErr = err
				return
			}
			result, err := p.summarizer.Summarize(summaryText, summarize.SummarizeOptions{
				Prompt: prompt,
			})
			if err != nil {
				kwErr = err
				return
			}
			keywords = summarize.ParseKeywords(summarize.StripThinkingTags(result.Text))
			slog.Info("keywords extracted", "count", len(keywords))
		}()
	}

	if p.config.Summary.Mermaid.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			prompt, err := summarize.MermaidPrompt(summaryText, p.config.Summary.Language)
			if err != nil {
				mermaidErr = err
				return
			}
			result, err := p.summarizer.Summarize(summaryText, summarize.SummarizeOptions{
				Prompt: prompt,
			})
			if err != nil {
				mermaidErr = err
				return
			}
			raw := summarize.StripThinkingTags(result.Text)
			blocks := summarize.ValidateMermaidBlocks(raw)
			for _, b := range blocks {
				mermaidBlocks = append(mermaidBlocks, output.MermaidBlock{
					Title: b.Title,
					Code:  b.Code,
				})
			}
			slog.Info("mermaid blocks extracted", "count", len(mermaidBlocks))
		}()
	}

	wg.Wait()

	// Log non-fatal errors from stage 2/3.
	if kwErr != nil {
		slog.Warn("keyword extraction failed", "err", kwErr)
	}
	if mermaidErr != nil {
		slog.Warn("mermaid generation failed", "err", mermaidErr)
	}

	// ── 6. Assemble + write summary ──
	contentTier := summarize.CalculateTier(len(markdown), p.config.Summary.Language)

	summaryDoc := output.AssembleSummary(output.SummaryParams{
		Title:         item.Title,
		URL:           item.URL,
		Domain:        domain,
		Source:        item.Source,
		DateAdded:     item.DateAdded,
		ProcessedDate: now,
		LLMProvider:   summaryResult.Provider,
		LLMModel:      summaryResult.Model,
		ContentLength: len(markdown),
		ContentTier:   contentTier,
		SummaryText:   summaryText,
		Keywords:      keywords,
		MermaidBlocks: mermaidBlocks,
	})

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("create dir %s: %w", outDir, err)
	}

	summaryFile := output.SummaryFilename(now, sha8)
	summaryPath := filepath.Join(outDir, summaryFile)
	if err := os.WriteFile(summaryPath, []byte(summaryDoc), 0644); err != nil {
		return fmt.Errorf("write summary: %w", err)
	}
	slog.Info("wrote summary", "path", summaryPath)

	// ── 7. CopyTo ──
	copyVars := output.CopyToVars{
		OutputDir: p.config.OutputDir,
		Date:      now.Format("2006-01-02"),
		DateAdded: item.DateAdded.Format("2006-01-02"),
		Title:     output.SanitizeTitleForDisplay(item.Title),
		SHA8:      sha8,
		Source:    item.Source,
		Domain:    domain,
		DomainDir: domainDir,
	}
	if err := output.ExecuteCopyTo(p.config.CopyTo, outDir, sha8, copyVars); err != nil {
		slog.Warn("copy_to failed", "err", err)
	}

	return nil
}

// delayBetweenItems sleeps for a random duration between DelayMin and DelayMax seconds.
func (p *Pipeline) delayBetweenItems() {
	min := p.config.Pipeline.DelayMin
	max := p.config.Pipeline.DelayMax
	if min <= 0 || max <= 0 || max <= min {
		return
	}
	d := time.Duration(min+rand.Intn(max-min+1)) * time.Second
	slog.Debug("delay between items", "seconds", d.Seconds())

	select {
	case <-time.After(d):
	case <-p.ctx.Done():
	}
}

// extractDomain returns the hostname from a URL, or "unknown" on failure.
func extractDomain(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return "unknown"
	}
	return u.Hostname()
}

// isBlockedPage detects anti-bot protection pages (Cloudflare, CAPTCHA, etc.)
// that were extracted instead of actual content.
func isBlockedPage(content string) bool {
	lower := strings.ToLower(content)
	blockedPatterns := []string{
		"performing security verification",
		"security challenge",
		"checking your browser",
		"please wait while we verify",
		"just a moment",
		"ray id:",
		"cloudflare",
		"captcha",
		"access denied",
		"please verify you are a human",
		"bot protection",
	}
	matchCount := 0
	for _, pattern := range blockedPatterns {
		if strings.Contains(lower, pattern) {
			matchCount++
		}
	}
	// Require at least 2 matches to avoid false positives
	return matchCount >= 2
}

// stripFrontmatter removes YAML frontmatter (--- ... ---) from the beginning of content.
func stripFrontmatter(content string) string {
	if !strings.HasPrefix(content, "---\n") {
		return content
	}
	end := strings.Index(content[4:], "\n---\n")
	if end < 0 {
		return content
	}
	return strings.TrimSpace(content[4+end+5:])
}
