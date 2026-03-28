package main

import (
	"github.com/spf13/cobra"
)

var (
	flagConfigPath string
	flagOutputDir  string
	flagProvider   string
	flagVerbose    bool
	flagLimit      int
	flagSince      string
	flagUnreadOnly bool
	flagWatch      bool
	flagDryRun     bool
	flagNoSummary  bool
)

var rootCmd = &cobra.Command{
	Use:     "rlss",
	Short:   "Reading List Summarize Scraper",
	Long:    "Extract and summarize articles from Safari/Chrome reading lists.",
	Version: version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setupLogging(flagVerbose)
	},
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.StringVarP(&flagConfigPath, "config", "c", "~/.config/rlss/config.yaml", "config file path")
	pf.StringVarP(&flagOutputDir, "output", "o", "", "output directory (overrides config)")
	pf.StringVarP(&flagProvider, "provider", "p", "", "LLM provider (overrides config)")
	pf.BoolVarP(&flagVerbose, "verbose", "v", false, "enable debug logging")
	pf.IntVarP(&flagLimit, "limit", "n", 0, "max items to process")
	pf.StringVar(&flagSince, "since", "", "only items added after this date (YYYY-MM-DD)")
	pf.BoolVar(&flagUnreadOnly, "unread-only", false, "only process unread items")
	pf.BoolVar(&flagWatch, "watch", false, "enable watch mode")
	pf.BoolVar(&flagDryRun, "dry-run", false, "list items without processing")
	pf.BoolVar(&flagNoSummary, "no-summary", false, "extract only, skip summarization")

	rootCmd.AddCommand(processCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(urlCmd)
	rootCmd.AddCommand(configCmd)
}
