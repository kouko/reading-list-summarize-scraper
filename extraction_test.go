package rlss_test

import (
	"os/exec"
	"testing"
)

func TestExtraction(t *testing.T) {
	output, err := exec.Command("node", "--test", "extraction.test.mjs", "facebook-identity.test.mjs").CombinedOutput()
	t.Logf("JavaScript extraction tests:\n%s", output)
	if err != nil {
		t.Fatalf("JavaScript extraction tests failed: %v", err)
	}
}
