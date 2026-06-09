package pipeline

import (
	"strings"
	"testing"

	"github.com/kouko/reading-list-summarize-scraper/internal/summarize"
)

func TestEmptyResponseError_NamesProviderAndModel(t *testing.T) {
	err := emptyResponseError("stage 1", "qwen-code", "coder-model")
	if err == nil {
		t.Fatal("expected a non-nil error")
	}
	msg := err.Error()
	for _, want := range []string{"stage 1", "empty response", "qwen-code", "coder-model"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error message %q is missing %q", msg, want)
		}
	}
}

func TestEmptyResponseError_StillReadsWithUnknownProvider(t *testing.T) {
	// Defensive: even if provider/model are empty, the message must still
	// identify it as an empty-response failure rather than crash or read blank.
	err := emptyResponseError("stage 1", "", "")
	if err == nil || !strings.Contains(err.Error(), "empty response") {
		t.Errorf("expected an empty-response error, got: %v", err)
	}
}

func TestValidateSummaryText_EmptyReturnsNamedError(t *testing.T) {
	// A blank or whitespace-only LLM response is not a usable summary; it must
	// produce a named error (so the caller can mark the item partial/resumable)
	// rather than be written out as an empty summary file.
	for _, empty := range []string{"", "   ", "\n\t  \n"} {
		err := validateSummaryText(empty, "ollama", "llama3")
		if err == nil {
			t.Errorf("validateSummaryText(%q) = nil, want a named empty-response error", empty)
			continue
		}
		msg := err.Error()
		if !strings.Contains(msg, "ollama") || !strings.Contains(msg, "llama3") {
			t.Errorf("validateSummaryText(%q) error %q should name provider/model", empty, msg)
		}
	}
}

func TestValidateSummaryText_NonEmptyReturnsNil(t *testing.T) {
	if err := validateSummaryText("## Summary\nreal content", "claude-api", "haiku"); err != nil {
		t.Errorf("validateSummaryText(non-empty) = %v, want nil", err)
	}
}

func TestValidateSummaryText_ThinkingOnlyResponseIsEmpty(t *testing.T) {
	// The real-world #51 trigger: a thinking-model response that is ONLY a
	// <think>…</think> block strips to empty, so the post-strip text the
	// pipeline validates is "" and must be treated as an empty response
	// (partial), not written out as a summary.
	stripped := summarize.StripThinkingTags("<think>reasoning, but no actual answer produced</think>")
	if err := validateSummaryText(stripped, "qwen-code", "coder-model"); err == nil {
		t.Errorf("thinking-only response stripped to %q should be a named empty-response error", stripped)
	}
}
