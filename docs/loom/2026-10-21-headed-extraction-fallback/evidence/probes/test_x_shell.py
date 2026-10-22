# concern: An X page shell heading must not turn authentication UI into article content.
import json
from pathlib import Path
import subprocess
import tempfile
import unittest


class XShell(unittest.TestCase):
    def test_extraction_xshell_rejectswall(self):
        """Reject authentication-only X output even when it retains its Post heading."""
        root = Path(__file__).resolve().parents[5]
        source = '''package pipeline
import "testing"
func TestAdversarialXShell(t *testing.T) {
    wall := "## Post\\n\\n## Log in or sign up for X\\n\\n[Log in](https://x.com/i/flow/login)\\n[Sign up](https://x.com/i/flow/signup)"
    s := &stubExtractor{first: wall, retry: wall}
    got, err := extractWithFallback(s, "https://x.com/example/status/123", true)
    if err == nil || got != "" {
        t.Fatalf("authentication-only shell accepted: body=%q err=%v", got, err)
    }
}
'''
        with tempfile.TemporaryDirectory() as tmp:
            trial = Path(tmp) / "shell_test.go"
            trial.write_text(source)
            overlay = Path(tmp) / "overlay.json"
            overlay.write_text(json.dumps({"Replace": {
                str(root / "internal/pipeline/adversarial_shell_test.go"): str(trial)
            }}))
            result = subprocess.run(
                ["go", "test", "-overlay", str(overlay), "./internal/pipeline",
                 "-run", "^TestAdversarialXShell$", "-count=1"],
                cwd=root, capture_output=True, text=True, timeout=90,
            )
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
