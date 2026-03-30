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
	flagSafari     bool
	flagChrome     bool
	flagAll        bool
	flagForce        bool
	flagProfile      string
	flagCloneProfile bool
	flagEmbedContent bool
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
	pf.StringVar(&flagConfigPath, "config", "./config.yaml", "config file path")
	pf.StringVarP(&flagOutputDir, "output", "o", "", "output directory (overrides config)")
	pf.StringVar(&flagProvider, "provider", "", "LLM provider (overrides config)")
	pf.BoolVarP(&flagVerbose, "verbose", "v", false, "enable debug logging")
	pf.IntVarP(&flagLimit, "limit", "n", 0, "max items to process")
	pf.StringVar(&flagSince, "since", "", "only items added after this date (YYYY-MM-DD)")
	pf.BoolVar(&flagUnreadOnly, "unread-only", false, "only process unread items")
	pf.BoolVar(&flagWatch, "watch", false, "enable watch mode")
	pf.BoolVar(&flagDryRun, "dry-run", false, "list items without processing")
	pf.BoolVar(&flagNoSummary, "no-summary", false, "extract only, skip summarization")
	pf.BoolVarP(&flagSafari, "safari", "s", false, "process Safari Reading List")
	pf.BoolVarP(&flagChrome, "chrome", "c", false, "process Chrome Reading List")
	pf.BoolVarP(&flagAll, "all", "a", false, "process all sources")
	pf.BoolVar(&flagForce, "force", false, "force reprocess existing items")
	pf.StringVar(&flagProfile, "profile", "", "Chrome Reading List profile override (UI name)")
	pf.BoolVar(&flagCloneProfile, "clone-profile", false, "clone Chrome profile if locked by SingletonLock")
	pf.BoolVar(&flagEmbedContent, "embed-content", false, "embed original article content in summary file")

	rootCmd.AddCommand(processCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(urlCmd)
	rootCmd.AddCommand(configCmd)
}
