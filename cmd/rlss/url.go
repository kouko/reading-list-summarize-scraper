package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/kouko/reading-list-summarize-scraper/internal/extract"
	"github.com/kouko/reading-list-summarize-scraper/internal/pipeline"
	"github.com/kouko/reading-list-summarize-scraper/internal/source"
	"github.com/kouko/reading-list-summarize-scraper/internal/summarize"
)

var (
	flagHeaded         bool
	flagExtractProfile string
)

var urlCmd = &cobra.Command{
	Use:   "url <url>",
	Short: "Process a single URL",
	Args:  cobra.ExactArgs(1),
	RunE:  runURL,
}

func init() {
	urlCmd.Flags().BoolVar(&flagHeaded, "headed", false, "use headed (visible) browser")
	urlCmd.Flags().StringVar(&flagExtractProfile, "extract-profile", "", "Chrome profile for extraction")
}

func runURL(cmd *cobra.Command, args []string) error {
	rawURL := args[0]

	cfg, _, err := loadAndOverrideConfig()
	if err != nil {
		return err
	}

	// Apply url-specific overrides.
	if flagHeaded {
		cfg.Extract.Headless = false
	}
	if flagExtractProfile != "" {
		cfg.Extract.ChromeProfile = flagExtractProfile
	}

	// Force processing (don't skip existing).
	cfg.Pipeline.SkipExisting = false

	// Create profile resolver (best-effort).
	resolver, err := extract.NewProfileResolver(cfg.Extract.UserDataDir)
	if err != nil {
		slog.Warn("chrome profile resolver unavailable", "err", err)
	}

	// Create browser pool.
	pool := extract.NewPool(&cfg.Extract, resolver, extract.GetDefuddleJS())
	defer pool.CloseAll()

	// Create summarizer (nil if --no-summary).
	var sum summarize.Summarizer
	if !flagNoSummary {
		sum, err = summarize.NewSummarizer(cfg.LLM)
		if err != nil {
			return fmt.Errorf("create summarizer: %w", err)
		}
	}

	// Create pipeline.
	p := pipeline.New(cfg, pool, sum)
	defer p.Shutdown()

	// Create a manual source item.
	manual := source.NewManualSource(rawURL)
	items, err := manual.Fetch()
	if err != nil {
		return fmt.Errorf("create manual source: %w", err)
	}

	stats := p.ProcessBatch(items)
	fmt.Print(stats.Report())

	if stats.Failed > 0 {
		return fmt.Errorf("%d item(s) failed", stats.Failed)
	}
	return nil
}
