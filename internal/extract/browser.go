package extract

import (
	"context"
	"time"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

type Browser struct {
	allocCtx context.Context
	cancel   context.CancelFunc
	headed   bool
	profile  string
}

func NewBrowser(headed bool, profileDir string, userDataDir string) (*Browser, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", !headed),
		chromedp.Flag("disable-gpu", true),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"),
	)

	if profileDir != "" {
		opts = append(opts, chromedp.Flag("profile-directory", profileDir))
	}
	if userDataDir != "" {
		opts = append(opts, chromedp.UserDataDir(userDataDir))
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	return &Browser{
		allocCtx: allocCtx,
		cancel:   cancel,
		headed:   headed,
		profile:  profileDir,
	}, nil
}

func (b *Browser) Extract(url string, jsCode string, timeout time.Duration, waitAfterLoad time.Duration) (string, error) {
	ctx, cancel := chromedp.NewContext(b.allocCtx)
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
