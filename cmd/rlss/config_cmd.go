package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/kouko/reading-list-summarize-scraper/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show current configuration",
	RunE:  runConfig,
}

func runConfig(cmd *cobra.Command, args []string) error {
	cfgPath := config.ExpandPath(flagConfigPath)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Apply CLI overrides so the user sees what would actually be used.
	if flagOutputDir != "" {
		cfg.OutputDir = config.ExpandPath(flagOutputDir)
	}
	if flagProvider != "" {
		cfg.LLM.Provider.SetPrimary(flagProvider)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	fmt.Printf("# Config loaded from: %s\n", cfgPath)
	fmt.Println(string(data))
	return nil
}
