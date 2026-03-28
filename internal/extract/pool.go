package extract

import (
	"fmt"
	"sync"

	"github.com/kouko/reading-list-summarize-scraper/internal/config"
)

type poolKey struct {
	headed  bool
	profile string
}

type Pool struct {
	mu        sync.Mutex
	instances map[poolKey]*Browser
	resolver  *ProfileResolver
	cfg       *config.ExtractConfig
	jsCode    string
}

func NewPool(cfg *config.ExtractConfig, resolver *ProfileResolver, jsCode string) *Pool {
	return &Pool{
		instances: make(map[poolKey]*Browser),
		resolver:  resolver,
		cfg:       cfg,
		jsCode:    jsCode,
	}
}

func (p *Pool) ExtractURL(rawURL string) (string, error) {
	headed, profile := p.resolveForURL(rawURL)

	browser, err := p.getBrowser(headed, profile)
	if err != nil {
		return "", fmt.Errorf("get browser (headed=%v, profile=%q): %w", headed, profile, err)
	}

	return browser.Extract(rawURL, p.jsCode, p.cfg.Timeout, p.cfg.WaitAfterLoad)
}

// ExtractURLHeaded forces headed (non-headless) mode for the given URL,
// using the same profile as resolveForURL would pick.
func (p *Pool) ExtractURLHeaded(rawURL string) (string, error) {
	_, profile := p.resolveForURL(rawURL)

	browser, err := p.getBrowser(true, profile)
	if err != nil {
		return "", fmt.Errorf("get browser (headed=true, profile=%q): %w", profile, err)
	}

	return browser.Extract(rawURL, p.jsCode, p.cfg.Timeout, p.cfg.WaitAfterLoad)
}

func (p *Pool) resolveForURL(rawURL string) (bool, string) {
	headed, profileName, matched := MatchDomainRules(rawURL, p.cfg.DomainRules)
	if !matched {
		headed = !p.cfg.Headless
		profileName = p.cfg.ChromeProfile
	}

	if p.resolver != nil && profileName != "" {
		if folder, err := p.resolver.Resolve(profileName); err == nil {
			profileName = folder
		}
	}
	return headed, profileName
}

func (p *Pool) getBrowser(headed bool, profile string) (*Browser, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := poolKey{headed: headed, profile: profile}
	if b, ok := p.instances[key]; ok {
		return b, nil
	}

	b, err := NewBrowser(headed, profile, p.cfg.UserDataDir)
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
	p.instances = make(map[poolKey]*Browser)
}
