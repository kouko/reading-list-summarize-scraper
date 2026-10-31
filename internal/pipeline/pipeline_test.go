package pipeline

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kouko/reading-list-summarize-scraper/internal/config"
	"github.com/kouko/reading-list-summarize-scraper/internal/extract"
	"github.com/kouko/reading-list-summarize-scraper/internal/output"
	"github.com/kouko/reading-list-summarize-scraper/internal/source"
	"github.com/kouko/reading-list-summarize-scraper/internal/summarize"
)

func TestProcessItemRejectsCachedWall(t *testing.T) {
	for _, wall := range []string{
		"## Log in or sign up for X\n\n[Log in](https://x.com/i/flow/login)",
		"Please sign in to read the full article.\n\n[Privacy](https://example.com/privacy)",
		"Please sign in to read the full article.\n\n[Forgot password?](https://example.com/reset)",
		"## Log in or sign up for X\n\n[Log in](https://x.com/i/flow/login)\n\n[Privacy](https://example.com/privacy)",
	} {
		t.Run(wall, func(t *testing.T) {
			s := &stubExtractor{first: wall, retry: wall}
			if got, err := extractWithFallback(s, "https://example.com/article", true); err == nil || got != "" || len(s.calls) != 2 {
				t.Errorf("expected initial/retry rejection, got %q, %v, calls=%v", got, err, s.calls)
			}
			dir := t.TempDir()
			item := source.ReadingItem{URL: "https://x.com/example/status/123"}
			domainDir := filepath.Join(dir, output.DomainDir(item.URL))
			if err := os.MkdirAll(domainDir, 0755); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(domainDir, output.ContentFilename(time.Now(), output.SHA8(item.URL)))
			content := "---\ntype: reading-list-content\n---\n\n" + wall
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				t.Fatal(err)
			}
			p := New(&config.Config{OutputDir: dir}, s, nil)
			if err := p.ProcessItem(item); err == nil {
				t.Fatalf("expected extraction failure, got %v", err)
			}
			if got, err := os.ReadFile(path); err != nil || string(got) != content {
				t.Fatalf("existing note changed: %q, %v", got, err)
			}
		})
	}
}

type recordingSummary struct {
	bodies []string
	result *summarize.SummarizeResult
}

func (s *recordingSummary) Summarize(body string, _ summarize.SummarizeOptions) (summarize.SummarizeResult, error) {
	s.bodies = append(s.bodies, body)
	if s.result != nil {
		return *s.result, nil
	}
	return summarize.SummarizeResult{}, errors.New("summary unavailable")
}

func TestProcessItemSocialCache(t *testing.T) {
	for _, tt := range []struct {
		name, url                               string
		completed, force, dryRun, fail, extract bool
	}{
		{"verified-social-resume", "https://x.com/example/status/123", false, false, false, false, true},
		{"failed-refresh-preserves-files", "https://www.threads.com/@example/post/abc", false, true, false, true, true},
		{"completed-skip", "https://twitter.com/example/status/123", true, false, false, false, false},
		{"completed-force", "https://mobile.twitter.com/example/status/123", true, true, false, true, true},
		{"dry-run", "https://www.threads.net/@example/post/abc", false, true, true, false, false},
		{"ordinary-cache-summary-partial", "https://example.com/article", false, true, false, false, false},
		{"spoof-host-cache", "https://x.com.example.org/article", false, true, false, false, false},
		{"scoped-cache-fresh-facebook", "https://facebook.com/writer/posts/456/", false, false, false, false, true},
		{"scoped-cache-fresh-facebook-www", "https://www.facebook.com/writer/posts/456/", false, false, false, false, true},
		{"scoped-cache-fresh-felix", "https://felixrieseberg.com/another-article/", false, false, false, false, true},
		{"scoped-cache-fresh-felix-www", "https://www.felixrieseberg.com/another-article/", false, false, false, false, true},
		{"scoped-cache-failure-preserves", "https://facebook.com/writer/posts/456/", false, false, false, true, true},
		{"scoped-force-failure-preserves", "https://felixrieseberg.com/another-article/", true, true, false, true, true},
		{"scoped-completed-skip", "https://facebook.com/writer/posts/456/", true, false, false, false, false},
		{"scoped-dry-run", "https://felixrieseberg.com/another-article/", false, false, true, false, false},
		{"scoped-spoof-host-cache", "https://facebook.com.example.org/article", false, false, false, false, false},
		{"scoped-unrelated-subdomain-cache", "https://other.felixrieseberg.com/article", false, false, false, false, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			item := source.ReadingItem{URL: tt.url}
			out := filepath.Join(dir, output.DomainDir(item.URL))
			if err := os.MkdirAll(out, 0755); err != nil {
				t.Fatal(err)
			}
			sha, now := output.SHA8(item.URL), time.Now()
			files := map[string]string{output.ContentFilename(now, sha): "---\ntype: reading-list-content\n---\n\nOld partial body"}
			if tt.completed {
				files[output.SummaryFilename(now, sha)] = "Existing summary"
			}
			for name, body := range files {
				if err := os.WriteFile(filepath.Join(out, name), []byte(body), 0644); err != nil {
					t.Fatal(err)
				}
			}
			s := &stubExtractor{first: "Verified complete author article", retry: "Verified complete author article"}
			if tt.name == "verified-social-resume" {
				s.first, s.retry = verifiedXPost, verifiedXPost
			}
			if tt.fail {
				s.firstErr = errors.New("Inaccessible social target: Cloudflare CAPTCHA")
				s.retryErr = s.firstErr
			}
			sum := &recordingSummary{}
			cfg := &config.Config{OutputDir: dir, Pipeline: config.PipelineConfig{SkipExisting: !tt.force, DryRun: tt.dryRun}}
			err := New(cfg, s, sum).ProcessItem(item)
			if (len(s.calls) > 0) != tt.extract {
				t.Fatalf("extraction calls=%v want extract=%v", s.calls, tt.extract)
			}
			if tt.fail {
				if err == nil || IsPartial(err) || len(sum.bodies) != 0 || len(s.calls) != 1 {
					t.Fatalf("failed refresh summarized or retried blocked page: err=%v calls=%v summaries=%v", err, s.calls, sum.bodies)
				}
			} else if tt.dryRun || tt.completed {
				if !IsSkipped(err) || len(sum.bodies) != 0 {
					t.Fatalf("skip changed: err=%v summaries=%v", err, sum.bodies)
				}
			} else {
				want := "Old partial body"
				if tt.extract {
					want = s.first
				}
				if !IsPartial(err) || len(sum.bodies) != 1 || sum.bodies[0] != want {
					t.Fatalf("summary input/classification changed: err=%v summaries=%v", err, sum.bodies)
				}
			}
			if tt.fail || !tt.extract {
				entries, readErr := os.ReadDir(out)
				if readErr != nil || len(entries) != len(files) {
					t.Fatalf("files changed: %v %v", entries, readErr)
				}
				for name, want := range files {
					got, readErr := os.ReadFile(filepath.Join(out, name))
					if readErr != nil || string(got) != want {
						t.Fatalf("existing file changed: %s %q %v", name, got, readErr)
					}
				}
			}
		})
	}
}

// --- isBlockedPage tests ---

func TestIsBlockedPage_CloudflareChallenge(t *testing.T) {
	content := "Please wait while we verify your browser. Checking your browser before accessing the site. Powered by Cloudflare."
	if !extract.IsBlockedPage(content) {
		t.Error("expected blocked detection for Cloudflare challenge page")
	}
}

func TestIsBlockedPage_CaptchaAndAccessDenied(t *testing.T) {
	content := "Access Denied. Please complete the CAPTCHA to continue."
	if !extract.IsBlockedPage(content) {
		t.Error("expected blocked detection for CAPTCHA + Access Denied")
	}
}

func TestIsBlockedPage_BotProtection(t *testing.T) {
	content := "Bot protection is enabled. Please verify you are a human to proceed."
	if !extract.IsBlockedPage(content) {
		t.Error("expected blocked detection for bot protection page")
	}
}

func TestIsBlockedPage_NormalContent(t *testing.T) {
	content := `# Introduction to Go Programming

Go is a statically typed, compiled language designed at Google.
It provides excellent concurrency support through goroutines and channels.

## Getting Started

To install Go, visit the official website at golang.org.`

	if extract.IsBlockedPage(content) {
		t.Error("normal article content should not be detected as blocked")
	}
}

func TestIsBlockedPage_SinglePatternMatch(t *testing.T) {
	// Only one match should NOT trigger blocked detection (requires >= 2).
	content := "This article discusses cloudflare as a CDN provider for websites."
	if extract.IsBlockedPage(content) {
		t.Error("single pattern match should not be detected as blocked")
	}
}

func TestIsBlockedPage_EmptyContent(t *testing.T) {
	if extract.IsBlockedPage("") {
		t.Error("empty content should not be detected as blocked")
	}
}

func TestIsBlockedPage_CaseInsensitive(t *testing.T) {
	content := "CLOUDFLARE Security Challenge detected"
	if !extract.IsBlockedPage(content) {
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
		if err := extract.ValidateContent(c); err == nil {
			t.Errorf("ValidateContent should reject a wall page: %q", c)
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
		if err := extract.ValidateContent(c); err != nil {
			t.Errorf("ValidateContent should accept normal content %q: %v", c, err)
		}
	}
}

func TestIsBlockedPage_MultiLanguage(t *testing.T) {
	// Localized anti-bot challenge pages (no English Cloudflare branding) — each
	// has >=2 challenge-specific UI phrases, so they should be detected.
	blocked := []string{
		"ブラウザを確認しています。あなたが人間であることを確認します。",   // ja
		"アクセスが拒否されました。ロボットではないことを確認してください。", // ja (other phrases)
		"正在檢查您的瀏覽器，請完成安全驗證後繼續。",             // zh-Hant
		"拒絕存取，請確認您是真人後再試。",                  // zh-Hant (other phrases)
		"正在检查您的浏览器，请完成安全验证。",                // zh-Hans
		"拒绝访问，请确认您是真人。",                     // zh-Hans (other phrases)
	}
	for _, c := range blocked {
		if !extract.IsBlockedPage(c) {
			t.Errorf("isBlockedPage should detect a localized challenge page: %q", c)
		}
	}

	// Real articles that merely discuss these topics must NOT be flagged
	// (one topic word + an existing English pattern must not reach the >=2 bar).
	normal := []string{
		"本文討論人機驗證（CAPTCHA）的演進與存取控制機制。",    // captcha(1) only → not blocked
		"セキュリティチェックの自動化とアクセス制御について解説します。", // no challenge-UI phrase
		"## to, err:= human()\n\nThis article explains Go error handling and cookie policies.",
	}
	for _, c := range normal {
		if extract.IsBlockedPage(c) {
			t.Errorf("isBlockedPage should NOT flag a normal article: %q", c)
		}
	}
}

func TestProcessItemFocalScope(t *testing.T) {
	for _, kind := range []string{"summary", "embedded-summary", "unlabelled", "wrong-target", "thread-scope", "changed-body", "missing-notice", "threads"} {
		t.Run(kind, func(t *testing.T) {
			url := "https://x.com/example/status/123"
			body := verifiedXPost
			switch kind {
			case "unlabelled":
				body = "An unverified legacy cached body."
			case "wrong-target":
				body = strings.Replace(body, `"target":"123"`, `"target":"456"`, 1)
			case "thread-scope":
				body = strings.Replace(body, `"scope":"single_post"`, `"scope":"complete_thread"`, 1)
			case "changed-body":
				body = strings.Replace(body, "workers exchange files", "someone else's post", 1)
			case "missing-notice":
				body = strings.Replace(body, extract.SocialScopeNotice, "", 1)
			case "threads":
				url = "https://www.threads.com/@example/post/abc"
			}
			cfg := config.DefaultConfig()
			cfg.OutputDir = t.TempDir()
			cfg.Pipeline.SkipExisting = false
			cfg.Summary.Keywords.Enabled = false
			cfg.Summary.Mermaid.Enabled = false
			cfg.Summary.EmbedContent = kind == "embedded-summary"
			cfg.Extract.MinContentLength = 0
			item := source.ReadingItem{URL: url, Title: "Focal post"}
			out := filepath.Join(cfg.OutputDir, output.DomainDir(url))
			if err := os.MkdirAll(out, 0755); err != nil {
				t.Fatal(err)
			}
			now, sha := time.Now(), output.SHA8(url)
			contentPath, summaryPath := filepath.Join(out, output.ContentFilename(now, sha)), filepath.Join(out, output.SummaryFilename(now, sha))
			for path, text := range map[string]string{contentPath: "---\ntype: reading-list-content\n---\n\nOld body", summaryPath: "Old summary"} {
				if err := os.WriteFile(path, []byte(text), 0644); err != nil {
					t.Fatal(err)
				}
			}
			s := &stubExtractor{first: body, retry: body}
			sum := &recordingSummary{result: &summarize.SummarizeResult{Text: "Generated summary without any scope notice.", Provider: "fixture", Model: "fixture"}}
			err := New(&cfg, s, sum).ProcessItem(item)
			good := kind == "summary" || kind == "embedded-summary"
			if !good {
				if err == nil || IsPartial(err) || len(sum.bodies) != 0 {
					t.Fatalf("invalid extraction summarized: %v %v", err, sum.bodies)
				}
				for path, want := range map[string]string{contentPath: "---\ntype: reading-list-content\n---\n\nOld body", summaryPath: "Old summary"} {
					got, readErr := os.ReadFile(path)
					if readErr != nil || string(got) != want {
						t.Fatalf("existing file changed: %q %v", got, readErr)
					}
				}
				return
			}
			if err != nil || len(sum.bodies) != 1 || sum.bodies[0] != verifiedXPost {
				t.Fatalf("verified summary failed: %v %v", err, sum.bodies)
			}
			for _, path := range []string{contentPath, summaryPath} {
				got, readErr := os.ReadFile(path)
				if readErr != nil || !strings.Contains(string(got), extract.SocialScopeNotice) || !strings.Contains(string(got), "[Source]("+url+")") {
					t.Fatalf("missing scope/source: %q %v", got, readErr)
				}
			}
		})
	}
}
