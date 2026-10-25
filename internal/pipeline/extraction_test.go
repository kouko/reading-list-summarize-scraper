package pipeline

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/kouko/reading-list-summarize-scraper/internal/config"
	"github.com/kouko/reading-list-summarize-scraper/internal/extract"
	"github.com/kouko/reading-list-summarize-scraper/internal/output"
	"github.com/kouko/reading-list-summarize-scraper/internal/source"
)

type stubExtractor struct {
	headed   bool
	calls    []string
	urls     []string
	first    string
	firstErr error
	retry    string
	retryErr error
}

const verifiedTranscript = `<!-- rlss-video-v1 {"platform":"youtube","target":"abc","language":"ja","source":"automatic","evidence":"original-audio-asr-association"} -->` + "\n\nSubtitle language: ja\nSubtitle source: automatic\n\n今日は字幕について話します。"

func (s *stubExtractor) UsesHeaded(string) bool { return s.headed }
func (s *stubExtractor) ExtractURL(url string) (string, error) {
	s.calls = append(s.calls, "initial")
	s.urls = append(s.urls, url)
	return s.first, s.firstErr
}
func (s *stubExtractor) ExtractURLHeaded(url string) (string, error) {
	s.calls = append(s.calls, "headed")
	s.urls = append(s.urls, url)
	return s.retry, s.retryErr
}

func TestProcessItemWithFallback(t *testing.T) {
	loadErr := errors.New("net::ERR_HTTP_RESPONSE_CODE_FAILURE")
	transcript := verifiedTranscript
	for _, tt := range []struct {
		name          string
		s             stubExtractor
		want          string
		wantCalls     []string
		wantErr       bool
		blockDisabled bool
	}{
		{"success", stubExtractor{first: "article"}, "article", []string{"initial"}, false, false},
		{"video verified", stubExtractor{headed: true, first: transcript}, transcript, []string{"initial"}, false, false},
		{"video description", stubExtractor{headed: true, first: "A description without subtitles"}, "", []string{"initial"}, true, false},
		{"video metadata only", stubExtractor{headed: true, first: strings.Split(transcript, "今日は")[0]}, "", []string{"initial"}, true, false},
		{"video wrong target", stubExtractor{headed: true, first: transcript}, "", []string{"initial"}, true, false},
		{"social body", stubExtractor{first: "The target post explains how workers exchange files."}, "The target post explains how workers exchange files.", []string{"initial"}, false, false},
		{"specialized challenge disabled", stubExtractor{firstErr: errors.New("Inaccessible social target: Cloudflare CAPTCHA"), retry: "article"}, "", []string{"initial"}, true, true},
		{"load error", stubExtractor{firstErr: loadErr, retry: "article"}, "article", []string{"initial", "headed"}, false, false},
		{"pagination retry replaces partial body", stubExtractor{first: "partial first attempt", firstErr: fmt.Errorf("pagination page 2: %w", loadErr), retry: "page one\n\npage two\n\npage three"}, "page one\n\npage two\n\npage three", []string{"initial", "headed"}, false, false},
		{"pagination wrapped challenge disabled", stubExtractor{first: "partial first page", firstErr: fmt.Errorf("pagination page 2: %w", errors.New("blocked by anti-bot protection (Cloudflare/CAPTCHA)")), retry: "article"}, "", []string{"initial"}, true, true},
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
			rawURL := "https://example.com/article"
			if strings.HasPrefix(tt.name, "pagination") {
				rawURL += "?page=3"
			}
			if strings.HasPrefix(tt.name, "video") {
				rawURL = "https://www.youtube.com/watch?v=abc"
				if tt.name == "video wrong target" {
					rawURL = "https://youtu.be/other"
				}
			}
			if tt.name == "social body" || tt.name == "specialized challenge disabled" {
				rawURL = "https://x.com/example/status/123"
			}
			err := p.ProcessItem(source.ReadingItem{URL: rawURL})
			if (err != nil) != tt.wantErr {
				t.Fatalf("got error %v, want error=%v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.s.calls, tt.wantCalls) {
				t.Fatalf("calls=%v, want %v", tt.s.calls, tt.wantCalls)
			}
			for _, gotURL := range tt.s.urls {
				if gotURL != rawURL {
					t.Fatalf("attempt URL=%q, want original %q", gotURL, rawURL)
				}
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
			if strings.HasPrefix(tt.name, "pagination") {
				if !strings.Contains(string(data), fmt.Sprintf("url: %q\n", rawURL)) || !strings.Contains(filepath.Base(files[0]), output.SHA8(rawURL)) {
					t.Fatalf("saved output lost supplied URL identity: %s\n%s", files[0], data)
				}
				p.RebuildIndex()
				if p.index.ContentPath(output.SHA8(rawURL)) != files[0] || p.index.Has(output.SHA8("https://example.com/article")) {
					t.Fatal("output index replaced the supplied page URL with the first page")
				}
			}
		})
	}
}

func TestProcessBatchFailedPaginationPreservesFiles(t *testing.T) {
	dir := t.TempDir()
	item := source.ReadingItem{URL: "https://example.com/article?page=3"}
	out := filepath.Join(dir, output.DomainDir(item.URL))
	if err := os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	now, sha := time.Now(), output.SHA8(item.URL)
	files := map[string]string{
		output.ContentFilename(now, sha): "---\ntype: reading-list-content\n---\n\nExisting complete article",
		output.SummaryFilename(now, sha): "---\ntype: reading-list-summary\n---\n\nExisting summary",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(out, name), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	cause := errors.New("later page navigation failed")
	s := &stubExtractor{
		first: "partial initial body", firstErr: fmt.Errorf("pagination page 2: %w", cause),
		retry: "partial retry body", retryErr: fmt.Errorf("pagination page 3: %w", cause),
	}
	cfg := &config.Config{OutputDir: dir, Extract: config.ExtractConfig{HeadedOnBlock: true}}
	stats := New(cfg, s, nil).ProcessBatch([]source.ReadingItem{item})
	if stats.Success != 0 || stats.Partial != 0 || stats.Skipped != 0 || stats.Failed != 1 || len(stats.Errors) != 1 {
		t.Fatalf("incomplete pagination misreported: %+v", stats)
	}
	if !errors.Is(stats.Errors[0].Err, cause) || !strings.Contains(stats.Errors[0].Err.Error(), "pagination page 3") {
		t.Fatalf("lost pagination failure: %v", stats.Errors[0].Err)
	}
	if !reflect.DeepEqual(s.calls, []string{"initial", "headed"}) {
		t.Fatalf("unexpected retries: %v", s.calls)
	}
	if !reflect.DeepEqual(s.urls, []string{item.URL, item.URL}) {
		t.Fatalf("retry did not restart supplied page: %v", s.urls)
	}
	entries, err := os.ReadDir(out)
	if err != nil || len(entries) != len(files) {
		t.Fatalf("output files changed: %v, err=%v", entries, err)
	}
	for name, want := range files {
		body, err := os.ReadFile(filepath.Join(out, name))
		if err != nil || string(body) != want {
			t.Fatalf("existing file %s changed: %q, err=%v", name, body, err)
		}
	}
}

func TestProcessItemVideoResume(t *testing.T) {
	for _, mode := range []string{"pending", "verified", "completed", "dry-run"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			item := source.ReadingItem{URL: "https://youtu.be/abc"}
			out := filepath.Join(dir, output.DomainDir(item.URL))
			if err := os.MkdirAll(out, 0755); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(out, output.ContentFilename(time.Now(), output.SHA8(item.URL)))
			if err := os.WriteFile(path, []byte("---\ntype: reading-list-content\n---\n\nOld video description"), 0644); err != nil {
				t.Fatal(err)
			}
			if mode == "verified" {
				if err := os.WriteFile(path, []byte("---\ntype: reading-list-content\n---\n\n"+verifiedTranscript), 0644); err != nil {
					t.Fatal(err)
				}
			}
			if mode == "completed" {
				if err := os.WriteFile(filepath.Join(out, output.SummaryFilename(time.Now(), output.SHA8(item.URL))), []byte("---\ntype: reading-list-summary\n---\n\nDone"), 0644); err != nil {
					t.Fatal(err)
				}
			}
			s := &stubExtractor{headed: true, firstErr: errors.New("Unverified original language")}
			cfg := &config.Config{OutputDir: dir, Pipeline: config.PipelineConfig{SkipExisting: true, DryRun: mode == "dry-run"}}
			err := New(cfg, s, nil).ProcessItem(item)
			if mode == "pending" {
				if err == nil || !reflect.DeepEqual(s.calls, []string{"initial"}) {
					t.Fatalf("unverified cache reused: err=%v calls=%v", err, s.calls)
				}
			} else if mode == "verified" {
				if err != nil || len(s.calls) != 0 {
					t.Fatalf("verified cache not resumed: err=%v calls=%v", err, s.calls)
				}
			} else if !IsSkipped(err) || len(s.calls) != 0 {
				t.Fatalf("skip changed: err=%v calls=%v", err, s.calls)
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
		if err := extract.ValidateContent(content); err == nil {
			t.Errorf("accepted wall: %q", content)
		}
	}
	article := "## Designing worker teams\n\nWorkers exchange files and review results.\n\n" + xAuth
	if err := extract.ValidateContent(article); err != nil {
		t.Fatalf("rejected article with login chrome: %v", err)
	}
}
