package rlss_test

// concern: The native suite entry executes the unchanged earned Facebook identity regression and fails when its JavaScript tests fail.

import (
	"os/exec"
	"testing"
)

func TestExtraction(t *testing.T) {
	output, err := exec.Command("node", "--test", "extraction.test.mjs", "test_facebook_conflictingPermalink_rejected.mjs").CombinedOutput()
	t.Logf("JavaScript extraction tests:\n%s", output)
	if err != nil {
		t.Fatalf("JavaScript extraction tests failed: %v", err)
	}
}
