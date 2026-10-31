# Final embedded integration evidence

Date: 2026-10-31 (Asia/Taipei). Implementation input: `3f8baaa4e64f`.
This records implementer evidence, not an independent acceptance verdict.

## Build and focused checks

- `npm run build:defuddle`: passed; regenerated the committed browser bundle.
- `env GOCACHE=/private/tmp/rlss-go-build go build -o /private/tmp/rlss-complete-articles-20261031 ./cmd/rlss`: exit 0 and executable produced. Go printed a denied optional module stat-cache write outside the workspace; no dependency or production binary was changed.
- `node --test --test-name-pattern='dated|facebook' extraction.test.mjs`: 4 passed, 0 failed. Defuddle printed fixture URL diagnostics; assertions passed.
- Memory `regenerate-index` and `validate`: passed.

## Isolation and method

All actual CLI runs used `url URL --no-summary --config TEMP_CONFIG`, temporary
output under `/private/tmp/rlss-w4-20261031/output`, disabled Safari/Chrome list
sources, watch and copy-to, and no LLM. Scoped public samples used a fresh logged-out
CLI profile. Independent source reloads used separate fresh headless and headed
profiles and public DOM/native fields only. Browser commands required sandbox
escalation after isolated Chrome startup failed inside the sandbox.

X regression runs used the previously authorized isolated session clone,
sequentially: final CLI first, independent source browser afterward. No live user
profile, production config, production binary or existing notes were modified.
Full source captures, signed media URLs and browser state remain uncommitted.

## Facebook

Both the original slugged reading-list URL and
`https://www.facebook.com/fatality1117/posts/10234719952645505/` saved successfully.
Independent headless and headed source reloads matched both saved bodies:

- Native authored text: 4,400 Unicode characters and 148 lines, equal after
  reversing six Markdown link labels to original URL text and whitespace layout.
- Six hashtags and all symbols retained: 16 `◾`, two `▪` (18 small bullet glyphs),
  plus eight `■`. These categories explain the earlier 18-glyph shorthand.
- One native target photo, identity `10234719952085491`, corresponds to the owned
  rendered photo and saved image path. Signed CDN parameters vary per reload;
  stable photo identity and asset path agree.
- Equality of the authored body excludes comments, reactions and login controls;
  source-owned interface images were excluded from the one-photo count.

## Felix

`https://felixrieseberg.com/how-to-be-more-ambitious-a-tutorial-for-humans-with-models/`
saved successfully. Fresh headless and headed source reloads retained all 42
paragraph/heading/list/caption selector units in order, plus the independently
captured blockquote (43 visible units total), four headings, eight authored
links and one original image URL. The source helper's 42-unit selector omits
blockquote, so it must not be reported as a 43-unit selector result.

Comparison removes Markdown emphasis/link syntax and layout whitespace, preserving
prose, punctuation, order and destinations. Heading comparison uses visible text;
CSS-hidden `PART` labels are excluded. The headed source helper also observed four
empty section-anchor controls; these are interface links, not missing authored
links. Generated image-alt captions and emphasis formatting are layout differences.

## X regression sources

All 11 original X URLs were re-exercised with the final CLI and independently
reloaded afterward; every CLI run saved successfully and every source was
available. Comparison results are in local `x-comparison.json`.

Ten saved bodies equal the native authored source after Markdown and whitespace
normalization. Target `2114704821852991771` has one explicit URL serialization
difference: `http://amitaro.net` becomes `http://amitaro.net/`; all other authored
text and all 13 line breaks match. The destination is equivalent, but this is
not counted as literal URL equality. A focused Node WHATWG `new URL` check found
equal serialized `href` and equal scheme, host, port, path, query and fragment for
only this pair; no broader URL normalization was applied. The other two plain posts retain all 20
and 72 authored line breaks respectively.

All eight rich articles retain matching owned text and 52 owned image/thumbnail
references. Target `2113959093220335871` excludes three independently identified
other-post cards and their six images; these are not counted as owned content.
Saved rich heading counts include the article title, explaining the additional
one heading compared with source body headings. No cached source report was used
as a substitute for a fresh comparison, and no unavailable source was marked pass.

## Growth and limits

Against confirmed-intent commit `fc19803`, handwritten production changes add 287
and remove 8 lines (net +279); tests add 209 and remove 0 (net +209); regenerated
bundle adds 167 and removes 161 (net +6 minified physical lines). The production
estimate of 100–200 lines was exceeded, including request binding and failure
classification across browser and pipeline boundaries. Physical generated lines
are not a useful complexity measure and are reported separately.

These checks cover the studied public Facebook post and Felix article, not every
site, Facebook URL shape, attachment type or full X thread. Cached earlier proofs
were not substituted for final CLI verification. Independent closing review and
full package verification remain the parent station's responsibility.

Local audit artifacts: `/private/tmp/rlss-w4-20261031/` contains command logs,
isolated configs/output, public-source captures and comparison scripts/results.
