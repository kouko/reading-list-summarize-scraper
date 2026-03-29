package main

import (
	"log/slog"
	"os"

	rlssembed "github.com/kouko/reading-list-summarize-scraper/embed"
	"github.com/kouko/reading-list-summarize-scraper/internal/extract"
)

var version = "dev"

func init() {
	extract.SetDefuddleJS(rlssembed.DefuddleJS)
}

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
