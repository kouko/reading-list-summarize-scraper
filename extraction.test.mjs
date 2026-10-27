import test, { after } from 'node:test';
import assert from 'node:assert/strict';
import { build } from 'esbuild';
import { parseHTML } from 'linkedom';
import { mkdtemp, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { pathToFileURL } from 'node:url';
const bundle = await build({ entryPoints: ['extraction.ts'], bundle: true, platform: 'node', format: 'esm', write: false });
const temporary = await mkdtemp(join(tmpdir(), 'rlss-extraction-'));
after(() => rm(temporary, {recursive: true, force: true}));
const modulePath = join(temporary, 'extraction.mjs');
await writeFile(modulePath, bundle.outputFiles[0].contents);
const browser = parseHTML('<html><head></head><body></body></html>');
globalThis.window = browser.window;
globalThis.document = browser.document;
globalThis.DOMParser = browser.window.DOMParser;
const { extractArticle } = await import(pathToFileURL(modulePath).href);
const doc = (body) => parseHTML(`<html><head></head><body>${body}</body></html>`).document;
// Sanitized shape from the public dQw4w9WgXcQ player response: no invented
// caption-audio label or ID; discard all transient signed URL parameters.
const yt = (verified = true, tracks = [{baseUrl:'https://www.youtube.com/api/timedtext?v=abc',languageCode:'en',...(verified ? {kind:'asr'} : {})}], extra = {}) => doc(`<script>var ytInitialPlayerResponse = ${JSON.stringify({videoDetails:{videoId:'abc'}, captions:{playerCaptionsTracklistRenderer:{captionTracks:tracks,audioTracks:[{captionTrackIndices:tracks.map((_, i) => i),defaultCaptionTrackIndex:0,hasDefaultTrack:true,captionsInitialState:'CAPTIONS_INITIAL_STATE_OFF_RECOMMENDED'}],defaultAudioTrackIndex:0,...extra}}})};</script>`);
const cues = '<transcript><text start="0" dur="2">Actual spoken subtitle text.</text></transcript>';
test('A1/A3 native player refresh recovers empty inline captions with bounded first-party requests', async () => {
 const fresh = JSON.parse(yt().querySelector('script').textContent.split(' = ')[1].slice(0,-1));
 fresh.playabilityStatus = {status:'OK'};
 fresh.captions.playerCaptionsTracklistRenderer.captionTracks[0].baseUrl += '&fresh=1';
 const calls = [];
 const out = await extractArticle(yt(), 'https://www.youtube.com/watch?v=abc', async (u, options) => {
  calls.push(String(u));
  if (String(u).includes('/youtubei/')) {
   assert.equal(String(u), 'https://www.youtube.com/youtubei/v1/player?prettyPrint=false');
   assert.equal(options.method, 'POST'); assert.equal(options.credentials, 'include'); assert.equal(options.redirect, 'error'); assert.ok(options.signal);
   assert.deepEqual(JSON.parse(options.body), {context:{client:{clientName:'IOS',clientVersion:'20.10.3'}},videoId:'abc'});
   return Response.json(fresh);
  }
  return new Response(String(u).includes('fresh=1') ? cues : '');
 });
 assert.match(out, /Actual spoken subtitle text/); assert.equal(calls.length, 3);
});
test('A4 captions work when Trusted Types rejects HTML parsing sinks', async () => {
 const d = yt();
 d.createElement = () => { throw Error('TrustedHTML required'); };
 const out = await extractArticle(d, 'https://www.youtube.com/watch?v=abc', async () => new Response('<timedtext><body><p t="320" d="14260"><s>Actual &amp; spoken</s> subtitle text.</p></body></timedtext>'));
 assert.match(out, /Actual & spoken subtitle text/);
});
test('A1 verified original audio association', async () => {
 const tracks = ['en','de-DE','ja','pt-BR','es-419'].map(languageCode => ({baseUrl:`https://www.youtube.com/api/timedtext?v=abc&lang=${languageCode}`,languageCode}));
 tracks.push({baseUrl:'https://www.youtube.com/api/timedtext?v=abc&lang=en&kind=asr',languageCode:'en',kind:'asr'});
 const out = await extractArticle(yt(true, tracks), 'https://www.youtube.com/watch?v=abc', async () => new Response(cues));
 assert.match(out, /rlss-video-v1/); assert.match(out, /Actual spoken subtitle text/); assert.match(out, /Subtitle language: en/);
});
test('A1 unverified default language rejected', async () => {
 for (const d of [yt(false), yt(true, undefined, {audioTracks:[{captionTrackIndices:[0]},{captionTrackIndices:[0]}]}), yt(true,[{baseUrl:'https://www.youtube.com/api/timedtext?v=abc&tlang=ja',languageCode:'ja',kind:'asr'},{baseUrl:'https://www.youtube.com/api/timedtext?v=abc&lang=ja',languageCode:'ja'}])]) await assert.rejects(extractArticle(d, 'https://www.youtube.com/watch?v=abc', async () => new Response(cues)), /original language|unusable/i);
});
test('A1/A3 refreshed invalid identity, language and API failures stay unsuccessful', async () => {
 const valid = JSON.parse(yt().querySelector('script').textContent.split(' = ')[1].slice(0,-1));
 valid.playabilityStatus = {status:'OK'};
 for (const change of [
  d => { d.videoDetails.videoId = 'other'; },
  d => { d.captions.playerCaptionsTracklistRenderer.audioTracks.push({captionTrackIndices:[0]}); },
  d => { d.captions.playerCaptionsTracklistRenderer.captionTracks[0].baseUrl += '&tlang=ja'; },
  d => { d.captions.playerCaptionsTracklistRenderer.captionTracks[0].languageCode = 'ja'; },
 ]) {
  const fresh = structuredClone(valid); change(fresh); let apiCalls = 0;
  await assert.rejects(extractArticle(yt(), 'https://www.youtube.com/watch?v=abc', async u => {
   if (String(u).includes('/youtubei/')) { apiCalls++; return Response.json(fresh); }
   return new Response('');
  }), /unusable original-language/i);
  assert.equal(apiCalls, 2);
 }
 for (const response of [() => new Response('', {status:503}), () => new Response(''), () => new Response('{'), () => Response.json({})]) {
  let calls = 0;
  await assert.rejects(extractArticle(yt(), 'https://www.youtube.com/watch?v=abc', async u => {
   if (String(u).includes('/youtubei/')) { calls++; return response(); }
   return new Response('');
  }), /unusable original-language/i); assert.equal(calls, 2);
 }
});
test('A2 target bodies on three platforms exclude replies', async () => {
 for (const [url, html] of [
 ['https://x.com/a/status/123','<article data-testid="tweet"><a href="/a/status/123"><time>today</time></a><div data-testid="tweetText">Target main post text.<br>Next line.<div>Separate block.</div></div></article><article data-testid="tweet"><div data-testid="tweetText">reply</div></article>'],
 ['https://x.com/a/article/123','<link rel="canonical" href="https://x.com/a/article/123"><div data-testid="twitterArticleReadView"><div data-testid="twitterArticleRichTextView"><div class="longform-unstyled">Target main post text.<br>Next line.</div><div class="longform-unstyled">Separate block.</div></div></div><article data-testid="tweet"><div data-testid="tweetText">reply</div></article>'],
 ['https://www.threads.net/@a/post/ABC','<script type="application/ld+json">{"@type":"SocialMediaPosting","url":"https://www.threads.net/@a/post/ABC","articleBody":"Target main post text.\\nNext line.\\n\\nSeparate block."}</script>'],
 ['https://www.threads.com/@a/post/ABC','<div data-pagelet="threads_post_page_0"><div data-pressable-container><a href="/@a/post/ABC"><time>today</time></a><span dir="auto">Target main post text.<br>Next line.<div>Separate block.</div></span><div data-pressable-container><a href="/@b/post/QUOTE"><time>then</time></a><span dir="auto">quoted reply</span></div></div><div data-pressable-container><a href="/@b/post/REPLY"><time>later</time></a><span dir="auto">reply</span></div></div>'],
 ['https://www.threads.com/@a/post/ABC','<div role="region"><a href="/@a/post/ABC"><time>today</time></a><span dir="auto">Target main post text.<br>Next line.<div>Separate block.</div></span><a href="/@b/post/QUOTE"><span dir="auto">quoted reply</span></a></div>'],
 ['https://medium.com/@a/title-abcdef123456','<link rel="canonical" href="https://medium.com/@a/title-abcdef123456"><article><p>Target main post text.<br>Next line.</p><div>Separate block.</div></article>']]) {
  const out = await extractArticle(doc(html), url); assert.match(out, /Target main post text\.\s+Next line\.\s+Separate block\./); assert.doesNotMatch(out, /reply/);
 }
});
test('A2/A3 explicit target previews fail while complete own bodies remain usable', async () => {
 const marker = '<a data-testid="tweet-text-show-more-link">Show more</a>';
 const own = '<a href="/a/status/123"><time>today</time></a><div data-testid="tweetText">Substantive target preview.</div>';
 const canonical = '<link rel="canonical" href="https://x.com/a/status/123">';
 for (const html of [
  `<article data-testid="tweet">${own}${marker}</article>`,
  `${canonical}<div data-testid="twitterArticleReadView"><div data-testid="twitterArticleRichTextView">Substantive article preview.${marker}</div></div>`,
 ]) await assert.rejects(extractArticle(doc(html), 'https://x.com/a/status/123'), /preview|truncat/i);
 const complete = await extractArticle(doc(`<article data-testid="tweet">${own}<div data-testid="quoteTweet">Quoted preview.${marker}</div></article>`), 'https://x.com/a/status/123');
 assert.equal(complete, 'Substantive target preview.');
 const fullArticle = await extractArticle(doc(`${canonical}<div data-testid="twitterArticleReadView"><div data-testid="twitterArticleRichTextView">Complete target article.</div></div><article data-testid="tweet">${own}${marker}</article>`), 'https://x.com/a/status/123');
 assert.equal(fullArticle, 'Complete target article.');
});
test('A2 wrong target rejected', async () => {
 await assert.rejects(extractArticle(doc('<article data-testid="tweet"><a href="/a/status/999"><time>today</time></a><div data-testid="tweetText">Wrong post</div></article>'), 'https://x.com/a/status/123'), /target/i);
 await assert.rejects(extractArticle(doc('<div data-pagelet="threads_post_page_0"><div data-pressable-container><a href="/@b/post/REPLY"><time>today</time></a><span dir="auto">Other reply</span><div data-pressable-container><a href="/@a/post/ABC"><time>yesterday</time></a><span dir="auto">Quoted target</span></div></div></div>'), 'https://www.threads.com/@a/post/ABC'), /target/i);
});
test('A1 x-public-target-body preserves current public post and article bodies', async () => {
 const canonical = '<link rel="canonical" href="https://x.com/Author/status/123">';
 const identity = '<a href="/Author" aria-current="page">@Author</a><a href="/Author/status/123" aria-current="page">today</a>';
 for (const body of ['<div class="text-lead whitespace-pre-wrap" dir="auto">Complete target.\n\nLast paragraph.</div>', '<h1 data-article-part="title">Target title</h1><div data-article-part="body"><div class="x-article-body"><h2>First section</h2><p>Complete target.</p><div data-slot="embedded-tweet-frame"><article>Quoted unrelated text.</article></div><p>Last paragraph.</p></div></div>']) {
  const out = await extractArticle(doc(`${canonical}<article>${identity}${body}</article><article><div dir="auto" class="text-lead">Unrelated reply.</div></article>`), 'https://x.com/author/status/123');
  assert.match(out, /Complete target\./); assert.match(out, /Last paragraph\./); assert.doesNotMatch(out, /Unrelated|Quoted|@Author|<div/);
  if (body.includes('x-article-body')) { assert.match(out, /# Target title/); assert.match(out, /## First section/); }
 }
});
test('A3 x-preview-or-reply rejects unverified current bodies', async () => {
 const canonical = '<link rel="canonical" href="https://x.com/a/status/123">';
 const identity = '<a href="/a" aria-current="page">@a</a><a href="/a/status/123" aria-current="page">today</a>';
 const body = '<div class="text-lead whitespace-pre-wrap" dir="auto">Complete target.</div>';
 const own = `<article>${identity}${body}</article>`;
 for (const html of [own, `${canonical}${own}${own}`, `${canonical}<article>${identity}${body}<button>Show more</button></article>`, `${canonical}<article><a href="/b/status/999" aria-current="page">today</a>${body}</article>`, `${canonical}<article><div data-slot="embedded-tweet-frame">${own}</div></article>`, `${canonical}<article>${identity}<a href="/i/article/123">Article preview</a>${body}</article>`, `${canonical}<article>${identity}<div class="text-lead" dir="auto">Log in to continue</div></article>`, `${canonical}<article>${identity}<h1 data-article-part="title">Real title</h1><div data-article-part="body"><div class="x-article-body">Log in to continue</div></div></article>`]) {
  await assert.rejects(extractArticle(doc(html), 'https://x.com/a/status/123'), /target|preview|ambiguous/i);
 }
});
test('A3 verified social content succeeds without network fallback', async () => {
 assert.equal(await extractArticle(doc('<article data-testid="tweet"><a href="/a/status/123"><time>now</time></a><div data-testid="tweetText">Actual target</div></article>'), 'https://x.com/a/status/123', () => {throw Error('external request');}), 'Actual target');
 const article = await extractArticle(doc(`<article><h1>Ordinary article</h1><p>${'This ordinary article explains a concrete subject with substantial detail. '.repeat(20)}</p></article>`), 'https://example.com/article');
 assert.match(article, /ordinary article explains/);
 const paid = await extractArticle(doc('<link rel="canonical" href="https://medium.com/@a/title-abcdef123456"><script type="application/ld+json">{"@type":"Article","isAccessibleForFree":false}</script><article><h1>Member-only story</h1><p>Full legitimately accessible article text.</p></article>'), 'https://medium.com/@a/title-abcdef123456');
 assert.match(paid, /Full legitimately accessible article text/);
});
test('A3 challenge identity survives actual adapter failures without classifying article mentions', async () => {
 const page = (title, body) => parseHTML(`<html><head><title>${title}</title></head><body>${body}</body></html>`).document;
 for (const url of ['https://x.com/a/status/123', 'https://medium.com/@a/title-abcdef123456', 'https://www.youtube.com/watch?v=abc']) {
  await assert.rejects(extractArticle(page('Attention Required! | Cloudflare', '<main>Private challenge details</main>'), url), error => /Cloudflare security challenge/.test(error.message) && !/Private challenge details/.test(error.message));
  await assert.rejects(extractArticle(page('Verify you are human', '<form id="captcha"><iframe src="https://hcaptcha.com/challenge"></iframe></form>'), url), /CAPTCHA security challenge/);
  await assert.rejects(extractArticle(page('Article about Cloudflare CAPTCHA', '<article>Cloudflare and CAPTCHA explain internet access.</article>'), url), error => !/security challenge/.test(error.message));
 }
 const out = await extractArticle(doc('<article data-testid="tweet"><a href="/a/status/123"><time>today</time></a><div data-testid="tweetText">Cloudflare CAPTCHA security challenge is this article topic.</div></article>'), 'https://x.com/a/status/123');
 assert.equal(out, 'Cloudflare CAPTCHA security challenge is this article topic.');
});
test('A3 login, preview and unsupported video fail', async () => {
 for (const [url, html] of [['https://medium.com/@a/title-abcdef123456','<link rel="canonical" href="https://medium.com/@a/title-abcdef123456"><article><p>Preview</p></article><div data-testid="paywall">Sign up to read the rest.</div>'],['https://x.com/a/status/123','<main>Log in to X</main>'],['https://vimeo.com/123','<article>Video description</article>'],['https://www.bilibili.com/video/BV123','<article>description</article>']]) await assert.rejects(extractArticle(doc(html),url), /preview|target|video|original language/i);
});
test('A4 blank human track does not block usable automatic track', async () => {
 const d = yt(true,[{baseUrl:'https://www.youtube.com/api/timedtext?v=abc&manual=1',languageCode:'en'},{baseUrl:'https://www.youtube.com/api/timedtext?v=abc',languageCode:'en',kind:'asr'}]);
 const out = await extractArticle(d,'https://www.youtube.com/watch?v=abc',async u => new Response(String(u).includes('manual')?'<transcript></transcript>':cues)); assert.match(out,/automatic/);
 const fresh = JSON.parse(d.querySelector('script').textContent.split(' = ')[1].slice(0,-1)); fresh.playabilityStatus = {status:'OK'};
 for (const track of fresh.captions.playerCaptionsTracklistRenderer.captionTracks) track.baseUrl += '&fresh=1';
 const requests = [];
 const recovered = await extractArticle(d, 'https://www.youtube.com/watch?v=abc', async (u, options) => {
  if (String(u).includes('/youtubei/')) {
   requests.push(JSON.parse(options.body).context.client.clientName);
   return requests.length === 1 ? new Response('', {status:500}) : Response.json(fresh);
  }
  return new Response(String(u).includes('fresh=1') && !String(u).includes('manual') ? cues : '');
 });
 assert.match(recovered, /Subtitle source: automatic/); assert.deepEqual(requests, ['IOS','ANDROID']);
});
test('A4 malformed, truncated and timestamp-only cues fail', async () => {
 for (const xml of ['','<transcript><text start="-1" dur="2">speech</text></transcript>','<transcript><text start="0" dur="2">speech</text>','<transcript><text start="0" dur="2">00:01</text></transcript>','<!DOCTYPE transcript SYSTEM "https://example.com/external.dtd"><transcript><text start="0" dur="2">speech</text></transcript>','<transcript><text start="0" dur="2">speech &unknown;</text></transcript>']) {
  const fresh = JSON.parse(yt().querySelector('script').textContent.split(' = ')[1].slice(0,-1)); fresh.playabilityStatus = {status:'OK'};
  await assert.rejects(extractArticle(yt(),'https://www.youtube.com/watch?v=abc',async u => String(u).includes('/youtubei/') ? Response.json(fresh) : new Response(xml)), /unusable/i);
 }
});
