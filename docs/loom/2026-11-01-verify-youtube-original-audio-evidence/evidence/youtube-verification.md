# YouTube source verification — 2026-11-01

Fresh public-source checks used a separate logged-out browser profile in headless mode. The study imported `extractArticle` exactly once from this worktree's `extraction.ts` at `aa16398`; `db5df11` subsequently changed only the consumer. This report verifies extraction source evidence, not the complete save/summary pipeline or word-for-word audio accuracy.

The local study source `/private/tmp/rlss-youtube-original-first-study.ts`, bundled `.js`, driver `/private/tmp/rlss-youtube-study.go` and executable are research artifacts, outside Git. The completed browser command exited 0; all eight result files had `browserError: <nil>`:

```sh
RLSS_STUDY_METHODS=current /private/tmp/rlss-youtube-study \
  /private/tmp/rlss-youtube-original-first-study.js \
  /private/tmp/rlss-youtube-original-first-live \
  /private/tmp/rlss-youtube-original-first-profile headless \
  rfscVS0vtbw aircAruvnKk DHjqpvDnNGE dQw4w9WgXcQ \
  gBumdOWWMhY N0zhdMwD2Z8 jNQXAC9IVRw aqz-KE-bpKQ
```

## Source correspondence

Independent verification (`python3 /private/tmp/verify-youtube-live-evidence.py`) parsed all eight `*-current.json` files, checked each watch URL and every initial/refreshed player ID against the requested video, and parsed actual downloaded timedtext XML. For each successful result, the delivered body after the two metadata blocks exactly equaled the downloaded cues in order: XML-decoded text, collapsed whitespace, only cues containing a Unicode letter, separated by blank lines. Counts below distinguish all downloaded cues from delivered substantive cues; characters count the delivered body including separators, excluding metadata. Every matched caption request returned HTTP 200, had the recorded source language, had no `tlang`, and agreed with `platformTranslation: false`.

| Public target | Language / source | Language status / evidence | XML cues / delivered cues | Body characters | Exact source match |
|---|---|---|---:|---:|---|
| [Python course](https://www.youtube.com/watch?v=rfscVS0vtbw) | en / human | verified-original / original-audio-metadata | 2,935 / 2,935 | 262,335 | yes |
| [Neural network introduction](https://www.youtube.com/watch?v=aircAruvnKk) | en / human | verified-original / original-audio-metadata | 286 / 286 | 18,715 | yes |
| [Automatic-caption sample](https://www.youtube.com/watch?v=DHjqpvDnNGE) | en / automatic | verified-original / original-audio-asr-association | 161 / 81 | 3,149 | yes |
| [Human-caption sample](https://www.youtube.com/watch?v=dQw4w9WgXcQ) | en / human | verified-original / original-audio-asr-association | 61 / 60 | 2,142 | yes |
| [Japanese talk](https://www.youtube.com/watch?v=gBumdOWWMhY) | en / human | original-unconfirmed / validated-platform-caption | 463 / 463 | 19,121 | yes |
| [Chinese talk](https://www.youtube.com/watch?v=N0zhdMwD2Z8) | en / human | original-unconfirmed / validated-platform-caption | 545 / 545 | 19,021 | yes |
| [Me at the zoo](https://www.youtube.com/watch?v=jNQXAC9IVRw) | en / human | original-unconfirmed / validated-platform-caption | 6 / 6 | 222 | yes |
| [Big Buck Bunny](https://www.youtube.com/watch?v=aqz-KE-bpKQ) | none | no eligible tracks | 0 / 0 | 0 | unavailable |

Totals: eight targets; four verified-original results, three explicitly unconfirmed fallbacks, one no-subtitles failure. All seven delivered bodies matched source payloads: 4,457 downloaded cues, 4,376 delivered cues and 324,705 body characters. The no-subtitles target exposed no caption tracks in initial, IOS or ANDROID player data and returned `Unusable original-language subtitles: no eligible tracks`, without transcript content.

## Original-language evidence

- `rfscVS0vtbw`: streaming audio IDs were `ar.3`, `es.3`, `hi.3`, `pt.3`, `en.4`. The observed original label was `English original` on `en.4`. Exactly one caption-audio association had that exact ID, with indices 0–13; human English `.en` was index 3. Initial and IOS data supported this association; the successful downloaded human English payload came after an empty initial response.
- `aircAruvnKk`: streaming IDs were `ar.3`, `de.3`, `es.3`, `fr.3`, `hi.3`, `ru.3`, `th.3`, `uz.3`, `en-US.4`. The original label was `English (US) original` on `en-US.4`. Exactly one association had that ID and included all indices 0–30, including human English `.en` at 6 and automatic English `a.en` at 7. Initial and IOS data supported this association; human English succeeded after empty initial human and automatic requests.
- `DHjqpvDnNGE`: no explicit streaming audio ID or original label was supplied. One caption-audio association included index 0, the English automatic track. This supported the existing single-audio automatic-caption association evidence; delivered captions were automatic English.
- `dQw4w9WgXcQ`: no explicit streaming audio ID or original label was supplied. One association covered indices 0–5, including English human and automatic tracks. The unambiguous English automatic association established the language; the delivered track was human English.
- `gBumdOWWMhY`, `N0zhdMwD2Z8`, `jNQXAC9IVRw`: no explicit streaming original ID/label or automatic-language proof was supplied. Their single associations covered indices 0–7, 0–2 and 0–1 respectively. These are validated human-caption fallbacks, not verified originals. In particular, the Japanese and Chinese samples selected platform-default English; this is the permitted fallback policy, not evidence that English is their original spoken language or an independently proven translation.

These are first-party platform metadata assertions, not independent audio recognition. A default flag, display label alone, or audio ID suffix alone is insufficient proof.

## Limits and privacy

This fresh run found no naturally supplied known-translation endpoint, permanently blank human track with successful automatic recovery, or malformed live subtitle payload. Those boundaries rely on focused synthetic tests, not claimed live coverage. Empty initial timedtext responses followed by successful refreshed human captions demonstrate bounded retry recovery only. No successful `known-translation` result was observed. Source availability and undocumented metadata may change; older captures are replay evidence, not this run's fresh results.

Full captions, signed endpoints, account/cookie data and browser profile artifacts remain outside Git. Only public watch URLs, bounded source metadata and aggregate outcomes are recorded here. No user login state or existing notes were changed. This documentation task added no production code or test harness and required no artificial failing test.

## Configured-channel regression expansion

After the user approved fixing the expanded sample failures, a fresh isolated logged-out headless run exercised the same 18 public targets at `8a8ccf2`. The cohort is a bounded sample, not coverage of every configured channel. The private configuration and full preference list are not published. Local evidence: `/private/tmp/rlss-configured-channel-study/repaired`; source bundle `study-repaired.js` imported this worktree's exact `extraction.ts` and the existing research driver ran only its `current` method.

Before repair, 10 targets succeeded, five failed cue timing, and three exposed no accessible caption tracks. After repair, **15 succeeded and three still exposed no tracks** across inline, IOS and ANDROID player data. Nine successful results were verified-original; six remained original-unconfirmed. Every successful body exactly matched decoded downloaded substantive cues in order: **10,214 cues and 311,840 body characters**, including paragraph separators and excluding provenance labels. The independent local checker parsed XML and validated timing of every nonempty cue, skipping only normalized whitespace; all player IDs and delivered target markers matched the requested target. This verifies source correspondence, not audio transcription accuracy or guaranteed full spoken coverage.

| Public regression target | Delivered language / source | Language status | Delivered cues | Body characters |
|---|---|---|---:|---:|
| [Previously rejected Japanese captions](https://www.youtube.com/watch?v=Wz4Qy4BU7ks) | ja / automatic | verified-original | 366 | 13,662 |
| [Previously rejected Japanese captions](https://www.youtube.com/watch?v=gLcnqPTGT_w) | ja / automatic | verified-original | 96 | 1,971 |
| [Previously rejected English captions](https://www.youtube.com/watch?v=jOlPOxqFD5s) | en / automatic | verified-original | 902 | 32,890 |
| [Previously rejected short English captions](https://www.youtube.com/watch?v=o4D1J1YWjhg) | en / automatic | verified-original | 10 | 93 |
| [Previously rejected English captions](https://www.youtube.com/watch?v=bjE6HYuEA1k) | en / automatic | verified-original | 855 | 32,314 |
| [Traditional Chinese original metadata](https://www.youtube.com/watch?v=cQLKVzbwN7I) | zh-TW / human | verified-original | 2,115 | 28,823 |

The five timing failures contained whitespace-only srv3 paragraphs without duration; nonempty cue timing was valid. Normalized-empty paragraphs are now excluded before timing checks; malformed/truncated XML, blank-only responses, missing duration on nonempty text and decreasing nonempty cue timestamps remain rejected by regression tests. No duration is invented.

The Chinese target supplied `Chinese (Traditional) original` on audio ID `zh-Hant.4`, bound uniquely to a caption-audio association including human `zh-TW`. A default auto-dubbed English track was also present; the default flag did not establish originality. The exact affirmative label now supplies the `zh` family only when the existing ID-family and association checks agree. Negated/qualified Chinese labels and English-ID contradictions remain unconfirmed in focused tests. Other unrecognized labels still supply no proof.

The three remaining failures had no offered caption tracks in all bounded player responses; they were not counted as successful and no permanent absence of captions is inferred. Three successful samples still delivered unconfirmed `en-US` automatic captions despite their Japanese sampling group; grouping is not proof of spoken language. No naturally supplied translated endpoint was observed. Production configuration, running processes, browser login state and existing notes were unchanged.
