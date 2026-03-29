package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ────────────────────────────────────────────────────────────────────────────
// LLM provider types — ported from ytss config/config.go
// ────────────────────────────────────────────────────────────────────────────

// ProviderList holds one or more LLM provider names.
// The first entry is the primary provider; subsequent entries are fallbacks
// tried in order when the primary is unavailable (e.g., quota exhausted).
//
// YAML accepts both a scalar string and a list:
//
//	provider: "gemini-cli"          # single provider
//	provider:                       # provider chain
//	  - "gemini-cli"
//	  - "claude-code"
type ProviderList []string

func (p *ProviderList) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		*p = ProviderList{value.Value}
		return nil
	}
	var list []string
	if err := value.Decode(&list); err != nil {
		return err
	}
	*p = list
	return nil
}

// Primary returns the first (highest priority) provider name.
func (p ProviderList) Primary() string {
	if len(p) == 0 {
		return ""
	}
	return p[0]
}

// Fallbacks returns all providers after the primary.
func (p ProviderList) Fallbacks() []string {
	if len(p) <= 1 {
		return nil
	}
	return p[1:]
}

// String returns the primary provider name for display/frontmatter.
func (p ProviderList) String() string {
	return p.Primary()
}

// SetPrimary replaces the primary provider while keeping fallbacks.
func (p *ProviderList) SetPrimary(name string) {
	if len(*p) == 0 {
		*p = ProviderList{name}
		return
	}
	// Rebuild list: new primary + all others (deduplicating name).
	result := ProviderList{name}
	for _, existing := range *p {
		if existing != name {
			result = append(result, existing)
		}
	}
	*p = result
}

// Contains reports whether the list includes the named provider.
func (p ProviderList) Contains(name string) bool {
	for _, v := range p {
		if strings.EqualFold(v, name) {
			return true
		}
	}
	return false
}

// MarshalYAML serializes ProviderList back to YAML.
// A single-element list is written as a scalar string for backward compatibility.
func (p ProviderList) MarshalYAML() (interface{}, error) {
	if len(p) == 1 {
		return p[0], nil
	}
	return []string(p), nil
}

// Equal compares two ProviderList values for testing.
func (p ProviderList) Equal(other ProviderList) bool {
	if len(p) != len(other) {
		return false
	}
	for i := range p {
		if p[i] != other[i] {
			return false
		}
	}
	return true
}

// FallbackStrategyConfig controls how the provider fallback chain behaves.
type FallbackStrategyConfig struct {
	CooldownSeconds  int `yaml:"cooldown_seconds"`  // Seconds before retrying a failed provider (default: 300)
	FailureThreshold int `yaml:"failure_threshold"`  // Quota errors before skipping a provider (default: 1)
}

type LLMConfig struct {
	Provider                 ProviderList           `yaml:"provider"`
	ProviderFallbackStrategy FallbackStrategyConfig `yaml:"provider_fallback_strategy"`
	Ollama                   OllamaConfig           `yaml:"ollama"`
	LlamaCpp                 LlamaCppConfig         `yaml:"llamacpp"`
	ClaudeAPI                ClaudeAPIConfig        `yaml:"claude-api"`
	ClaudeCode               ClaudeCodeConfig       `yaml:"claude-code"`
	GeminiCLI                GeminiCLIConfig        `yaml:"gemini-cli"`
	QwenCode                 QwenCodeConfig         `yaml:"qwen-code"`
	OpenAICompat             OpenAICompatConfig     `yaml:"openai-compat"`
}

type OllamaConfig struct {
	Model    string `yaml:"model"`
	Endpoint string `yaml:"endpoint"`
	Think    *bool  `yaml:"think,omitempty"`
	Timeout  int    `yaml:"timeout"` // Seconds per LLM request (default: 900)
}

type LlamaCppConfig struct {
	Endpoint string `yaml:"endpoint"`
}

type ClaudeAPIConfig struct {
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
}

type ClaudeCodeConfig struct {
	Model   string `yaml:"model"`   // e.g. "sonnet", "opus", "claude-sonnet-4-6"
	Path    string `yaml:"path"`    // Path to claude binary (default: search in PATH)
	Timeout int    `yaml:"timeout"` // Seconds per LLM request (default: 900)
}

type GeminiCLIConfig struct {
	Model   string `yaml:"model"`
	Path    string `yaml:"path"`
	Timeout int    `yaml:"timeout"` // Seconds per LLM request (default: 900)
}

type QwenCodeConfig struct {
	Model   string `yaml:"model"`   // e.g. "coder-model" (free tier), "qwen3-coder-plus" (paid)
	Path    string `yaml:"path"`    // Path to qwen binary (default: search in PATH)
	Timeout int    `yaml:"timeout"` // Seconds per LLM request (default: 900)
}

type OpenAICompatConfig struct {
	Endpoint string `yaml:"endpoint"` // e.g. "http://localhost:8000/v1"
	Model    string `yaml:"model"`
	APIKey   string `yaml:"api_key"` // optional
	Timeout  int    `yaml:"timeout"` // Seconds per LLM request (default: 900)
}

// ────────────────────────────────────────────────────────────────────────────
// rlss-specific config structs
// ────────────────────────────────────────────────────────────────────────────

type Config struct {
	OutputDir string         `yaml:"output_dir"`
	LLM       LLMConfig      `yaml:"llm"`
	Summary   SummaryConfig  `yaml:"summary"`
	Safari    SafariConfig   `yaml:"safari"`
	Chrome    ChromeConfig   `yaml:"chrome"`
	Extract   ExtractConfig  `yaml:"extract"`
	Pipeline  PipelineConfig `yaml:"pipeline"`
	Filter    FilterConfig   `yaml:"filter"`
	Watch     WatchConfig    `yaml:"watch"`
	CopyTo    CopyToConfig   `yaml:"copy_to"`
	Obsidian  ObsidianConfig `yaml:"obsidian"`
}

type SummaryConfig struct {
	Language          string         `yaml:"language"`
	Prompt            string         `yaml:"prompt"`
	SummaryPromptFile string         `yaml:"summary_prompt_file"`
	MaxTokens         int            `yaml:"max_tokens"`
	Keywords          KeywordsConfig `yaml:"keywords"`
	Mermaid           MermaidConfig  `yaml:"mermaid"`
}

type KeywordsConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Language string `yaml:"language"`
	Count    int    `yaml:"count"`
}

type MermaidConfig struct {
	Enabled bool `yaml:"enabled"`
}

type SafariConfig struct {
	Enabled   bool   `yaml:"enabled"`
	PlistPath string `yaml:"plist_path"`
}

type ChromeConfig struct {
	Enabled         bool   `yaml:"enabled"`
	GoogleAccount   string `yaml:"google_account"`
	Profile         string `yaml:"profile"`
	UserDataDir     string `yaml:"user_data_dir"`
	ForceQuitChrome bool   `yaml:"force_quit_chrome"`
}

type DomainRule struct {
	Domains       []string `yaml:"domains"`
	Headed        bool     `yaml:"headed"`
	GoogleAccount string   `yaml:"google_account"`
	ChromeProfile string   `yaml:"chrome_profile"`
}

type ExtractConfig struct {
	Headless         bool          `yaml:"headless"`
	ChromeProfile    string        `yaml:"chrome_profile"`
	GoogleAccount    string        `yaml:"google_account"`
	UserDataDir      string        `yaml:"user_data_dir"`
	Timeout          time.Duration `yaml:"timeout"`
	HeadedTimeout    time.Duration `yaml:"headed_timeout"`
	WaitAfterLoad    time.Duration `yaml:"wait_after_load"`
	MinContentLength int           `yaml:"min_content_length"`
	DomainRules      []DomainRule  `yaml:"domain_rules"`
	HeadedOnBlock    bool          `yaml:"headed_on_block"`
}

type PipelineConfig struct {
	SkipExisting bool `yaml:"skip_existing"`
	DryRun       bool `yaml:"dry_run"`
	DelayMin     int  `yaml:"delay_min"`
	DelayMax     int  `yaml:"delay_max"`
}

type FilterConfig struct {
	UnreadOnly bool   `yaml:"unread_only"`
	Since      string `yaml:"since"`
	Limit      int    `yaml:"limit"`
}

type WatchConfig struct {
	Enabled  bool `yaml:"enabled"`
	Interval int  `yaml:"interval"`
}

type CopyToConfig struct {
	Enabled   bool     `yaml:"enabled"`
	Path      string   `yaml:"path"`
	Filename  string   `yaml:"filename"`
	Files     []string `yaml:"files"`
	Overwrite bool     `yaml:"overwrite"`
}

type ObsidianConfig struct {
	AutoTags  bool `yaml:"auto_tags"`
	Wikilinks bool `yaml:"wikilinks"`
}

// ────────────────────────────────────────────────────────────────────────────
// Loading
// ────────────────────────────────────────────────────────────────────────────

// Load reads and parses a YAML config file. If the file does not exist,
// it returns DefaultConfig without error.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			expandPaths(&cfg)
			return &cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	expandPaths(&cfg)
	return &cfg, nil
}

// expandPaths resolves ~ prefixes and environment variables in all path fields.
func expandPaths(cfg *Config) {
	cfg.OutputDir = ExpandPath(cfg.OutputDir)
	if cfg.Safari.PlistPath != "" {
		cfg.Safari.PlistPath = ExpandPath(cfg.Safari.PlistPath)
	}
	if cfg.Chrome.UserDataDir != "" {
		cfg.Chrome.UserDataDir = ExpandPath(cfg.Chrome.UserDataDir)
	}
	if cfg.Extract.UserDataDir != "" {
		cfg.Extract.UserDataDir = ExpandPath(cfg.Extract.UserDataDir)
	}
	if cfg.CopyTo.Path != "" {
		cfg.CopyTo.Path = ExpandPath(cfg.CopyTo.Path)
	}
	cfg.LLM.ClaudeAPI.APIKey = os.ExpandEnv(cfg.LLM.ClaudeAPI.APIKey)
	cfg.LLM.ClaudeCode.Path = ExpandPath(cfg.LLM.ClaudeCode.Path)
	cfg.LLM.GeminiCLI.Path = ExpandPath(cfg.LLM.GeminiCLI.Path)
	cfg.LLM.QwenCode.Path = ExpandPath(cfg.LLM.QwenCode.Path)
	cfg.Summary.SummaryPromptFile = ExpandPath(cfg.Summary.SummaryPromptFile)
}

// ExpandPath expands a leading ~ to the user's home directory.
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
