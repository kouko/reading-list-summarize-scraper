package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/kouko/reading-list-summarize-scraper/internal/extract"
	"github.com/kouko/reading-list-summarize-scraper/internal/output"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List reading list items without processing",
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, _, err := loadAndOverrideConfig()
	if err != nil {
		return err
	}

	// Create profile resolver (best-effort).
	resolver, err := extract.NewProfileResolver(cfg.Extract.UserDataDir)
	if err != nil {
		slog.Warn("chrome profile resolver unavailable", "err", err)
	}

	items, err := fetchAndFilter(cfg, resolver)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		fmt.Println("No items found.")
		return nil
	}

	// Build file index to show existing status.
	idx := output.NewFileIndex()
	idx.Build(cfg.OutputDir)

	for i, item := range items {
		sha8 := output.SHA8(item.URL)
		status := "new"
		if idx.Has(sha8) {
			info := idx.Get(sha8)
			if info.SummaryExists {
				status = "done"
			} else if info.ContentExists {
				status = "extracted"
			}
		}

		unread := " "
		if item.IsUnread {
			unread = "*"
		}

		fmt.Printf("%3d. [%s] %s %s\n", i+1, status, unread, item.Title)
		fmt.Printf("     %s  (%s, %s)\n", item.URL, item.Source, item.DateAdded.Format("2006-01-02"))
	}

	fmt.Printf("\nTotal: %d items\n", len(items))
	return nil
}
