// concern: The declared Go suite must execute the earned wrong-target and malformed-subtitle regressions.
package rlss_test

import (
	"os/exec"
	"testing"
)

// TestExtractionBoundaries runs the unchanged JavaScript boundary regressions.
func TestExtractionBoundaries(t *testing.T) {
	output, err := exec.Command("node", "--test", "extraction-boundaries.test.mjs").CombinedOutput()
	t.Logf("JavaScript extraction regressions:\n%s", output)
	if err != nil {
		t.Fatalf("JavaScript extraction regressions failed: %v", err)
	}
}
