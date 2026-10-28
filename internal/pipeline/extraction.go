package pipeline

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"github.com/kouko/reading-list-summarize-scraper/internal/extract"
)

type articleExtractor interface {
	UsesHeaded(string) bool
	ExtractURL(string) (string, error)
	ExtractURLHeaded(string) (string, error)
}

var htmlTags = regexp.MustCompile(`<[^>]*>`)

func authorThreadTarget(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	for _, domain := range []string{"x.com", "twitter.com", "threads.com", "threads.net"} {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}
	return false
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
	if headed || (!headedOnBlock && (extract.IsBlockedPage(content) || extract.IsBlockedPage(err.Error()))) {
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
	if err := extract.ValidateContent(content); err != nil {
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
	if !extract.HasArticleBody(body) || !strings.ContainsFunc(htmlTags.ReplaceAllString(body, ""), unicode.IsLetter) {
		return fmt.Errorf("unusable subtitles: no substantive cue text")
	}
	return nil
}
