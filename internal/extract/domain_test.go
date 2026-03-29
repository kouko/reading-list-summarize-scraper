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
