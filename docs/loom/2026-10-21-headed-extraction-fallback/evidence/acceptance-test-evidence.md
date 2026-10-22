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

## Re-run on 2026-10-22, at a8eab1e

Reviewed fix range `ae5de46..a8eab1e`: shared body detection now discards link-only authentication/navigation controls; pipeline dependency accepts the extraction interface; fallback/flag cases now enter ProcessItem and inspect output files. Because that boundary serves every row, all rows were re-tested rather than carried over. Prior live-browser observations remain historical evidence, not claims of repeated browser checks where this re-run uses focused pipeline fixtures.

Fresh `git archive a8eab1e` copy: `/private/tmp/rlss-acceptance-round2.TbcVUA`. README quick build `go build -o rlss ./cmd/rlss/` exit 0. Temporary profiles/output only; no AI, vault writes, login, or authentication changes.

Focused command:

```sh
go test -v ./internal/pipeline ./internal/extract -run 'Test(ProcessItemWithFallback|ProcessItemRejectsCachedWall|ExtractionRejectsObservedWalls|IsUnavailablePage|IsBlockedPage|Adversarial|Pool)' -count=1
```

Exit 0; pipeline 0.402s, extract 0.719s, all selected tests PASS. `TestProcessItemWithFallback` covers 14 cases: initial success, load error, empty, challenge, login wall, already-headed failure/challenge, failed/empty/challenge/login retry, and disabled-switch challenge/navigation/empty. Successful cases assert exactly one saved content file with expected body and no summary; failed cases assert no files. Calls assert initial-only or exactly one headed retry. `TestProcessItemRejectsCachedWall` tests plain X auth plus generic privacy/reset links and X privacy links on initial/retry/cache, rejecting them and preserving cached bytes. Both permanent adversarial regressions and wall/body recognizers PASS. Pool profile/pagination/timeout/lifecycle cases PASS.

Independent temporary probe `test_pipeline_auth_controls_rejected_and_real_body_preserved` (`TestAcceptanceRound2BodyStates`) ran through ProcessItem for 15 combinations: three wall-only bodies (generic privacy, generic password reset, X privacy), two genuine bodies with generic/X login prompts, each as initial response, retry response after navigation error, and existing cache. `go test -v ./internal/pipeline -run '^TestAcceptanceRound2BodyStates$' -count=1`: exit 0, all 15 PASS, 0.393s. Walls return errors and create no new files; genuine bodies remain eligible and produce content only; cached bytes remain identical and make no extraction calls. Probe docstring is the English name above; no product file changed.

Temporary batch probe `test_batch_cached_wall_continues` repeated with generic wall plus privacy link followed by real cached prose. `go test -v ./internal/pipeline -run '^TestAcceptanceBatchCachedWallContinues$' -count=1`: exit 0, PASS, 0.379s. Captured `cached content: page unavailable` for first item, then `processing item index=2 total=2`; Total 2 / Failed 1 / Success 1, no summarizer configured.

Live target repeated in fresh isolated profile with the same temporary config settings as before:

```sh
./rlss url https://x.com/Sukiea1008/status/2112084040362897441 --config acceptance.yaml --no-summary
```

```text
13:49:26 WARN headless extraction failed, retrying with headed mode
err="page load error net::ERR_HTTP_RESPONSE_CODE_FAILURE"
13:49:32 INFO wrote content
13:49:32 INFO no summarizer configured, skipping summary
Pipeline completed in 8s
Total: 1  Success: 1  Skipped: 0  Partial: 0  Failed: 0
```

Exit 0. Read-back: content_length 17621, target article heading and all eight scenario headings remain present. This is live public recovery proof, while auth/privacy/reset distinctions above remain synthetic fixtures.

- 1: re-tested — generic ProcessItem recovery, pool routing and live X recovery all PASS.
- 2: re-tested — live X initial failure followed by one successful retry saved target body in one invocation.
- 3: re-tested — ProcessItem final load/empty/wall failures return errors and assert zero output files.
- 4: re-tested — initial/retry/cache privacy/reset/X walls rejected, real bodies preserved, cached bytes unchanged; verification recognizers PASS.
- 5: re-tested — ProcessItem success case asserts one initial call and expected saved content, no summary.
- 6: re-tested — live warning and successful visible retry; focused pool timeout/profile/pagination cases PASS.
- 7: re-tested — live target saved; independent retry-state genuine-body cases save content only.
- 8: re-tested — final-error/empty ProcessItem cases assert errors and no files; batch witness verifies next-item continuation. CLI exit mapping is unchanged in the inspected diff; previous actual nonzero CLI result remains supporting history.
- 9: re-tested — all observed wall recognizers, new link-control wall variants and ProcessItem challenge/login cases PASS.
- 10: re-tested — generic and X genuine-body+login prompts preserved in initial/retry/cache states; live X saved again.
- 11: re-tested — headed failure/challenge cases call initial only, return errors and create no files; explicit pool/domain-rule routing PASS.
- 12: re-tested — disabled challenge fails without retry/files, while disabled navigation/empty recover with exactly one retry and save expected body.
- 13: re-tested — enabled challenge can recover; terminal retry-challenge fails with no files.
- 14: re-tested — four permanent cached-wall shapes and independent body-state matrix preserve bytes; batch proceeds after rejected cache.
- 15: re-tested — focused tests PASS; required `go test ./...` remains the finalize-review gate, not a suite run claimed by this tester.

No dismissed important findings were supplied. No new product findings. Main-profile state remains unobservable and heuristic limitations remain as described above.
