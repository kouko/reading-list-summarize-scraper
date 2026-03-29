package source

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	cu "github.com/Davincible/chromedp-undetected"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"

	"github.com/kouko/reading-list-summarize-scraper/internal/extract"
)

// ChromeSource fetches Chrome's Reading List via a Chrome Extension that uses
// the chrome.readingList.query({}) API. The extension is loaded at runtime
// through CDP and communicates back via service worker evaluation.
type ChromeSource struct {
	profileDir          string // resolved folder name (e.g., "Profile 5")
	userDataDir         string
	googleAccount       string
	cloneProfile        bool
	extensionManifest   []byte
	extensionBackground []byte
}

// NewChromeSource creates a ChromeSource. The manifest and background parameters
// are the raw bytes of the extension's manifest.json and background.js files.
// These should be embedded at the cmd level and passed in, since go:embed only
// works with relative paths.
func NewChromeSource(profileDir, userDataDir, googleAccount string, cloneProfile bool, manifest, background []byte) *ChromeSource {
	if userDataDir == "" {
		home, _ := os.UserHomeDir()
		userDataDir = filepath.Join(home, ".config", "rlss", "chrome-data")
	}
	return &ChromeSource{
		profileDir:          profileDir,
		userDataDir:         userDataDir,
		googleAccount:       googleAccount,
		cloneProfile:        cloneProfile,
		extensionManifest:   manifest,
		extensionBackground: background,
	}
}

func (c *ChromeSource) Name() string { return "chrome" }

func (c *ChromeSource) Fetch() ([]ReadingItem, error) {
	userDataDir := c.userDataDir

	// If locked, try clone_profile fallback.
	if extract.IsLocked(userDataDir) {
		if c.cloneProfile {
			cloneDir, err := extract.CloneProfile(userDataDir, c.profileDir)
			if err != nil {
				return nil, fmt.Errorf("clone profile for Chrome Reading List: %w", err)
			}
			slog.Info("using cloned profile for Chrome Reading List",
				"original", userDataDir, "clone", cloneDir)
			userDataDir = cloneDir
		} else {
			return nil, fmt.Errorf("Chrome user data dir %q is locked (SingletonLock); set clone_profile: true", userDataDir)
		}
	}

	// Check if profile dir exists; create and hint if not.
	if _, err := os.Stat(userDataDir); os.IsNotExist(err) {
		if mkErr := os.MkdirAll(userDataDir, 0755); mkErr != nil {
			return nil, fmt.Errorf("create chrome data dir %s: %w", userDataDir, mkErr)
		}
		slog.Warn("Chrome automation profile not found, creating new one",
			"path", userDataDir,
			"hint", "Sign in to your Google account in the Chrome window to sync your Reading List, then re-run rlss")
	}

	// Extract embedded extension to temp dir.
	extDir, err := c.extractExtension()
	if err != nil {
		return nil, fmt.Errorf("extract extension: %w", err)
	}
	defer os.RemoveAll(extDir)

	// Launch headed Chrome with extension + profile using chromedp-undetected.
	cfg := cu.NewConfig(
		cu.WithUserDataDir(userDataDir),
		cu.WithExtensions(extDir),
	)
	cfg.ChromeFlags = append(cfg.ChromeFlags,
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-extensions-except", extDir),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	)
	if c.profileDir != "" {
		cfg.ChromeFlags = append(cfg.ChromeFlags,
			chromedp.Flag("profile-directory", c.profileDir),
		)
	}

	ctx, cancel, err := cu.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("launch chrome: %w", err)
	}
	defer cancel()

	ctx, timeoutCancel := context.WithTimeout(ctx, 30*time.Second)
	defer timeoutCancel()

	// Navigate to trigger browser startup.
	if err := chromedp.Run(ctx, chromedp.Navigate("about:blank")); err != nil {
		return nil, fmt.Errorf("launch chrome: %w", err)
	}

	// Wait for extension service worker to register.
	time.Sleep(3 * time.Second)

	// Find extension service worker target.
	swTargetID, err := c.findServiceWorker(ctx)
	if err != nil {
		return nil, err
	}

	// Create a new context attached to the service worker target.
	swCtx, swCancel := chromedp.NewContext(ctx, chromedp.WithTargetID(swTargetID))
	defer swCancel()

	// Evaluate chrome.readingList.query({}) in the service worker.
	var jsonStr string
	err = chromedp.Run(swCtx, chromedp.Evaluate(`
		(async () => {
			const entries = await chrome.readingList.query({});
			return JSON.stringify(entries);
		})()
	`, &jsonStr, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
		return p.WithAwaitPromise(true)
	}))
	if err != nil {
		return nil, fmt.Errorf("evaluate in service worker: %w", err)
	}

	var entries []chromeEntry
	if err := json.Unmarshal([]byte(jsonStr), &entries); err != nil {
		return nil, fmt.Errorf("parse reading list entries: %w", err)
	}

	slog.Info("fetched chrome reading list", "count", len(entries))
	return toReadingItems(entries), nil
}

// findServiceWorker scans browser targets for the extension's service worker.
// Retries up to 5 times with 2-second intervals since the SW may take time to register.
func (c *ChromeSource) findServiceWorker(ctx context.Context) (target.ID, error) {
	for attempt := 0; attempt < 5; attempt++ {
		targets, err := chromedp.Targets(ctx)
		if err != nil {
			return "", fmt.Errorf("get targets: %w", err)
		}

		for _, t := range targets {
			slog.Debug("browser target", "type", t.Type, "url", t.URL, "attempt", attempt+1)
			if t.Type == "service_worker" {
				slog.Info("found extension service worker", "url", t.URL, "attempt", attempt+1)
				return t.TargetID, nil
			}
		}

		if attempt < 4 {
			time.Sleep(2 * time.Second)
		}
	}

	// Final attempt — list all target types for debugging
	targets, _ := chromedp.Targets(ctx)
	var types []string
	for _, t := range targets {
		types = append(types, fmt.Sprintf("%s(%s)", t.Type, t.URL))
	}
	return "", fmt.Errorf("extension service worker not found after 5 attempts; targets: %v", types)
}

type chromeEntry struct {
	URL            string  `json:"url"`
	Title          string  `json:"title"`
	HasBeenRead    bool    `json:"hasBeenRead"`
	CreationTime   float64 `json:"creationTime"`
	LastUpdateTime float64 `json:"lastUpdateTime"`
}

func toReadingItems(entries []chromeEntry) []ReadingItem {
	items := make([]ReadingItem, 0, len(entries))
	for _, e := range entries {
		items = append(items, ReadingItem{
			Title:     e.Title,
			URL:       e.URL,
			DateAdded: time.UnixMilli(int64(e.CreationTime)),
			IsUnread:  !e.HasBeenRead,
			Source:    "chrome",
		})
	}
	return items
}

func (c *ChromeSource) extractExtension() (string, error) {
	dir, err := os.MkdirTemp("", "rlss-ext-*")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), c.extensionManifest, 0o644); err != nil {
		os.RemoveAll(dir)
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, "background.js"), c.extensionBackground, 0o644); err != nil {
		os.RemoveAll(dir)
		return "", err
	}
	return dir, nil
}

// Ensure ChromeSource implements Source at compile time.
var _ Source = (*ChromeSource)(nil)
