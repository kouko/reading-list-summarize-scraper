package extract

import (
	"errors"
	"testing"
	"time"

	"github.com/kouko/reading-list-summarize-scraper/internal/config"
)

type recordingBrowser struct {
	closed            bool
	timeout, wait     time.Duration
	url, js, selector string
	maxPages          int
}

func (b *recordingBrowser) Extract(url, js string, timeout, wait time.Duration) (string, error) {
	b.url, b.js, b.timeout, b.wait = url, js, timeout, wait
	return "article", nil
}

func (b *recordingBrowser) ExtractPaginated(url, js, selector string, maxPages int, timeout, wait time.Duration) (string, error) {
	b.selector, b.maxPages = selector, maxPages
	return b.Extract(url, js, timeout, wait)
}

func (b *recordingBrowser) Close() { b.closed = true }

func TestPoolHeadedRecoveryLifecycle(t *testing.T) {
	cfg := &config.ExtractConfig{Headless: true, ChromeProfile: "Default", UserDataDir: t.TempDir(), Timeout: time.Second, HeadedTimeout: time.Minute}
	p := NewPool(cfg, nil, "extract JS")
	var previous *recordingBrowser
	var modes []bool
	p.newBrowser = func(headed bool, profile, dir string) (poolBrowser, error) {
		if previous != nil && !previous.closed {
			t.Fatal("previous browser still owns user data directory at startup")
		}
		if profile != cfg.ChromeProfile || dir != cfg.UserDataDir {
			t.Fatalf("profile/directory = %q/%q", profile, dir)
		}
		modes = append(modes, headed)
		previous = &recordingBrowser{}
		return previous, nil
	}
	defer p.CloseAll()
	for _, headed := range []bool{false, true, false, true} {
		var body string
		var err error
		if headed {
			body, err = p.ExtractURLHeaded("https://ordinary.example/article")
		} else {
			body, err = p.ExtractURL("https://ordinary.example/article")
		}
		if err != nil || body != "article" {
			t.Fatalf("body=%q error=%v", body, err)
		}
		wantTimeout := cfg.Timeout
		if headed {
			wantTimeout = cfg.HeadedTimeout
		}
		if modes[len(modes)-1] != headed || previous.timeout != wantTimeout {
			t.Fatalf("mode=%v timeout=%v, want mode=%v timeout=%v", modes[len(modes)-1], previous.timeout, headed, wantTimeout)
		}
	}
	if len(modes) != 4 {
		t.Fatalf("starts=%v", modes)
	}
	// A compatible browser is reused rather than opening another process.
	if _, err := p.ExtractURLHeaded("https://another.example/article"); err != nil {
		t.Fatal(err)
	}
	if len(modes) != 4 {
		t.Fatalf("compatible browser was not reused: %v", modes)
	}
}

func TestPoolExplicitHeaded(t *testing.T) {
	for _, domainOverride := range []bool{false, true} {
		for _, headedTimeout := range []time.Duration{0, time.Minute} {
			cfg := &config.ExtractConfig{Headless: domainOverride, Timeout: time.Second, HeadedTimeout: headedTimeout}
			if domainOverride {
				cfg.DomainRules = []config.DomainRule{{Domains: []string{"example.com"}, Headed: true}}
			}
			p := NewPool(cfg, nil, "")
			b := &recordingBrowser{}
			p.newBrowser = func(headed bool, _, _ string) (poolBrowser, error) {
				if !headed {
					t.Fatal("explicit headed setting was lost")
				}
				return b, nil
			}
			if !p.UsesHeaded("https://example.com/article") {
				t.Fatal("explicit headed mode not reported")
			}
			if _, err := p.ExtractURL("https://example.com/article"); err != nil {
				t.Fatal(err)
			}
			want := headedTimeout
			if want == 0 {
				want = cfg.Timeout
			}
			if b.timeout != want {
				t.Fatalf("timeout=%v, want %v (domain override=%v)", b.timeout, want, domainOverride)
			}
			p.CloseAll()
		}
	}
}

func TestPoolRecoveryPreservesPaginationProfile(t *testing.T) {
	cfg := &config.ExtractConfig{Headless: true, ChromeProfile: "Default", Timeout: time.Second, HeadedTimeout: time.Minute, WaitAfterLoad: 2 * time.Second,
		DomainRules: []config.DomainRule{{Domains: []string{"example.com"}, ChromeProfile: "Profile 2", NextPageSelector: "a.next", MaxPages: 3}}}
	p := NewPool(cfg, nil, "extract JS")
	previous := &recordingBrowser{}
	p.instances[poolKey{profile: "Default"}] = previous
	b := &recordingBrowser{}
	p.newBrowser = func(headed bool, profile, _ string) (poolBrowser, error) {
		if !previous.closed {
			t.Fatal("different profile still owns shared user data directory")
		}
		if !headed || profile != "Profile 2" {
			t.Fatalf("headed=%v profile=%q", headed, profile)
		}
		return b, nil
	}
	defer p.CloseAll()
	url := "https://example.com/article"
	if body, err := p.ExtractURLHeaded(url); err != nil || body != "article" {
		t.Fatalf("body=%q error=%v", body, err)
	}
	if b.url != url || b.js != p.jsCode || b.selector != "a.next" || b.maxPages != 3 || b.timeout != cfg.HeadedTimeout || b.wait != cfg.WaitAfterLoad {
		t.Fatalf("extraction settings lost: %+v", b)
	}
}

func TestPoolRecoveryBrowserFailure(t *testing.T) {
	p := NewPool(&config.ExtractConfig{Headless: true}, nil, "")
	previous := &recordingBrowser{}
	p.instances[poolKey{}] = previous
	want := errors.New("browser startup failed")
	p.newBrowser = func(bool, string, string) (poolBrowser, error) {
		if !previous.closed {
			t.Fatal("headless browser not closed before failed startup")
		}
		return nil, want
	}
	if body, err := p.ExtractURLHeaded("https://example.com/article"); body != "" || !errors.Is(err, want) {
		t.Fatalf("body=%q error=%v", body, err)
	}
	if len(p.instances) != 0 {
		t.Fatalf("failed startup left stale browser: %v", p.instances)
	}
}
