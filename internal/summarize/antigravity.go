package summarize

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// sourceInjectionRe matches （來源：…） tail-notes that the Antigravity CLI (agy)
// occasionally appends. Full-width parens （）, prefix 來源, non-greedy match to
// the closing ）. (?s) lets the group span newlines.
var sourceInjectionRe = regexp.MustCompile(`(?s)（來源.*?）`)

// stripSourceInjection removes every （來源…） note from s and trims surrounding
// whitespace. All other text is left intact.
func stripSourceInjection(s string) string {
	result := sourceInjectionRe.ReplaceAllString(s, "")
	return strings.TrimSpace(result)
}

// AntigravityCLISummarizer invokes the Antigravity CLI (agy) in headless mode.
// Prompt is passed via stdin and via -p. agy headless cannot select or report
// a model, so Model is always empty.
type AntigravityCLISummarizer struct {
	binaryPath string
	timeout    time.Duration
}

func (a *AntigravityCLISummarizer) Summarize(text string, opts SummarizeOptions) (SummarizeResult, error) {
	combinedPrompt := resolvePrompt(text, opts)

	binary := a.binaryPath
	if binary == "" {
		var err error
		binary, err = exec.LookPath("agy")
		if err != nil {
			return SummarizeResult{}, fmt.Errorf("antigravity-cli: binary not found in PATH: %w", err)
		}
	}

	timeout := a.timeout
	if timeout == 0 {
		timeout = 15 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// agy expects --print-timeout in whole minutes. Round UP so agy's budget
	// always covers the Go context deadline (a sub-minute timeout like 90s must
	// not truncate to 1m and make agy abort before ctx fires).
	minutes := int((timeout + time.Minute - 1) / time.Minute)
	if minutes < 1 {
		minutes = 1
	}

	// combinedPrompt is passed BOTH as the -p value and via stdin (below): agy
	// reads stdin as additional context while -p carries the instruction, so
	// the long prompt survives even if it would exceed -p's arg limits. This
	// duplication is intentional (verified: brief §Boundary), not redundant.
	// agy has no -m/-o/--approval-mode/MCP flags (verified: brief §Boundary).
	args := []string{"-p", combinedPrompt, "--print-timeout", fmt.Sprintf("%dm", minutes)}
	cmd := exec.CommandContext(ctx, binary, args...)

	// Isolate side effects: agy writes .antigravitycli/ into its cwd. Point it
	// at a per-call temp dir so the user's working directory stays clean.
	tmpDir, err := os.MkdirTemp("", "rlss-agy-*")
	if err != nil {
		return SummarizeResult{}, fmt.Errorf("antigravity-cli: failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	cmd.Dir = tmpDir

	var stdout, stderr bytes.Buffer
	cmd.Stdin = strings.NewReader(combinedPrompt)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// agy is agent-first and may write its failure reason to stdout; fall
		// back to stdout when stderr is empty (mirrors claude_code.go).
		errMsg := stderr.String()
		if errMsg == "" {
			errMsg = stdout.String()
		}
		baseErr := fmt.Errorf("antigravity-cli: execution failed: %w\noutput: %s", err, errMsg)
		if isQuotaMessage(stderr.String() + "\n" + stdout.String()) {
			return SummarizeResult{}, &QuotaError{Provider: "antigravity-cli", Err: baseErr}
		}
		return SummarizeResult{}, baseErr
	}

	return SummarizeResult{
		Text:     stripSourceInjection(StripThinkingTags(strings.TrimSpace(stdout.String()))),
		Provider: "antigravity-cli",
		Model:    "",
	}, nil
}
