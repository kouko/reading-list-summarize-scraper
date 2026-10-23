# Real Japanese news pagination validation

Latest validation on 2026-10-24 (Asia/Taipei), using code commit `62c4765c7c670530356403f201b7b70af455896e`: twelve first-page articles passed body and ordering checks; two incomplete controls and all four later-page entries exited 1 without content or summaries. The historical runs below at `07bf97e59b93899fb50030bd98617b74291ef8c4` preserve the original later-page false-success defect. The final section records its resolved behavior; neither run set guarantees detection of every arbitrary entry page.

The corpus contains twelve articles, ten distinct site groups and 34 observed source pages. ITmedia PC USER, Business and Mobile count as one group. Yahoo is a distribution site; ten site groups does not mean ten independent original publishers. Public article URLs, selectors, per-page counts, output hashes and individual results are in [news-pagination-cases.json](news-pagination-cases.json).

| Site / article | Source pages | CLI exit | Seconds | Body blocks present / checked | Outcome |
| --- | ---: | ---: | ---: | ---: | --- |
| ITmedia PC USER (original) | 5 | 0 | 21.68 | 61 / 61 | PASS |
| PRESIDENT Online | 6 | 0 | 29.53 | 38 / 38 | PASS |
| Gendai Business | 2 | 0 | 6.26 | 14 / 14 | PASS |
| Bunshun Online | 2 | 0 | 17.21 | 14 / 14 | PASS |
| JBpress | 2 | 0 | 8.95 | 21 / 21 | PASS |
| Toyokeizai Online | 3 | 0 | 20.79 | 22 / 22 | PASS |
| Yahoo News | 2 | 0 | 14.22 | 5 / 5 | PASS |
| ITmedia Business (additional) | 3 | 0 | 13.67 | 32 / 32 | PASS |
| ITmedia Mobile (additional) | 3 | 0 | 10.79 | 10 / 10 | PASS |
| ASCII.jp (additional group) | 2 | 0 | 17.83 | 34 / 34 | PASS |
| Mynavi News (additional group) | 2 | 0 | 15.57 | 27 / 27 | PASS |
| Nikkan Gendai (additional group) | 2 | 0 | 10.71 | 2 / 2 | PASS |

PASS is bounded by the checks below. A zero exit code alone was not sufficient. No blocked, incomplete or mismatched case was reclassified as successful verification. These selected examples did not encounter an access restriction during the current runs; this is not a guarantee of future access or a new extractor for every publisher.

## Execution and isolation

Built a fresh executable with `go build -o /private/tmp/rlss-pagination-fix/rlss-w3 ./cmd/rlss`. The actual run pattern was `/private/tmp/rlss-pagination-fix/rlss-w3 --config /private/tmp/rlss-pagination-fix/w3/<case>/config.yaml --no-summary url <public-url>`. Fourteen runs used separate fresh configs, browser-data directories and output directories under that temporary root, with at most two CLI processes at once. No root executable, existing notes, real browser profile or production configuration was modified. Safari, Chrome Reading List, RSS and copy-to sources were disabled. Summary generation was disabled.

The twelve corpus domain rules used the selectors recorded in the JSON, `max_pages: 10`, headless mode, `timeout: 30s`, `headed_timeout: 60s`, `wait_after_load: 2s` and `headed_on_block: false`. The two negative controls changed the cap or selector as shown below. The existing generic headed retry remained available for other failures. Per-case CLI logs and raw outputs remain at `/private/tmp/rlss-pagination-fix/w3/<case>/cli.log` and the relative content paths recorded in the JSON. Temporary helpers `run-w3.py`, `check-w3.py` and `prepare-evidence.py` orchestrated runs and comparisons; they are not repository tests.

## Body and ordering checks

The 34 source-page observations were collected separately using direct browser navigation. The original seven cases had already been observed before this fix; five additional articles were subsequently observed the same way. Those source observations are not runs of the fixed CLI. Their temporary `inspection.json`, page Markdown and DOM text live under `/private/tmp/rlss-pagination-check/results/dom-<case>/`. The current CLI outputs were compared with those per-page observations.

For all three ITmedia articles, a further independent HTTP fetch of every source page extracted target-body `p`, `h2` and `h3` text between `cmsBodyMainStart` and `cmsBodyMainEnd`, excluding scripts and styles. Every complete qualifying block longer than 35 characters appeared in the fresh CLI output: original article 61/61, Business 32/32 and Mobile 10/10. This check did not use Defuddle to choose the target body. The original article also had an earlier fixed-CLI run completing in approximately 15 seconds with 61/61 independent body blocks; the table reports the separate fresh run above.

For the other nine articles, every qualifying prose block with at least 60 normalized characters in the earlier per-page Markdown was checked against the fresh output. Images, HTML, tables, headings and list blocks were excluded. This checks substantial prose, not every short sentence, caption or source element. Nikkan's two large prose blocks each contain multiple paragraphs; its small block count is not a claim that its article has only two paragraphs. Unicode compatibility normalization, Markdown link destinations, formatting marks, whitespace and quote prefixes were normalized before matching.

Each of the 34 pages also supplied a matching prose anchor appearing exactly once in the combined output. Anchor offsets increase in source-page order; their hashes, occurrence counts and offsets are recorded in the JSON. This detects missing or duplicated page bodies through representative prose anchors, without claiming exhaustive byte identity or excluding all repeated publisher boilerplate.

Broad ITmedia Defuddle snapshots also contained changing advertising links, recommendation lists and comment-registration prompts. Their differences were inspected and separated from target article text. They were neither accepted as article-body evidence nor treated as missing target paragraphs. Raw article text and transient advertising URLs are not committed.

## Intentional incomplete results

| Real article control | CLI exit | Seconds | Saved content | Outcome |
| --- | ---: | ---: | ---: | --- |
| Original ITmedia, `max_pages: 1` | 1 | 9.78 | 0 files | PASS: continuation beyond cap rejected |
| Original ITmedia, selector `a.nonexistent-pagination-w3` | 1 | 6.31 | 0 files | PASS: missing configured next link contradicted by standard next hint |

Both controls used fresh isolated profiles and outputs. Logs identify the cap or missing-selector reason, including the existing retry attempt, and the final single-URL failure. No partial content or summary was saved.

A missing configured selector without an independent standard next-link hint remains indistinguishable from a terminal page. No automatic pagination discovery is claimed. The controlled version of that limitation belongs to the package-level verification; this report does not present it as a successful real-site detection. Live pages, extraction noise and access controls can change after this observation. Full package tests, adversarial checks and closing review are separate evidence.

## Historical failure at 07bf97e: entry URL is not the first page

After the user asked about starting from a later page, four additional actual CLI runs at that code revision reproduced false completion. Every run exited 0, reported Success 1 / Partial 0 / Failed 0, and wrote one content file despite missing earlier target-body pages. This violated the confirmed requirement to avoid presenting known incomplete extraction as success and was recorded as an unresolved correctness failure before the subsequent fix. The historical FAIL results below remain unchanged; current resolution is recorded separately.

| Entry URL | Seconds | Target prose saved from pages | Missing pages | Outcome |
| --- | ---: | --- | --- | --- |
| [ITmedia page 3](https://www.itmedia.co.jp/pcuser/articles/2606/08/news096_3.html) | 14.39 | 3–5 | 1–2 | FAIL, unresolved |
| [ITmedia page 5](https://www.itmedia.co.jp/pcuser/articles/2606/08/news096_5.html) | 3.74 | 5 | 1–4 | FAIL, unresolved |
| [PRESIDENT page 3](https://president.jp/articles/-/120600?page=3) | 20.54 | 3–6 | 1–2 | FAIL, unresolved |
| [Yahoo page 2](https://news.yahoo.co.jp/articles/04860bcd34a7f82d074799f216f7f76dbfd12129?page=2) | 4.67 | 2 | 1 | FAIL, unresolved |

These are additional entry URLs for existing corpus articles, not four new articles or publishers. That historical run set contained eighteen actual CLI runs: twelve first-page corpus runs, two intentional incomplete controls and four failing later-page entries. The earlier twelve PASS results remain valid within their stated first-page coverage.

The additional runs used isolated configs and outputs under `/private/tmp/rlss-pagination-fix/nonfirst/<case>/`, following the same `--config ... --no-summary url <public-url>` command pattern. Their `cli.log`, `result.json` and `output/` artifacts were inspected, and per-page qualifying target-prose comparisons confirmed the saved and missing page ranges above. Source pages and raw article content remain temporary; the JSON records sanitized results and content hashes. No automatic navigation back to page one has been added. This finding records the pre-fix behavior before independent adversarial review; no production code, tests or acceptance scope changed in this evidence update.

## Current verification at 62c4765

After the later-page fix, all eighteen cases were rerun with a new binary and fresh per-case profiles, configs and outputs. Historical logs and failing outputs were preserved. Twelve first-page articles across the same ten site groups still passed: 280/280 qualifying target-body blocks and all 34 unique page anchors matched in source order. This specifically checks that detection of earlier pages did not reject the working first-page corpus.

| Case | Source pages | CLI exit | Seconds | Body blocks present / checked | Outcome |
| --- | ---: | ---: | ---: | ---: | --- |
| pcuser | 5 | 0 | 24.58 | 61 / 61 | PASS |
| president | 6 | 0 | 26.18 | 38 / 38 | PASS |
| gendai | 2 | 0 | 7.37 | 14 / 14 | PASS |
| bunshun | 2 | 0 | 11.64 | 14 / 14 | PASS |
| jbpress | 2 | 0 | 6.15 | 21 / 21 | PASS |
| toyokeizai | 3 | 0 | 13.18 | 22 / 22 | PASS |
| yahoo | 2 | 0 | 6.89 | 5 / 5 | PASS |
| itmedia-business | 3 | 0 | 10.77 | 32 / 32 | PASS |
| itmedia-mobile | 3 | 0 | 9.35 | 10 / 10 | PASS |
| ascii | 2 | 0 | 6.86 | 34 / 34 | PASS |
| mynavi | 2 | 0 | 6.71 | 27 / 27 | PASS |
| nikkan | 2 | 0 | 13.31 | 2 / 2 | PASS |

The same prior per-page observations and normalization rules were used. For ITmedia, all eleven target HTML pages were independently fetched again and compared using the same target-body boundary parser: PC USER 61/61, Business 32/32 and Mobile 10/10. Other sites retained the earlier independent per-page prose comparisons. Current output hashes, per-page counts, unique anchor hashes and offsets are in the JSON's `current_validation` section. Broad snapshot differences remain confined to non-target page furniture, including pagination navigation, ads, recommendations and comment prompts; the target-body check did not waive missing article prose.

| Current incomplete control | CLI exit | Seconds | Content / summaries saved | Outcome |
| --- | ---: | ---: | --- | --- |
| ITmedia cap1 | 1 | 7.50 | 0 / 0 | PASS |
| ITmedia missing selector with standard next hint | 1 | 6.65 | 0 / 0 | PASS |
| ITmedia page 3 | 1 | 6.67 | 0 / 0 | PASS; historical false success resolved |
| ITmedia page 5 | 1 | 6.35 | 0 / 0 | PASS; historical false success resolved |
| PRESIDENT page 3 | 1 | 11.73 | 0 / 0 | PASS; historical false success resolved |
| Yahoo page 2 | 1 | 9.56 | 0 / 0 | PASS; historical false success resolved |

All four later-page logs report Success 0 / Partial 0 / Failed 1. ITmedia rejects an initial page advertising an earlier page; PRESIDENT and Yahoo reject an initial page linking to a lower article page. Existing generic headed retry ran and retained the failure; no partial body or summary was saved. Rejection is the verified outcome: the tool does not automatically return to page one. A later-page entry without recognizable earlier-page evidence remains outside this demonstrated detection coverage.

Actual commands were `go build -o /private/tmp/rlss-pagination-fix/rlss-w3-final ./cmd/rlss` and `/private/tmp/rlss-pagination-fix/rlss-w3-final --config /private/tmp/rlss-pagination-fix/w3-final/<case>/config.yaml --no-summary url <public-url>`. The same 30s/60s timeouts, 2s wait, `headed_on_block: false`, per-domain selectors, disabled reading-list/RSS/copy sources and max-pages settings were retained. Temporary helpers `run-w3-final.py`, `check-w3-final.py`, `check-itmedia-final.py` and `prepare-evidence-final.py` produced the current observations and checks. Logs, `run.json` and raw outputs remain under the fresh temporary case directories. Output-directory inspection found no files for the six rejected controls.

There are eighteen current CLI runs and eighteen historical runs, thirty-six in total. They cover the same twelve articles and ten site groups, not additional publishers. This evidence-only update changes no production code or tests; the stated threshold, live-access and missing-hint limitations still apply.
