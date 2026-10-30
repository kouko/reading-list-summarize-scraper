package extract

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// SocialScopeNotice is written by the program into content and summaries.
const SocialScopeNotice = "Only the specified X post was extracted; author follow-up replies are not included."

const socialMarkerPrefix = "<!-- rlss-social-v1 "

type socialProvenance struct {
	Platform   string `json:"platform"`
	Target     string `json:"target"`
	Scope      string `json:"scope"`
	Source     string `json:"source"`
	BodySHA256 string `json:"body_sha256"`
}

var xStatusPath = regexp.MustCompile(`^/(?:[^/]+/status|i/web/status)/([0-9]+)/?$`)

func xTarget(address string) (*url.URL, string) {
	u, err := url.Parse(address)
	if err != nil || u.Scheme != "https" || u.User != nil {
		return nil, ""
	}
	host := strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
	if host != "x.com" && host != "twitter.com" {
		return nil, ""
	}
	match := xStatusPath.FindStringSubmatch(u.Path)
	if match == nil {
		return nil, ""
	}
	return u, match[1]
}

// Snapshot bodies are owned-source verified. Conversation extent and ancestry
// are deliberately irrelevant to this explicitly single-post result.
func focalSocial(address string, snapshot socialSnapshot) (string, error) {
	requested, target := xTarget(address)
	fail := func() (string, error) {
		return "", fmt.Errorf("Inaccessible social target: verified complete specified X post unavailable")
	}
	if requested == nil || snapshot.Version != 1 || snapshot.Platform != "x" || snapshot.TargetID != target {
		return fail()
	}
	var chosen *socialPost
	for _, p := range snapshot.Posts {
		if p.ID != target {
			continue
		}
		source, id := xTarget(p.URL)
		if chosen != nil || source == nil || id != target || source.Port() != requested.Port() || p.AuthorID == "" || !p.BodyComplete || ValidateContent(p.Body) != nil || !HasArticleBody(p.Body) {
			return fail()
		}
		copy := p
		chosen = &copy
	}
	if chosen == nil {
		return fail()
	}
	marker, _ := json.Marshal(socialProvenance{Platform: "x", Target: target, Scope: "single_post", Source: chosen.URL, BodySHA256: fmt.Sprintf("%x", sha256.Sum256([]byte(chosen.Body)))})
	content := socialMarkerPrefix + string(marker) + " -->\n" + SocialScopeNotice + "\n\n" + chosen.Body + "\n\n[Source](" + chosen.URL + ")"
	return content, ValidateSocialContent(address, content)
}

// ValidateSocialContent binds the adapter's exact body and source to the
// requested post. Legacy caches cannot supply an unlabelled success.
func ValidateSocialContent(address, content string) error {
	fail := func() error {
		return fmt.Errorf("Inaccessible social target: missing or invalid target-bound single-post provenance")
	}
	requested, target := xTarget(address)
	if requested == nil {
		return fail()
	}
	line, body, ok := strings.Cut(content, "\n")
	var p socialProvenance
	if !ok || !strings.HasPrefix(line, socialMarkerPrefix) || !strings.HasSuffix(line, " -->") || json.Unmarshal([]byte(strings.TrimSuffix(strings.TrimPrefix(line, socialMarkerPrefix), " -->")), &p) != nil || p.Platform != "x" || p.Target != target || p.Scope != "single_post" {
		return fail()
	}
	source, id := xTarget(p.Source)
	if source == nil || id != target || source.Port() != requested.Port() || !strings.HasPrefix(body, SocialScopeNotice+"\n\n") || !strings.HasSuffix(body, "\n\n[Source]("+p.Source+")") {
		return fail()
	}
	body = strings.TrimSuffix(strings.TrimPrefix(body, SocialScopeNotice+"\n\n"), "\n\n[Source]("+p.Source+")")
	if p.BodySHA256 != fmt.Sprintf("%x", sha256.Sum256([]byte(body))) || !HasArticleBody(body) {
		return fail()
	}
	return ValidateContent(body)
}

type socialPost struct {
	ID           string          `json:"id"`
	URL          string          `json:"url"`
	AuthorID     string          `json:"authorID"`
	Body         string          `json:"body"`
	BodyComplete bool            `json:"bodyComplete"`
	ParentID     json.RawMessage `json:"parentID"`
	RootID       string          `json:"rootID"`
	IsReply      *bool           `json:"isReply"`
	ReplyCount   *int            `json:"replyCount"`
}
type socialSequence struct {
	RootID     string   `json:"rootID"`
	IDs        []string `json:"ids"`
	End        bool     `json:"end"`
	Cursor     string   `json:"cursor"`
	MissingIDs []string `json:"missingIDs"`
}
type socialSnapshot struct {
	Version    int              `json:"version"`
	Platform   string           `json:"platform"`
	TargetID   string           `json:"targetID"`
	Posts      []socialPost     `json:"posts"`
	Sequences  []socialSequence `json:"sequences"`
	Branches   []socialSequence `json:"branches"`
	ReplyPages []socialSequence `json:"replyPages"`
}

func collectSocial(address string, limit int, fetch func(string) (socialSnapshot, error)) (string, error) {
	posts := map[string]socialPost{}
	loaded := map[string]bool{}
	var sequences, branches, pages []socialSequence
	platform, requested := "", ""
	fail := func(reason string) error { return fmt.Errorf("Inaccessible social target: %s", reason) }
	load := func(address string) error {
		if loaded[address] {
			return fail("unresolved member or navigation loop")
		}
		if len(loaded) >= limit {
			return fail("complete thread exceeds navigation limit")
		}
		loaded[address] = true
		snapshot, err := fetch(address)
		if err != nil {
			return fmt.Errorf("Inaccessible social target: member navigation: %w", err)
		}
		if snapshot.Version != 1 || (snapshot.Platform != "x" && snapshot.Platform != "threads") || (platform != "" && snapshot.Platform != platform) {
			return fail("unsupported source snapshot")
		}
		platform = snapshot.Platform
		matched := false
		for _, p := range snapshot.Posts {
			if p.ID == snapshot.TargetID && sameSocialURL(address, p.URL) {
				matched = true
			}
			if p.ID == "" || p.AuthorID == "" {
				return fail("missing source identity")
			}
			if old, ok := posts[p.ID]; ok {
				if old.ReplyCount != nil && p.ReplyCount != nil && *old.ReplyCount != *p.ReplyCount {
					return fail("conflicting source reply count")
				}
				if old.IsReply != nil && p.IsReply != nil && *old.IsReply != *p.IsReply {
					return fail("conflicting source reply identity")
				}
				if old.AuthorID != p.AuthorID || !sameSocialURL(old.URL, p.URL) || (old.BodyComplete && p.BodyComplete && old.Body != p.Body) || (len(old.ParentID) > 0 && len(p.ParentID) > 0 && string(old.ParentID) != string(p.ParentID)) || (old.RootID != "" && p.RootID != "" && old.RootID != p.RootID) {
					return fail("conflicting source member")
				}
				if old.BodyComplete && !p.BodyComplete {
					p.Body, p.BodyComplete = old.Body, true
				}
				if len(p.ParentID) == 0 {
					p.ParentID = old.ParentID
				}
				if p.RootID == "" {
					p.RootID = old.RootID
				}
				if p.IsReply == nil {
					p.IsReply = old.IsReply
				}
				if p.ReplyCount == nil {
					p.ReplyCount = old.ReplyCount
				}
			}
			posts[p.ID] = p
		}
		if !matched {
			return fail("source target identity unavailable")
		}
		if requested == "" {
			requested = snapshot.TargetID
		}
		sequences = append(sequences, snapshot.Sequences...)
		branches = append(branches, snapshot.Branches...)
		pages = append(pages, snapshot.ReplyPages...)
		return nil
	}
	if err := load(address); err != nil {
		return "", err
	}
	current := requested
	ancestors := map[string]bool{}
	for {
		if ancestors[current] {
			return "", fail("ancestor loop")
		}
		ancestors[current] = true
		p := posts[current]
		parent := ""
		if platform == "threads" {
			if p.RootID != "" && p.RootID != current {
				parent = p.RootID
			} else {
				// A local self-thread can begin at a reply. Only explicit
				// source root identity establishes the article beginning.
				if p.IsReply == nil || *p.IsReply {
					return "", fail("article beginning is unknown")
				}
			}
		} else {
			if len(p.ParentID) == 0 {
				return "", fail("article beginning is unknown")
			}
			if string(p.ParentID) != "null" {
				if json.Unmarshal(p.ParentID, &parent) != nil || parent == "" {
					return "", fail("invalid parent identity")
				}
			}
		}
		if parent == "" {
			break
		}
		if _, ok := posts[parent]; !ok {
			if platform != "x" {
				return "", fail("ancestor URL unavailable")
			}
			u, err := url.Parse(p.URL)
			if err != nil {
				return "", fail("ancestor URL unavailable")
			}
			u.Path = u.Path[:strings.LastIndex(u.Path, "/")+1] + parent
			u.RawQuery = ""
			u.Fragment = ""
			if err := load(u.String()); err != nil {
				return "", err
			}
		}
		if posts[parent].AuthorID != posts[requested].AuthorID {
			return "", fail("target belongs to a reader conversation")
		}
		current = parent
	}
	root := current
	if !loaded[posts[root].URL] {
		if err := load(posts[root].URL); err != nil {
			return "", err
		}
	}
	var ids []string
	if platform == "threads" {
		var selected *socialSequence
		for i := range sequences {
			s := &sequences[i]
			if s.RootID != root {
				continue
			}
			if selected != nil && !reflect.DeepEqual(*selected, *s) {
				return "", fail("conflicting self-thread sequence")
			}
			selected = s
		}
		if selected == nil {
			if p := posts[root]; p.ReplyCount == nil || *p.ReplyCount != 0 {
				return "", fail("self-thread extent unavailable")
			}
			ids = []string{root}
		} else {
			if !selected.End || len(selected.MissingIDs) > 0 {
				return "", fail("self-thread pagination incomplete")
			}
			ids = append(ids, selected.IDs...)
		}
		// Direct author branches may contain prose omitted by self_thread. A
		// partially paginated reply collection cannot establish their extent.
		for _, p := range pages {
			if p.RootID == root && !p.End {
				return "", fail("direct reply pagination incomplete")
			}
		}
		for _, b := range branches {
			if b.RootID != root || len(b.IDs) == 0 {
				continue
			}
			first, ok := posts[b.IDs[0]]
			if !ok {
				return "", fail("missing direct reply member")
			}
			if first.AuthorID != posts[root].AuthorID {
				continue
			}
			if !b.End {
				return "", fail("author reply branch incomplete")
			}
			for _, id := range b.IDs {
				p, ok := posts[id]
				if !ok || p.AuthorID != posts[root].AuthorID || (p.RootID != "" && p.RootID != root) {
					return "", fail("unproven direct author branch")
				}
				ids = append(ids, id)
			}
		}
	} else {
		seen := map[string]bool{}
		pending := []string{root}
		for len(pending) > 0 {
			current = pending[0]
			pending = pending[1:]
			if seen[current] {
				return "", fail("continuation loop")
			}
			seen[current] = true
			p := posts[current]
			if !loaded[p.URL] {
				if err := load(p.URL); err != nil {
					return "", err
				}
				p = posts[current]
			}
			ids = append(ids, current)
			var children []string
			for id, candidate := range posts {
				var parent string
				if candidate.AuthorID == posts[root].AuthorID && json.Unmarshal(candidate.ParentID, &parent) == nil && parent == current {
					// X Snowflake IDs encode creation order. Relationships prove
					// membership; numeric ordering only orders proven siblings.
					// https://docs.x.com/fundamentals/x-ids
					childNumber, err := strconv.ParseUint(id, 10, 64)
					parentNumber, parentErr := strconv.ParseUint(current, 10, 64)
					if err != nil || parentErr != nil || childNumber <= parentNumber {
						return "", fail("contradictory author continuation order")
					}
					children = append(children, id)
				}
			}
			pending = append(pending, children...)
			sort.Slice(pending, func(i, j int) bool {
				a, _ := strconv.ParseUint(pending[i], 10, 64)
				b, _ := strconv.ParseUint(pending[j], 10, 64)
				return a < b
			})
		}
	}
	seen := map[string]bool{}
	var bodies []string
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		p, ok := posts[id]
		if !ok || p.AuthorID != posts[root].AuthorID {
			return "", fail("missing or unrelated member")
		}
		if !p.BodyComplete {
			if err := load(p.URL); err != nil {
				return "", err
			}
			p = posts[id]
		}
		if !p.BodyComplete || strings.TrimSpace(p.Body) == "" {
			return "", fail("member body remains incomplete")
		}
		if err := ValidateContent(p.Body); err != nil {
			return "", fmt.Errorf("Inaccessible social target: member body: %w", err)
		}
		bodies = append(bodies, p.Body+"\n\n[Source]("+p.URL+")")
	}
	if !seen[requested] || len(ids) == 0 || ids[0] != root {
		return "", fail("sequence omits requested member or beginning")
	}
	// Count every known explicit direct child, including reader replies, after
	// body recovery has had its final chance to reveal more source evidence.
	children := map[string]int{}
	for _, p := range posts {
		var parent string
		if json.Unmarshal(p.ParentID, &parent) == nil && parent != "" {
			children[parent]++
		}
	}
	for id, p := range posts {
		if p.ReplyCount != nil && children[id] > *p.ReplyCount {
			return "", fail("reply count contradicts known direct children")
		}
	}
	// Leaf replyCount=0 proves only that leaf ends. A rendered author chain
	// cannot prove there are no hidden author siblings earlier in the article.
	// replyPages must describe exhaustive first-party direct-reply pagination
	// for the named post, not display-group or visible-list exhaustion.
	for id := range seen {
		p := posts[id]
		if platform == "threads" && id != root {
			if (p.RootID != "" && p.RootID != root) || (p.IsReply != nil && !*p.IsReply) {
				return "", fail("contradictory article membership")
			}
			continue
		}
		if p.ReplyCount != nil && *p.ReplyCount == 0 {
			continue
		}
		complete := false
		for _, page := range pages {
			if page.RootID != id {
				continue
			}
			if !page.End {
				return "", fail("direct reply pagination incomplete")
			}
			complete = true
		}
		if !complete {
			return "", fail("complete reply extent is unknown")
		}
	}
	if platform == "threads" {
		for _, s := range sequences {
			if s.RootID != root {
				continue
			}
			if !s.End || len(s.MissingIDs) > 0 {
				return "", fail("self-thread pagination incomplete")
			}
			for _, id := range s.IDs {
				if !seen[id] {
					return "", fail("new member discovered during body recovery")
				}
			}
		}
		for _, b := range branches {
			if b.RootID != root || len(b.IDs) == 0 || posts[b.IDs[0]].AuthorID != posts[root].AuthorID {
				continue
			}
			if !b.End {
				return "", fail("author reply branch incomplete")
			}
			for _, id := range b.IDs {
				if !seen[id] {
					return "", fail("new author branch discovered during body recovery")
				}
			}
		}
	} else {
		// Navigating a child can reveal another author sibling after its
		// parent was walked. Earlier pagination observations cannot license
		// returning a body that omits this newly known component member.
		for id, p := range posts {
			var parent string
			if p.AuthorID == posts[root].AuthorID && json.Unmarshal(p.ParentID, &parent) == nil && seen[parent] && !seen[id] {
				return "", fail("new author member discovered during body recovery")
			}
		}
	}
	return strings.Join(bodies, "\n\n"), nil
}

func sameSocialURL(a, b string) bool {
	x, e := url.Parse(a)
	if e != nil {
		return false
	}
	y, e := url.Parse(b)
	if e != nil {
		return false
	}
	host := func(s string) string {
		s = strings.TrimPrefix(strings.ToLower(s), "www.")
		if s == "twitter.com" {
			return "x.com"
		}
		if s == "threads.net" {
			return "threads.com"
		}
		return s
	}
	return host(x.Hostname()) == host(y.Hostname()) && strings.TrimSuffix(x.Path, "/") == strings.TrimSuffix(y.Path, "/")
}
