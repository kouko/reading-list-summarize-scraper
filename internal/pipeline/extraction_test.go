package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/kouko/reading-list-summarize-scraper/internal/config"
	"github.com/kouko/reading-list-summarize-scraper/internal/source"
)

type stubExtractor struct {
	headed   bool
	calls    []string
	first    string
	firstErr error
	retry    string
	retryErr error
}

func (s *stubExtractor) UsesHeaded(string) bool { return s.headed }
func (s *stubExtractor) ExtractURL(string) (string, error) {
	s.calls = append(s.calls, "initial")
	return s.first, s.firstErr
}
func (s *stubExtractor) ExtractURLHeaded(string) (string, error) {
	s.calls = append(s.calls, "headed")
	return s.retry, s.retryErr
}

func TestProcessItemWithFallback(t *testing.T) {
	loadErr := errors.New("net::ERR_HTTP_RESPONSE_CODE_FAILURE")
	for _, tt := range []struct {
		name          string
		s             stubExtractor
		want          string
		wantCalls     []string
		wantErr       bool
		blockDisabled bool
	}{
		{"success", stubExtractor{first: "article"}, "article", []string{"initial"}, false, false},
		{"load error", stubExtractor{firstErr: loadErr, retry: "article"}, "article", []string{"initial", "headed"}, false, false},
		{"empty", stubExtractor{first: " \n", retry: "article"}, "article", []string{"initial", "headed"}, false, false},
		{"challenge", stubExtractor{first: "Cloudflare CAPTCHA", retry: "article"}, "article", []string{"initial", "headed"}, false, false},
		{"login wall", stubExtractor{first: "Please sign in to read", retry: "article"}, "article", []string{"initial", "headed"}, false, false},
		{"headed failure", stubExtractor{headed: true, firstErr: loadErr}, "", []string{"initial"}, true, false},
		{"headed challenge", stubExtractor{headed: true, first: "Cloudflare CAPTCHA"}, "", []string{"initial"}, true, false},
		{"retry failure", stubExtractor{firstErr: loadErr, retryErr: loadErr}, "", []string{"initial", "headed"}, true, false},
		{"retry empty", stubExtractor{firstErr: loadErr}, "", []string{"initial", "headed"}, true, false},
		{"retry challenge", stubExtractor{firstErr: loadErr, retry: "Cloudflare CAPTCHA"}, "", []string{"initial", "headed"}, true, false},
		{"retry login wall", stubExtractor{firstErr: loadErr, retry: "Please sign in to read"}, "", []string{"initial", "headed"}, true, false},
		{"block disabled challenge", stubExtractor{first: "Cloudflare CAPTCHA", retry: "article"}, "", []string{"initial"}, true, true},
		{"block disabled navigation", stubExtractor{firstErr: loadErr, retry: "article"}, "article", []string{"initial", "headed"}, false, true},
		{"block disabled empty", stubExtractor{retry: "article"}, "article", []string{"initial", "headed"}, false, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			cfg := &config.Config{OutputDir: dir, Extract: config.ExtractConfig{HeadedOnBlock: !tt.blockDisabled}}
			p := New(cfg, &tt.s, nil)
			err := p.ProcessItem(source.ReadingItem{URL: "https://example.com/article"})
			if (err != nil) != tt.wantErr {
				t.Fatalf("got error %v, want error=%v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.s.calls, tt.wantCalls) {
				t.Fatalf("calls=%v, want %v", tt.s.calls, tt.wantCalls)
			}
			files, err := filepath.Glob(filepath.Join(dir, "*", "*"))
			if err != nil {
				t.Fatal(err)
			}
			if tt.wantErr {
				if len(files) != 0 {
					t.Fatalf("failed extraction wrote files: %v", files)
				}
				return
			}
			if len(files) != 1 || !strings.HasSuffix(files[0], "__content.md") {
				t.Fatalf("expected content only, no summary; got %v", files)
			}
			data, err := os.ReadFile(files[0])
			if err != nil || stripFrontmatter(string(data)) != tt.want {
				t.Fatalf("saved content=%q, err=%v; want %q", data, err, tt.want)
			}
		})
	}
}

func TestExtractionRejectsObservedWalls(t *testing.T) {
	// Derived from captured X authentication chrome; no post body included.
	xAuth := "## Log in or sign up for X\n\nSee what’s happening and join the conversation\n\n" +
		"[Continue with phone](https://x.com/i/jf/onboarding/web?mode=signup&redirect_after_login=%2Fexample)\n\nContinue with Google\n\n" +
		"透過 Google 帳戶繼續操作透過 Google 帳戶繼續操作。在新分頁中開啟\n\n<iframe src=\"https://accounts.google.com/gsi/button?theme=filled_black\"></iframe>\n\nor\n\n" +
		"[Log in with username or email](https://x.com/i/jf/onboarding/web?mode=login&redirect_after_login=%2Fexample)\n\n" +
		"[Terms](https://x.com/tos)·[Privacy](https://x.com/privacy)·© 2026 X Corp.\n\n<iframe style=\"display: none;\"></iframe>"
	for _, content := range []string{
		"## to, err:= human()\n\ngo.dev uses cookies from Google to deliver and enhance the quality of its services and to analyze traffic.",
		"## Log in or sign up for X\n\n[Log in](https://x.com/i/flow/login)\n[Sign up](https://x.com/i/flow/signup)",
		"## Log in or sign up for X\n\n[Log in](https://x.com/login)\n[Sign up](https://x.com/i/flow/signup)",
		"[Log in](https://x.com/i/flow/login)\n[Sign up](https://x.com/i/flow/signup)",
		xAuth,
	} {
		if err := validateExtractedContent(content); err == nil {
			t.Errorf("accepted wall: %q", content)
		}
	}
	article := "## Designing worker teams\n\nWorkers exchange files and review results.\n\n" + xAuth
	if err := validateExtractedContent(article); err != nil {
		t.Fatalf("rejected article with login chrome: %v", err)
	}
}
