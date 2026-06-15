package extract

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"

	cu "github.com/Davincible/chromedp-undetected"
)

type Browser struct {
	ctx     context.Context
	cancel  context.CancelFunc
	headed  bool
	profile string
}

func NewBrowser(headed bool, profileDir string, userDataDir string) (*Browser, error) {
	cfg := cu.NewConfig(
		cu.WithUserDataDir(userDataDir),
	)

	// Add stealth flags.
	cfg.ChromeFlags = append(cfg.ChromeFlags,
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-gpu", true),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	)

	if profileDir != "" {
		cfg.ChromeFlags = append(cfg.ChromeFlags,
			chromedp.Flag("profile-directory", profileDir),
		)
	}

	if !headed {
		// On macOS, chromedp-undetected's WithHeadless() uses Xvfb (Linux only).
		// Fall back to standard headless flag for macOS.
		cfg.ChromeFlags = append(cfg.ChromeFlags,
			chromedp.Flag("headless", true),
		)
	}

	ctx, cancel, err := cu.New(cfg)
	if err != nil {
		return nil, err
	}

	return &Browser{
		ctx:     ctx,
		cancel:  cancel,
		headed:  headed,
		profile: profileDir,
	}, nil
}

// Extract navigates to the URL, injects Defuddle JS, and returns extracted Markdown.
func (b *Browser) Extract(url string, jsCode string, timeout time.Duration, waitAfterLoad time.Duration) (string, error) {
	ctx, cancel := chromedp.NewContext(b.ctx)
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	var content string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(waitAfterLoad),
		chromedp.Evaluate(jsCode, nil),
		chromedp.Evaluate(`window.extractArticle()`, &content, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
			return p.WithAwaitPromise(true)
		}),
	)
	return content, err
}

// ExtractPaginated extracts a multi-page article by following the next-page
// link (matched by nextSelector) page-by-page and concatenating the per-page
// markdown. It reuses this Browser across sequential navigations. The loop
// stops at the first page with no next link, an already-visited URL (loop
// guard), or maxPages (<=0 → defaultMaxPages). It is fail-soft: a page that
// errors mid-sequence stops the loop but keeps the pages gathered so far;
// only a total failure (nothing extracted) returns the error.
func (b *Browser) ExtractPaginated(startURL, jsCode, nextSelector string, maxPages int, timeout, waitAfterLoad time.Duration) (string, error) {
	if maxPages <= 0 {
		maxPages = defaultMaxPages
	}
	// Embed the selector safely as a JS string literal.
	selLit, _ := json.Marshal(nextSelector)
	nextJS := `(function(){var e=document.querySelector(` + string(selLit) + `);return e?e.href:"";})()`

	var pages []string
	visited := map[string]bool{}
	current := startURL
	var firstErr error

	for pageCount := 0; ; pageCount++ {
		visited[current] = true

		var content, nextHref string
		ctx, cancel := chromedp.NewContext(b.ctx)
		ctx, cancelT := context.WithTimeout(ctx, timeout)
		err := chromedp.Run(ctx,
			chromedp.Navigate(current),
			chromedp.WaitVisible(`body`, chromedp.ByQuery),
			chromedp.Sleep(waitAfterLoad),
			chromedp.Evaluate(jsCode, nil),
			chromedp.Evaluate(`window.extractArticle()`, &content, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
				return p.WithAwaitPromise(true)
			}),
			chromedp.Evaluate(nextJS, &nextHref),
		)
		cancelT()
		cancel()
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			// Fail-soft: keep the pages gathered so far, but surface the
			// failure so a truncated article isn't silently indistinguishable
			// from a genuinely short one.
			slog.Warn("paginated extract: page failed, keeping pages gathered so far",
				"url", current, "pages_gathered", len(pages), "err", err)
			break
		}
		pages = append(pages, content)

		next, ok := resolveNextURL(current, nextHref)
		if !ok || !shouldFollowNext(next, visited, pageCount+1, maxPages) {
			break
		}
		current = next
	}

	combined := joinPages(pages)
	if combined == "" && firstErr != nil {
		return "", firstErr
	}
	return combined, nil
}

func (b *Browser) Close() {
	b.cancel()
}
