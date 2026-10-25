package extract

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
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
	metadataJS := `(function(){
var links=Array.from(document.querySelectorAll('link[rel~=first],a[rel~=first],link[rel~=prev],a[rel~=prev]'));
var canonical=document.querySelector('link[rel~=canonical]');
return {url:location.href,canonical:canonical&&canonical.href||'',text:document.title+'\n'+document.body.innerText,earlier:links.filter(e=>(e.getAttribute('href')||'').trim()).map(e=>e.href)};
})()`
	var metadata struct {
		URL, Canonical, Text string
		Earlier              []string
	}
	challengePrompt := regexp.MustCompile(`(?i)^(?:please )?verify (?:that )?you are (?:a )?human[.!]?$`)
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
		responses := make(chan *network.EventResponseReceived, 16)
		chromedp.ListenTarget(ctx, func(ev any) {
			if e, ok := ev.(*network.EventResponseReceived); ok && e.Type == network.ResourceTypeDocument {
				select {
				case responses <- e:
				default:
				}
			}
			if e, ok := ev.(*page.EventLifecycleEvent); ok && e.Name == "DOMContentLoaded" {
				select {
				case ready <- e:
				default:
				}
			}
		})
		err := chromedp.Run(ctx,
			page.SetLifecycleEventsEnabled(true),
			network.Enable(),
			chromedp.ActionFunc(func(ctx context.Context) error {
				frame, loader, errorText, download, err := page.Navigate(pageURL).Do(ctx)
				if err != nil {
					return err
				}
				if errorText != "" {
					return fmt.Errorf("navigate: %s", errorText)
				}
				if download || loader == "" {
					return fmt.Errorf("navigate: no article document loaded")
				}
				// CDP Network.responseReceived identifies the document response by
				// frameId and loaderId; page.navigate returns those same identifiers.
				// https://chromedevtools.github.io/devtools-protocol/tot/Network/#event-responseReceived
				documentReady, responseReceived := false, false
				for !documentReady || !responseReceived {
					select {
					case e := <-responses:
						if e.LoaderID == loader && e.FrameID == frame {
							if e.Response.Status >= 400 {
								return fmt.Errorf("navigate: HTTP %d", e.Response.Status)
							}
							responseReceived = true
						}
					case e := <-ready:
						if e.LoaderID == loader && e.FrameID == frame {
							documentReady = true
						}
					case <-ctx.Done():
						return ctx.Err()
					}
				}
				return nil
			}),
			chromedp.WaitVisible(`body`, chromedp.ByQuery),
			chromedp.Sleep(waitAfterLoad),
			chromedp.Evaluate(metadataJS, &metadata),
			chromedp.Evaluate(jsCode, nil),
			chromedp.Evaluate(`window.extractArticle()`, &content, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
				return p.WithAwaitPromise(true)
			}),
			chromedp.Evaluate(nextJS, &nextHref),
		)
		if err == nil {
			err = ValidateContent(content)
		}
		// Defuddle may discard the challenge heading and leave only its prompt.
		// Retain original-document evidence, but allow independent article prose
		// beside security terminology instead of rejecting raw DOM keywords.
		if err == nil && IsBlockedPage(metadata.Text) {
			var prose []string
			for _, line := range strings.Split(content, "\n") {
				if !challengePrompt.MatchString(strings.TrimSpace(line)) && !IsBlockedPage(line) {
					prose = append(prose, line)
				}
			}
			if !HasArticleBody(strings.Join(prose, "\n")) {
				err = fmt.Errorf("blocked by anti-bot protection (Cloudflare/CAPTCHA)")
			}
		}
		return content, nextHref, err
	}

	identity, requestedPage, _, err := articlePage(startURL)
	if err != nil {
		return "", err
	}
	if maxPages <= 0 {
		maxPages = defaultMaxPages
	}
	current := startURL
	visited := map[string]bool{}
	var firstContent, firstNext string
	for visits := 0; ; visits++ {
		if visited[current] {
			return "", fmt.Errorf("cannot establish article first page: recovery loop at %s", current)
		}
		if visits >= maxPages {
			return "", fmt.Errorf("cannot establish article first page: recovery exceeds maxPages %d", maxPages)
		}
		visited[current] = true
		firstContent, firstNext, err = fetch(current)
		if err != nil {
			return "", fmt.Errorf("paginated extract: page %s: %w", current, err)
		}
		actualIdentity, position, candidate, parseErr := articlePage(metadata.URL)
		if parseErr != nil || actualIdentity != identity {
			return "", fmt.Errorf("article identity mismatch: page %s", current)
		}
		if metadata.Canonical != "" {
			if !compatibleCanonical(identity, metadata.Canonical, position) {
				return "", fmt.Errorf("article identity mismatch: canonical on %s", current)
			}
		}
		for _, earlier := range metadata.Earlier {
			earlierIdentity, earlierPage, _, earlierErr := articlePage(earlier)
			if earlierErr != nil || earlierIdentity != identity {
				return "", fmt.Errorf("article identity mismatch: earlier link on %s", current)
			}
			if earlierPage < position {
				candidate = earlier
			} else if earlierPage != position {
				return "", fmt.Errorf("cannot establish article first page: contradictory earlier link on %s", current)
			}
		}
		if position == 1 {
			break
		}
		current = candidate
	}
	// Discovery bodies are discarded except the established first page. The
	// normal complete-only walker starts afresh with its own page allowance.
	first := true
	reachedRequested := false
	forwardPage := 0
	content, err := paginate(current, maxPages, func(pageURL string) (string, string, error) {
		body, next := firstContent, firstNext
		var fetchErr error
		if !first {
			body, next, fetchErr = fetch(pageURL)
		}
		first = false
		forwardPage++
		if fetchErr != nil {
			return "", "", fetchErr
		}
		actualIdentity, position, _, identityErr := articlePage(metadata.URL)
		if requestedPage > 1 && (actualIdentity != identity || identityErr != nil || position != forwardPage) {
			return "", "", fmt.Errorf("article identity mismatch: page %s", pageURL)
		}
		if requestedPage > 1 && metadata.Canonical != "" {
			if !compatibleCanonical(identity, metadata.Canonical, position) {
				return "", "", fmt.Errorf("article identity mismatch: canonical on %s", pageURL)
			}
		}
		reachedRequested = reachedRequested || requestedPage == 1 || position == requestedPage
		return body, next, nil
	})
	if err != nil {
		return "", err
	}
	if !reachedRequested {
		return "", fmt.Errorf("cannot establish article first page: forward sequence omits requested page %s", startURL)
	}
	return content, nil
}

func (b *Browser) Close() {
	b.cancel()
}
