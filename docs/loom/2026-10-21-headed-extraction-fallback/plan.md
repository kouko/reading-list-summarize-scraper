# Headed extraction fallback — plan
intent: 2026-10-21-headed-extraction-fallback@ca2e7cd
spec: docs/loom/2026-10-21-headed-extraction-fallback/spec.md@de2eec2
charter: 1.1

## Task DAG
### Wave 0 — browser recovery
**W0-01 Preserve browser recovery settings** after: none acceptance: 1, 2
- Files: internal/extract/pool.go (entry: ExtractURLHeaded), internal/extract/pool_test.go
- Test: A1 positive: headed-any-domain; boundary: explicit-headed. A2 positive: preserved-pagination-profile; negative: browser-failure.
- Risk: agent-decided — REQ-1/REQ-2 preserve selected profile and pagination; close a conflicting headless browser before headed startup; verify timeout routing without a live browser dependency.

### Wave 1 — validate and retry
**W1-01 Validate article content and apply bounded retry** after: W0-01 acceptance: 1, 2, 3, 4
- Files: internal/pipeline/extraction.go (entry: extractWithFallback), internal/pipeline/extraction_test.go, internal/pipeline/runner.go, internal/pipeline/pipeline_test.go, internal/config/config.go, config.example.yaml, README.md
- Test: A1 positive: navigation-recovery; negative: block-disabled. A2 positive: recovered-body; boundary: already-headed. A3 positive: terminal-error; negative: invalid-retry. A4 positive: genuine-x-body; negative: login-verification-cache.
- Risk: agent-decided — REQ-1/REQ-4 preserve existing pipeline_test.go recognizer cases, widen wall detection conservatively, validate cached content, and retain headed_on_block semantics. Adopt existing uncommitted changes only after test-first verification.

### Wave 2 — article/authentication distinction
**W2-01 Preserve genuine bodies and reject authentication shells** after: W1-01 acceptance: 4
- Files: internal/pipeline/runner.go (entry: isUnavailablePage), internal/pipeline/pipeline_test.go
- Test: A4 positive: incidental-generic-login-prompt; negative: x-post-shell.
- Risk: agent-decided — REQ-4 fix the whole wall/body confusion class across generic and X checks; preserve existing wall-only rejection and quoted-topic negative cases. Adversarial programs remain unchanged.

**W2-02 Graduate adversarial regressions into the package suite** after: W2-01 acceptance: 4
- Files: internal/pipeline/adversarial_test.go, docs/loom/2026-10-21-headed-extraction-fallback/evidence/probes/test_article_prompt.py, docs/loom/2026-10-21-headed-extraction-fallback/evidence/probes/test_x_shell.py
- Test: A4 positive: article-prompt-preserved; negative: authentication-shell-rejected.
- Risk: agent-decided — move the two adversary-authored injected Go test bodies into a permanent package test without changing assertions; retire only their temporary Python overlay runners after equivalent RED/GREEN verification.

### Wave 3 — closing review repairs
**W3-01 Reject authentication navigation as body evidence** after: W2-02 acceptance: 4
- Files: internal/pipeline/runner.go (entry: isUnavailablePage), internal/pipeline/pipeline_test.go
- Test: A4 positive: substantive-body-with-prompt; negative: cached-wall-with-ui-links.
- Risk: agent-decided — REQ-4 exclude link-only, navigation, and authentication chrome from body evidence; cover privacy and password-reset examples across initial, retry, and cache validation without weakening real-body cases.

**W3-02 Verify recovery through the pipeline entry** after: W3-01 acceptance: 1, 2, 3, 4
- Files: internal/pipeline/runner.go (entry: ProcessItem), internal/pipeline/extraction_test.go
- Test: A1 positive: entry-recovery; boundary: block-disabled. A2 positive: saved-body; boundary: headed-once. A3 positive: returned-error; negative: output-on-failure. A4 positive: real-body; negative: wall-saved.
- Risk: agent-decided — existing extraction interface becomes the pipeline dependency; move the fallback table to ProcessItem and assert file results, removing duplicate helper-only cases.

## Simplicity check
- none found

## Questions asked
① — what — 正確性、速度、費用、隱私，你的優先順序是什麼？最不能接受的失敗結果是什麼？這次修改有沒有排擠其他工作？
① — consequence — 依 Loom 流程，確認後代表授權在審查通過後推送分支並建立 PR，合併另行決定；若只要本機完成，可以直接說「不要發布」。
① — behaviour — 若網站封鎖、缺少登入權限或網路中斷，重試後仍無法取得正文，就明確回報失敗，不產生錯誤摘要——這個處理界線可以嗎？
① — what — 另外，這次修改是否需要暫緩其他工作？沒有的話回答「沒有」即可。
① — what — 最後確認：所有網站先用 headless，失敗改用 headed；仍抓不到正文就回報失敗，不把登入／驗證頁當正文或產生摘要。無法保證每次成功，但以正確性優先。這樣符合你的需求嗎？
② — behaviour — 以上行為符合你的預期嗎？

## Risks
1. Detection is evidence-based rather than a guarantee against arbitrary website changes; recognized wall pages must never be summarized, and false positives must be tested against real article fixtures.
2. Existing headed overrides and headed_on_block false are user-controlled constraints; retries must not silently override them.
3. Verification uses temporary output directories and public content; no vault writes or automated website login.
