package probes

// concern: Contradictory explicit thread membership must not merge another article's body.

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Test_collectSocial_conflictingRoot_rejects exercises the real collector with contradictory source relationships.
func Test_collectSocial_conflictingRoot_rejects(t *testing.T) {
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(root, "internal", "extract", "threads.go")); err == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			t.Fatal("repository root not found")
		}
		root = parent
	}
	dir := t.TempDir()
	// Copy production sources unchanged and reuse the repository's fixture helpers.
	for _, name := range []string{"threads.go", "content.go", "threads_test.go"} {
		data, err := os.ReadFile(filepath.Join(root, "internal", "extract", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	const attack = `package extract
import "testing"
// Test_collectSocial_conflictingRoot_rejects rejects a member explicitly assigned to another article.
func Test_collectSocial_conflictingRoot_rejects(t *testing.T) {
 root := socialPost{ID:"1",URL:"https://www.threads.com/@a/post/ROOT",AuthorID:"a",Body:"Root body.",BodyComplete:true,IsReply:new(bool),ReplyCount:intPointer(1)}
 child := socialPost{ID:"2",URL:"https://www.threads.com/@a/post/CHILD",AuthorID:"a",Body:"Other article body.",BodyComplete:true,RootID:"99"}
 body,err:=collectSocial(root.URL,8,func(string)(socialSnapshot,error){return socialSnapshot{Version:1,Platform:"threads",TargetID:"1",Posts:[]socialPost{root,child},Sequences:[]socialSequence{{RootID:"1",IDs:[]string{"1","2"},End:true}},ReplyPages:[]socialSequence{{RootID:"1",End:true}}},nil})
 if err==nil || body!="" {t.Fatalf("accepted contradictory membership: body=%q err=%v",body,err)}
}
`
	if err := os.WriteFile(filepath.Join(dir, "attack_test.go"), []byte(attack), 0600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("go", "test", "-json", "-run", "^Test_collectSocial_conflictingRoot_rejects$", "threads.go", "content.go", "threads_test.go", "attack_test.go")
	command.Dir = dir
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("collector attack failed: %v\n%s", err, output)
	}
}
