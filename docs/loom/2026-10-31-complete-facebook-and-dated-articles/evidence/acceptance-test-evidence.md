# Final embedded integration evidence

Date: 2026-10-31 (Asia/Taipei). Initial implementation input: `3f8baaa4e64f`;
Facebook identity follow-up input: `8c37cac`.
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
  restoring the six hashtag Markdown link labels and normalizing whitespace.
- Six hashtags and all symbols retained: 16 `◾`, two `▪` (18 small bullet glyphs),
  plus eight `■`. These categories explain the earlier 18-glyph shorthand.
- One native target photo, identity `10234719952085491`, corresponds to the owned
  rendered photo and saved image path. Signed CDN parameters vary per reload;
  stable photo identity and asset path agree.
- Equality of the authored body excludes comments, reactions and login controls;
  source-owned interface images were excluded from the one-photo count.

## Felix

The Felix and X evidence below was collected against the initial implementation;
those extraction paths were unchanged by the Facebook identity repair. They were
not reloaded during this bounded follow-up; independent acceptance will reload
them against the final tree.

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

Against confirmed-intent commit `fc19803`, tree `8c37cac` handwritten production changes
add 300 and remove 8 lines (net +292); tests add 232 and remove 1 (net +231);
regenerated bundle adds 149 and removes 143 (net +6 minified physical lines).
These counts exclude documentation and classify root `.test.mjs`, `_test.go`, and
browser extraction tests as tests. The production
estimate of 100–200 lines was exceeded, including request binding and failure
classification across browser and pipeline boundaries. Physical generated lines
are not a useful complexity measure and are reported separately.

`git diff --check` reports upstream template-literal trailing whitespace in the
generated bundle at lines 241 and 336. These are generated string contents and
were retained without manual bundle edits.

These checks cover the studied public Facebook post and Felix article, not every
site, Facebook URL shape, attachment type or full X thread. Cached earlier proofs
were not substituted for final CLI verification. Independent closing review and
full package verification remain the parent station's responsibility.

Local audit artifacts: `/private/tmp/rlss-w4-20261031/` contains command logs,
isolated configs/output, public-source captures and comparison scripts/results.

## Facebook identity follow-up

Repair `4b52a3f` rejects a matching native post ID when its explicit permalink
contradicts the requested post or author. Its regenerated bundle was included in
`8c37cac`, from which the final isolated CLI was rebuilt with
`env GOCACHE=/private/tmp/rlss-go-build go build -o /private/tmp/rlss-w4-final-20261031 ./cmd/rlss`
(exit 0; the same optional module stat-cache warning remained).

Both numeric and original slugged URLs were rerun with `--no-summary`, a fresh
temporary profile/config and temporary output. Each saved successfully. Fresh
independent logged-out headless and headed source reloads again matched both
saved authored bodies after restoring six hashtag labels and whitespace layout:
4,400 characters, 148 lines, six hashtags, 16 `◾`, two `▪`, eight `■`, one native
owned photo `10234719952085491` and the same stable saved image asset path.
No prior source capture was substituted for these reloads.

The earned identity probe went RED at `83df5fa`, then GREEN after `4b52a3f`, and
was graduated to root `facebook-identity.test.mjs` in `8c37cac`. Its original and
maintained file blobs are identical (`94602b5707815c79969c9d010fbf2704498ab653`).
`node --test facebook-identity.test.mjs` passed the parent and both contradictory
permalink subtests (3 tests total) during this follow-up.

Documentation defect class: ambiguous comparison wording and claims exceeding
observed behavior. Evidence and memory were corrected after searching those
artifacts plus the retained plan, spec and README; no other same-class wording
was found. The memory now captures expected units before cleanup and verifies
the resulting extraction before saving; it records the observed Felix empty
result without describing it as a nonempty result. No spec/hash was changed.

Follow-up audit artifacts: `/private/tmp/rlss-w4-final-20261031-evidence/` contains
fresh source captures, both CLI logs/output, `comparison.json` and
`media-check.json`. No full suite was run here; the parent station owns it.
