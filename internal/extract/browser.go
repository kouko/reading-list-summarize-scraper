package extract

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chromedp/cdproto/page"
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
// markdown. Only a complete sequence returns content; failures, unusable
// continuations, loops and a cap with remaining pages return an error.
func (b *Browser) ExtractPaginated(startURL, jsCode, nextSelector string, maxPages int, timeout, waitAfterLoad time.Duration) (string, error) {
	// This guard applies only to the entry document. Previous links on pages
	// reached by our own forward traversal do not imply missing content.
	entryJS := `(function(){
var current=new URL(location.href);current.hash='';
var earlier=document.querySelectorAll('link[rel~=first],a[rel~=first],link[rel~=prev],a[rel~=prev]');
for(var e of earlier){
 if(!(e.getAttribute('href')||'').trim())continue;
 var target;try{target=new URL(e.getAttribute('href'),document.baseURI);}catch(_){continue;}
 target.hash='';
 if(target.origin===current.origin&&target.href!==current.href)
  throw new Error('incomplete pagination: initial page advertises an earlier page');
}
var numbers=current.searchParams.getAll('page');
if(numbers.length===1&&/^\d+$/.test(numbers[0])&&Number(numbers[0])>1)
 throw new Error('incomplete pagination: initial page number is greater than one');
})()`
	// Embed the selector safely as a JS string literal.
	selLit, _ := json.Marshal(nextSelector)
	nextJS := `(function(){
var e=document.querySelector(` + string(selLit) + `);
if(e&&(e.getAttribute('href')||'').trim())return e.href||e.getAttribute('href');
var hints=document.querySelectorAll('link[rel~=next],a[rel~=next]');
for(var h of hints){
 if((h.getAttribute('href')||'').trim())throw new Error('incomplete pagination: configured next link missing or empty with standard next hint');
}
return '';
})()`

	// fetch loads one page and returns its markdown + the next-page href. The
	// loop, stop conditions, and concatenation live in paginate (tested there);
	// this closure is the only browser-bound part.
	fetch := func(pageURL string) (string, string, error) {
		var content, nextHref string
		ctx, cancel := chromedp.NewContext(b.ctx)
		defer cancel()
		ctx, cancelT := context.WithTimeout(ctx, timeout)
		defer cancelT()
		// Match readiness to this navigation's loader: an old document or a
		// visible body during parsing is not sufficient. DOMContentLoaded waits
		// for deferred scripts, but not ancillary images or ads to finish.
		ready := make(chan *page.EventLifecycleEvent, 16)
		chromedp.ListenTarget(ctx, func(ev any) {
			if e, ok := ev.(*page.EventLifecycleEvent); ok && e.Name == "DOMContentLoaded" {
				select {
				case ready <- e:
				default:
				}
			}
		})
		err := chromedp.Run(ctx,
			page.SetLifecycleEventsEnabled(true),
			chromedp.ActionFunc(func(ctx context.Context) error {
				_, loader, errorText, download, err := page.Navigate(pageURL).Do(ctx)
				if err != nil {
					return err
				}
				if errorText != "" {
					return fmt.Errorf("navigate: %s", errorText)
				}
				if download || loader == "" {
					return fmt.Errorf("navigate: no article document loaded")
				}
				for {
					select {
					case e := <-ready:
						if e.LoaderID == loader {
							return nil
						}
					case <-ctx.Done():
						return ctx.Err()
					}
				}
			}),
			chromedp.WaitVisible(`body`, chromedp.ByQuery),
			chromedp.Sleep(waitAfterLoad),
			chromedp.ActionFunc(func(ctx context.Context) error {
				if pageURL == startURL {
					return chromedp.Evaluate(entryJS, nil).Do(ctx)
				}
				return nil
			}),
			chromedp.Evaluate(jsCode, nil),
			chromedp.Evaluate(`window.extractArticle()`, &content, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
				return p.WithAwaitPromise(true)
			}),
			chromedp.Evaluate(nextJS, &nextHref),
		)
		return content, nextHref, err
	}

	return paginate(startURL, maxPages, fetch)
}

func (b *Browser) Close() {
	b.cancel()
}
