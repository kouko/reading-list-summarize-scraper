# Headed extraction fallback — acceptance test evidence

Tested 2026-10-22 in a fresh `git archive 60c3b0b475e2addd5e93d2c938af696d05f79119` copy at `/private/tmp/rlss-acceptance.GYZ6eU`. No product fixes. No AI calls, vault writes, automated login, or authentication changes. All browser user-data and output paths were under that temporary directory.

Setup followed README's quick source build: `go build -o rlss ./cmd/rlss/`, exit 0. Initial sandbox attempts could not write the existing Go cache; host-approved build and focused tests succeeded. Committed Defuddle bundle was used.

## Commands and captured results

Focused criteria command, run in the clean copy:

```sh
go test -v ./internal/pipeline ./internal/extract ./internal/config -run 'Test(ExtractWithFallback|ExtractionRejectsObservedWalls|ProcessItemRejectsCachedWall|IsUnavailablePage|IsBlockedPage|Adversarial|Pool|DefaultConfig_SanityCheck|Load_ValidYAML|Load_DomainRulePagination|ClassifyResult|Stats_Partial)' -count=1
```

Exit 0; all selected tests PASS. Packages: pipeline 0.403s, extract 0.747s, config 1.113s. These include load-error/empty/challenge/login recovery, successful first extraction, already-headed failures, failed/empty/invalid retry, disabled verification retry, recognized wall fixtures, cached-wall preservation, both permanent adversarial regressions, profile/pagination/timeout/lifecycle routing, and configuration defaults.

Temporary browser config: headless true, `chrome_profile: Default`, `user_data_dir` temporary `/profile`, timeout 25s, headed timeout 40s, post-load wait 3s, headed_on_block true, output temporary `/output`, copy_to false. Every CLI call used `--no-summary`.

### Live public target

```sh
./rlss url https://x.com/Sukiea1008/status/2112084040362897441 --config acceptance.yaml --no-summary
```

```text
13:34:41 WARN headless extraction failed, retrying with headed mode
err="page load error net::ERR_HTTP_RESPONSE_CODE_FAILURE"
13:34:46 INFO wrote content
13:34:46 INFO no summarizer configured, skipping summary
Pipeline completed in 7s
Total: 1  Success: 1  Skipped: 0  Partial: 0  Failed: 0
```

Exit 0. One retry warning. Saved `x_com/2026-10-22__62a601b4__content.md`; frontmatter content_length 17621. Read-back found the requested article heading “搭建你的 Bot 团队：8 种身份的 Grok Bot 多 bot 实战手册”, substantive introductory paragraphs, all eight scenario sections, and concluding article paragraphs, rather than an authentication shell. No full copyrighted article is copied into this evidence.

### First-attempt success

`./rlss url https://example.com --config acceptance.yaml --no-summary`: exit 0, 4s, Success 1 / Failed 0, wrote `example_com/2026-10-22__100680ad__content.md`, no retry warning.

### Real-browser synthetic verification wall

Fixture URL:

```text
data:text/html,<html><body><article><h1>Access denied</h1><p>Cloudflare CAPTCHA. Please verify you are human.</p></article></body></html>
```

`./rlss url '<fixture URL>' --config acceptance.yaml --no-summary`: exit 1, 7s; exactly one `headless extraction failed, retrying with headed mode` warning; both attempts reported `blocked by anti-bot protection (Cloudflare/CAPTCHA)`; Total 1 / Success 0 / Failed 1.

Same command with `--headed`: exit 1, 3s; blocked error, no retry warning; Success 0 / Failed 1.

Same command with `disabled.yaml` (headed_on_block false, separate temporary profile/output, post-load wait 1s): exit 1, 1s; blocked error, no retry warning; Success 0 / Failed 1.

File inventory after these runs: only the live X and example.com content files exist; no summaries and no synthetic-wall files. The disabled-output directory was never created.

### Temporary batch witness

`go test -v ./internal/pipeline -run '^TestAcceptanceBatchCachedWallContinues$' -count=1`: exit 0, PASS, pipeline 0.477s. The temporary clean-copy-only probe `test_batch_cached_wall_continues` wrote two temporary cached contents: `Please sign in to read`, followed by a genuine wetland-report paragraph. It constructed the normal pipeline with no browser or summarizer, then called ProcessBatch. Captured output:

```text
processing item index=1 total=2
resuming: content exists, skipping extraction
item failed ... cached content: page unavailable (maintenance / paywall / login required)
processing item index=2 total=2
resuming: content exists, skipping extraction
no summarizer configured, skipping summary
Total: 2  Success: 1  Skipped: 0  Partial: 0  Failed: 1
PASS
```

The first version of this temporary probe referenced a nonexistent Stats.Total field and did not compile; corrected only the probe assertion to check Failed and Success. No product source changed. Existing `TestProcessItemRejectsCachedWall` separately asserts original cached bytes remain unchanged.

## 1. All-site recovery

Live X navigation failure recovered. `TestExtractWithFallback` exercises generic example.com without a domain rule; `TestPoolHeadedRecoveryLifecycle` verifies mode lifecycle. This verifies generic routing, not universal website availability.

## 2. Single-invocation public recovery

Live target command, captured warning, exit 0 and read-back above demonstrate the actual requested article was saved after automatic retry, without a manual second invocation.

## 3. Explicit terminal failure

Synthetic browser wall: exit 1, Failed 1, no file. Focused retry failure/empty/wall subcases all PASS. Temporary batch witness confirms next-item continuation.

## 4. Reject wall-only content

Synthetic browser verification rejection and no-file inventory above. `TestExtractionRejectsObservedWalls`, `TestIsUnavailablePage_MaintenanceAndWalls`, `TestIsBlockedPage_*`, `TestProcessItemRejectsCachedWall`, `TestAdversarialXShell`, and `TestAdversarialArticlePrompt` PASS. Authentication shell fixtures are synthetic/previously observed shapes, not a live logged-out X failure in this run.

## 5. Initial success flow

Live example.com command above: saved content without retry warning. `TestExtractWithFallback/success` confirms one initial call only.

## 6. First-failure progress flow

Live X warning includes initial navigation error and announces headed retry. Pool lifecycle and profile/pagination focused tests PASS. Headed timeout 40s selected in temporary config; timeout routing verified by pool test, not by measuring a timeout on live X.

## 7. Retry-success flow

Live X saved substantive target body in the same run. AI summarization intentionally disabled for safe acceptance execution; normal continuation is evidenced by `no summarizer configured, skipping summary` after valid save.

## 8. Final error / empty flow

Synthetic wall command exit 1 with Failed 1; retry-empty focused test PASS. Batch witness above: failed cached item followed by successful second item. No summary provider was invoked.

## 9. Login or verification only flow

Real-browser synthetic verification page rejected. Login-only shapes rejected by focused fixtures. No automatic login attempted.

## 10. Login prompt plus article flow

`TestAdversarialArticlePrompt` and `TestExtractionRejectsObservedWalls` preserve genuine bodies alongside generic/X login wording. Live requested X article also saved. Live response need not contain the exact fixture login chrome; the fixed-content regression supplies that specific distinction.

## 11. Already-visible flow

Synthetic fixture `--headed` run: one attempt, exit 1, no retry warning. `TestPoolExplicitHeaded` and already-headed fallback subcases PASS, including domain-rule routing.

## 12. Disabled anti-bot switch flow

Disabled config browser run: no retry warning, exit 1, no output directory. `TestExtractWithFallbackBlockDisabled/navigation` and `/empty` PASS, confirming these still receive a generic retry while `/challenge` does not.

## 13. Enabled anti-bot switch flow

Enabled config browser run: exactly one retry warning, wall rejected both times, no saved fixture.

## 14. Resume flow

`TestProcessItemRejectsCachedWall` PASS: cached wall returns extraction failure and old bytes remain unchanged. Temporary batch witness PASS: one rejected cache does not block the next cached article.

## 15. Automated suite gate

Focused tests above PASS. Complete required command is `go test ./...`; acceptance tester deliberately leaves this to finalize-review, which runs before acceptance and blocks attestation on failure. This row does not claim a full-suite run by this tester. Station additionally reports a previous full-suite and vet pass at the tested product commit; final gate remains authoritative.

## Limits and discarded harness attempts

- New isolated profile contains no inherited main-profile credentials. Main-profile state was not observed; inherited/default/explicit profile and pagination routing are covered by focused pool tests, not a live credential comparison.
- Localhost HTTP fixture server was unreachable (`curl --max-time 3 -I http://127.0.0.1:18764/article` also timed out). Its CLI attempt correctly failed both navigations but does not count as wall or initial-success evidence. Switched to a data URL for deterministic real-browser wall validation.
- One wall CLI was accidentally started while that localhost attempt still held the same temporary profile; it failed with SingletonLock. Discarded as a harness collision, not product wall behavior. Usable wall checks were rerun sequentially or with separate profile directories after completion.
- Fixed-content recognition does not guarantee arbitrary future pages are classified perfectly. No important findings were dismissed by the station; both adversarial defects are fixed and permanent regression cases passed.
