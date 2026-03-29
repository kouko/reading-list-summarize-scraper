package extract

import (
	"net/url"
	"strings"

	"github.com/kouko/reading-list-summarize-scraper/internal/config"
)

func MatchDomainRules(rawURL string, rules []config.DomainRule) (headed bool, profileName string, googleAccount string, matched bool) {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return false, "", "", false
	}
	host := strings.ToLower(u.Hostname())

	for _, rule := range rules {
		for _, pattern := range rule.Domains {
			pattern = strings.ToLower(pattern)
			if matchDomain(host, pattern) {
				return rule.Headed, rule.ChromeProfile, rule.GoogleAccount, true
			}
		}
	}
	return false, "", "", false
}

func matchDomain(host, pattern string) bool {
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[1:] // ".example.com"
		return strings.HasSuffix(host, suffix)
	}
	return host == pattern || strings.HasSuffix(host, "."+pattern)
}
