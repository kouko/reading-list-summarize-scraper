# concern: Login prompts accompanying genuine article content must not reject the body.
import json
from pathlib import Path
import subprocess
import tempfile
import unittest


class ArticlePrompt(unittest.TestCase):
    def test_extraction_loginprompt_acceptsbody(self):
        """Exercise the real fallback with an article followed by incidental login UI."""
        root = Path(__file__).resolve().parents[5]
        source = '''package pipeline
import "testing"
func TestAdversarialArticlePrompt(t *testing.T) {
    article := "## How coastal wetlands recover\\n\\nResearchers measured sediment accumulation across twelve restored wetlands. Native vegetation returned after tidal flow was restored, and the report compares annual monitoring results with nearby reference sites.\\n\\nPlease sign in to read more articles."
    s := &stubExtractor{first: article, retry: article}
    got, err := extractWithFallback(s, "https://example.com/article", true)
    if err != nil || got != article || len(s.calls) != 1 {
        t.Fatalf("real body with incidental prompt rejected: body=%q err=%v calls=%v", got, err, s.calls)
    }
}
'''
        with tempfile.TemporaryDirectory() as tmp:
            trial = Path(tmp) / "article_test.go"
            trial.write_text(source)
            overlay = Path(tmp) / "overlay.json"
            overlay.write_text(json.dumps({"Replace": {
                str(root / "internal/pipeline/adversarial_article_test.go"): str(trial)
            }}))
            result = subprocess.run(
                ["go", "test", "-overlay", str(overlay), "./internal/pipeline",
                 "-run", "^TestAdversarialArticlePrompt$", "-count=1"],
                cwd=root, capture_output=True, text=True, timeout=90,
            )
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
