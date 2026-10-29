package extract

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
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
	if isSocialPostURL(url) {
		return b.extractSocial(ctx, url, jsCode, waitAfterLoad)
	}

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

func isSocialPostURL(address string) bool {
	u, err := url.Parse(address)
	if err != nil {
		return false
	}
	host := strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
	return ((host == "x.com" || host == "twitter.com") && strings.Contains(u.Path, "/status/")) || ((host == "threads.com" || host == "threads.net") && strings.Contains(u.Path, "/post/"))
}

// Every member uses this attempt's tab and deadline. The walker returns no
// content until both source membership and every member body are complete.
func (b *Browser) extractSocial(ctx context.Context, address, jsCode string, waitAfterLoad time.Duration) (string, error) {
	return collectSocial(address, 32, func(memberURL string) (socialSnapshot, error) {
		member, _ := url.Parse(memberURL)
		threads := strings.HasPrefix(strings.TrimPrefix(member.Hostname(), "www."), "threads.")
		// A navigation owns its listener and captures; no response crosses documents.
		listenCtx, stop := context.WithCancel(ctx)
		defer stop()
		events := make(chan any, 64)
		if threads {
			chromedp.ListenTarget(listenCtx, func(ev any) {
				switch e := ev.(type) {
				case *network.EventResponseReceived:
					u, err := url.Parse(e.Response.URL)
					if err != nil || u.Scheme != "https" || strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.") != strings.TrimPrefix(strings.ToLower(member.Hostname()), "www.") || u.Port() != member.Port() || u.Path != "/api/graphql" || e.Response.Status != 200 || (e.Type != network.ResourceTypeXHR && e.Type != network.ResourceTypeFetch) {
						return
					}
					ev = e.RequestID // Retain no headers or unrelated response metadata.
				case *network.EventLoadingFinished:
				default:
					return
				}
				select {
				case events <- ev:
				default:
				}
			})
		}
		var snapshot socialSnapshot
		err := chromedp.Run(ctx,
			chromedp.ActionFunc(func(ctx context.Context) error {
				if threads {
					return network.Enable().WithMaxResourceBufferSize(16 << 20).WithMaxTotalBufferSize(16 << 20).Do(ctx)
				}
				return nil
			}),
			chromedp.Navigate(memberURL),
			chromedp.WaitVisible(`body`, chromedp.ByQuery),
			chromedp.Sleep(waitAfterLoad),
			chromedp.Evaluate(jsCode, nil),
			chromedp.Evaluate(`(async()=>{
try { return window.socialSnapshot(); }
catch (sourceError) {
 // Reuse the existing extractor's verification-page classification only.
 // A successful single-post fallback never establishes a complete thread.
 try { await window.extractArticle(); } catch (classifiedError) { throw classifiedError; }
 throw sourceError;
}
})()`, &snapshot, func(p *runtime.EvaluateParams) *runtime.EvaluateParams { return p.WithAwaitPromise(true) }),
		)
		if err == nil && threads {
			snapshot, err = loadThreadsReplies(ctx, events, snapshot)
		} else if err == nil {
			snapshot, err = loadXReplies(ctx, snapshot)
		}
		stop()
		return snapshot, err
	})
}

// Only source-proven target extent ends loading. Movement and native controls
// acquire observations; neither proves completeness or article membership.
func loadXReplies(ctx context.Context, snapshot socialSnapshot) (socialSnapshot, error) {
	idle := 0
	for round := 0; round < 32 && idle < 8; round++ {
		for _, page := range snapshot.ReplyPages {
			if page.RootID == snapshot.TargetID && page.End {
				return snapshot, nil
			}
		}
		before, _ := json.Marshal(snapshot)
		var activity struct {
			Clicked bool    `json:"clicked"`
			Top     float64 `json:"top"`
		}
		var top float64
		err := chromedp.Run(ctx,
			page.BringToFront(),
			chromedp.Evaluate(`(()=>{
let progress=false;
const labels=new Set(['Show probable spam','Show probable spam replies','スパムの可能性がある返信を表示','顯示可能的垃圾訊息']);
for(const e of document.querySelectorAll('button,[role="button"]')){
 if(e.closest('article,a[href],[role="link"]')||e.hasAttribute('href')||e.disabled||e.__rlssReplyExpanded)continue;
 const rect=e.getBoundingClientRect();
 if(rect.width<=0||rect.height<=0||getComputedStyle(e).visibility==='hidden')continue;
 if(!labels.has((e.innerText||e.getAttribute('aria-label')||'').trim()))continue;
 e.__rlssReplyExpanded=true;e.click();progress=true;
}
const top=scrollY;window.scrollBy(0,Math.max(innerHeight*.8,400));
return {clicked:progress,top};
})()`, &activity),
			chromedp.Sleep(750*time.Millisecond),
			chromedp.Evaluate(`window.socialSnapshot()`, &snapshot),
			chromedp.Evaluate(`scrollY`, &top),
		)
		if err != nil {
			return snapshot, err
		}
		after, _ := json.Marshal(snapshot)
		if activity.Clicked || top != activity.Top || string(before) != string(after) {
			idle = 0
		} else {
			idle++
		}
	}
	return snapshot, nil // Existing collector rejects unresolved reply extent.
}

// Native requests are observed, never replayed. The source parser remains the
// authority for response ownership and cursor continuity.
func loadThreadsReplies(ctx context.Context, events <-chan any, snapshot socialSnapshot) (socialSnapshot, error) {
	type supplement struct {
		PostID    string          `json:"postID"`
		After     string          `json:"after"`
		SortOrder string          `json:"sortOrder"`
		Data      json.RawMessage `json:"data"`
	}
	pages := make([]supplement, 0)
	responses := make(map[network.RequestID]bool)
	total := 0
	idleScrolls := 0
	for len(pages) < 32 {
		pending := map[string]string{}
		for _, page := range snapshot.ReplyPages {
			if !page.End && page.Cursor != "" {
				pending[page.RootID] = page.Cursor
			}
		}
		if len(pending) == 0 {
			return snapshot, nil
		}
		if err := chromedp.Run(ctx, page.BringToFront(), chromedp.Evaluate(`(()=>{window.scrollBy(0,Math.max(innerHeight,600));for(const e of document.querySelectorAll('main,div,section')){if(e.scrollHeight>e.clientHeight+50&&e.clientHeight>0&&['auto','scroll'].includes(getComputedStyle(e).overflowY))e.scrollTop+=Math.max(e.clientHeight,600)}})()`, nil)); err != nil {
			return snapshot, err
		}
		timer := time.NewTimer(750 * time.Millisecond)
		progress := false
		waiting := true
		for waiting && !progress {
			select {
			case ev := <-events:
				switch e := ev.(type) {
				case network.RequestID:
					if len(responses) < 32 {
						responses[e] = true
					}
				case *network.EventLoadingFinished:
					if !responses[e.RequestID] {
						continue
					}
					delete(responses, e.RequestID)
					if e.EncodedDataLength > float64(16<<20-total) {
						timer.Stop()
						return snapshot, fmt.Errorf("Threads reply payload limit")
					}
					var body []byte
					var request []byte
					err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
						var err error
						request, err = network.GetRequestPostData(e.RequestID).Do(ctx)
						if err != nil {
							return err
						}
						body, err = network.GetResponseBody(e.RequestID).Do(ctx)
						return err
					}))
					if err != nil {
						continue
					}
					if len(request)+len(body) > (16<<20)-total {
						timer.Stop()
						return snapshot, fmt.Errorf("Threads reply payload limit")
					}
					var page supplement
					form, _ := url.ParseQuery(string(request))
					variables := form.Get("variables")
					if variables == "" {
						var envelope struct {
							Variables json.RawMessage `json:"variables"`
						}
						_ = json.Unmarshal([]byte(request), &envelope)
						variables = string(envelope.Variables)
					}
					if json.Unmarshal([]byte(variables), &page) != nil || page.PostID == "" || pending[page.PostID] != page.After || page.After == "" || page.SortOrder != "TOP" {
						continue
					}
					body = []byte(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(string(body)), "for (;;);")))
					if !json.Valid(body) {
						timer.Stop()
						return snapshot, fmt.Errorf("invalid Threads reply response")
					}
					page.Data = json.RawMessage(body)
					total += len(body) + len(request)
					pages = append(pages, page)
					encoded, _ := json.Marshal(pages)
					err = chromedp.Run(ctx, chromedp.Evaluate(`window.socialSnapshot(`+string(encoded)+`)`, &snapshot))
					if err != nil {
						timer.Stop()
						return snapshot, err
					}
					progress = true
				}
			case <-timer.C:
				waiting = false
			case <-ctx.Done():
				timer.Stop()
				return snapshot, ctx.Err()
			}
		}
		timer.Stop()
		if progress {
			idleScrolls = 0
		} else {
			idleScrolls++
			if idleScrolls >= 8 {
				return snapshot, nil // collectSocial rejects the unresolved source.
			}
		}
	}
	for _, page := range snapshot.ReplyPages {
		if !page.End {
			return snapshot, fmt.Errorf("Threads reply page limit")
		}
	}
	return snapshot, nil
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
