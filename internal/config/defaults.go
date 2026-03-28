package config

import "time"

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() Config {
	return Config{
		OutputDir: "~/reading-list-summaries",
		LLM: LLMConfig{
			Provider: ProviderList{"claude-code"},
			ProviderFallbackStrategy: FallbackStrategyConfig{
				CooldownSeconds:  300,
				FailureThreshold: 1,
			},
			Ollama: OllamaConfig{
				Model:    "llama3",
				Endpoint: "http://localhost:11434",
				Think:    ptrBool(false),
				Timeout:  900,
			},
			LlamaCpp: LlamaCppConfig{
				Endpoint: "http://localhost:8080",
			},
			ClaudeCode: ClaudeCodeConfig{
				Model:   "haiku",
				Timeout: 900,
			},
			GeminiCLI: GeminiCLIConfig{
				Model:   "auto",
				Timeout: 900,
			},
			QwenCode: QwenCodeConfig{
				Model:   "coder-model",
				Timeout: 900,
			},
			OpenAICompat: OpenAICompatConfig{
				Timeout: 900,
			},
		},
		Summary: SummaryConfig{
			Language:  "zh-Hant",
			MaxTokens: 10000,
			Keywords: KeywordsConfig{
				Enabled:  true,
				Language: "en",
				Count:    5,
			},
			Mermaid: MermaidConfig{
				Enabled: true,
			},
		},
		Safari: SafariConfig{
			Enabled: true,
		},
		Chrome: ChromeConfig{
			Enabled:     true,
			Profile:     "Default",
			UserDataDir: "~/.config/rlss/chrome-data",
		},
		Extract: ExtractConfig{
			Headless:      true,
			ChromeProfile: "Default",
			Timeout:       30 * time.Second,
			WaitAfterLoad: 2 * time.Second,
		},
		Pipeline: PipelineConfig{
			SkipExisting: true,
			DelayMin:     3,
			DelayMax:     8,
		},
		Filter: FilterConfig{},
		Watch: WatchConfig{
			Interval: 10,
		},
		CopyTo: CopyToConfig{
			Files: []string{"summary", "content"},
		},
		Obsidian: ObsidianConfig{
			AutoTags:  true,
			Wikilinks: true,
		},
	}
}

func ptrBool(b bool) *bool {
	return &b
}
