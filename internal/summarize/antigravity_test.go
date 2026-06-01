package summarize

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// writeFakeAgy writes an executable shell script to the test's temp dir and
// returns its path. The script is the injection seam: AntigravityCLISummarizer
// runs binaryPath directly (LookPath is skipped when binaryPath != ""), so the
// fake stands in for the real agy CLI. body is the script after the shebang.
func writeFakeAgy(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "agy")
	script := "#!/bin/sh\n" + body
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake agy: %v", err)
	}
	return path
}

// TestAntigravityCLI_Summarize_SuccessPostProcessing pins the success path:
// stdout flows through StripThinkingTags then stripSourceInjection, and the
// result carries the fixed provider/model identity. agy headless cannot report
// a model, so Model is always "".
func TestAntigravityCLI_Summarize_SuccessPostProcessing(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake-binary test is POSIX-only; rlss targets macOS")
	}
	// printf keeps the literal \n as a real newline in stdout. The <think>
	// block and （來源…） tail-note must both be stripped from Text.
	fake := writeFakeAgy(t, `cat >/dev/null; printf '<think>ignore</think>### 概要\n結論（來源：[x.com]）'`)
	s := &AntigravityCLISummarizer{binaryPath: fake, timeout: time.Minute}

	got, err := s.Summarize("article body", SummarizeOptions{})
	if err != nil {
		t.Fatalf("Summarize returned error: %v", err)
	}
	if strings.Contains(got.Text, "<think>") {
		t.Errorf("Text still contains <think> tag: %q", got.Text)
	}
	if strings.Contains(got.Text, "（來源") {
		t.Errorf("Text still contains source injection note: %q", got.Text)
	}
	if want := "### 概要\n結論"; got.Text != want {
		t.Errorf("Text = %q, want %q", got.Text, want)
	}
	if got.Provider != "antigravity-cli" {
		t.Errorf("Provider = %q, want %q", got.Provider, "antigravity-cli")
	}
	if got.Model != "" {
		t.Errorf("Model = %q, want empty", got.Model)
	}
}

// TestAntigravityCLI_Summarize_ExecutionFailure pins the non-quota failure
// path: a generic stderr message that matches no quota pattern yields a plain
// error, NOT a *QuotaError, so the circuit breaker treats it as a hard failure.
func TestAntigravityCLI_Summarize_ExecutionFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake-binary test is POSIX-only; rlss targets macOS")
	}
	fake := writeFakeAgy(t, `echo "agy: unexpected internal error" >&2; exit 1`)
	s := &AntigravityCLISummarizer{binaryPath: fake, timeout: time.Minute}

	_, err := s.Summarize("article body", SummarizeOptions{})
	if err == nil {
		t.Fatal("Summarize returned nil error on agy exit 1")
	}
	if IsQuotaError(err) {
		t.Errorf("IsQuotaError = true for generic failure; want false. err: %v", err)
	}
}

// TestAntigravityCLI_Summarize_QuotaError pins quota detection: a stderr
// message matching a quota pattern ("rate limit") is wrapped as *QuotaError
// tagged with the antigravity-cli provider, so the breaker can fall through.
func TestAntigravityCLI_Summarize_QuotaError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake-binary test is POSIX-only; rlss targets macOS")
	}
	fake := writeFakeAgy(t, `echo "Error 429: rate limit exceeded" >&2; exit 1`)
	s := &AntigravityCLISummarizer{binaryPath: fake, timeout: time.Minute}

	_, err := s.Summarize("article body", SummarizeOptions{})
	if err == nil {
		t.Fatal("Summarize returned nil error on quota failure")
	}
	if !IsQuotaError(err) {
		t.Fatalf("IsQuotaError = false; want true. err: %v", err)
	}
	var qe *QuotaError
	if !errors.As(err, &qe) {
		t.Fatalf("error is not *QuotaError: %v", err)
	}
	if qe.Provider != "antigravity-cli" {
		t.Errorf("QuotaError.Provider = %q, want %q", qe.Provider, "antigravity-cli")
	}
}

// TestAntigravityCLI_Summarize_BinaryNotFound pins behavior when binaryPath
// points at a nonexistent file: LookPath is skipped (binaryPath is set), so the
// failure surfaces at cmd.Run() as a returned error rather than a panic.
func TestAntigravityCLI_Summarize_BinaryNotFound(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake-binary test is POSIX-only; rlss targets macOS")
	}
	s := &AntigravityCLISummarizer{binaryPath: "/nonexistent/agy-xyz", timeout: time.Minute}

	_, err := s.Summarize("article body", SummarizeOptions{})
	if err == nil {
		t.Fatal("Summarize returned nil error for nonexistent binary")
	}
}

func TestStripSourceInjection(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "tail note removed",
			input: "提升開發效率與體驗（來源：[ai-souken.com]、[antigravity.google]）",
			want:  "提升開發效率與體驗",
		},
		{
			name:  "multiple notes removed",
			input: "第一段（來源：a）中間文字第二段（來源：b）",
			want:  "第一段中間文字第二段",
		},
		{
			name:  "no note unchanged",
			input: "純文字沒有來源備註",
			want:  "純文字沒有來源備註",
		},
		{
			name:  "no note trailing whitespace trimmed",
			input: "純文字沒有來源備註\n\n  ",
			want:  "純文字沒有來源備註",
		},
		{
			name:  "inline note within paragraph removed",
			input: "開頭（來源：[x.com]）結尾繼續寫",
			want:  "開頭結尾繼續寫",
		},
		{
			// Pins the intentional broad match: 來源 prefix triggers removal
			// even without a colon, since agy emits both （來源：…） and （來源…）.
			name:  "note without colon removed",
			input: "結論（來源x.com）",
			want:  "結論",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripSourceInjection(tt.input)
			if got != tt.want {
				t.Errorf("stripSourceInjection(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestAntigravityCLI_InterfaceConformance(t *testing.T) {
	var _ Summarizer = (*AntigravityCLISummarizer)(nil)
}
