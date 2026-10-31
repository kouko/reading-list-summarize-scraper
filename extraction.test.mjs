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
const datedProse = `<p>${'A concrete authored observation deserves careful reading. '.repeat(12)}<a href="https://example.com/reference">Reference</a></p><h2><span aria-hidden="true">PART 2/5</span>Next section</h2><p>${'A second observation preserves its original wording. '.repeat(12)}</p><figure><img src="https://example.com/owned.jpg" alt="Owned illustration"></figure>`;
test('dated-body-complete retains owned prose, headings, references and images on a copy', async () => {
 const d=doc(`<article><h1>Authored article</h1><time data-date="2026-10-13">DATE ONLY METADATA</time><div class="gh-content" data-date="2026-10-13">${datedProse}</div><form>SUBSCRIBE NOW</form></article>`);
 const original=d.documentElement.outerHTML;
 const out=await extractArticle(d,'https://example.com/article');
 for (const text of ['A concrete authored observation','A second observation','## Next section','[Reference](https://example.com/reference)','![Owned illustration](https://example.com/owned.jpg)']) assert.ok(out.includes(text),text);
 assert.doesNotMatch(out,/PART 2\/5|DATE ONLY METADATA|SUBSCRIBE NOW/);
 assert.equal(d.documentElement.outerHTML,original);
});
test('dated-body-unit-loss rejects ambiguous or lost owned units and ignores dated side content', async () => {
 for (const extra of [`<p class="advertisement">Unique authored sentence that must survive.</p>`,`<p>Case-sensitive authored SENTENCE.</p><div data-date="2026-10-13" class="gh-content">${datedProse}</div>`]) {
  await assert.rejects(extractArticle(doc(`<article><h1>Authored article</h1><div class="gh-content" data-date="2026-10-13">${datedProse}${extra}</div></article>`),'https://example.com/article'),/Article incomplete|Article target unverified/);
 }
 const out=await extractArticle(doc(`<article><h1>Real article</h1><p>${'The genuine ordinary article remains readable. '.repeat(30)}</p><aside><div class="gh-content" data-date="2026-10-13">${datedProse}</div></aside><section class="comments"><div data-date="2026-10-13">${datedProse}</div></section></article>`),'https://example.com/article');
 assert.doesNotMatch(out,/A concrete authored observation/);
});
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
  const root=first.__reactFiber$fixture.memoizedState.next.memoizedState.snapshot.data;
  root.counts.reply_count=1;
  const second=d.querySelectorAll('article')[1];
  second.__reactFiber$fixture.memoizedProps.tweet.core.user_results.result.user.rest_id='reader';
  const duplicate=second.cloneNode(true);duplicate.__reactFiber$fixture=second.__reactFiber$fixture;d.body.append(duplicate);
  assert.deepEqual(socialSnapshot(d,'https://x.com/writer/status/1').replyPages,[{rootID:'1',end:true,cursor:null}]);
  for (const reply_count of [2,-1,1.5,'1',undefined]) {
   root.counts.reply_count=reply_count;
   assert.deepEqual(socialSnapshot(d,'https://x.com/writer/status/1').replyPages,[]);
  }
  root.counts.reply_count=0;
  assert.deepEqual(socialSnapshot(d,'https://x.com/writer/status/1').replyPages,[]);
  root.counts.reply_count=undefined;
  const last=d.querySelectorAll('article')[count-1].__reactFiber$fixture.memoizedProps.tweet;
  last.counts={reply_count:0};
  assert.deepEqual(socialSnapshot(d,'https://x.com/writer/status/1').replyPages,[{rootID:String(count),end:true,cursor:null}]);
  delete last.counts;
  root.counts.reply_count=1;
  first.__reactFiber$fixture.memoizedProps.tweet.counts={reply_count:2};
  assert.throws(()=>socialSnapshot(d,'https://x.com/writer/status/1'),/conflicting.*count/i);
 }
 const legacyPage=doc('<article><a href="/writer/status/1">post</a><div dir="auto" class="whitespace-pre-wrap">Synthetic root.</div></article><article><a href="/writer/status/2">post</a><div dir="auto" class="whitespace-pre-wrap">Synthetic reply.</div></article>');
 const legacyTweets=[{rest_id:'1',core:{user_results:{result:{rest_id:'author'}}},legacy:{reply_count:1,conversation_id_str:'1'}},{rest_id:'2',core:{user_results:{result:{rest_id:'author'}}},legacy:{reply_count:0,conversation_id_str:'1',in_reply_to_status_id_str:'1'}}];
 [...legacyPage.querySelectorAll('article')].forEach((a,i)=>{a.__reactFiber$fixture={memoizedProps:{tweet:legacyTweets[i]}};});
 const snapshot=()=>socialSnapshot(legacyPage,'https://x.com/writer/status/1');
 assert.deepEqual(snapshot().posts.map(p=>[p.id,p.authorID,p.parentID,p.replyCount]),[['1','author',null,1],['2','author','1',0]]);
 assert.deepEqual(snapshot().replyPages,[{rootID:'1',end:true,cursor:null},{rootID:'2',end:true,cursor:null}]);
 delete legacyTweets[1].legacy.in_reply_to_status_id_str;
 assert.equal(snapshot().posts[1].parentID,undefined);
 delete legacyTweets[0].legacy.conversation_id_str;
 assert.equal(snapshot().posts[0].parentID,undefined);
 for (const conflict of [
  {core:{user_results:{result:{rest_id:'author',user:{rest_id:'other'}}}}},
  {counts:{reply_count:2}},
  {reply_to_results:{rest_id:'9'},legacy:{reply_count:1,conversation_id_str:'1'}},
  {reply_to_results:null,legacy:{reply_count:1,in_reply_to_status_id_str:'9'}},
  {legacy:{reply_count:1,conversation_id_str:'1',in_reply_to_status_id_str:'9'}},
 ]) {
  const prior=legacyTweets[0];legacyTweets[0]={...prior,...conflict};
  legacyPage.querySelector('article').__reactFiber$fixture.memoizedProps.tweet=legacyTweets[0];
  assert.throws(snapshot,/conflicting.*(identity|count|parent|relationship)/i);
  legacyTweets[0]=prior;legacyPage.querySelector('article').__reactFiber$fixture.memoizedProps.tweet=prior;
 }
 const normalizedTweets=[{id_str:'1',user:{id_str:'author'},full_text:'Synthetic root.',reply_count:1,conversation_id_str:'1'},{id_str:'2',user:{id_str:'author'},text:'Synthetic reply.',reply_count:0,conversation_id_str:'1',in_reply_to_status_id_str:'1'}];
 const retained=[];
 [...legacyPage.querySelectorAll('article')].forEach((a,i)=>{
  let fiber={memoizedProps:{tweet:normalizedTweets[i],analytics:{items:[{id:normalizedTweets[i].id_str,author_id:'analytics',reply_count:99}]}}};retained.push(fiber);
  for(let depth=0;depth<11;depth++) fiber={return:fiber};
  a.__reactFiber$fixture=fiber;
 });
 assert.deepEqual(snapshot().posts.map(p=>[p.id,p.authorID,p.parentID,p.replyCount]),[['1','author',null,1],['2','author','1',0]]);
 assert.deepEqual(snapshot().replyPages,[{rootID:'1',end:true,cursor:null},{rootID:'2',end:true,cursor:null}]);
 delete normalizedTweets[1].in_reply_to_status_id_str;
 assert.equal(snapshot().posts[1].parentID,undefined);
 delete normalizedTweets[0].conversation_id_str;
 assert.equal(snapshot().posts[0].parentID,undefined);
 for(const conflict of [{rest_id:'9'},{core:{user_results:{result:{rest_id:'other'}}}},{legacy:{reply_count:2}},{reply_to_results:null,in_reply_to_status_id_str:'9'},{conversation_id_str:'1',in_reply_to_status_id_str:'9'}]) {
  retained[0].memoizedProps.tweet={...normalizedTweets[0],...conflict};
  assert.throws(snapshot,/conflicting.*(identity|count|parent|relationship)/i);
 }
 retained[0].memoizedProps.tweet={id:'1',user:{id_str:'author'},full_text:'Synthetic analytics.'};
 assert.throws(snapshot,/target.*identity/i);
});
test('thread A1/A2/A3 owned X source text preserves authored breaks and reference links', () => {
 const text='First authored line.\nSecond authored line.\n\nReference https://t.co/ref';
 const entity={url:'https://t.co/ref',expanded_url:'https://x.com/foreign/status/99',display_url:'x.com/foreign/status/99',indices:[text.indexOf('https://t.co/ref'),text.length]};
 const expected='First authored line.  \nSecond authored line.  \n  \nReference [https://x.com/foreign/status/99](https://x.com/foreign/status/99)';
 for(const normalized of [false,true]) for(const note of [false,true]) {
  const d=doc('<article><a href="/writer/status/1">post</a><div data-testid="tweetText">First authored line. Second authored line. Reference <a href="/foreign/status/99">x.com/foreign/status/99</a></div><div data-testid="quoteTweet"><a href="/foreign/status/99">quote</a><div data-testid="tweetText">QUOTED BODY</div></div></article>');
  const legacy={full_text:note?'Legacy preview.':text,display_text_range:[0,note?7:text.length],entities:{urls:[entity]},reply_count:0};
  const tweet=normalized?{id_str:'1',user:{id_str:'author'},...legacy}:{rest_id:'1',core:{user_results:{result:{rest_id:'author'}}},legacy};
  if(note) tweet.note_tweet=normalized?{text,is_expandable:false,entity_set:{urls:[entity]}}:{note_tweet_results:{result:{text,entity_set:{urls:[entity]}}}};
  const quote=normalized?{id_str:'99',user:{id_str:'foreign'},full_text:'QUOTED SOURCE BODY'}:{rest_id:'99',core:{user_results:{result:{rest_id:'foreign'}}},legacy:{full_text:'QUOTED SOURCE BODY'}};
  tweet[normalized?'quoted_status':'quoted_status_result']=normalized?quote:{result:quote};
  const fiber={memoizedProps:{tweet},memoizedState:{memoizedState:{snapshot:{data:{rest_id:'1',reply_to_results:null,counts:{reply_count:0}}}}}};
  d.querySelector('article').__reactFiber$fixture=fiber;
  const capture=()=>socialSnapshot(d,'https://x.com/writer/status/1');
  assert.equal(capture().posts[0].body,expected);assert.equal(capture().posts[0].bodyComplete,true);
  assert.equal(capture().posts.length,1);
  const canonical=capture();fiber.memoizedProps.tweet={rest_id:'1'};
  assert.equal(socialSnapshot(d,'https://x.com/writer/status/1',[],{document:d,address:'https://x.com/writer/status/1',posts:canonical.posts,counts:canonical.observedReplyCounts}).posts[0].body,expected);
  fiber.memoizedProps.tweet=tweet;
  const marker=d.createElement('button');marker.textContent='Show more';d.querySelector('article').append(marker);
  assert.equal(capture().posts[0].bodyComplete,false);marker.remove();
  (normalized?tweet:tweet.legacy).truncated=true;
  assert.equal(capture().posts[0].bodyComplete,note);
  delete (normalized?tweet:tweet.legacy).truncated;
  const contradictory=structuredClone(tweet);
  if(note) (normalized?contradictory.note_tweet:contradictory.note_tweet.note_tweet_results.result).text='Contradictory complete authored text.';
  else (normalized?contradictory:contradictory.legacy).full_text='Contradictory complete authored text.';
  fiber.memoizedProps.otherObservation=contradictory;
  assert.throws(capture,/conflicting.*body/i);
 }
 for(const normalized of [false,true]) for(const truncated of [undefined,false]) {
  const complete='First complete paragraph.\nSecond complete paragraph.';
  const d=doc(`<article><a href="/writer/status/1">post</a><div data-testid="tweetText">${complete}</div></article>`);
  const preview={full_text:'First complete paragraph. https://t.co/media',truncated,display_text_range:[0,25]};
  const tweet=normalized?{id_str:'1',user:{id_str:'author'},...preview}:{rest_id:'1',core:{user_results:{result:{rest_id:'author'}}},legacy:preview};
  d.querySelector('article').__reactFiber$fixture={memoizedProps:{tweet}};
  const capture=()=>socialSnapshot(d,'https://x.com/writer/status/1');
  assert.equal(capture().posts[0].body,'First complete paragraph.  \nSecond complete paragraph.');
  assert.equal(capture().posts[0].bodyComplete,true);
  tweet.note_tweet={is_expandable:true,unknown_fragment:{}};
  assert.throws(capture,/note.*body|note.*source/i);
 }
});
test('thread A1/A3/A4 injected X snapshots retain bounded observations only within navigation', async () => {
 const injected=await build({entryPoints:['inject.ts'],bundle:true,platform:'node',format:'esm',write:false});
 const injectedPath=join(temporary,'injected.mjs');await writeFile(injectedPath,injected.outputFiles[0].contents);
 const saved={window:globalThis.window,document:globalThis.document,location:globalThis.location};
 const render=(d,posts)=>{
  d.body.innerHTML=posts.map(p=>`<article><a href="/writer/status/${p.id_str}">post</a><div data-testid="tweetText">${p.full_text}</div>${p.preview?'<button>Show more</button>':''}</article>`).join('');
  [...d.querySelectorAll('article')].forEach((a,i)=>{a.__reactFiber$fixture={memoizedProps:{tweet:posts[i]}};});
 };
 const root={id_str:'1',user:{id_str:'author'},full_text:'Complete synthetic root.',conversation_id_str:'1',reply_count:2};
 const child=id=>({id_str:id,user:{id_str:'author'},full_text:`Synthetic child ${id}.`,conversation_id_str:'1',in_reply_to_status_id_str:'1',reply_count:0});
 try {
  const d=doc('');render(d,[root,child('2')]);
  globalThis.window=d.defaultView;globalThis.document=d;globalThis.location={href:'https://x.com/writer/status/1'};
  await import(pathToFileURL(injectedPath).href);
  const capture=globalThis.window.socialSnapshot;
  const initial=capture();assert.equal(initial.posts.length,2);assert.equal(initial.replyPages.some(p=>p.rootID==='1'),false);
  initial.posts[0].body='Caller mutation';
  render(d,[{...child('3'),in_reply_to_status_id_str:undefined}]);
  assert.equal(capture().replyPages.some(p=>p.rootID==='1'),false);
  render(d,[child('3'),child('3')]);
  const cumulative=capture();assert.deepEqual(cumulative.posts.map(p=>p.id),['1','2','3']);
  assert.equal(cumulative.posts[0].body,root.full_text);
  assert.equal(cumulative.replyPages.some(p=>p.rootID==='1'&&p.end),true);
  assert.throws(()=>socialSnapshot(d,globalThis.location.href),/target.*identity/i);
  render(d,[{...root,full_text:'Short preview.',preview:true}]);
  assert.equal(capture().posts[0].body,root.full_text);assert.equal(capture().posts[0].bodyComplete,true);
  for(const conflict of [{user:{id_str:'other'}},{reply_count:3},{in_reply_to_status_id_str:'9'},{full_text:'Different complete body.'}]) {
   render(d,[{...root,...conflict}]);assert.throws(capture,/conflicting/i);
  }
  render(d,[root]);d.querySelector('article').__reactFiber$fixture.memoizedProps.tweet={rest_id:'1',reply_to_results:{rest_id:'9'}};
  assert.throws(capture,/conflicting/i);
  render(d,[child('3')]);globalThis.location.href='https://x.com/writer/status/9';
  assert.throws(capture,/target.*identity/i);
  globalThis.location.href='https://x.com/writer/status/1';assert.throws(capture,/target.*identity/i);
  render(d,[root]);assert.equal(capture().posts.length,1);
  const other=doc('');render(other,[child('3')]);globalThis.document=other;
  assert.throws(capture,/target.*identity/i);
  globalThis.document=d;render(d,[child('3')]);assert.throws(capture,/target.*identity/i);
  render(d,[root]);assert.equal(capture().posts.length,1);
  render(d,[{...root,reply_count:'2'}]);assert.throws(capture,/conflicting.*count/i);
  globalThis.location.href='https://x.com/writer/status/4';
  render(d,[{...root,id_str:'4',conversation_id_str:'4',reply_count:'0'}]);capture();
  render(d,[{...root,id_str:'4',conversation_id_str:'4',reply_count:0}]);assert.throws(capture,/conflicting.*count/i);
  globalThis.location.href='https://x.com/writer/status/1';render(d,[root]);capture();
  render(d,[...Array.from({length:2000},(_,i)=>child(String(i+2)))]);
  assert.throws(capture,/snapshot limit/i);
  globalThis.location.href='https://x.com/writer/status/5';
  render(d,[{...root,id_str:'5',conversation_id_str:'5'}]);
  d.querySelector('[data-testid="tweetText"]').textContent='x'.repeat(16000001);
  d.querySelector('article').__reactFiber$fixture.memoizedProps.tweet.full_text=d.querySelector('[data-testid="tweetText"]').textContent;
  assert.throws(capture,/snapshot limit/i);
 } finally {Object.assign(globalThis,saved);}
});
test('thread A1/A3 missing members and unknown pagination never prove completeness', () => {
 const posts = [threadPost('ROOT','1')];
 const s = socialSnapshot(threadPage(posts,{self_thread:{post_ids:{edges:[{node:{pk:'2'}}]},posts:{edges:[],page_info:{has_next_page:true,end_cursor:'cursor'}}}}),'https://www.threads.com/@writer/post/ROOT');
 assert.equal(s.sequences[0].end,false); assert.deepEqual(s.sequences[0].missingIDs,['2']);
 assert.throws(()=>socialSnapshot(doc(''),'https://www.threads.com/@writer/post/ROOT'),/target|identity/i);
 const mid=threadPost('MID','2',true), root=threadPost('ROOT','1');
 const separate=threadPage([mid]);
 const script=separate.createElement('script');script.type='application/json';script.textContent=JSON.stringify({data:{media:{text_post_app_info:{id:mid.text_post_app_info.id,is_reply:true,containing_thread:{posts:{edges:[{node:root}]}}}}}});separate.body.append(script);
 const recovered=socialSnapshot(separate,'https://www.threads.com/@writer/post/MID');
 assert.equal(recovered.posts.find(p=>p.id==='2').rootID,'1');assert.equal(recovered.posts.find(p=>p.id==='2').isReply,true);
 assert.equal(socialSnapshot(threadPage([mid,root]),'https://www.threads.com/@writer/post/MID').posts.find(p=>p.id==='2').rootID,undefined);
 const initial=threadPage([root,mid],{direct_replies:{edges:[],page_info:{has_next_page:true,end_cursor:'next'}}});
 const branch=node=>({node:{posts:{edges:[{node}],page_info:{has_next_page:false,end_cursor:null}}}});
 const supplement=(after='next',owner=root.text_post_app_info.id)=>({postID:'1',after,sortOrder:'TOP',data:{data:{media:{text_post_app_info:{id:owner,direct_replies:{edges:[branch(threadPost('LATE','3',true)),branch({...mid,caption:null})],page_info:{has_next_page:false,end_cursor:null}}}}}}});
 const merged=socialSnapshot(initial,'https://www.threads.com/@writer/post/ROOT',[supplement()]);
 assert.deepEqual(merged.sequences[0].ids,['1','2']);
 assert.deepEqual(merged.replyPages,[{rootID:'1',end:true,cursor:null}]);
 assert.deepEqual(merged.posts.map(p=>p.id),['1','2','3']);
 assert.equal(merged.posts.find(p=>p.id==='2').body,'Section MID.');
 const advancing=supplement();advancing.data.data.media.text_post_app_info.direct_replies.page_info={has_next_page:true,end_cursor:'last'};
 const terminal=supplement('last');terminal.data.data.media.text_post_app_info.direct_replies.edges=[branch(threadPost('FINAL','4',true))];
 assert.deepEqual(socialSnapshot(initial,'https://www.threads.com/@writer/post/ROOT',[advancing,terminal]).posts.map(p=>p.id),['1','2','3','4']);
 assert.deepEqual(socialSnapshot(initial,'https://www.threads.com/@writer/post/ROOT',[advancing]).replyPages,[{rootID:'1',end:false,cursor:'last'}]);
 const loop=supplement('last');loop.data.data.media.text_post_app_info.direct_replies.page_info={has_next_page:true,end_cursor:'next'};
 for(const pages of [[supplement('gap')],[supplement('next','foreign')],[{...supplement(),postID:'9'}],[{...supplement(),sortOrder:'RECENT'}],[advancing,loop],[supplement(),supplement()],Array(33).fill(supplement())]) {
  assert.throws(()=>socialSnapshot(initial,'https://www.threads.com/@writer/post/ROOT',pages),/source|cursor|page/i);
 }
 const conflicting=supplement();conflicting.data.data.media.text_post_app_info.direct_replies.edges[1].node.posts.edges[0].node.caption={text:'Conflicting complete body.'};
 assert.throws(()=>socialSnapshot(initial,'https://www.threads.com/@writer/post/ROOT',[conflicting]),/conflicting/i);
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
 const quotedReadView='<article data-testid="twitterArticleReadView"><div><div data-testid="twitterArticleRichTextView">FOREIGN RICH BODY</div></div></article>';
 const quotedCards=`<div data-testid="quoteTweet">${quotedReadView}</div><div data-testid="simpleTweet">${quotedReadView}</div><div data-slot="embedded-tweet-frame">${quotedReadView}</div><div aria-labelledby="id__quote"><div data-testid="User-Name">Foreign author</div>${quotedReadView}</div><article>${quotedReadView}</article>`;
 const classic=doc(`<article data-testid="tweet"><a href="/a/status/123"><time>today</time></a><article data-testid="twitterArticleReadView"><h1 data-testid="twitter-article-title">Own long title</h1><div><div data-testid="twitterArticleRichTextView"><p>Complete authored rich body.</p><a href="https://example.com/source">Reference</a><article>QUOTED ARTICLE</article><div data-slot="embedded-tweet-frame">QUOTED BODY</div></div></div></article>${quotedCards}<div data-testid="tweetText">OUTSIDE BODY</div></article>`);
 classic.querySelector('article').__reactFiber$fixture={memoizedProps:{tweet:{id_str:'123',user:{id_str:'author'},full_text:'Article reference.',conversation_id_str:'123',reply_count:2}}};
 const classicSnapshot=()=>socialSnapshot(classic,'https://x.com/a/status/123');
 assert.equal(classicSnapshot().posts[0].bodyComplete,true);
 assert.equal(classicSnapshot().posts.length,1);
 assert.equal(classicSnapshot().posts[0].body,snapshot.posts[0].body);
 assert.deepEqual(classicSnapshot().replyPages,[]);
 const container=classic.querySelector('[data-testid="twitterArticleRichTextView"]');
 const marker=classic.createElement('button');marker.textContent='Show more';container.append(marker);
 assert.equal(classicSnapshot().posts[0].bodyComplete,false);marker.remove();
 const duplicate=container.cloneNode(true);container.parentElement.append(duplicate);
 assert.equal(classicSnapshot().posts[0].body,'');assert.equal(classicSnapshot().posts[0].bodyComplete,false);duplicate.remove();
 container.innerHTML='';
 assert.equal(classicSnapshot().posts[0].body,'');assert.equal(classicSnapshot().posts[0].bodyComplete,false);
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
