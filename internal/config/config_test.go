package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestExpandPath_Tilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}

	got := ExpandPath("~/foo")
	want := filepath.Join(home, "foo")
	if got != want {
		t.Errorf("ExpandPath(~/foo) = %q, want %q", got, want)
	}
}

func TestExpandPath_Absolute(t *testing.T) {
	got := ExpandPath("/abs/path")
	if got != "/abs/path" {
		t.Errorf("ExpandPath(/abs/path) = %q, want /abs/path", got)
	}
}

func TestExpandPath_Empty(t *testing.T) {
	got := ExpandPath("")
	if got != "" {
		t.Errorf("ExpandPath(\"\") = %q, want empty", got)
	}
}

func TestExpandPath_NoTildePrefix(t *testing.T) {
	// A tilde not followed by / should be unchanged.
	got := ExpandPath("~notapath")
	if got != "~notapath" {
		t.Errorf("ExpandPath(~notapath) = %q, want ~notapath", got)
	}
}

func TestLoad_ValidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")

	content := `output_dir: /tmp/test-output
llm:
  provider: gemini-cli
summary:
  language: en
pipeline:
  skip_existing: false
  dry_run: true
`
	if err := os.WriteFile(cfgFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgFile)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.OutputDir != "/tmp/test-output" {
		t.Errorf("OutputDir = %q, want /tmp/test-output", cfg.OutputDir)
	}
	if cfg.LLM.Provider.Primary() != "gemini-cli" {
		t.Errorf("Provider.Primary() = %q, want gemini-cli", cfg.LLM.Provider.Primary())
	}
	if cfg.Summary.Language != "en" {
		t.Errorf("Summary.Language = %q, want en", cfg.Summary.Language)
	}
	if cfg.Pipeline.SkipExisting != false {
		t.Error("Pipeline.SkipExisting should be false")
	}
	if cfg.Pipeline.DryRun != true {
		t.Error("Pipeline.DryRun should be true")
	}
}

func TestLoad_FallbackStrategy(t *testing.T) {
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")

	content := `output_dir: /tmp/test-output
llm:
  provider: gemini-cli
  provider_fallback_strategy:
    cooldown_seconds: 600
    failure_threshold: 3
    rate_limit_cooldown_seconds: 90
    empty_response_threshold: 5
`
	if err := os.WriteFile(cfgFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgFile)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	s := cfg.LLM.ProviderFallbackStrategy
	if s.CooldownSeconds != 600 {
		t.Errorf("CooldownSeconds = %d, want 600", s.CooldownSeconds)
	}
	if s.FailureThreshold != 3 {
		t.Errorf("FailureThreshold = %d, want 3", s.FailureThreshold)
	}
	if s.RateLimitCooldownSeconds != 90 {
		t.Errorf("RateLimitCooldownSeconds = %d, want 90", s.RateLimitCooldownSeconds)
	}
	if s.EmptyResponseThreshold != 5 {
		t.Errorf("EmptyResponseThreshold = %d, want 5", s.EmptyResponseThreshold)
	}
}

func TestLoad_DomainRulePagination(t *testing.T) {
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")

	content := `output_dir: /tmp/test-output
extract:
  domain_rules:
    - domains: ["www.itmedia.co.jp"]
      next_page_selector: "a[rel=next]"
      max_pages: 5
`
	if err := os.WriteFile(cfgFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgFile)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(cfg.Extract.DomainRules) != 1 {
		t.Fatalf("DomainRules len = %d, want 1", len(cfg.Extract.DomainRules))
	}
	r := cfg.Extract.DomainRules[0]
	if r.NextPageSelector != "a[rel=next]" {
		t.Errorf("NextPageSelector = %q, want a[rel=next]", r.NextPageSelector)
	}
	if r.MaxPages != 5 {
		t.Errorf("MaxPages = %d, want 5", r.MaxPages)
	}
}

func TestLoad_SummaryStyle(t *testing.T) {
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")

	content := `output_dir: /tmp/test-output
summary:
  style: article
`
	if err := os.WriteFile(cfgFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(cfgFile)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Summary.Style != "article" {
		t.Errorf("Summary.Style = %q, want article", cfg.Summary.Style)
	}

	// Default config leaves Style unset; the resolver maps "" → outline.
	if def := DefaultConfig(); def.Summary.Style != "" {
		t.Errorf("DefaultConfig Summary.Style = %q, want \"\" (outline default)", def.Summary.Style)
	}
}

func TestLoad_RSSConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")

	content := `output_dir: /tmp/test-output
rss:
  enabled: true
  count: 5
  feeds:
    - https://example.com/feed.xml
    - https://blog.example.org/atom
`
	if err := os.WriteFile(cfgFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgFile)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if !cfg.RSS.Enabled {
		t.Error("RSS.Enabled should be true")
	}
	if cfg.RSS.Count != 5 {
		t.Errorf("RSS.Count = %d, want 5", cfg.RSS.Count)
	}
	if len(cfg.RSS.Feeds) != 2 {
		t.Fatalf("RSS.Feeds len = %d, want 2", len(cfg.RSS.Feeds))
	}
	if cfg.RSS.Feeds[0] != "https://example.com/feed.xml" {
		t.Errorf("RSS.Feeds[0] = %q", cfg.RSS.Feeds[0])
	}
}

func TestLoad_AntigravityCLIConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")

	content := `output_dir: /tmp/test-output
llm:
  provider: antigravity-cli
  antigravity-cli:
    path: "~/bin/agy"
    timeout: 600
`
	if err := os.WriteFile(cfgFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgFile)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.LLM.AntigravityCLI.Timeout != 600 {
		t.Errorf("AntigravityCLI.Timeout = %d, want 600", cfg.LLM.AntigravityCLI.Timeout)
	}

	// Path must be tilde-expanded to an absolute path.
	if strings.HasPrefix(cfg.LLM.AntigravityCLI.Path, "~") {
		t.Errorf("AntigravityCLI.Path = %q, want tilde-expanded (no leading ~)", cfg.LLM.AntigravityCLI.Path)
	}
	home, err := os.UserHomeDir()
	if err == nil {
		want := filepath.Join(home, "bin/agy")
		if cfg.LLM.AntigravityCLI.Path != want {
			t.Errorf("AntigravityCLI.Path = %q, want %q", cfg.LLM.AntigravityCLI.Path, want)
		}
	}
}

func TestLLMConfig_OpenAICompat_Map_Parse(t *testing.T) {
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")

	content := `output_dir: /tmp/test-output
llm:
  openai-compat:
    default:
      endpoint: "http://localhost:8000/v1"
      model: "m-default"
    box1:
      endpoint: "http://192.168.1.10:1234/v1"
`
	if err := os.WriteFile(cfgFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgFile)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if got := cfg.LLM.OpenAICompat["default"].Endpoint; got != "http://localhost:8000/v1" {
		t.Errorf(`OpenAICompat["default"].Endpoint: got %q, want %q`, got, "http://localhost:8000/v1")
	}
	if got := cfg.LLM.OpenAICompat["default"].Model; got != "m-default" {
		t.Errorf(`OpenAICompat["default"].Model: got %q, want %q`, got, "m-default")
	}
	if got := cfg.LLM.OpenAICompat["box1"].Endpoint; got != "http://192.168.1.10:1234/v1" {
		t.Errorf(`OpenAICompat["box1"].Endpoint: got %q, want %q`, got, "http://192.168.1.10:1234/v1")
	}
}

func TestLLMConfig_OpenAICompat_OldSingleBlockShape_FailsToParse(t *testing.T) {
	// Pre-migration configs wrote openai-compat as a single struct block
	// (endpoint/model/api_key/timeout as direct siblings). That shape no
	// longer decodes into map[string]OpenAICompatConfig; Load() must
	// surface a parse error rather than silently dropping the fields.
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")

	content := `output_dir: /tmp/test-output
llm:
  openai-compat:
    endpoint: "http://127.0.0.1:8000/v1"
    model: "some-model"
    api_key: ""
    timeout: 900
`
	if err := os.WriteFile(cfgFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(cfgFile)
	if err == nil {
		t.Fatal("Load() with old single-block openai-compat shape: expected parse error, got nil")
	}
}

func TestLLMConfig_OpenAICompat_NoPhantomDefault_ThroughLoad(t *testing.T) {
	// DefaultConfig() seeds no "default" instance (TestDefaultConfig_OpenAICompat_NoSeed
	// guards that in isolation). This guards the actual risk: yaml.v3 merges into an
	// existing map rather than replacing it, so if DefaultConfig() ever seeded a
	// "default" instance again, a user config containing only box1 would still end up
	// with a phantom "default" key after Load() — silently defeating the resolver's
	// "bare openai-compat with no configured default -> error" contract.
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")

	content := `output_dir: /tmp/test-output
llm:
  openai-compat:
    box1:
      endpoint: "http://192.168.1.10:1234/v1"
`
	if err := os.WriteFile(cfgFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgFile)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if _, ok := cfg.LLM.OpenAICompat["default"]; ok {
		t.Error(`OpenAICompat: got phantom "default" instance after Load() with only box1 configured, want none`)
	}
	if len(cfg.LLM.OpenAICompat) != 1 {
		t.Errorf("OpenAICompat: got %d instances, want 1 (box1 only)", len(cfg.LLM.OpenAICompat))
	}
}

func TestLoad_NonexistentFile(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("Load() should not error for nonexistent file, got: %v", err)
	}
	// Should return DefaultConfig.
	def := DefaultConfig()
	if cfg.OutputDir == "" {
		t.Error("expected non-empty OutputDir from default config")
	}
	if cfg.LLM.Provider.Primary() != def.LLM.Provider.Primary() {
		t.Errorf("Provider = %q, want default %q", cfg.LLM.Provider.Primary(), def.LLM.Provider.Primary())
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "bad.yaml")

	// Write invalid YAML (tab-indented mapping value is typically fine,
	// but we use a truly broken structure).
	if err := os.WriteFile(cfgFile, []byte(":\n  - :\n  bad:: ["), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(cfgFile)
	if err == nil {
		t.Fatal("Load() should return error for invalid YAML")
	}
}

// --- ProviderList tests ---

func TestProviderList_UnmarshalYAML_Scalar(t *testing.T) {
	var cfg struct {
		Provider ProviderList `yaml:"provider"`
	}
	input := `provider: "gemini-cli"`
	if err := yaml.Unmarshal([]byte(input), &cfg); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(cfg.Provider) != 1 || cfg.Provider[0] != "gemini-cli" {
		t.Errorf("got %v, want [gemini-cli]", cfg.Provider)
	}
}

func TestProviderList_UnmarshalYAML_List(t *testing.T) {
	var cfg struct {
		Provider ProviderList `yaml:"provider"`
	}
	input := `provider:
  - gemini-cli
  - claude-code
  - ollama
`
	if err := yaml.Unmarshal([]byte(input), &cfg); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(cfg.Provider) != 3 {
		t.Fatalf("got %d providers, want 3", len(cfg.Provider))
	}
	if cfg.Provider[0] != "gemini-cli" || cfg.Provider[1] != "claude-code" || cfg.Provider[2] != "ollama" {
		t.Errorf("got %v, want [gemini-cli claude-code ollama]", cfg.Provider)
	}
}

func TestProviderList_Primary(t *testing.T) {
	p := ProviderList{"a", "b", "c"}
	if p.Primary() != "a" {
		t.Errorf("Primary() = %q, want a", p.Primary())
	}

	empty := ProviderList{}
	if empty.Primary() != "" {
		t.Errorf("Primary() on empty = %q, want empty", empty.Primary())
	}
}

func TestProviderList_Fallbacks(t *testing.T) {
	p := ProviderList{"a", "b", "c"}
	fb := p.Fallbacks()
	if len(fb) != 2 || fb[0] != "b" || fb[1] != "c" {
		t.Errorf("Fallbacks() = %v, want [b c]", fb)
	}

	single := ProviderList{"only"}
	if single.Fallbacks() != nil {
		t.Errorf("Fallbacks() on single = %v, want nil", single.Fallbacks())
	}

	empty := ProviderList{}
	if empty.Fallbacks() != nil {
		t.Errorf("Fallbacks() on empty = %v, want nil", empty.Fallbacks())
	}
}

func TestProviderList_SetPrimary(t *testing.T) {
	p := ProviderList{"a", "b", "c"}
	p.SetPrimary("d")
	if p.Primary() != "d" {
		t.Errorf("Primary() after SetPrimary = %q, want d", p.Primary())
	}
	if len(p) != 4 {
		t.Errorf("len after SetPrimary(new) = %d, want 4", len(p))
	}

	// SetPrimary with existing element should deduplicate.
	q := ProviderList{"a", "b", "c"}
	q.SetPrimary("b")
	if q.Primary() != "b" {
		t.Errorf("Primary() = %q, want b", q.Primary())
	}
	if len(q) != 3 {
		t.Errorf("len after SetPrimary(existing) = %d, want 3", len(q))
	}

	// SetPrimary on empty list.
	var r ProviderList
	r.SetPrimary("x")
	if r.Primary() != "x" || len(r) != 1 {
		t.Errorf("SetPrimary on empty: got %v", r)
	}
}

func TestProviderList_Contains(t *testing.T) {
	p := ProviderList{"Gemini-CLI", "Claude-Code"}
	if !p.Contains("gemini-cli") {
		t.Error("Contains should be case-insensitive")
	}
	if !p.Contains("CLAUDE-CODE") {
		t.Error("Contains should find CLAUDE-CODE")
	}
	if p.Contains("ollama") {
		t.Error("Contains should not find ollama")
	}
}

func TestDefaultConfig_SanityCheck(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.OutputDir == "" {
		t.Error("OutputDir should not be empty")
	}
	if cfg.LLM.Provider.Primary() == "" {
		t.Error("LLM.Provider should have a primary")
	}
	if cfg.Summary.Language == "" {
		t.Error("Summary.Language should not be empty")
	}
	if cfg.Summary.MaxTokens <= 0 {
		t.Error("Summary.MaxTokens should be positive")
	}
	if cfg.Extract.Timeout <= 0 {
		t.Error("Extract.Timeout should be positive")
	}
	if cfg.Pipeline.DelayMin <= 0 || cfg.Pipeline.DelayMax <= 0 {
		t.Error("Pipeline delays should be positive")
	}
	if !cfg.Pipeline.SkipExisting {
		t.Error("Pipeline.SkipExisting should default to true")
	}
}

func TestDefaultConfig_OpenAICompat_NoSeed(t *testing.T) {
	// DefaultConfig must NOT seed a "default" openai-compat instance: a phantom
	// default would survive yaml.v3 map-merge into user configs and make the
	// resolver's "bare openai-compat with no default -> error" contract
	// unreachable. Omitting the seed keeps code, docs, and tests consistent.
	cfg := DefaultConfig()
	if len(cfg.LLM.OpenAICompat) != 0 {
		t.Errorf("DefaultConfig OpenAICompat: got %d instances, want 0 (no seeded default)", len(cfg.LLM.OpenAICompat))
	}
}
