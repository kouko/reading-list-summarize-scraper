package config

import (
	"os"
	"path/filepath"
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
