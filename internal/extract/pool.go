package extract

import (
	"fmt"
	"sync"
	"time"

	"github.com/kouko/reading-list-summarize-scraper/internal/config"
)

type poolKey struct {
	headed  bool
	profile string
}

type poolBrowser interface {
	Extract(string, string, time.Duration, time.Duration) (string, error)
	ExtractPaginated(string, string, string, int, time.Duration, time.Duration) (string, error)
	Close()
}

type Pool struct {
	mu         sync.Mutex
	instances  map[poolKey]poolBrowser
	newBrowser func(bool, string, string) (poolBrowser, error)
	resolver   *ProfileResolver
	cfg        *config.ExtractConfig
	jsCode     string
}

func NewPool(cfg *config.ExtractConfig, resolver *ProfileResolver, jsCode string) *Pool {
	return &Pool{
		instances: make(map[poolKey]poolBrowser),
		newBrowser: func(headed bool, profile, dir string) (poolBrowser, error) {
			return NewBrowser(headed, profile, dir)
		},
		resolver: resolver,
		cfg:      cfg,
		jsCode:   jsCode,
	}
}

func (p *Pool) ExtractURL(rawURL string) (string, error) {
	headed, profile := p.resolveForURL(rawURL)

	browser, err := p.getBrowser(headed, profile)
	if err != nil {
		return "", fmt.Errorf("get browser (headed=%v, profile=%q): %w", headed, profile, err)
	}
	timeout := p.extractionTimeout(headed)

	// Paginated article: when a DomainRule configures a next-page selector,
	// follow it and concatenate pages. Otherwise single-page extraction.
	if selector, maxPages := MatchPaginationRule(rawURL, p.cfg.DomainRules); selector != "" {
		return browser.ExtractPaginated(rawURL, p.jsCode, selector, maxPages, timeout, p.cfg.WaitAfterLoad)
	}

	return browser.Extract(rawURL, p.jsCode, timeout, p.cfg.WaitAfterLoad)
}

// ExtractURLHeaded forces headed (non-headless) mode for the given URL,
// using the same profile as resolveForURL would pick.
// Uses HeadedTimeout (longer) to allow anti-bot challenges to resolve.
func (p *Pool) ExtractURLHeaded(rawURL string) (string, error) {
	_, profile := p.resolveForURL(rawURL)
	browser, err := p.getBrowser(true, profile)
	if err != nil {
		return "", fmt.Errorf("get browser (headed=true, profile=%q): %w", profile, err)
	}

	timeout := p.extractionTimeout(true)

	// Mirror ExtractURL's pagination routing so a paginated site that falls
	// back to the headed path (anti-bot retry) still gets the full article.
	if selector, maxPages := MatchPaginationRule(rawURL, p.cfg.DomainRules); selector != "" {
		return browser.ExtractPaginated(rawURL, p.jsCode, selector, maxPages, timeout, p.cfg.WaitAfterLoad)
	}

	return browser.Extract(rawURL, p.jsCode, timeout, p.cfg.WaitAfterLoad)
}

func (p *Pool) extractionTimeout(headed bool) time.Duration {
	if headed && p.cfg.HeadedTimeout != 0 {
		return p.cfg.HeadedTimeout
	}
	return p.cfg.Timeout
}

// UsesHeaded reports the initial mode after applying per-domain overrides.
func (p *Pool) UsesHeaded(rawURL string) bool {
	headed, _ := p.resolveForURL(rawURL)
	return headed
}

func (p *Pool) resolveForURL(rawURL string) (bool, string) {
	headed, profileName, googleAccount, matched := MatchDomainRules(rawURL, p.cfg.DomainRules)
	if !matched {
		headed = !p.cfg.Headless
		profileName = p.cfg.ChromeProfile
		googleAccount = p.cfg.GoogleAccount
	}

	// Use SmartResolve if email is available
	if p.resolver != nil && googleAccount != "" {
		folder, _, err := p.resolver.SmartResolve(googleAccount, profileName, p.cfg.UserDataDir, false)
		if err == nil {
			profileName = folder
		}
	} else if p.resolver != nil && profileName != "" {
		if folder, err := p.resolver.Resolve(profileName); err == nil {
			profileName = folder
		}
	}
	return headed, profileName
}

func (p *Pool) getBrowser(headed bool, profile string) (poolBrowser, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := poolKey{headed: headed, profile: profile}
	if b, ok := p.instances[key]; ok {
		return b, nil
	}
	// Every instance shares UserDataDir, even across Chrome profiles. Release
	// its current owner before changing mode or profile, including after retry.
	for existingKey, b := range p.instances {
		b.Close()
		delete(p.instances, existingKey)
	}

	b, err := p.newBrowser(headed, profile, p.cfg.UserDataDir)
	if err != nil {
		return nil, err
	}
	p.instances[key] = b
	return b, nil
}

func (p *Pool) CloseAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, b := range p.instances {
		b.Close()
	}
	p.instances = make(map[poolKey]poolBrowser)
}
