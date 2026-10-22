package pipeline

import "testing"

// Graduated verbatim from evidence/probes/test_article_prompt.py at 3448625.
// concern: Login prompts accompanying genuine article content must not reject the body.
func TestAdversarialArticlePrompt(t *testing.T) {
	article := "## How coastal wetlands recover\n\nResearchers measured sediment accumulation across twelve restored wetlands. Native vegetation returned after tidal flow was restored, and the report compares annual monitoring results with nearby reference sites.\n\nPlease sign in to read more articles."
	s := &stubExtractor{first: article, retry: article}
	got, err := extractWithFallback(s, "https://example.com/article", true)
	if err != nil || got != article || len(s.calls) != 1 {
		t.Fatalf("real body with incidental prompt rejected: body=%q err=%v calls=%v", got, err, s.calls)
	}
}

// Graduated verbatim from evidence/probes/test_x_shell.py at 3448625.
// concern: An X page shell heading must not turn authentication UI into article content.
func TestAdversarialXShell(t *testing.T) {
	wall := "## Post\n\n## Log in or sign up for X\n\n[Log in](https://x.com/i/flow/login)\n[Sign up](https://x.com/i/flow/signup)"
	s := &stubExtractor{first: wall, retry: wall}
	got, err := extractWithFallback(s, "https://x.com/example/status/123", true)
	if err == nil || got != "" {
		t.Fatalf("authentication-only shell accepted: body=%q err=%v", got, err)
	}
}
