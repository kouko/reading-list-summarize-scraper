package extract

import (
	"context"
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

func (b *Browser) Close() {
	b.cancel()
}
