# Real-site recovery validation

At code `0ecf509` on 2026-10-25 (Asia/Taipei), five of six later-page entries recovered the complete checked target prose. Toyokeizai page two failed safely without output, so Acceptance #1 is not yet satisfied. All twelve first-page entries exited successfully. The individual URLs, selectors, page counts, hashes and results are in [recovery-cases.json](recovery-cases.json).

| Entry | Source pages | CLI exit | Target blocks present / checked | Unique page anchors in order |
|---|---:|---:|---:|---|
| ITmedia page 3 | 5 | 0 | 61 / 61 | Yes |
| ITmedia page 5, max_pages 5 | 5 | 0 | 61 / 61 | Yes |
| PRESIDENT page 2 | 6 | 0 | 38 / 38 | Yes |
| PRESIDENT page 3 | 6 | 0 | 38 / 38 | Yes |
| Yahoo page 2 | 2 | 0 | 5 / 5 | Yes |
| Toyokeizai page 2 | 3 | 1 | 0 / 22; no saved body | No output |

The Toyokeizai failure was reproduced in both the initial headless attempt and the existing headed retry: `cannot establish article first page: forward sequence omits requested page`. The successful first-page baseline uses `?display=b`; recovery from `?page=2` does not establish that complete chain. This difference is an unresolved product failure, not a waived content difference. No production change is included in this evidence task.

The twelve first-page articles cover ten site groups and 34 source pages. They matched 280/280 qualifying blocks from the separately collected per-page browser observations, with all 34 page anchors appearing once in source order. Current independent HTTP observations re-confirmed 279 of those 280 blocks. One 290-normalized-character block from Mynavi page two was not found in the fresh HTTP text and remains unconfirmed; its presence in the CLI output does not prove fresh-source correctness. All other checked blocks and all chosen page-order anchors are grounded in the fresh source observations. ITmedia PC USER, Business and Mobile count as one site group; Yahoo is a distributor, not an additional independent original publisher.

ITmedia checks extract every target-body `p`, `h2` and `h3` block longer than 35 characters between the live HTML `cmsBodyMainStart` and `cmsBodyMainEnd` boundaries, excluding scripts and styles. The other nine articles reuse independently collected per-page browser prose blocks of at least 60 normalized characters, then check their presence in freshly fetched HTTP text as well as CLI output. Images, HTML, tables, headings and lists are excluded from that prose sample. Unicode compatibility, whitespace, Markdown formatting, quote prefixes and link destinations are normalized. The five successful recovery outputs contain the same checked target prose as their first-page outputs, in the same page order; changing ads and other page furniture are outside the target-body comparison. This is not an exhaustive identity guarantee for every source element.

Three negative controls passed: a real article with max_pages 1 exited 1 and saved no content; a missing configured next selector with standard next metadata exited 1 and saved no content; forced failed extraction with synthetic existing content and summary preserved both files byte for byte. The JSON records their SHA256 values. No run timed out. The initial twenty-case harness allowed at most two simultaneous CLI processes; starting the separate file-preservation control briefly raised the total to three, contrary to the planned limit. It completed normally, and no unrelated successful case was rerun to hide that deviation.

Commands executed:

```text
GOCACHE=/private/tmp/rlss-go-build go build -o /private/tmp/rlss-first-page-recovery-cli ./cmd/rlss
python3 /private/tmp/rlss-recovery-live.py
python3 /private/tmp/rlss-recovery-check.py
python3 /private/tmp/rlss-recovery-evidence.py
python3 /private/tmp/rlss-recovery-preserve.py
```

The actual CLI pattern was `/private/tmp/rlss-first-page-recovery-cli --config /private/tmp/rlss-recovery-live/<case>/config.yaml --no-summary url <public-url>`; the preservation control additionally used `--force`. All 21 runs used isolated profiles, configs and outputs. Timeout was 30s, headed timeout 60s, post-load wait 2s, and headed_on_block false. max_pages was 10 except the first-page cap-one control and ITmedia final-page positive control at five. Summaries, reading lists, RSS and copy-to were disabled. The root executable, production configuration, existing notes and real browser profiles were untouched.

Temporary source HTML, source-block lists, raw saved bodies, logs, run metadata and helpers remain under `/private/tmp/rlss-recovery-live/` and the named temporary helper paths. Only sanitized metadata and this account are committed. An initial `go build .` failed because the root has no Go files; the correct command above succeeded. An initial source-check helper stopped when a later case directory did not exist, then completed after directory creation was fixed; neither tooling error was counted as a product result. No full package suite or adversarial probe was run by this evidence task.
