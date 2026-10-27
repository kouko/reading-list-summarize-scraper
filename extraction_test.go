package rlss_test

import (
	"os/exec"
	"testing"
)

// TestExtractionBodies runs the unchanged JavaScript body regressions.
func TestExtractionBodies(t *testing.T) {
	output, err := exec.Command("node", "--test", "extraction.test.mjs").CombinedOutput()
	t.Logf("JavaScript extraction regressions:\n%s", output)
	if err != nil {
		t.Fatalf("JavaScript extraction regressions failed: %v", err)
	}
}
