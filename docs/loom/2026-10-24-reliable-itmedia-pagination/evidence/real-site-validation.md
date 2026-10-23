# Real Japanese news pagination validation

Validated on 2026-10-24 (Asia/Taipei), using code commit `07bf97e59b93899fb50030bd98617b74291ef8c4`. All twelve real CLI runs completed successfully and passed the stated body and page-order checks. Two intentional incomplete-pagination runs failed with exit code 1 and saved no content.

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
