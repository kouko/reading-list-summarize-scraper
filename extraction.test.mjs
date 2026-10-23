import test from 'node:test';
import assert from 'node:assert/strict';
import { build } from 'esbuild';
import { parseHTML } from 'linkedom';
import { writeFile } from 'node:fs/promises';
const bundle = await build({ entryPoints: ['extraction.ts'], bundle: true, platform: 'node', format: 'esm', write: false });
await writeFile('/private/tmp/rlss-extraction-test.mjs', bundle.outputFiles[0].contents);
const { extractArticle } = await import('file:///private/tmp/rlss-extraction-test.mjs');
const doc = (body) => parseHTML(`<html><head></head><body>${body}</body></html>`).document;
const yt = (original = true, tracks) => doc(`<script>var ytInitialPlayerResponse = ${JSON.stringify({videoDetails:{videoId:'abc'}, captions:{playerCaptionsTracklistRenderer:{captionTracks:tracks ?? [{baseUrl:'https://www.youtube.com/api/timedtext?v=abc',languageCode:'ja',kind:'asr'}],audioTracks:[{audioTrackId:'ja.4',displayName:original?'Japanese (original)':'Japanese',captionTrackIndices:[0]}]}}})};</script>`);
const cues = '<transcript><text start="0" dur="2">これは実際の字幕です。</text></transcript>';
test('A1 verified original audio association', async () => {
 const out = await extractArticle(yt(), 'https://www.youtube.com/watch?v=abc', async () => new Response(cues));
 assert.match(out, /rlss-video-v1/); assert.match(out, /これは実際の字幕/); assert.match(out, /automatic/);
});
test('A1 unverified default language rejected', async () => {
 await assert.rejects(extractArticle(yt(false), 'https://www.youtube.com/watch?v=abc'), /original language/i);
});
test('A2 target bodies on three platforms exclude replies', async () => {
 for (const [url, html] of [
 ['https://x.com/a/status/123','<article data-testid="tweet"><a href="/a/status/123"><time>today</time></a><div data-testid="tweetText">Target main post text.</div></article><article data-testid="tweet"><div data-testid="tweetText">reply</div></article>'],
 ['https://www.threads.net/@a/post/ABC','<script type="application/ld+json">{"@type":"SocialMediaPosting","url":"https://www.threads.net/@a/post/ABC","articleBody":"Target main post text."}</script>'],
 ['https://medium.com/@a/title-abcdef123456','<link rel="canonical" href="https://medium.com/@a/title-abcdef123456"><article><p>Target main post text.</p></article>']]) {
  const out = await extractArticle(doc(html), url); assert.match(out, /Target main post text/); assert.doesNotMatch(out, /reply/);
 }
});
test('A2 wrong target rejected', async () => {
 await assert.rejects(extractArticle(doc('<article data-testid="tweet"><a href="/a/status/999"><time>today</time></a><div data-testid="tweetText">Wrong post</div></article>'), 'https://x.com/a/status/123'), /target/i);
});
test('A3 verified social content succeeds without network fallback', async () => {
 assert.equal(await extractArticle(doc('<article data-testid="tweet"><a href="/a/status/123"><time>now</time></a><div data-testid="tweetText">Actual target</div></article>'), 'https://x.com/a/status/123', () => {throw Error('external request');}), 'Actual target');
 const article = await extractArticle(doc(`<article><h1>Ordinary article</h1><p>${'This ordinary article explains a concrete subject with substantial detail. '.repeat(20)}</p></article>`), 'https://example.com/article');
 assert.match(article, /ordinary article explains/);
});
test('A3 login, preview and unsupported video fail', async () => {
 for (const [url, html] of [['https://medium.com/@a/title-abcdef123456','<article>Member-only story<p>Preview</p></article>'],['https://x.com/a/status/123','<main>Log in to X</main>'],['https://vimeo.com/123','<article>Video description</article>'],['https://www.bilibili.com/video/BV123','<article>description</article>']]) await assert.rejects(extractArticle(doc(html),url));
});
test('A4 blank human track does not block usable automatic track', async () => {
 const d = yt(true,[{baseUrl:'https://www.youtube.com/api/timedtext?v=abc&manual=1',languageCode:'ja'},{baseUrl:'https://www.youtube.com/api/timedtext?v=abc',languageCode:'ja',kind:'asr'}]);
 d.querySelector('script').textContent = d.querySelector('script').textContent.replace('"captionTrackIndices":[0]','"captionTrackIndices":[0,1]');
 const out = await extractArticle(d,'https://www.youtube.com/watch?v=abc',async u => new Response(String(u).includes('manual')?'<transcript></transcript>':cues)); assert.match(out,/automatic/);
});
test('A4 malformed, truncated and timestamp-only cues fail', async () => {
 for (const xml of ['<transcript><text start="-1" dur="2">speech</text></transcript>','<transcript><text start="0" dur="2">speech</text>','<transcript><text start="0" dur="2">00:01</text></transcript>']) await assert.rejects(extractArticle(yt(),'https://www.youtube.com/watch?v=abc',async()=>new Response(xml)), /unusable/i);
});
