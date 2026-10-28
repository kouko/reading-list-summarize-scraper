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
const { extractArticle, socialSnapshot } = await import(pathToFileURL(modulePath).href);
const doc = (body) => parseHTML(`<html><head></head><body>${body}</body></html>`).document;
// Public Threads response shape, with synthetic captions and identities only.
const threadPost = (code, id, reply = false) => ({pk:id, code, user:{id:'author',username:'writer'}, caption:{text:`Section ${code}.`},text_post_app_info:{id:`XDTTextPostAppMediaInfo:${id}`,is_reply:reply}});
const threadPage = (posts, extra = {}) => doc(`<script type="application/json">${JSON.stringify({data:{post:posts[0],connection:{id:posts[0].text_post_app_info.id,self_thread:{posts:{edges:posts.slice(1).map(node=>({node})),page_info:{has_next_page:false,end_cursor:null}},post_ids:{edges:posts.slice(1).map(p=>({node:{pk:p.pk}}))}},direct_replies:{edges:[],page_info:{has_next_page:false,end_cursor:null}},...extra}}})}</script>`);
test('thread A1/A2 source-ordered first and middle snapshots retain distinct members', () => {
 // Counts/shapes from Open Design, AI colleague, pstack, Claude Code,
 // adoption gap; captions and user identities are deliberately synthetic.
 for (const continuations of [3,1,5,4,4]) {
  const sample=[threadPost('ROOT','1'),...Array.from({length:continuations},(_,i)=>threadPost(`PART${i}`,String(i+2),true))];
  assert.equal(socialSnapshot(threadPage(sample),'https://www.threads.com/@writer/post/ROOT').sequences[0].ids.length,continuations+1);
 }
 const posts = [threadPost('ROOT','1'),threadPost('MID','2',true),threadPost('END','3',true)];
 const d = threadPage(posts);
 for (const target of ['ROOT','MID']) {
  const s = socialSnapshot(d,`https://www.threads.com/@writer/post/${target}`);
  assert.equal(s.targetID,target==='ROOT'?'1':'2');
  assert.deepEqual(s.sequences[0].ids,['1','2','3']);
  assert.equal(s.sequences[0].end,true);
  assert.deepEqual(s.posts.map(p=>p.body),['Section ROOT.','Section MID.','Section END.']);
 }
 // Dashboard, Japan MCP and Google source use the public React tweet shape.
 // These bounded synthetic windows do not assert the live thread lengths.
 for (const count of [6,2,8]) {
  const d=doc(Array.from({length:count},(_,i)=>`<article><a aria-current="page" href="/writer/status/${i+1}">post</a><div dir="auto" class="text-lead whitespace-pre-wrap">Section ${i+1}.</div></article>`).join(''));
  [...d.querySelectorAll('article')].forEach((a,i)=>{a.__reactFiber$fixture={memoizedProps:{tweet:{rest_id:String(i+1),core:{user_results:{result:{user:{rest_id:'author'}}}},reply_to_results:i?{rest_id:String(i)}:null},conversationPosition:{isStart:i===0,isEnd:i===count-1}}};});
  const s=socialSnapshot(d,'https://x.com/writer/status/2');
  assert.equal(s.posts.length,count);assert.equal(s.posts[0].parentID,null);assert.equal(s.posts[1].parentID,'1');assert.equal(s.posts.at(-1).conversationEnd,true);
  const first=d.querySelector('article');
  delete first.__reactFiber$fixture.memoizedProps.tweet.reply_to_results;
  first.__reactFiber$fixture.memoizedState={next:{memoizedState:{snapshot:{data:{rest_id:'1',reply_to_results:null,counts:{reply_count:2}}}}}};
  const merged=socialSnapshot(d,'https://x.com/writer/status/1').posts[0];
  assert.equal(merged.parentID,null);assert.equal(merged.replyCount,2);
 }
});
test('thread A1/A3 missing members and unknown pagination never prove completeness', () => {
 const posts = [threadPost('ROOT','1')];
 const s = socialSnapshot(threadPage(posts,{self_thread:{post_ids:{edges:[{node:{pk:'2'}}]},posts:{edges:[],page_info:{has_next_page:true,end_cursor:'cursor'}}}}),'https://www.threads.com/@writer/post/ROOT');
 assert.equal(s.sequences[0].end,false); assert.deepEqual(s.sequences[0].missingIDs,['2']);
 assert.throws(()=>socialSnapshot(doc(''),'https://www.threads.com/@writer/post/ROOT'),/target|identity/i);
});
test('thread A2 reader branches are separate from explicit self-thread membership', () => {
 const root=threadPost('ROOT','1'), reader={...threadPost('READER','9',true),user:{id:'reader',username:'reader'}};
 const next=threadPost('NEXT','2',true);next.text_post_app_info.containing_thread={posts:{edges:[{node:root}]}};
 const d=threadPage([root,next],{direct_replies:{edges:[{node:{posts:{edges:[{node:reader},{node:threadPost('CHAT','10',true)}],page_info:{has_next_page:false,end_cursor:null}}}}],page_info:{has_next_page:true,end_cursor:'more-replies'}}});
 const s=socialSnapshot(d,'https://www.threads.com/@writer/post/ROOT');
 assert.deepEqual(s.sequences[0].ids,['1','2']); assert.equal(s.branches.length,1); assert.deepEqual(s.branches[0].ids,['9','10']);
 assert.equal(s.posts.find(p=>p.id==='2').rootID,'1'); assert.equal(s.replyPages[0].end,false);assert.equal(s.replyPages[0].cursor,'more-replies');
});
test('thread A3 previews remain unresolved rather than full bodies', () => {
 const p=threadPost('ROOT','1');p.text_post_app_info.text_fragments={fragments:[{fragment_type:'plaintext',plaintext:'Short preview.'}]};p.caption={text:'Complete authored body.'};
 const s=socialSnapshot(threadPage([p]),'https://www.threads.com/@writer/post/ROOT');
 assert.equal(s.posts[0].body,'Complete authored body.'); assert.equal(s.posts[0].bodyComplete,true);
 p.caption=null;
 const preview=socialSnapshot(threadPage([p]),'https://www.threads.com/@writer/post/ROOT');
 assert.equal(preview.posts[0].bodyComplete,false);
});
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
 ['https://x.com/a/status/123','<article><a aria-current="page" href="/a">author</a><a aria-current="page" href="/a/status/123">date</a><div dir="auto" class="text-lead whitespace-pre-wrap">Target main post text.<br>Next line.<div>Separate block.</div></div><div data-slot="embedded-tweet-frame">quoted reply</div></article>'],
 ['https://x.com/a/article/123','<link rel="canonical" href="https://x.com/a/article/123"><div data-testid="twitterArticleReadView"><div data-testid="twitterArticleRichTextView"><div class="longform-unstyled">Target main post text.<br>Next line.</div><div class="longform-unstyled">Separate block.</div></div></div><article data-testid="tweet"><div data-testid="tweetText">reply</div></article>'],
 ['https://www.threads.net/@a/post/ABC','<script type="application/ld+json">{"@type":"SocialMediaPosting","url":"https://www.threads.net/@a/post/ABC","articleBody":"Target main post text.\\nNext line.\\n\\nSeparate block."}</script>'],
 ['https://www.threads.com/@a/post/ABC','<div data-pagelet="threads_post_page_0"><div data-pressable-container><a href="/@a/post/ABC"><time>today</time></a><span dir="auto">Target main post text.<br>Next line.<div>Separate block.</div></span><div data-pressable-container><a href="/@b/post/QUOTE"><time>then</time></a><span dir="auto">quoted reply</span></div></div><div data-pressable-container><a href="/@b/post/REPLY"><time>later</time></a><span dir="auto">reply</span></div></div>'],
 ['https://www.threads.com/@a/post/ABC','<div role="region"><a href="/@a/post/ABC"><time>today</time></a><span dir="auto">Target main post text.<br>Next line.<div>Separate block.</div></span><a href="/@b/post/QUOTE"><span dir="auto">quoted reply</span></a></div>'],
 ['https://medium.com/@a/title-abcdef123456','<link rel="canonical" href="https://medium.com/@a/title-abcdef123456"><article><p>Target main post text.<br>Next line.</p><div>Separate block.</div></article>']]) {
  const out = await extractArticle(doc(html), url); assert.match(out, /Target main post text\.\s+Next line\.\s+Separate block\./); assert.doesNotMatch(out, /reply/);
 }
 const rich=doc('<article><a aria-current="page" href="/a">author</a><a aria-current="page" href="/a/status/123">date</a><h1 data-article-part="title">Own long title</h1><div data-article-part="body"><style>UNRELATED STYLES</style><div>OUTSIDE BODY</div><div class="x-article-body"><p>Complete authored rich body.</p><a href="https://example.com/source">Reference</a><div data-slot="embedded-tweet-frame">QUOTED BODY</div></div></div></article>');
 rich.querySelector('article').__reactFiber$fixture={memoizedProps:{tweet:{rest_id:'123',core:{user_results:{result:{user:{rest_id:'author'}}}},reply_to_results:null,counts:{reply_count:2}}}};
 const snapshot=socialSnapshot(rich,'https://x.com/a/status/123');
 assert.equal(snapshot.posts[0].bodyComplete,true);assert.match(snapshot.posts[0].body,/# Own long title\s+Complete authored rich body\./);assert.match(snapshot.posts[0].body,/\[Reference\]\(https:\/\/example.com\/source\)/);assert.doesNotMatch(snapshot.posts[0].body,/OUTSIDE BODY|STYLES|QUOTED BODY/);assert.deepEqual(snapshot.replyPages,[]);
 assert.equal(await extractArticle(rich,'https://x.com/a/status/123'),snapshot.posts[0].body);
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
 const modern=doc('<article><a aria-current="page" href="/a">author</a><a aria-current="page" href="/a/status/123">date</a><div dir="auto" class="text-lead whitespace-pre-wrap">Partial authored body.<button>Show more</button></div></article>');
 modern.querySelector('article').__reactFiber$fixture={memoizedProps:{tweet:{rest_id:'123',core:{user_results:{result:{user:{rest_id:'author'}}}},reply_to_results:null}}};
 assert.equal(socialSnapshot(modern,'https://x.com/a/status/123').posts[0].bodyComplete,false);
 await assert.rejects(extractArticle(modern,'https://x.com/a/status/123'),/preview|incomplete/i);
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
