package pipeline

import (
	"context"
	"errors"
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

// emptyResponseError builds the stage-1 "empty response" error, naming the
// provider/model that returned empty text so the failure identifies the actual
// LLM backend used (an empty response counts as success to the fallback chain,
// so summaryResult still carries the provider that answered).
func emptyResponseError(stage, provider, model string) error {
	return fmt.Errorf("%s: LLM returned empty response from provider %q (model %q) — if using a thinking model (e.g., Qwen3.5), ensure think mode is disabled or increase max_tokens", stage, provider, model)
}

// validateSummaryText returns a named empty-response error when the LLM
// produced no usable (non-whitespace) summary, else nil.
func validateSummaryText(summaryText, provider, model string) error {
	if strings.TrimSpace(summaryText) == "" {
		return emptyResponseError("stage 1", provider, model)
	}
	return nil
}

// errPartial is a sentinel signaling that an item was partially processed:
// its content was extracted and saved but summarization failed (e.g. all LLM
// providers out of quota). It is wrapped with the underlying cause via %w, so
// detection uses errors.Is rather than equality.
var errPartial = errors.New("partial: content extracted but summarization failed")

// IsPartial reports whether err is (or wraps) the partial sentinel.
func IsPartial(err error) bool { return errors.Is(err, errPartial) }

// resultBucket classifies a single item's processing outcome for Stats.
type resultBucket int

const (
	bucketSuccess resultBucket = iota
	bucketSkipped
	bucketPartial
	bucketFailed
)

// classifyResult maps a ProcessItem return value to its stats bucket.
// Order matters: skipped and partial are sentinels checked before the
// catch-all failed bucket.
func classifyResult(err error) resultBucket {
	switch {
	case err == nil:
		return bucketSuccess
	case IsSkipped(err):
		return bucketSkipped
	case IsPartial(err):
		return bucketPartial
	default:
		return bucketFailed
	}
}

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

	for i, item := range items {
		// Check for shutdown.
		select {
		case <-p.ctx.Done():
			slog.Info("pipeline: shutdown requested, stopping batch")
			stats.End = time.Now()
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
		switch classifyResult(err) {
		case bucketSuccess:
			stats.Success++
		case bucketSkipped:
			stats.Skipped++
		case bucketPartial:
			// Content saved; only summarization failed (e.g. quota). Not lost
			// and not a hard failure — a later run resumes the summary.
			stats.Partial++
			slog.Warn("item partial: content saved, summary failed", "url", item.URL, "err", err)
		case bucketFailed:
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

	stats.End = time.Now()
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

		// Maintenance / paywall / login-wall page served instead of the article.
		// Fail (not a silent short-content "success", not a garbage summary) and
		// write no content file, so the item is re-attempted on the next run.
		if isUnavailablePage(markdown) {
			return fmt.Errorf("extract %s: page unavailable (maintenance / paywall / login required) — not summarized; will retry next run", item.URL)
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

	contentLen := len([]rune(markdown))
	minLen := p.config.Extract.MinContentLength
	if minLen > 0 && contentLen < minLen {
		slog.Warn("content too short, skipping summary",
			"sha8", sha8, "length", contentLen, "min", minLen)
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
		// Content was already extracted and written above, so this is a
		// partial outcome (resumable), not a hard failure.
		return fmt.Errorf("%w: summarize: %v", errPartial, err)
	}
	summaryText := summarize.StripThinkingTags(summaryResult.Text)
	if err := validateSummaryText(summaryText, summaryResult.Provider, summaryResult.Model); err != nil {
		// Content was already written above; an empty summary is resumable, so
		// mark partial rather than writing out an empty summary file.
		return fmt.Errorf("%w: %v", errPartial, err)
	}
	slog.Info("summary stage 1 complete",
		"sha8", sha8,
		"provider", summaryResult.Provider,
		"model", summaryResult.Model,
		"bytes", len(summaryText),
	)

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
				Prompt:     prompt,
				AllowEmpty: true, // keywords can legitimately be empty
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
				Prompt:     prompt,
				AllowEmpty: true, // a summary may legitimately need no diagram
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

	var embedContent string
	if p.config.Summary.EmbedContent {
		embedContent = markdown
	}

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
		EmbedContent:  embedContent,
		Language:      p.config.Summary.Language,
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

// isUnavailablePage detects pages that were served INSTEAD of the article and
// that won't be fixed by retrying headless→headed: maintenance pages, paywalls,
// and login/subscription walls. These otherwise slip through (they aren't
// anti-bot challenges) and get silently mis-handled — counted as a short-content
// "success" or summarized into a "I can't, it's a paywall" non-summary. Treat
// them as an extraction failure instead (the caller does not write a content
// file, so the item is re-attempted on the next run).
//
// Patterns are deliberately specific multi-word/wall strings (not bare topic
// words like "subscribe" or "登入") so a real article *about* paywalls or
// maintenance is not misflagged. One match is enough.
func isUnavailablePage(content string) bool {
	lower := strings.ToLower(content)
	unavailablePatterns := []string{
		// English — maintenance / login / subscription walls
		"under maintenance",
		"sign in to read",
		"log in to read",
		"please sign in to",
		"subscribe to read",
		"subscribe to continue reading",
		// Chinese (Traditional + Simplified) — maintenance / login / member walls
		"系統維護中",
		"系统维护中",
		"登入後閱讀全文",
		"請登入後閱讀",
		"訂閱成為會員",
		// Japanese — maintenance / login walls
		"メンテナンス中",
		"ログインしてください",
		"ログインが必要",
	}
	for _, pattern := range unavailablePatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

// isBlockedPage detects anti-bot protection pages (Cloudflare, CAPTCHA, etc.)
// that were extracted instead of actual content. Patterns include localized
// (ja / zh-Hant / zh-Hans) challenge-UI strings, since Cloudflare/anti-bot
// interstitials render in the page's language; the language-neutral branding
// ("cloudflare", "ray id:", "captcha") is kept too. The >=2 threshold guards
// against false-positives on real articles that merely discuss these topics —
// so the localized entries are challenge-UI phrases, not generic security words.
func isBlockedPage(content string) bool {
	lower := strings.ToLower(content)
	blockedPatterns := []string{
		// English / language-neutral
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
		// Japanese challenge-UI
		"ブラウザを確認しています",
		"あなたが人間であることを確認",
		"ロボットではないことを確認",
		"アクセスが拒否されました",
		// Chinese (Traditional) challenge-UI
		"正在檢查您的瀏覽器",
		"請完成安全驗證",
		"確認您是真人",
		"拒絕存取",
		// Chinese (Simplified) challenge-UI
		"正在检查您的浏览器",
		"请完成安全验证",
		"确认您是真人",
		"拒绝访问",
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
