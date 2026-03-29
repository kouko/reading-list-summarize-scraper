package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"

	rlssembed "github.com/kouko/reading-list-summarize-scraper/embed"
	"github.com/kouko/reading-list-summarize-scraper/internal/config"
	"github.com/kouko/reading-list-summarize-scraper/internal/extract"
	"github.com/kouko/reading-list-summarize-scraper/internal/pipeline"
	"github.com/kouko/reading-list-summarize-scraper/internal/source"
	"github.com/kouko/reading-list-summarize-scraper/internal/summarize"
)

var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Process reading list items (default command)",
	RunE:  runProcess,
}

func init() {
	// Make "process" the default command when no subcommand is given.
	rootCmd.RunE = runProcess
}

func runProcess(cmd *cobra.Command, args []string) error {
	cfg, cfgPath, err := loadAndOverrideConfig()
	if err != nil {
		return err
	}

	// Create profile resolver scanning multiple userDataDirs (best-effort, nil on failure).
	resolver, err := extract.NewProfileResolver(
		config.ExpandPath("~/.config/rlss/chrome-data"),
		config.ExpandPath("~/Library/Application Support/Google/Chrome"),
	)
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

	// Watch mode.
	if flagWatch || cfg.Watch.Enabled {
		fetchFn := func() ([]source.ReadingItem, error) {
			return fetchAndFilter(cfg, resolver)
		}
		pipeline.Watch(p, cfgPath, fetchFn)
		return nil
	}

	// Single run.
	items, err := fetchAndFilter(cfg, resolver)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		slog.Info("no items to process")
		return nil
	}

	stats := p.ProcessBatch(items)
	fmt.Print(stats.Report())
	return nil
}

// loadAndOverrideConfig loads config from file and applies CLI flag overrides.
func loadAndOverrideConfig() (*config.Config, string, error) {
	cfgPath := config.ExpandPath(flagConfigPath)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, "", fmt.Errorf("load config: %w", err)
	}

	// Apply CLI overrides.
	if flagOutputDir != "" {
		cfg.OutputDir = config.ExpandPath(flagOutputDir)
	}
	if flagProvider != "" {
		cfg.LLM.Provider.SetPrimary(flagProvider)
	}
	if flagLimit > 0 {
		cfg.Filter.Limit = flagLimit
	}
	if flagSince != "" {
		cfg.Filter.Since = flagSince
	}
	if flagUnreadOnly {
		cfg.Filter.UnreadOnly = true
	}
	if flagDryRun {
		cfg.Pipeline.DryRun = true
	}

	// Source selection overrides.
	if flagAll {
		cfg.Safari.Enabled = true
		cfg.Chrome.Enabled = true
	} else if flagSafari || flagChrome {
		// When any source flag is explicitly set, disable all first,
		// then enable only what was requested.
		cfg.Safari.Enabled = false
		cfg.Chrome.Enabled = false
		if flagSafari {
			cfg.Safari.Enabled = true
		}
		if flagChrome {
			cfg.Chrome.Enabled = true
		}
	}
	// If none of --safari, --chrome, --all set, use config defaults.

	if flagForce {
		cfg.Pipeline.SkipExisting = false
	}
	if flagProfile != "" {
		cfg.Chrome.Profile = flagProfile
	}
	if flagCloneProfile {
		cfg.Chrome.CloneProfile = true
	}

	return cfg, cfgPath, nil
}

// fetchAndFilter fetches items from all enabled sources, deduplicates, and applies filters.
func fetchAndFilter(cfg *config.Config, resolver *extract.ProfileResolver) ([]source.ReadingItem, error) {
	var sources []source.Source

	if cfg.Safari.Enabled {
		sources = append(sources, source.NewSafariSource(cfg.Safari.PlistPath))
	}
	if cfg.Chrome.Enabled {
		profileDir := cfg.Chrome.Profile
		userDataDir := cfg.Chrome.UserDataDir
		if resolver != nil && cfg.Chrome.GoogleAccount != "" {
			folder, dir, err := resolver.SmartResolve(
				cfg.Chrome.GoogleAccount, profileDir, userDataDir, cfg.Chrome.CloneProfile,
			)
			if err != nil {
				slog.Warn("smart resolve failed, falling back", "err", err)
			} else {
				profileDir = folder
				userDataDir = dir
			}
		} else if resolver != nil && profileDir != "" {
			if folder, err := resolver.Resolve(profileDir); err == nil {
				profileDir = folder
			}
		}
		sources = append(sources, source.NewChromeSource(
			profileDir, userDataDir, cfg.Chrome.GoogleAccount, cfg.Chrome.CloneProfile,
			rlssembed.ExtensionManifest, rlssembed.ExtensionBackground,
		))
	}

	var allItems []source.ReadingItem
	for _, src := range sources {
		items, err := src.Fetch()
		if err != nil {
			var fdaErr *source.FullDiskAccessError
			if errors.As(err, &fdaErr) {
				fmt.Fprint(os.Stderr, source.FormatFullDiskAccessBanner(fdaErr.Path))
				if openErr := source.OpenFullDiskAccessSettings(); openErr != nil {
					slog.Debug("could not open System Settings", "err", openErr)
				}
			} else {
				slog.Error("source fetch failed", "source", src.Name(), "err", err)
			}
			continue
		}
		slog.Info("fetched items", "source", src.Name(), "count", len(items))
		allItems = append(allItems, items...)
	}

	allItems = source.DeduplicateByURL(allItems)

	// Apply filters.
	allItems = applyFilters(allItems, cfg.Filter)

	return allItems, nil
}

// applyFilters applies the configured filters to the item list.
func applyFilters(items []source.ReadingItem, filter config.FilterConfig) []source.ReadingItem {
	var result []source.ReadingItem

	var sinceTime time.Time
	if filter.Since != "" {
		if t, err := time.Parse("2006-01-02", filter.Since); err == nil {
			sinceTime = t
		} else {
			slog.Warn("invalid --since date, ignoring", "value", filter.Since, "err", err)
		}
	}

	for _, item := range items {
		if filter.UnreadOnly && !item.IsUnread {
			continue
		}
		if !sinceTime.IsZero() && item.DateAdded.Before(sinceTime) {
			continue
		}
		result = append(result, item)
	}

	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}

	return result
}
