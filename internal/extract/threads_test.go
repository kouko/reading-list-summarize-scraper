package extract

import (
	"fmt"
	"strings"
	"testing"
)

// concern: A complete social article requires source relationships and complete bodies.
func TestCompleteSocialThread(t *testing.T) {
	root := socialPost{ID: "1", URL: "https://x.com/author/status/1", AuthorID: "a", Body: "first", BodyComplete: true, ParentID: jsonParent("null")}
	middle := socialPost{ID: "2", URL: "https://x.com/author/status/2", AuthorID: "a", Body: "preview", ParentID: jsonParent(`"1"`)}
	last := socialPost{ID: "3", URL: "https://x.com/author/status/3", AuthorID: "a", Body: "same [source](https://example.org)", BodyComplete: true, ParentID: jsonParent(`"2"`), ReplyCount: intPointer(0)}
	for _, entry := range []string{root.URL, middle.URL} {
		calls := 0
		body, err := collectSocial(entry, 8, func(address string) (socialSnapshot, error) {
			calls++
			posts := []socialPost{root, middle, last, {ID: "reader", AuthorID: "reader", ParentID: jsonParent(`"1"`), Body: "reader", BodyComplete: true}}
			posts = append(posts, socialPost{ID: "4", URL: "https://x.com/author/status/4", AuthorID: "a", ParentID: jsonParent(`"1"`), Body: "distinct ending", BodyComplete: true, ReplyCount: intPointer(0)})
			target := "1"
			if strings.HasSuffix(address, "/2") {
				target = "2"
				posts[1].Body, posts[1].BodyComplete = "same [source](https://example.org)", true
			}
			if strings.HasSuffix(address, "/3") {
				target = "3"
			}
			if strings.HasSuffix(address, "/4") {
				target = "4"
			}
			return socialSnapshot{Version: 1, Platform: "x", TargetID: target, Posts: posts, ReplyPages: []socialSequence{{RootID: "1", End: true}, {RootID: "2", End: true}}}, nil
		})
		if err != nil || body != "first\n\n[Source](https://x.com/author/status/1)\n\nsame [source](https://example.org)\n\n[Source](https://x.com/author/status/2)\n\nsame [source](https://example.org)\n\n[Source](https://x.com/author/status/3)\n\ndistinct ending\n\n[Source](https://x.com/author/status/4)" || calls > 8 {
			t.Fatalf("%s: %q, %v (%d calls)", entry, body, err, calls)
		}
	}
	t.Run("threads explicit self thread and author branch", func(t *testing.T) {
		for _, entry := range []string{"one", "two"} {
			body, err := collectSocial("https://www.threads.com/@author/post/"+entry, 8, func(address string) (socialSnapshot, error) {
				p, q, r := root, middle, last
				p.URL = "https://www.threads.com/@author/post/one"
				p.IsReply = new(bool)
				q.URL = "https://www.threads.com/@author/post/two"
				q.RootID = "1"
				q.Body = "second"
				q.BodyComplete = true
				r.URL = "https://www.threads.com/@author/post/three"
				r.Body = "author supplement"
				target := "1"
				if strings.HasSuffix(address, "/two") {
					target = "2"
				}
				return socialSnapshot{Version: 1, Platform: "threads", TargetID: target, Posts: []socialPost{p, q, r}, Sequences: []socialSequence{{RootID: "1", IDs: []string{"1", "2"}, End: true, Cursor: "terminal"}}, Branches: []socialSequence{{RootID: "1", IDs: []string{"3"}, End: true}}, ReplyPages: []socialSequence{{RootID: "1", End: true}}}, nil
			})
			if err != nil || body != "first\n\n[Source](https://www.threads.com/@author/post/one)\n\nsecond\n\n[Source](https://www.threads.com/@author/post/two)\n\nauthor supplement\n\n[Source](https://www.threads.com/@author/post/three)" {
				t.Fatalf("%s: %q, %v", entry, body, err)
			}
		}
	})
	t.Run("unproven boundaries", func(t *testing.T) {
		for _, kind := range []string{"local reply sequence", "unknown beginning", "conflicting reply identity", "standalone self-thread member", "standalone direct branch", "late X sibling", "zero with author child", "zero with reader child", "changed explicit count", "late reader exceeds count"} {
			t.Run(kind, func(t *testing.T) {
				p, q, r := root, middle, last
				p.ReplyCount = intPointer(1)
				q.Body = "second"
				q.BodyComplete = true
				q.ReplyCount = intPointer(0)
				platform := "threads"
				p.URL = "https://www.threads.com/@author/post/one"
				q.URL = "https://www.threads.com/@author/post/two"
				p.IsReply = new(bool)
				q.RootID = "1"
				if strings.HasPrefix(kind, "zero with ") {
					p.ReplyCount = intPointer(0)
				}
				if kind == "zero with reader child" {
					q.AuthorID = "reader"
				}
				if kind == "changed explicit count" || kind == "late reader exceeds count" {
					q.BodyComplete = false
				}
				if kind == "local reply sequence" {
					*p.IsReply = true
				}
				if kind == "unknown beginning" {
					p.IsReply = nil
				}
				if kind == "conflicting reply identity" {
					q.BodyComplete = false
				}
				if strings.HasPrefix(kind, "standalone ") {
					q.IsReply = new(bool)
				}
				if kind == "late X sibling" {
					platform = "x"
					p.URL = root.URL
					q.URL = middle.URL
					r.ParentID = jsonParent(`"1"`)
				}
				body, err := collectSocial(p.URL, 8, func(address string) (socialSnapshot, error) {
					target := "1"
					posts := []socialPost{p, q}
					if address == q.URL {
						target = "2"
						posts[1].BodyComplete = true
						if kind == "conflicting reply identity" {
							changed := true
							posts[0].IsReply = &changed
						}
						if kind == "late X sibling" {
							posts = append(posts, r)
						}
						if kind == "changed explicit count" {
							posts[0].ReplyCount = intPointer(2)
						}
						if kind == "late reader exceeds count" {
							r.AuthorID, r.ParentID = "reader", jsonParent(`"1"`)
							posts = append(posts, r)
						}
					}
					snapshot := socialSnapshot{Version: 1, Platform: platform, TargetID: target, Posts: posts, Sequences: []socialSequence{{RootID: "1", IDs: []string{"1", "2"}, End: true}}, ReplyPages: []socialSequence{{RootID: "1", End: true}}}
					if kind == "standalone direct branch" {
						snapshot.Sequences[0].IDs = []string{"1"}
						snapshot.Branches = []socialSequence{{RootID: "1", IDs: []string{"2"}, End: true}}
					}
					if kind == "zero with reader child" {
						snapshot.Sequences[0].IDs = []string{"1"}
					}
					return snapshot, nil
				})
				if err == nil || body != "" {
					t.Fatalf("%q, %v", body, err)
				}
			})
		}
	})
	t.Run("unavailable member", func(t *testing.T) {
		body, err := collectSocial(root.URL, 8, func(address string) (socialSnapshot, error) {
			if address != root.URL {
				return socialSnapshot{}, fmt.Errorf("unavailable")
			}
			return socialSnapshot{Version: 1, Platform: "x", TargetID: "1", Posts: []socialPost{root, middle, last}}, nil
		})
		if body != "" || err == nil || !strings.Contains(err.Error(), "Inaccessible social target") {
			t.Fatalf("%q, %v", body, err)
		}
	})
	t.Run("unrelated author branch", func(t *testing.T) {
		root.ReplyCount = intPointer(1)
		body, err := collectSocial(root.URL, 8, func(string) (socialSnapshot, error) {
			return socialSnapshot{Version: 1, Platform: "x", TargetID: "1", Posts: []socialPost{root, {ID: "reader", AuthorID: "b", ParentID: jsonParent(`"1"`)}, {ID: "aside", AuthorID: "a", ParentID: jsonParent(`"reader"`), Body: "conversation", BodyComplete: true}}, ReplyPages: []socialSequence{{RootID: "1", End: true}}}, nil
		})
		if err != nil || body != "first\n\n[Source]("+root.URL+")" {
			t.Fatalf("%q, %v", body, err)
		}
	})
	t.Run("cursor and cap", func(t *testing.T) {
		body, err := collectSocial("https://www.threads.com/@author/post/one", 1, func(string) (socialSnapshot, error) {
			p := root
			p.URL = "https://www.threads.com/@author/post/one"
			return socialSnapshot{Version: 1, Platform: "threads", TargetID: "1", Posts: []socialPost{p}, Sequences: []socialSequence{{RootID: "1", IDs: []string{"1"}, Cursor: "next"}}}, nil
		})
		if err == nil || body != "" {
			t.Fatalf("%q, %v", body, err)
		}
	})
	t.Run("complete bodies do not waive limits or unknown reply extent", func(t *testing.T) {
		for _, limit := range []int{1, 8} {
			body, err := collectSocial(root.URL, limit, func(address string) (socialSnapshot, error) {
				p, q := root, middle
				p.ReplyCount = intPointer(11)
				q.BodyComplete = true
				q.Body = "second"
				target := "1"
				if strings.HasSuffix(address, "/2") {
					target = "2"
				}
				if strings.HasSuffix(address, "/3") {
					target = "3"
				}
				return socialSnapshot{Version: 1, Platform: "x", TargetID: target, Posts: []socialPost{p, q, last}}, nil
			})
			if err == nil || body != "" {
				t.Fatalf("limit %d: %q, %v", limit, body, err)
			}
		}
	})
}

func jsonParent(s string) []byte { return []byte(s) }
func intPointer(n int) *int      { return &n }

func TestFocalSocialSnapshot(t *testing.T) {
	address := "https://x.com/author/status/2"
	for _, kind := range []string{"plain", "rich", "wrong-target", "preview", "empty", "wall", "foreign-url", "scheme", "port", "duplicate", "threads"} {
		t.Run(kind, func(t *testing.T) {
			p := socialPost{ID: "2", URL: address, AuthorID: "a", Body: "Owned text.\n\n[Reference](https://example.org)", BodyComplete: true}
			s := socialSnapshot{Version: 1, Platform: "x", TargetID: "2", Posts: []socialPost{{ID: "1", Body: "Excluded previous post"}, p, {ID: "3", Body: "Excluded next post"}}}
			switch kind {
			case "rich":
				s.Posts[1].Body = "# Owned title\n\nEntire owned rich article."
			case "wrong-target":
				s.TargetID = "1"
			case "preview":
				s.Posts[1].BodyComplete = false
			case "empty":
				s.Posts[1].Body = ""
			case "wall":
				s.Posts[1].Body = "Please sign in to read the full article."
			case "foreign-url":
				s.Posts[1].URL = "https://example.org/author/status/2"
			case "scheme":
				s.Posts[1].URL = "http://x.com/author/status/2"
			case "port":
				s.Posts[1].URL = "https://x.com:9999/author/status/2"
			case "duplicate":
				s.Posts = append(s.Posts, p)
			case "threads":
				s.Platform = "threads"
			}
			body, err := focalSocial(address, s)
			if kind != "plain" && kind != "rich" {
				if err == nil || body != "" {
					t.Fatalf("accepted %s: %q %v", kind, body, err)
				}
				return
			}
			if err != nil || strings.Contains(body, "Excluded") || !strings.Contains(body, s.Posts[1].Body) {
				t.Fatalf("focal %q %v", body, err)
			}
			if err := ValidateSocialContent(address, body); err != nil {
				t.Fatal(err)
			}
			for _, invalid := range []string{strings.Replace(body, "single_post", "complete_thread", 1), strings.Replace(body, "Owned", "Changed", 1), strings.Replace(body, "Only the specified X post", "All posts", 1), strings.TrimPrefix(body, strings.Split(body, "\n")[0]+"\n")} {
				if ValidateSocialContent(address, invalid) == nil {
					t.Fatalf("invalid provenance accepted: %q", invalid)
				}
			}
			if ValidateSocialContent("https://x.com/author/status/3", body) == nil {
				t.Fatal("wrong target accepted")
			}
		})
	}
}
