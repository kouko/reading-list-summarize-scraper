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
test('A1 verified original audio association', async () => {
 const tracks = ['en','de-DE','ja','pt-BR','es-419'].map(languageCode => ({baseUrl:`https://www.youtube.com/api/timedtext?v=abc&lang=${languageCode}`,languageCode}));
 tracks.push({baseUrl:'https://www.youtube.com/api/timedtext?v=abc&lang=en&kind=asr',languageCode:'en',kind:'asr'});
 const out = await extractArticle(yt(true, tracks), 'https://www.youtube.com/watch?v=abc', async () => new Response(cues));
 assert.match(out, /rlss-video-v1/); assert.match(out, /Actual spoken subtitle text/); assert.match(out, /Subtitle language: en/);
});
test('A1 unverified default language rejected', async () => {
 for (const d of [yt(false), yt(true, undefined, {audioTracks:[{captionTrackIndices:[0]},{captionTrackIndices:[0]}]}), yt(true,[{baseUrl:'https://www.youtube.com/api/timedtext?v=abc&tlang=ja',languageCode:'ja',kind:'asr'},{baseUrl:'https://www.youtube.com/api/timedtext?v=abc&lang=ja',languageCode:'ja'}])]) await assert.rejects(extractArticle(d, 'https://www.youtube.com/watch?v=abc', async () => new Response(cues)), /original language|unusable/i);
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
test('A2 wrong target rejected', async () => {
 await assert.rejects(extractArticle(doc('<article data-testid="tweet"><a href="/a/status/999"><time>today</time></a><div data-testid="tweetText">Wrong post</div></article>'), 'https://x.com/a/status/123'), /target/i);
 await assert.rejects(extractArticle(doc('<div data-pagelet="threads_post_page_0"><div data-pressable-container><a href="/@b/post/REPLY"><time>today</time></a><span dir="auto">Other reply</span><div data-pressable-container><a href="/@a/post/ABC"><time>yesterday</time></a><span dir="auto">Quoted target</span></div></div></div>'), 'https://www.threads.com/@a/post/ABC'), /target/i);
});
test('A3 verified social content succeeds without network fallback', async () => {
 assert.equal(await extractArticle(doc('<article data-testid="tweet"><a href="/a/status/123"><time>now</time></a><div data-testid="tweetText">Actual target</div></article>'), 'https://x.com/a/status/123', () => {throw Error('external request');}), 'Actual target');
 const article = await extractArticle(doc(`<article><h1>Ordinary article</h1><p>${'This ordinary article explains a concrete subject with substantial detail. '.repeat(20)}</p></article>`), 'https://example.com/article');
 assert.match(article, /ordinary article explains/);
 const paid = await extractArticle(doc('<link rel="canonical" href="https://medium.com/@a/title-abcdef123456"><script type="application/ld+json">{"@type":"Article","isAccessibleForFree":false}</script><article><h1>Member-only story</h1><p>Full legitimately accessible article text.</p></article>'), 'https://medium.com/@a/title-abcdef123456');
 assert.match(paid, /Full legitimately accessible article text/);
});
test('A3 login, preview and unsupported video fail', async () => {
 for (const [url, html] of [['https://medium.com/@a/title-abcdef123456','<link rel="canonical" href="https://medium.com/@a/title-abcdef123456"><article><p>Preview</p></article><div data-testid="paywall">Sign up to read the rest.</div>'],['https://x.com/a/status/123','<main>Log in to X</main>'],['https://vimeo.com/123','<article>Video description</article>'],['https://www.bilibili.com/video/BV123','<article>description</article>']]) await assert.rejects(extractArticle(doc(html),url), /preview|target|video|original language/i);
});
test('A4 blank human track does not block usable automatic track', async () => {
 const d = yt(true,[{baseUrl:'https://www.youtube.com/api/timedtext?v=abc&manual=1',languageCode:'en'},{baseUrl:'https://www.youtube.com/api/timedtext?v=abc',languageCode:'en',kind:'asr'}]);
 const out = await extractArticle(d,'https://www.youtube.com/watch?v=abc',async u => new Response(String(u).includes('manual')?'<transcript></transcript>':cues)); assert.match(out,/automatic/);
});
test('A4 malformed, truncated and timestamp-only cues fail', async () => {
 for (const xml of ['<transcript><text start="-1" dur="2">speech</text></transcript>','<transcript><text start="0" dur="2">speech</text>','<transcript><text start="0" dur="2">00:01</text></transcript>','<!DOCTYPE transcript SYSTEM "https://example.com/external.dtd"><transcript><text start="0" dur="2">speech</text></transcript>','<transcript><text start="0" dur="2">speech &unknown;</text></transcript>']) await assert.rejects(extractArticle(yt(),'https://www.youtube.com/watch?v=abc',async()=>new Response(xml)), /unusable/i);
});
