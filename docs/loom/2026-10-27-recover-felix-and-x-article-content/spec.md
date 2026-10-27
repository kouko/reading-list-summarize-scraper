# Recover Felix and X article content — spec
intent: 2026-10-27-recover-felix-and-x-article-content@02b94b0
confirmed-behavior: 2026-10-27 @f499ebc
pre-build-review: required — Correct target identity and completeness govern the saved-content contract; ambiguous social layouts can otherwise save unrelated or partial content.

## Requirements
REQ-1 — Complete target X content
  WHEN any of the eight X samples below is processed, the scraper shall extract that target's complete original body and existing sections, excluding other posts, replies and login interfaces, with each result independently compared against the actual source body rather than inferred from a success message → Acceptance #1

REQ-2 — Complete Felix article
  WHEN the Felix sample below is processed, the scraper shall extract its complete original body and existing sections despite incidental page metadata, with the result independently compared against the actual source body → Acceptance #2

REQ-3 — Reject unverified content safely
  IF the requested target identity or complete body cannot be established THEN the scraper shall report extraction failure without saving a preview, partial body, unrelated content, error page, login page or verification page as a successful article, and shall preserve existing article and summary files on that failure → Acceptance #3

REQ-4 — Preserve existing extraction behavior
  WHEN previously supported ordinary articles or X content are processed, the scraper shall retain their successful extraction and existing headless-first, headed-retry and verification-page setting effects, using isolated configuration, browser data, executable and outputs during this change's verification while preserving the running process, production configuration and notes → Acceptance #4

## Design decision
- agent-decided — Extend the existing extraction entry point and narrowly repair demonstrated body loss, retaining target checks and existing legacy X paths; the failure is in content selection rather than a need for another service.
- agent-decided — Preserve substantive article containers when incidental metadata matches cleanup selectors, without globally disabling metadata cleanup; the Felix body is under a `data-date` container that the pinned parser removes.
- agent-decided — Recognize the observed X article layout only with target identity and scoped body evidence, rejecting ambiguous or truncated representations; a nonempty social page does not prove the requested article was obtained.
- agent-decided — Verify every listed sample independently; only Felix and shota have been diagnosed so far, so the other seven must not be assumed to share the same cause or fix.
- agent-decided — Retain dependencies, configuration schema, file formats and CLI surfaces unless an implementation finding requires a documented revision; no paid service, account requirement or existing-data migration is authorized.
- agent-decided — Keep copyrighted full source captures outside version control; use synthetic regression fixtures and source URLs plus measured comparison results in committed evidence.

### Acceptance samples
1. https://x.com/romandevz/status/2114655026190618718?s=12
2. https://x.com/hitomiiim/status/2114704821852991771?s=12
3. https://x.com/adrianpunk115/status/2114648961399349608?s=12
4. https://x.com/0xluffy_eth/status/2114705266759569624?s=12
5. https://x.com/kenichiota0711/status/2113959093220335871?s=12
6. https://x.com/katelyn_lesse/status/2114542280535318557?s=12
7. https://x.com/thorstenball/status/2114386050638336082?s=12
8. https://x.com/shota7180/status/2113460983100530808?s=12
9. https://felixrieseberg.com/how-to-be-more-ambitious-a-tutorial-for-humans-with-models/

## Alternatives considered
- Treat any nonempty visible page as success: rejected because navigation, replies and previews can be substantial while the requested body is missing.
- Globally disable parser cleanup: rejected because ordinary articles would retain unrelated metadata and page furniture.
- Add a paid API or require an account: rejected because the diagnosed bodies are publicly accessible in an isolated visible browser and the intent does not authorize new obligations.
- Declare all nine fixed after checking two: rejected because seven samples have not yet been independently diagnosed.

## Current state evidence
- Forward: `cmd/rlss/url.go`, `runURL` starts single-URL processing and explicitly disables skip-existing.
- Reverse: `internal/pipeline/extraction.go`, `extractWithFallback` calls the browser pool and validates content before saving.
- Error: `extraction.ts`, `social` rejects missing target bodies; ordinary extraction rejects empty Defuddle content. Isolated captures reproduce both failures with source bodies present.
- Data: `extraction.ts`, `sameTarget` and `bodyMarkdown` bind social content to a URL and convert scoped DOM to Markdown; the observed new layout uses `.x-article-body`, while Felix's body container carries `data-date`.
- Boundary: `internal/pipeline/extraction.go`, `extractWithFallback` retains existing retry and `headed_on_block` policy; production configuration, browser profiles, running binary and notes are outside this change's writes.

## UI flows
| Case | What the user does | What they see |
|---|---|---|
| Complete source | Run `rlss url <sample-url> --no-summary` | The complete target body and original sections appear in the article output; the report counts `Success: 1`, and exit status is 0. |
| Summarization enabled | Process a complete source with the existing summary configuration | Existing summary behavior continues; summary failure remains `Partial` rather than an extraction success claim for a missing body. |
| In progress | Start ordinary extraction in the default headless configuration | Existing progress remains; if extraction fails, the existing headed retry can open a visible browser. An explicit headed invocation retains its existing behavior. |
| Verification restriction | Encounter a recognized verification page | With `headed_on_block: false`, that blocked attempt does not open a headed retry; with it enabled, the existing headed retry is allowed. Ordinary extraction failures retain their existing retry behavior. |
| Missing or unverified body | Run the single-URL command for an inaccessible, partial or mismatched target | Extraction reports the existing error category (such as `Inaccessible social target: main body unavailable`), the report counts `Failed: 1`, and the command exits nonzero. Existing body and summary files remain intact. Retry the URL after access or source availability is restored. |
| Batch or watch | Run the existing batch command or continue watch mode | Each item retains the existing `Success`, `Failed`, `Skipped` or `Partial` reporting; a failed item does not stop later items. Batch per-item failure retains its existing exit behavior, and watch can retry on its next cycle. |
| Already processed | Run batch processing with skip-existing enabled | Existing completed items remain `Skipped`; explicitly using `rlss url <url>` reprocesses that URL under the existing rules. No bulk rewrite or migration is performed by this change. |
| Empty or dry run | Run with no selected URLs, or use the existing dry-run option | Existing no-item or dry-run reporting remains; no new extraction or content write is introduced. |
| Isolated verification | Verify these nine samples during development | Test outputs are separate from production notes; the running production process, configuration and browser data are preserved. The report lists each sample and its completeness comparison, including failures honestly. |

The nine samples must all meet their corresponding requirements before this change is reported as complete. An unavailable source is a failed or blocked sample, never evidence of complete extraction. Existing successful ordinary and legacy X content must continue to pass.
