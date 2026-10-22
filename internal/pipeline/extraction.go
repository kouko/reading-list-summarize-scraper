package pipeline

import (
	"fmt"
	"log/slog"
	"strings"
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
		err = validateExtractedContent(content)
	}
	if err == nil {
		return content, nil
	}
	if headed || (!headedOnBlock && isBlockedPage(content)) {
		return "", err
	}

	slog.Warn("headless extraction failed, retrying with headed mode", "url", rawURL, "err", err)
	firstErr := err
	content, err = pool.ExtractURLHeaded(rawURL)
	if err == nil {
		err = validateExtractedContent(content)
	}
	if err != nil {
		return "", fmt.Errorf("headless attempt: %v; headed retry: %w", firstErr, err)
	}
	return content, nil
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
