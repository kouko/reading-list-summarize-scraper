// concern: Contradictory Facebook permalink identities must never establish target ownership.
import test from 'node:test';
import assert from 'node:assert/strict';
import { build } from 'esbuild';
import { parseHTML } from 'linkedom';
const bundle = await build({entryPoints:['facebook.ts'],bundle:true,platform:'node',format:'esm',write:false});
const { facebook } = await import(`data:text/javascript;base64,${Buffer.from(bundle.outputFiles[0].contents).toString('base64')}`);

test('test_facebook_conflictingPermalink_rejected', async (t) => {
    // A matching public post_id cannot override an explicitly contradictory permalink.
    for (const permalink of ['https://www.facebook.com/writer/posts/999/', 'https://www.facebook.com/other/posts/123/']) {
        await t.test(permalink, () => {
        const record = {id:'story',post_id:'123',actors:[{id:'789',url:'https://www.facebook.com/writer'}],permalink_url:permalink,attachments:[],comet_sections:{content:{story:{id:'story',message:{text:'A different target body.'},attachments:[]}}}};
        const {document} = parseHTML(`<html><body><script type="application/json">${JSON.stringify(record)}</script><div role="article"><a href="${permalink}">date</a><div data-ad-preview="message">A different target body.</div></div></body></html>`);
        assert.throws(() => facebook(document,new URL('https://www.facebook.com/writer/posts/123/')), /Article (target unverified|incomplete)/, permalink);
        });
    }
});
