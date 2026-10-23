package pipeline

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"unicode"
)

type articleExtractor interface {
	UsesHeaded(string) bool
	ExtractURL(string) (string, error)
	ExtractURLHeaded(string) (string, error)
}

// extractWithFallback retries a failed headless extraction once with a visible
// browser. Both attempts must pass the same content checks before being saved.
func extractWithFallback(pool articleExtractor, rawURL string, headedOnBlock bool) (string, error) {
	headed := pool.UsesHeaded(rawURL)
	content, err := pool.ExtractURL(rawURL)
	if err == nil {
		err = validateTargetContent(rawURL, content)
	}
	if err == nil {
		return content, nil
	}
	if headed || (!headedOnBlock && (isBlockedPage(content) || isBlockedPage(err.Error()))) {
		return "", err
	}

	slog.Warn("headless extraction failed, retrying with headed mode", "url", rawURL, "err", err)
	firstErr := err
	content, err = pool.ExtractURLHeaded(rawURL)
	if err == nil {
		err = validateTargetContent(rawURL, content)
	}
	if err != nil {
		return "", fmt.Errorf("headless attempt: %v; headed retry: %w", firstErr, err)
	}
	return content, nil
}

// videoTarget mirrors the browser adapter's supported/recognized video hosts.
// The marker is a versioned adapter result, not independent audio verification.
func videoTarget(rawURL string) (platform, target string) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", ""
	}
	host := strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
	switch host {
	case "youtube.com", "m.youtube.com", "youtu.be":
		if host == "youtu.be" {
			return "youtube", strings.Trim(u.Path, "/")
		}
		if u.Path == "/watch" {
			return "youtube", u.Query().Get("v")
		}
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) == 2 && (parts[0] == "shorts" || parts[0] == "embed" || parts[0] == "live") {
			return "youtube", parts[1]
		}
		return "youtube", ""
	}
	for _, domain := range []string{"bilibili.com", "vimeo.com", "dailymotion.com", "twitch.tv", "tiktok.com", "nicovideo.jp", "niconico.jp", "youku.com", "douyin.com"} {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return domain, ""
		}
	}
	return "", ""
}

func validateTargetContent(rawURL, content string) error {
	if err := validateExtractedContent(content); err != nil {
		return err
	}
	platform, target := videoTarget(rawURL)
	if platform == "" {
		return nil
	}
	if platform != "youtube" {
		return fmt.Errorf("unverified original language or unsupported video transcript extraction")
	}
	line, body, _ := strings.Cut(content, "\n")
	const prefix = "<!-- rlss-video-v1 "
	var provenance struct{ Platform, Target, Language, Source, Evidence string }
	if !strings.HasPrefix(line, prefix) || !strings.HasSuffix(line, " -->") ||
		json.Unmarshal([]byte(strings.TrimSuffix(strings.TrimPrefix(line, prefix), " -->")), &provenance) != nil ||
		provenance.Platform != platform || target == "" || provenance.Target != target ||
		strings.TrimSpace(provenance.Language) == "" || (provenance.Source != "human" && provenance.Source != "automatic") ||
		provenance.Evidence != "original-audio-asr-association" {
		return fmt.Errorf("unverified original language: missing or invalid target-bound transcript provenance")
	}
	metadata := "Subtitle language: " + provenance.Language + "\nSubtitle source: " + provenance.Source + "\n\n"
	body = strings.TrimLeft(body, "\n")
	if !strings.HasPrefix(body, metadata) {
		return fmt.Errorf("unusable subtitles: missing language/source labels")
	}
	body = strings.TrimPrefix(body, metadata)
	// Labels, timestamps and comments alone cannot establish subtitle contents.
	if !hasArticleBody(body) || !strings.ContainsFunc(htmlTags.ReplaceAllString(body, ""), unicode.IsLetter) {
		return fmt.Errorf("unusable subtitles: no substantive cue text")
	}
	return nil
}

func validateExtractedContent(content string) error {
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("empty content (Defuddle returned nothing)")
	}
	if isBlockedPage(content) {
		return fmt.Errorf("blocked by anti-bot protection (Cloudflare/CAPTCHA)")
	}
	if isUnavailablePage(content) {
		return fmt.Errorf("page unavailable (maintenance / paywall / login required)")
	}
	return nil
}
