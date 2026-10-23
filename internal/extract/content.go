package extract

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// ValidateContent rejects empty bodies and recognized challenge or access walls.
func ValidateContent(content string) error {
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("empty content (Defuddle returned nothing)")
	}
	if IsBlockedPage(content) {
		return fmt.Errorf("blocked by anti-bot protection (Cloudflare/CAPTCHA)")
	}
	if isUnavailablePage(content) {
		return fmt.Errorf("page unavailable (maintenance / paywall / login required)")
	}
	return nil
}

var xAuthLinks = regexp.MustCompile(`\[[^\]\n]*\]\(https://(?:x\.com/(?:i/jf/onboarding/web\?[^)\s]*|i/flow/(?:login|signup)|login|tos|privacy)|support\.x\.com/articles/20170514|help\.x\.com/resources/accessibility|business\.x\.com/en/help/troubleshooting/how-twitter-ads-work\.html[^)\s]*)\)`)
var xAuthIframe = regexp.MustCompile(`^<iframe (?:src="https://accounts\.google\.com/gsi/button\?[^"\n]*"[^>]*|style="display: none;")></iframe>$`)
var xCopyright = regexp.MustCompile(`^© [0-9]{4} x corp\.$`)
var markdownLinks = regexp.MustCompile(`!?\[[^\]\n]*\]\([^\)\n]*\)`)
var htmlTags = regexp.MustCompile(`<[^>]*>`)

// HasArticleBody ignores headings and navigation controls when deciding whether
// a recognized wall also contains independent prose. Link labels alone are not
// evidence of a body; prose beside links remains eligible, including short posts.
func HasArticleBody(text string) bool {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}
		line = markdownLinks.ReplaceAllString(line, "")
		line = strings.Trim(htmlTags.ReplaceAllString(line, ""), " \t·|>*_-")
		switch strings.ToLower(line) {
		case "", "log in", "sign in", "sign up", "forgot password?", "reset password",
			"privacy", "terms", "home", "back", "menu", "or", "continue with google":
			continue
		}
		if strings.ContainsFunc(line, unicode.IsLetter) {
			return true
		}
	}
	return false
}

// isUnavailablePage detects maintenance pages, paywalls, and login/subscription
// walls served INSTEAD of the article. These otherwise slip through (they aren't
// anti-bot challenges) and get silently mis-handled — counted as a short-content
// "success" or summarized into a "I can't, it's a paywall" non-summary. Treat
// them as an extraction failure instead (the caller does not write a content
// file, so the item is re-attempted on the next run).
//
// Patterns are deliberately specific multi-word/wall strings (not bare topic
// words like "subscribe" or "登入") so a real article *about* paywalls or
// maintenance is not misflagged. A wall phrase must occur without an independent
// body paragraph; a heading alone does not establish that an article was served.
func isUnavailablePage(content string) bool {
	lower := strings.ToLower(content)
	// X may return just its authentication UI. Remove known UI text, links,
	// and embedded sign-in controls; any other text preserves article content.
	if strings.Contains(lower, "https://x.com/") || strings.Contains(lower, "## log in or sign up for x") {
		authUI := xAuthLinks.MatchString(lower)
		body := false
		for _, line := range strings.Split(lower, "\n") {
			line = strings.TrimSpace(line)
			line = strings.Trim(xAuthLinks.ReplaceAllString(line, ""), " ·")
			switch {
			case line == "", line == "## post", xAuthIframe.MatchString(line), xCopyright.MatchString(line):
			case line == "## log in or sign up for x",
				line == "see what’s happening and join the conversation",
				line == "continue with google", line == "or",
				line == "透過 google 帳戶繼續操作透過 google 帳戶繼續操作。在新分頁中開啟":
				authUI = true
			default:
				body = body || HasArticleBody(line)
			}
		}
		if authUI && !body {
			return true
		}
	}
	unavailablePatterns := []string{
		// English — maintenance / login / subscription walls
		"under maintenance",
		"sign in to read",
		"log in to read",
		"please sign in to",
		"subscribe to read",
		"subscribe to continue reading",
		// Chinese (Traditional + Simplified) — maintenance / login / member walls
		"系統維護中",
		"系统维护中",
		"登入後閱讀全文",
		"請登入後閱讀",
		"訂閱成為會員",
		// Japanese — maintenance / login walls
		"メンテナンス中",
		"ログインしてください",
		"ログインが必要",
	}
	wall := false
	body := false
	for _, paragraph := range strings.Split(lower, "\n\n") {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}
		matched := false
		for _, pattern := range unavailablePatterns {
			if strings.Contains(paragraph, pattern) {
				matched = true
				break
			}
		}
		if matched {
			wall = true
		} else {
			body = body || HasArticleBody(paragraph)
		}
	}
	return wall && !body
}

// IsBlockedPage detects anti-bot protection pages (Cloudflare, CAPTCHA, etc.)
// that were extracted instead of actual content. Patterns include localized
// (ja / zh-Hant / zh-Hans) challenge-UI strings, since Cloudflare/anti-bot
// interstitials render in the page's language; the language-neutral branding
// ("cloudflare", "ray id:", "captcha") is kept too. The >=2 threshold guards
// against false-positives on real articles that merely discuss these topics —
// so the localized entries are challenge-UI phrases, not generic security words.
func IsBlockedPage(content string) bool {
	lower := strings.ToLower(content)
	// Observed go.dev interstitial: a challenge heading followed only by
	// Google's cookie notice. A heading mention in an article is insufficient.
	parts := strings.Split(strings.TrimSpace(lower), "\n\n")
	if len(parts) == 2 && parts[0] == "## to, err:= human()" &&
		strings.HasPrefix(parts[1], "go.dev uses cookies from google to deliver and enhance the quality of its services and to analyze traffic.") {
		return true
	}
	blockedPatterns := []string{
		// English / language-neutral
		"performing security verification",
		"security challenge",
		"checking your browser",
		"please wait while we verify",
		"just a moment",
		"ray id:",
		"cloudflare",
		"captcha",
		"access denied",
		"please verify you are a human",
		"bot protection",
		// Japanese challenge-UI
		"ブラウザを確認しています",
		"あなたが人間であることを確認",
		"ロボットではないことを確認",
		"アクセスが拒否されました",
		// Chinese (Traditional) challenge-UI
		"正在檢查您的瀏覽器",
		"請完成安全驗證",
		"確認您是真人",
		"拒絕存取",
		// Chinese (Simplified) challenge-UI
		"正在检查您的浏览器",
		"请完成安全验证",
		"确认您是真人",
		"拒绝访问",
	}
	matchCount := 0
	for _, pattern := range blockedPatterns {
		if strings.Contains(lower, pattern) {
			matchCount++
		}
	}
	// Require at least 2 matches to avoid false positives
	return matchCount >= 2
}
