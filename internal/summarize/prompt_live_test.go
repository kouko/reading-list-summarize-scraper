package summarize

import (
	"os"
	"strings"
	"testing"

	"github.com/kouko/reading-list-summarize-scraper/internal/config"
)

// TestSummaryPromptLive renders the summary prompt against a sample article and
// runs it through the real configured provider chain, printing the result so
// you can eyeball prompt edits (e.g. tuning the article-style built-ins). It is
// a manual dev harness, NOT part of the default suite: it is skipped unless
// RLSS_LIVE_PROMPT=1 because it makes a real LLM call (costs tokens, needs an
// authenticated provider) — so it never runs in `go test ./...` or CI.
//
// It goes through the exact app path — ResolvePrompt -> SubstituteVars ->
// NewSummarizer -> Summarize — so what you see matches the `rlss` summary
// stage (no extract / keywords / mermaid).
//
// Fast iteration loop (edit a prompt file, rerun this one test):
//
//	RLSS_LIVE_PROMPT=1 RLSS_PROMPT_FILE=./prompts/my-summary.md \
//	  go test ./internal/summarize/ -run TestSummaryPromptLive -v
//
// Tune a built-in style instead (article/outline) via your config's summary.style:
//
//	RLSS_LIVE_PROMPT=1 RLSS_CONFIG=/abs/path/config.yaml \
//	  go test ./internal/summarize/ -run TestSummaryPromptLive -v
//
// Env knobs (only the gate is required):
//
//	RLSS_LIVE_PROMPT=1           opt in to the real LLM call
//	RLSS_CONFIG=../../config.yaml  config for the provider chain + summary.style/language
//	                               (default; repo-root config relative to this package dir)
//	RLSS_PROMPT_FILE=path        prompt template to test; default = app resolution (built-in by language+style)
//	RLSS_ARTICLE=path            article text to summarize; default = a short sample
func TestSummaryPromptLive(t *testing.T) {
	if os.Getenv("RLSS_LIVE_PROMPT") == "" {
		t.Skip("set RLSS_LIVE_PROMPT=1 to run (makes a real LLM call)")
	}

	// `go test ./internal/summarize/` runs with CWD = the package dir, so the
	// repo-root config.yaml is two levels up. Override with an absolute
	// RLSS_CONFIG if your layout differs.
	cfgPath := liveEnvOr("RLSS_CONFIG", "../../config.yaml")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Skipf("need a config to pick the provider chain (%s): %v — set RLSS_CONFIG to an absolute path", cfgPath, err)
	}

	// Prompt template: explicit file override, else the app's own resolution
	// (summary_prompt_file / inline prompt / built-in by language+style).
	var tmpl string
	if pf := os.Getenv("RLSS_PROMPT_FILE"); pf != "" {
		b, rerr := os.ReadFile(pf)
		if rerr != nil {
			t.Fatalf("read RLSS_PROMPT_FILE %q: %v", pf, rerr)
		}
		tmpl = string(b)
	} else {
		tmpl, err = ResolvePrompt(cfg.Summary)
		if err != nil {
			t.Fatalf("resolve prompt: %v", err)
		}
	}

	article := sampleArticle
	if af := os.Getenv("RLSS_ARTICLE"); af != "" {
		b, rerr := os.ReadFile(af)
		if rerr != nil {
			t.Fatalf("read RLSS_ARTICLE %q: %v", af, rerr)
		}
		article = string(b)
	}

	// Mirror the pipeline's variable-building so the rendered prompt matches
	// what the app sends.
	prompt := SubstituteVars(tmpl, PromptVars{
		Title:         "Sample Article",
		Domain:        "example.com",
		Language:      cfg.Summary.Language,
		Content:       article,
		ContentLength: len([]rune(article)),
	})

	s, err := NewSummarizer(cfg.LLM)
	if err != nil {
		t.Fatalf("new summarizer: %v", err)
	}

	res, err := s.Summarize(article, SummarizeOptions{
		Prompt:    prompt,
		MaxTokens: cfg.Summary.MaxTokens,
	})
	if err != nil {
		t.Fatalf("summarize: %v", err)
	}

	t.Logf("provider=%q model=%q response_len=%d", res.Provider, res.Model, len(res.Text))
	t.Logf("\n----- summary -----\n%s\n-------------------", res.Text)

	if strings.TrimSpace(res.Text) == "" {
		t.Fatal("provider returned an EMPTY summary")
	}
}

func liveEnvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

const sampleArticle = "Small modular reactors (SMRs) are compact nuclear reactors that can be " +
	"factory-built and shipped to site, in contrast to bespoke gigawatt-scale plants. " +
	"Proponents argue they cut upfront capital cost and construction risk; critics note that, " +
	"as of 2026, almost none are operating commercially and the per-MWh economics remain unproven. " +
	"The article walks through the technology, the leading vendors, and the regulatory hurdles."
