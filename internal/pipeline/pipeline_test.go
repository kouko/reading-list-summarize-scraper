package pipeline

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

// --- isBlockedPage tests ---

func TestIsBlockedPage_CloudflareChallenge(t *testing.T) {
	content := "Please wait while we verify your browser. Checking your browser before accessing the site. Powered by Cloudflare."
	if !isBlockedPage(content) {
		t.Error("expected blocked detection for Cloudflare challenge page")
	}
}

func TestIsBlockedPage_CaptchaAndAccessDenied(t *testing.T) {
	content := "Access Denied. Please complete the CAPTCHA to continue."
	if !isBlockedPage(content) {
		t.Error("expected blocked detection for CAPTCHA + Access Denied")
	}
}

func TestIsBlockedPage_BotProtection(t *testing.T) {
	content := "Bot protection is enabled. Please verify you are a human to proceed."
	if !isBlockedPage(content) {
		t.Error("expected blocked detection for bot protection page")
	}
}

func TestIsBlockedPage_NormalContent(t *testing.T) {
	content := `# Introduction to Go Programming

Go is a statically typed, compiled language designed at Google.
It provides excellent concurrency support through goroutines and channels.

## Getting Started

To install Go, visit the official website at golang.org.`

	if isBlockedPage(content) {
		t.Error("normal article content should not be detected as blocked")
	}
}

func TestIsBlockedPage_SinglePatternMatch(t *testing.T) {
	// Only one match should NOT trigger blocked detection (requires >= 2).
	content := "This article discusses cloudflare as a CDN provider for websites."
	if isBlockedPage(content) {
		t.Error("single pattern match should not be detected as blocked")
	}
}

func TestIsBlockedPage_EmptyContent(t *testing.T) {
	if isBlockedPage("") {
		t.Error("empty content should not be detected as blocked")
	}
}

func TestIsBlockedPage_CaseInsensitive(t *testing.T) {
	content := "CLOUDFLARE Security Challenge detected"
	if !isBlockedPage(content) {
		t.Error("detection should be case-insensitive")
	}
}

// --- IsSkipped / errSkipped tests ---

func TestIsSkipped_True(t *testing.T) {
	if !IsSkipped(errSkipped) {
		t.Error("IsSkipped(errSkipped) should return true")
	}
}

func TestIsSkipped_False(t *testing.T) {
	if IsSkipped(errors.New("some other error")) {
		t.Error("IsSkipped should return false for other errors")
	}
	if IsSkipped(nil) {
		t.Error("IsSkipped(nil) should return false")
	}
}

// --- Stats tests ---

func TestStats_Duration(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(5 * time.Minute)
	s := Stats{Start: start, End: end}
	if s.Duration() != 5*time.Minute {
		t.Errorf("Duration() = %v, want 5m", s.Duration())
	}
}

func TestStats_Report_NoErrors(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(30 * time.Second)
	s := Stats{
		Success: 5,
		Skipped: 2,
		Failed:  0,
		Start:   start,
		End:     end,
	}

	report := s.Report()
	if !strings.Contains(report, "Total:   7") {
		t.Errorf("report missing total count, got:\n%s", report)
	}
	if !strings.Contains(report, "Success: 5") {
		t.Errorf("report missing success count, got:\n%s", report)
	}
	if !strings.Contains(report, "Skipped: 2") {
		t.Errorf("report missing skipped count, got:\n%s", report)
	}
	if !strings.Contains(report, "Failed:  0") {
		t.Errorf("report missing failed count, got:\n%s", report)
	}
	if strings.Contains(report, "Errors:") {
		t.Error("report should not contain Errors section when there are none")
	}
}

func TestStats_Report_WithErrors(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(1 * time.Minute)
	s := Stats{
		Success: 3,
		Skipped: 1,
		Failed:  2,
		Errors: []ItemError{
			{URL: "https://example.com/a", Title: "Article A", Err: fmt.Errorf("timeout")},
			{URL: "https://example.com/b", Title: "", Err: fmt.Errorf("blocked")},
		},
		Start: start,
		End:   end,
	}

	report := s.Report()
	if !strings.Contains(report, "Errors:") {
		t.Errorf("report should contain Errors section, got:\n%s", report)
	}
	if !strings.Contains(report, "Article A: timeout") {
		t.Errorf("report should contain error detail for Article A, got:\n%s", report)
	}
	// When title is empty, URL should be used.
	if !strings.Contains(report, "https://example.com/b: blocked") {
		t.Errorf("report should show URL when title is empty, got:\n%s", report)
	}
}

// --- stripFrontmatter tests ---

func TestStripFrontmatter_WithFrontmatter(t *testing.T) {
	input := "---\ntitle: Test\ndate: 2025-01-01\n---\n\nActual content here."
	got := stripFrontmatter(input)
	want := "Actual content here."
	if got != want {
		t.Errorf("stripFrontmatter() = %q, want %q", got, want)
	}
}

func TestStripFrontmatter_WithoutFrontmatter(t *testing.T) {
	input := "Just plain content without frontmatter."
	got := stripFrontmatter(input)
	if got != input {
		t.Errorf("stripFrontmatter() = %q, want unchanged input", got)
	}
}

func TestStripFrontmatter_Empty(t *testing.T) {
	got := stripFrontmatter("")
	if got != "" {
		t.Errorf("stripFrontmatter(\"\") = %q, want empty", got)
	}
}

func TestStripFrontmatter_OnlyOpeningDelimiter(t *testing.T) {
	input := "---\ntitle: Test\nNo closing delimiter"
	got := stripFrontmatter(input)
	if got != input {
		t.Errorf("stripFrontmatter() should return unchanged when no closing delimiter, got %q", got)
	}
}

func TestStripFrontmatter_DashesInContent(t *testing.T) {
	// Frontmatter followed by content that also contains ---
	input := "---\nkey: val\n---\n\nContent with --- dashes in it."
	got := stripFrontmatter(input)
	want := "Content with --- dashes in it."
	if got != want {
		t.Errorf("stripFrontmatter() = %q, want %q", got, want)
	}
}

func TestStripFrontmatter_EmptyFrontmatter(t *testing.T) {
	// "---\n---\n" has no content between delimiters. After skipping the
	// opening "---\n", content[4:] = "---\n\n..." and "\n---\n" is not found
	// (the closing --- is at the start, not after a newline). So the function
	// returns the input unchanged -- this is correct behavior since the
	// closing delimiter pattern requires a preceding newline.
	input := "---\n---\n\nContent after empty frontmatter."
	got := stripFrontmatter(input)
	if got != input {
		t.Errorf("stripFrontmatter() = %q, want unchanged input", got)
	}
}

func TestStripFrontmatter_MinimalFrontmatter(t *testing.T) {
	// Minimal valid frontmatter with a single key.
	input := "---\nk: v\n---\n\nBody text."
	got := stripFrontmatter(input)
	want := "Body text."
	if got != want {
		t.Errorf("stripFrontmatter() = %q, want %q", got, want)
	}
}

func TestIsUnavailablePage_MaintenanceAndWalls(t *testing.T) {
	unavailable := []string{
		// nikkei headless bot-served maintenance page (the dogfood case)
		"非常感謝您的訪問！現在日經中文網正在系統維護中。請在稍晚時間嘗試訪問本網站。",
		"现在日经中文网正在系统维护中。请在稍晚时间尝试访问。",
		"The site is currently under maintenance. Please try again later.",
		"ただいまメンテナンス中です。しばらくしてから再度お試しください。",
		// login / paywall walls
		"Please sign in to read the full article.",
		"Subscribe to read more.",
		"Subscribe to continue reading this article.",
		"この記事を読むにはログインしてください。",
		"請登入後閱讀全文，或訂閱成為會員。",
	}
	for _, c := range unavailable {
		if !isUnavailablePage(c) {
			t.Errorf("isUnavailablePage should be true for a wall page: %q", c)
		}
	}

	// Real article content (incl. prose that merely mentions the topics) must NOT match.
	normal := []string{
		"This article explains how scheduled server maintenance windows reduce downtime.",
		"本文討論付費牆（paywall）對新聞業的影響，以及訂閱制的興衰。",
		"ログイン機能の実装方法を解説します。セッション管理とCookieについて。",
		"A long, normal article body with several paragraphs of real content and no wall.",
	}
	for _, c := range normal {
		if isUnavailablePage(c) {
			t.Errorf("isUnavailablePage should be false for normal content: %q", c)
		}
	}
}
