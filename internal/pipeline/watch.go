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

// FetchFunc is the callback used by Watch to obtain reading items each cycle.
type FetchFunc func() ([]source.ReadingItem, error)

// Watch runs the pipeline in a loop, re-fetching items at the configured interval.
//
// Signal handling:
//   - SIGINT during processing: calls Shutdown(), finishes current item, then starts next cycle.
//   - SIGINT during sleep: exits the watch loop.
func Watch(p *Pipeline, configPath string, fetchItems FetchFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	for {
		// ── 1. Reset context for this cycle ──
		p.ResetContext()

		// ── 2. Partial config reload ──
		if configPath != "" {
			if newCfg, err := config.Load(configPath); err == nil {
				p.config.Filter = newCfg.Filter
				p.config.Watch = newCfg.Watch
				p.config.LLM = newCfg.LLM
				p.config.Summary = newCfg.Summary
				p.config.Pipeline = newCfg.Pipeline
				p.dryRun = newCfg.Pipeline.DryRun
				p.force = !newCfg.Pipeline.SkipExisting
				slog.Debug("watch: config reloaded")
			} else {
				slog.Warn("watch: config reload failed, using previous", "err", err)
			}
		}

		// ── 3. Rebuild index ──
		p.RebuildIndex()

		// ── 4. Fetch items ──
		items, err := fetchItems()
		if err != nil {
			slog.Error("watch: fetch items failed", "err", err)
		} else if len(items) > 0 {
			// Run batch in a goroutine so we can catch signals.
			doneCh := make(chan Stats, 1)
			go func() {
				stats := p.ProcessBatch(items)
				doneCh <- stats
			}()

			// Wait for batch or signal.
			select {
			case stats := <-doneCh:
				slog.Info("watch: cycle complete", "report", stats.Report())
			case sig := <-sigCh:
				slog.Info("watch: signal during processing, shutting down batch", "signal", sig)
				p.Shutdown()
				stats := <-doneCh // wait for batch to finish current item
				slog.Info("watch: batch stopped", "report", stats.Report())
				// Continue to sleep phase; next SIGINT will exit.
			}
		} else {
			slog.Info("watch: no items to process")
		}

		// ── 5. Sleep interval ──
		interval := time.Duration(p.config.Watch.Interval) * time.Minute
		if interval <= 0 {
			interval = 10 * time.Minute
		}
		slog.Info("watch: sleeping", "minutes", interval.Minutes())

		select {
		case <-time.After(interval):
			// Next cycle.
		case sig := <-sigCh:
			slog.Info("watch: signal during sleep, exiting", "signal", sig)
			return
		}
	}
}
