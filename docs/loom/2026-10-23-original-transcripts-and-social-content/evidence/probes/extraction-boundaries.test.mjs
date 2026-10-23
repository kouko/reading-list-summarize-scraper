// concern: Nested quoted-post identity and repaired malformed XML must not become successful target content.
import test, { after } from 'node:test';
import assert from 'node:assert/strict';
import { build } from 'esbuild';
import { parseHTML } from 'linkedom';
import { mkdtemp, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { pathToFileURL } from 'node:url';
const bundle = await build({ entryPoints: ['extraction.ts'], bundle: true, platform: 'node', format: 'esm', write: false });
const temporary = await mkdtemp(join(tmpdir(), 'rlss-adversary-'));
after(() => rm(temporary, {recursive: true, force: true}));
const modulePath = join(temporary, 'extraction.mjs');
await writeFile(modulePath, bundle.outputFiles[0].contents);
const browser = parseHTML('<html><head></head><body></body></html>');
globalThis.window = browser.window;
globalThis.document = browser.document;
globalThis.DOMParser = browser.window.DOMParser;
const { extractArticle } = await import(pathToFileURL(modulePath).href);
const doc = body => parseHTML(`<html><head></head><body>${body}</body></html>`).document;

test('test_social_quotedpermalink_rejectswrongbody', async () => {
 // A reply quoting the target must not establish ownership of the reply body.
 const document = doc('<article data-testid="tweet"><a href="/b/status/999"><time>today</time></a><div data-testid="tweetText">Other user reply.</div><div data-testid="quoteTweet"><a href="/a/status/123"><time>yesterday</time></a><div data-testid="tweetText">Target quoted text.</div></div></article>');
 await assert.rejects(extractArticle(document, 'https://x.com/a/status/123'), /target/i);
});

test('test_transcript_mismatchedmarkup_rejectsrepair', async () => {
 // XML is malformed even when its outer cue tag counts happen to match.
 const data = {videoDetails:{videoId:'abc'},captions:{playerCaptionsTracklistRenderer:{captionTracks:[{baseUrl:'https://www.youtube.com/api/timedtext?v=abc',languageCode:'en',kind:'asr'}],audioTracks:[{captionTrackIndices:[0]}]}}};
 const document = doc(`<script>var ytInitialPlayerResponse = ${JSON.stringify(data)};</script>`);
 const xml = '<transcript><text start="0" dur="2">Spoken <b>words</text></transcript>';
 await assert.rejects(extractArticle(document, 'https://www.youtube.com/watch?v=abc', async () => new Response(xml)), /unusable|malformed/i);
});
