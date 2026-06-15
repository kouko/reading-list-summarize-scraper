package extract

import (
	"testing"

	"github.com/kouko/reading-list-summarize-scraper/internal/config"
)

func TestMatchDomainRules(t *testing.T) {
	rules := []config.DomainRule{
		{Domains: []string{"medium.com"}, Headed: true, ChromeProfile: "Work"},
		{Domains: []string{"*.substack.com"}, Headed: true, ChromeProfile: "Default"},
		{Domains: []string{"github.com"}, Headed: false, ChromeProfile: "Dev"},
	}

	tests := []struct {
		url         string
		wantHeaded  bool
		wantProfile string
		wantMatch   bool
	}{
		{"https://medium.com/article", true, "Work", true},
		{"https://www.medium.com/article", true, "Work", true},
		{"https://foo.substack.com/post", true, "Default", true},
		{"https://github.com/repo", false, "Dev", true},
		{"https://example.com/page", false, "", false},
	}

	for _, tt := range tests {
		headed, profile, _, matched := MatchDomainRules(tt.url, rules)
		if matched != tt.wantMatch {
			t.Errorf("MatchDomainRules(%q) matched=%v, want %v", tt.url, matched, tt.wantMatch)
			continue
		}
		if matched {
			if headed != tt.wantHeaded {
				t.Errorf("MatchDomainRules(%q) headed=%v, want %v", tt.url, headed, tt.wantHeaded)
			}
			if profile != tt.wantProfile {
				t.Errorf("MatchDomainRules(%q) profile=%q, want %q", tt.url, profile, tt.wantProfile)
			}
		}
	}
}

func TestMatchPaginationRule(t *testing.T) {
	rules := []config.DomainRule{
		{Domains: []string{"www.itmedia.co.jp"}, NextPageSelector: "a[rel=next]", MaxPages: 5},
		{Domains: []string{"example.com"}}, // no pagination config
	}

	// Matching domain with pagination config.
	sel, mp := MatchPaginationRule("https://www.itmedia.co.jp/pcuser/articles/2606/08/news096.html", rules)
	if sel != "a[rel=next]" || mp != 5 {
		t.Errorf("itmedia: got (%q, %d), want (a[rel=next], 5)", sel, mp)
	}

	// Matching domain WITHOUT pagination config → zero values.
	if sel, mp := MatchPaginationRule("https://example.com/a", rules); sel != "" || mp != 0 {
		t.Errorf("example.com: got (%q, %d), want (\"\", 0)", sel, mp)
	}

	// No matching domain → zero values.
	if sel, mp := MatchPaginationRule("https://other.org/a", rules); sel != "" || mp != 0 {
		t.Errorf("no-match: got (%q, %d), want (\"\", 0)", sel, mp)
	}
}
