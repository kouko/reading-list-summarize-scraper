import Defuddle, { createMarkdownContent } from 'defuddle/full';
import { SaxesParser } from 'saxes';

const family = (language: string) => language.toLowerCase().replace(/_/g, '-').split('-')[0];
const substantive = (text: string) => /\p{L}/u.test(text);
const fail = (message: string): never => { throw new Error(message); };

// Read JSON without evaluating platform scripts. Track braces inside strings too.
function playerData(document: Document): any {
    for (const script of Array.from(document.querySelectorAll('script'))) {
        const source = script.textContent || '';
        const assignment = /(?:var\s+)?ytInitialPlayerResponse\s*=\s*/g.exec(source);
        if (!assignment) continue;
        const start = assignment.index + assignment[0].length;
        let depth = 0, quoted = false, escaped = false;
        for (let i = start; i < source.length; i++) {
            const c = source[i];
            if (quoted) {
                if (escaped) escaped = false;
                else if (c === '\\') escaped = true;
                else if (c === '"') quoted = false;
            } else if (c === '"') quoted = true;
            else if (c === '{') depth++;
            else if (c === '}' && --depth === 0) return JSON.parse(source.slice(start, i + 1));
        }
    }
    fail('Unavailable subtitles: missing player data');
}

function transcript(xml: string): string {
    const cues: {name: string; attributes: Record<string, string>; text: string}[] = [];
    let cue: typeof cues[number] | undefined;
    // Parse cue text directly: YouTube enforces Trusted Types on HTML sinks.
    // Saxes performs no resource loading; disallow DTDs/entities explicitly.
    try {
        const parser = new SaxesParser();
        parser.on('doctype', () => { throw new Error('DTD is not permitted'); });
        parser.on('opentag', tag => {
            if (tag.name === 'text' || tag.name === 'p') {
                if (cue) throw new Error('Nested cues');
                cue = {name: tag.name, attributes: tag.attributes as Record<string, string>, text: ''};
                cues.push(cue);
            }
        });
        const append = (text: string) => { if (cue) cue.text += text; };
        parser.on('text', append);
        parser.on('cdata', append);
        parser.on('closetag', tag => { if (cue?.name === tag.name) cue = undefined; });
        parser.write(xml).close();
    } catch (error) {
        fail(`Unusable subtitles: malformed XML (${error instanceof Error ? error.message : String(error)})`);
    }
    // Reject partial XML rather than letting the HTML parser repair a cut response.
    if (!/^\s*(?:<\?xml[^>]*>\s*)?<(transcript|timedtext)\b[^>]*>[\s\S]*<\/\1>\s*$/.test(xml)
        || /(?:truncated|incomplete)\s*=\s*["'](?:true|1)["']/i.test(xml)) fail('Unusable subtitles: incomplete payload');
    const opening = (xml.match(/<(?:text|p)\b/g) || []).length;
    const closing = (xml.match(/<\/(?:text|p)>/g) || []).length;
    if (opening !== closing || cues.length !== opening) fail('Unusable subtitles: malformed cues');
    let previous = -1;
    const lines: string[] = [];
    for (const cue of cues) {
        const startKey = cue.name === 'p' ? 't' : 'start';
        const durationKey = startKey === 't' ? 'd' : 'dur';
        const rawStart = cue.attributes[startKey], rawDuration = cue.attributes[durationKey];
        const start = !rawStart?.trim() ? NaN : Number(rawStart);
        const duration = !rawDuration?.trim() ? NaN : Number(rawDuration);
        if (!Number.isFinite(start) || !Number.isFinite(duration) || start < 0 || duration <= 0 || start < previous) fail('Unusable subtitles: invalid cue timing');
        previous = start;
        const text = cue.text.replace(/\s+/g, ' ').trim();
        if (substantive(text)) lines.push(text);
    }
    if (!lines.length) fail('Unusable subtitles: empty or timestamp-only cues');
    return lines.join('\n\n');
}

function youtubeCandidates(data: any, id: string): {originalLanguage: string; candidates: any[]} {
    if (data?.videoDetails?.videoId !== id) fail('Unavailable subtitles: wrong target player');
    const renderer = data?.captions?.playerCaptionsTracklistRenderer;
    const tracks = renderer?.captionTracks;
    if (!Array.isArray(tracks) || !tracks.length) fail('Unavailable subtitles');
    // YouTube documents ASR captions in the original audio language:
    // https://support.google.com/youtube/answer/13338784?hl=en
    // Multiple-language audio may instead use the default language, so require
    // a single caption-audio association and reject contrary streaming evidence.
    // Neither defaultCaptionTrackIndex nor a localized display label proves it.
    const audioTracks = renderer.audioTracks;
    const formats = [...(data.streamingData?.adaptiveFormats || []), ...(data.streamingData?.formats || [])];
    const audioIds = new Set(formats.map((f: any) => f.audioTrack?.id).filter(Boolean));
    const dubbed = formats.some((f: any) => /\bdubbed\b|\bdubbing\b/i.test(f.audioTrack?.displayName || ''));
    if (!Array.isArray(audioTracks) || audioTracks.length !== 1 || audioIds.size > 1 || dubbed) fail('Unverified original language: ambiguous or dubbed audio association');
    const audio = audioTracks[0];
    const associated = (audio.captionTrackIndices || []).map((i: number) => tracks[i]).filter(Boolean);
    const asr = associated.filter((t: any) => t.kind === 'asr' && typeof t.languageCode === 'string');
    const languages = new Set(asr.map((t: any) => family(t.languageCode)));
    if (asr.length !== 1 || languages.size !== 1) fail('Unverified original language: missing or conflicting original-audio automatic captions');
    try {
        if (new URL(asr[0].baseUrl).searchParams.has('tlang')) fail('Unverified original language: translated automatic caption association');
    } catch { fail('Unverified original language: invalid automatic caption association'); }
    const originalLanguage = asr[0].languageCode;
    // A caption explicitly translated by URL is never eligible.
    const candidates = associated.filter((t: any) => typeof t.languageCode === 'string' && family(t.languageCode) === family(originalLanguage));
    return {originalLanguage, candidates};
}

async function youtube(document: Document, url: URL, fetcher: typeof fetch): Promise<string> {
    const id = url.hostname === 'youtu.be' ? url.pathname.slice(1) : url.searchParams.get('v') || url.pathname.split('/')[2];
    if (!id) fail('Unavailable subtitles: missing video identity');
    const initial = youtubeCandidates(playerData(document), id);
    const errors: string[] = [];
    // One inline attempt followed by at most two official native-player requests.
    for (const client of [null, {clientName: 'IOS', clientVersion: '20.10.3'}, {clientName: 'ANDROID', clientVersion: '20.10.38'}]) {
        try {
            let candidates = initial.candidates;
            if (client) {
                const response = await fetcher('https://www.youtube.com/youtubei/v1/player?prettyPrint=false', {
                    method: 'POST', headers: {'Content-Type': 'application/json'}, credentials: 'include', redirect: 'error',
                    signal: AbortSignal.timeout(10000), body: JSON.stringify({context: {client}, videoId: id})
                });
                if (!response.ok) fail(`Unavailable subtitles: native player HTTP ${response.status}`);
                const data = await response.json();
                if (data?.playabilityStatus?.status !== 'OK') fail('Unavailable subtitles: native player not playable');
                const refreshed = youtubeCandidates(data, id);
                if (family(refreshed.originalLanguage) !== family(initial.originalLanguage)) fail('Unverified original language: refreshed player conflicts with inline evidence');
                candidates = refreshed.candidates;
            }
            for (const track of candidates) {
                try {
                    const endpoint = new URL(track.baseUrl);
                    if (endpoint.protocol !== 'https:' || !['www.youtube.com', 'youtube.com'].includes(endpoint.hostname)
                        || endpoint.pathname !== '/api/timedtext' || endpoint.searchParams.has('tlang') || endpoint.searchParams.get('v') !== id) fail('Unusable subtitles: disallowed or translated caption endpoint');
                    endpoint.searchParams.set('fmt', 'srv3');
                    const response = await fetcher(endpoint.href, {credentials: 'include', redirect: 'error', signal: AbortSignal.timeout(10000)});
                    if (!response.ok) fail(`Unusable subtitles: HTTP ${response.status}`);
                    const body = transcript(await response.text());
                    const source = track.kind === 'asr' ? 'automatic' : 'human';
                    const provenance = {platform: 'youtube', target: id, language: track.languageCode, source, evidence: 'original-audio-asr-association'};
                    return `<!-- rlss-video-v1 ${JSON.stringify(provenance)} -->\n\nSubtitle language: ${track.languageCode}\nSubtitle source: ${source}\n\n${body}`;
                } catch (error) { errors.push(error instanceof Error ? error.message : String(error)); }
            }
        } catch (error) { errors.push(error instanceof Error ? error.message : String(error)); }
    }
    fail(`Unusable original-language subtitles: ${errors.join('; ') || 'no eligible tracks'}`);
}

function sameTarget(candidate: string | null, target: URL): boolean {
    if (!candidate) return false;
    try {
        const other = new URL(candidate, target);
        return other.hostname.replace(/^www\./, '') === target.hostname.replace(/^www\./, '') && other.pathname.replace(/\/$/, '') === target.pathname.replace(/\/$/, '');
    } catch { return false; }
}

function bodyMarkdown(element: Element, url: URL): string {
    const clone = element.cloneNode(true) as Element;
    clone.querySelectorAll('nav, aside, footer, script, style, [role="button"], [data-testid="simpleTweet"], [data-testid="post-preview"], [data-pressable-container]').forEach(el => el.remove());
    return createMarkdownContent(clone.innerHTML, url.href).trim();
}

function ownTweet(article: Element): Element {
    const clone = article.cloneNode(true) as Element;
    clone.querySelectorAll('[data-testid="quoteTweet"], [data-testid="simpleTweet"], article, [data-slot="embedded-tweet-frame"]').forEach(el => el.remove());
    // Upstream TwitterExtractor identifies quoted cards by their labelled
    // container and nested author header. They cannot identify the outer post.
    clone.querySelectorAll('[aria-labelledby*="id__"]').forEach(el => {
        if (el.querySelector('[data-testid="User-Name"]')) el.remove();
    });
    return clone;
}

function postPreview(own: Element): boolean {
    return !!own.querySelector('[data-testid="tweet-text-show-more-link"], [data-testid="tweet-text-show-more"]')
        || Array.from(own.querySelectorAll('button, [role="button"]')).some(button => /^(?:show more|read more|顯示更多|もっと見る)$/i.test((button.textContent || '').trim()));
}

function modernXBody(own: Element, url: URL): string {
    const rich = own.querySelectorAll('[data-article-part="body"] .x-article-body');
    if (rich.length === 1) {
        const title = own.querySelector('[data-article-part="title"]');
        const body = bodyMarkdown(rich[0],url);
        return substantive(body) ? [title ? '# ' + bodyMarkdown(title,url) : '',body].filter(Boolean).join('\n\n') : '';
    }
    if (own.querySelector('[data-article-part]')) return '';
    const plain = own.querySelectorAll('[data-testid="tweetText"], [dir="auto"].whitespace-pre-wrap');
    return plain.length === 1 ? bodyMarkdown(plain[0],url) : '';
}

type SocialPost = {id: string; url: string; authorID: string; body: string; bodyComplete: boolean; parentID?: string | null; rootID?: string; isReply?: boolean; replyCount?: number; conversationStart?: boolean; conversationEnd?: boolean};
type SocialSequence = {rootID: string; ids: string[]; end: boolean; cursor: string | null; missingIDs: string[]};

// A snapshot is evidence for the browser walker, never a complete-article claim.
// Unknown parent/extent is deliberately distinct from an explicit null/end.
export function socialSnapshot(document: Document = globalThis.document, address: string = globalThis.location.href) {
    const url = new URL(address), host = url.hostname.replace(/^www\./, '');
    const posts = new Map<string, SocialPost>();
    const sequences: SocialSequence[] = [], branches: {rootID: string; ids: string[]; end: boolean}[] = [];
    const replyPages: {rootID: string; end: boolean; cursor: string | null}[] = [];
    const add = (post: SocialPost) => {
        const prior = posts.get(post.id);
        if (prior && (prior.authorID !== post.authorID || prior.url !== post.url || (prior.bodyComplete && post.bodyComplete && prior.body !== post.body)
            || (prior.parentID !== undefined && post.parentID !== undefined && prior.parentID !== post.parentID))) fail('Inaccessible social target: conflicting source identity or body');
        if (!prior || post.bodyComplete || !prior.bodyComplete) posts.set(post.id, {...prior,...post});
    };
    if (host === 'threads.com' || host === 'threads.net') {
        const rawPosts = new Map<string, any>(), collections: any[] = [];
        let remaining = 250000;
        const visit = (value: any, depth = 0) => {
            if (!value || typeof value !== 'object') return;
            if (--remaining < 0 || depth > 100) fail('Inaccessible social target: source snapshot limit');
            if (value.pk && value.code && value.user?.id && value.text_post_app_info?.id) rawPosts.set(String(value.pk), value);
            if (value.id && value.self_thread?.posts) collections.push(value);
            for (const child of Object.values(value)) visit(child, depth + 1);
        };
        for (const script of Array.from(document.querySelectorAll('script[type="application/json"]'))) {
            const text = script.textContent || '';
            if (text.length > 16000000) fail('Inaccessible social target: source snapshot limit');
            let parsed; try { parsed = JSON.parse(text); } catch { continue; }
            visit(parsed);
        }
        for (const [id, p] of rawPosts) {
            const caption = typeof p.caption?.text === 'string' ? p.caption.text.trim() : '';
            const preview = (p.text_post_app_info.text_fragments?.fragments || []).map((f: any) => f.plaintext || '').join('');
            const containing = p.text_post_app_info.containing_thread?.posts?.edges?.[0]?.node;
            add({id, url:`https://${host}/@${p.user.username}/post/${p.code}`, authorID:String(p.user.id), body:caption || preview,
                bodyComplete:substantive(caption), isReply:p.text_post_app_info.is_reply,
                ...(Number.isInteger(p.text_post_app_info.direct_reply_count) ? {replyCount:p.text_post_app_info.direct_reply_count} : {}),
                ...(containing?.pk ? {rootID:String(containing.pk)} : {})});
        }
        const ids = (connection: any) => (connection?.edges || []).map((e: any) => String(e.node?.pk || '')).filter(Boolean);
        for (const c of collections) {
            const roots = [...rawPosts.values()].filter(p => p.text_post_app_info.id === c.id);
            if (roots.length !== 1) continue;
            const rootID = String(roots[0].pk), connection = c.self_thread.posts;
            const members = ids(connection), expected = ids(c.self_thread.post_ids);
            const sequence = {rootID, ids:[rootID,...members], end:connection.page_info?.has_next_page === false,
                cursor:connection.page_info?.end_cursor || null, missingIDs:expected.filter((id: string) => !members.includes(id))};
            if (!sequences.some(s => JSON.stringify(s) === JSON.stringify(sequence))) sequences.push(sequence);
            replyPages.push({rootID,end:c.direct_replies?.page_info?.has_next_page === false,cursor:c.direct_replies?.page_info?.end_cursor || null});
            for (const edge of c.direct_replies?.edges || []) {
                const replies = edge.node?.posts;
                branches.push({rootID, ids:ids(replies), end:replies?.page_info?.has_next_page === false});
            }
        }
    } else if (host === 'x.com' || host === 'twitter.com') {
        // Only inspect bounded props belonging to rendered articles. Never return
        // React/app objects, account state, fetch tokens or arbitrary script data.
        for (const article of Array.from(document.querySelectorAll('article'))) {
            if (article.parentElement?.closest('article, [data-slot="embedded-tweet-frame"]')) continue;
            const key = Object.keys(article).find(k => k.startsWith('__reactFiber$'));
            let fiber = key ? (article as any)[key] : null;
            let budget = 5000;
            let position: any;
            const evidence = new Map<string, Partial<SocialPost>>();
            const seen = new Set<any>();
            const visit = (value: any, depth = 0) => {
                if (!value || typeof value !== 'object' || seen.has(value) || depth > 14 || --budget < 0) return;
                seen.add(value);
                if (typeof value.conversationPosition?.isStart === 'boolean' && typeof value.conversationPosition?.isEnd === 'boolean') position = value.conversationPosition;
                const user = value.core?.user_results?.result?.user;
                if (typeof value.rest_id === 'string') {
                    const id = value.rest_id, own = ownTweet(article);
                    const links = Array.from(own.querySelectorAll('a[href]'));
                    const permalink = links.find(a => new URL(a.getAttribute('href')!, url).pathname.endsWith('/status/' + id));
                    if (!permalink) return;
                    const body = modernXBody(own,url);
                    const fields: Partial<SocialPost> = {...evidence.get(id)};
                    if (typeof user?.rest_id === 'string') fields.authorID=user.rest_id;
                    if (Number.isInteger(value.counts?.reply_count)) fields.replyCount=value.counts.reply_count;
                    if (position) { fields.conversationStart=position.isStart; fields.conversationEnd=position.isEnd; }
                    if (value.reply_to_results === null) fields.parentID = null;
                    else if (typeof value.reply_to_results?.rest_id === 'string') fields.parentID = value.reply_to_results.rest_id;
                    evidence.set(id,fields);
                    if (fields.authorID) add({id,url:new URL(permalink.getAttribute('href')!,url).href,body,
                        bodyComplete:substantive(body) && !postPreview(own),...fields} as SocialPost);
                }
                // Props only; avoid React owner/DOM back references and functions.
                for (const [key, child] of Object.entries(value)) if (!['children','_owner','ref'].includes(key)) visit(child,depth+1);
            };
            for (let i=0; fiber && i<30; i++, fiber=fiber.return) {
                visit(fiber.memoizedProps || fiber.pendingProps);
                // useSyncExternalStore keeps public post query snapshots in hook
                // state; inspect only that named data and React's memo cache.
                let hook=fiber.memoizedState;
                for (let j=0; hook && j<20; j++, hook=hook.next) visit(hook.memoizedState?.snapshot?.data);
                visit(fiber.updateQueue?.memoCache?.data);
            }
        }
    } else fail('Inaccessible social target: unsupported snapshot platform');
    const targets = [...posts.values()].filter(p => sameTarget(p.url,url));
    if (targets.length !== 1) fail('Inaccessible social target: source target identity unavailable or ambiguous');
    return {version:1, platform:host.startsWith('threads')?'threads':'x', targetID:targets[0].id, posts:[...posts.values()], sequences, branches, replyPages};
}

function social(document: Document, url: URL): string {
    const host = url.hostname.replace(/^www\./, '');
    let text = '';
    if (host === 'x.com' || host === 'twitter.com') {
        const id = url.pathname.match(/^\/[^/]+\/(?:status|article)\/(\d+)(?:\/article)?\/?$/)?.[1];
        if (!id) fail('Inaccessible social target: invalid post URL');
        // XArticleExtractor 0.19.5 identifies these first-party read-view nodes.
        const richText = document.querySelector('[data-testid="twitterArticleReadView"] [data-testid="twitterArticleRichTextView"]');
        const canonical = document.querySelector('link[rel="canonical"]')?.getAttribute('href') || document.querySelector('meta[property="og:url"]')?.getAttribute('content') || null;
        if (richText && sameTarget(canonical, url)) {
            const own = ownTweet(richText);
            if (own.querySelector('[data-testid="tweet-text-show-more-link"]')) fail('Inaccessible social target: truncated article preview');
            text = bodyMarkdown(own, url);
        }
        for (const article of Array.from(document.querySelectorAll('article[data-testid="tweet"]'))) {
            if (text) break;
            const own = ownTweet(article);
            const permalinks = Array.from(own.querySelectorAll('a')).filter(a => a.querySelector('time') && !a.closest('[data-testid="tweetText"], [data-testid="twitterArticleRichTextView"]'));
            const target = new URL(url.origin + url.pathname.replace(/\/article\/?$/, ''));
            const ownsTarget = permalinks.length > 0 && permalinks.every(a => sameTarget(a.getAttribute('href'), target));
            if (!ownsTarget) continue;
            if (own.querySelector('[data-testid="tweet-text-show-more-link"]')) fail('Inaccessible social target: truncated post preview');
            const body = own.querySelector('[data-testid="twitterArticleRichTextView"]') || own.querySelector('[data-testid="tweetText"]');
            if (body) text = bodyMarkdown(body, url);
            break;
        }
        if (!text) {
            const target = new URL(url.origin + url.pathname.replace(/\/article\/?$/, ''));
            const author = new URL('/' + url.pathname.split('/')[1], url);
            const candidates = Array.from(document.querySelectorAll('article')).filter(a => !a.parentElement?.closest('article, [data-slot="embedded-tweet-frame"]')).map(ownTweet).filter(own => {
                const links = Array.from(own.querySelectorAll('a[aria-current="page"]')).filter(a => !a.closest('[data-article-part="body"], [dir="auto"].whitespace-pre-wrap'));
                const posts = links.filter(a => /\/(?:status|article)\/\d+/.test(a.getAttribute('href') || ''));
                return posts.length > 0 && posts.every(a => sameTarget(a.getAttribute('href'),target)) && links.some(a => sameTarget(a.getAttribute('href'),author));
            });
            if (candidates.length > 1) fail('Inaccessible social target: ambiguous post identity');
            if (candidates.length === 1) {
                const own = candidates[0];
                text = modernXBody(own,url);
                if (postPreview(own) || !substantive(text)) fail('Inaccessible social target: incomplete post preview or body');
            }
        }
    } else if (host === 'threads.net' || host === 'threads.com') {
        if (!/^\/@[^/]+\/post\/[^/]+\/?$/.test(url.pathname)) fail('Inaccessible social target: invalid post URL');
        // Scope to the exact permalink's post, never the entire conversation.
        // These pagelet/region/pressable/span selectors follow upstream 0.19.5.
        for (const link of Array.from(document.querySelectorAll('a[href*="/post/"]'))) {
            if (!sameTarget(link.getAttribute('href'), url)) continue;
            const container = link.closest('[data-pressable-container]') || link.closest('[data-pagelet^="threads_post_page_"], div[role="region"]');
            if (!container) continue;
            if (container.parentElement?.closest('[data-pressable-container]')) continue;
            const clone = container.cloneNode(true) as Element;
            clone.querySelectorAll('[data-pressable-container], a[href*="/post/"], a[href^="/@"], [role="button"], time').forEach(el => el.remove());
            const spans = Array.from(clone.querySelectorAll('span[dir="auto"]')).filter(span => !span.parentElement?.closest('span[dir="auto"]'));
            text = spans.map(span => bodyMarkdown(span, url)).filter(part => part && !['Author', '·', 'Top', 'View activity'].includes(part)).join('\n\n');
            if (text) break;
        }
        for (const script of Array.from(document.querySelectorAll('script[type="application/ld+json"]'))) {
            if (text) break;
            let parsed;
            try { parsed = JSON.parse(script.textContent || ''); } catch { continue; }
            const objects = Array.isArray(parsed) ? parsed : parsed['@graph'] || [parsed];
            for (const item of objects) if (item['@type'] === 'SocialMediaPosting' && sameTarget(item.url, url) && typeof item.articleBody === 'string') text = item.articleBody;
        }
    } else {
        if (!sameTarget(document.querySelector('link[rel="canonical"]')?.getAttribute('href') || null, url)) fail('Inaccessible social target: canonical identity missing or different');
        const visible = (element: Element) => !element.closest('[hidden], [aria-hidden="true"], [style*="display: none"], [style*="display:none"]');
        const denial = Array.from(document.querySelectorAll('[data-testid="paywall"], [data-testid="paywall-content"], [data-member-only="true"]')).some(visible);
        const denialText = Array.from(document.querySelectorAll('p, div, span')).some(el => visible(el) && !el.children.length && /read the rest.*member|sign up to read|purchase.*membership/i.test(el.textContent || ''));
        if (denial || denialText) fail('Inaccessible social target: subscription or preview');
        // isAccessibleForFree describes the resource, not this viewer's access.
        const article = document.querySelector('article.meteredContent') || document.querySelector('article');
        if (article) text = bodyMarkdown(article, url);
    }
    text = text.trim();
    if (!substantive(text) || /^(?:log in|sign in|sign up)\b/i.test(text)) fail('Inaccessible social target: main body unavailable');
    return text;
}

async function extractContent(document: Document, address: string, fetcher: typeof fetch): Promise<string> {
    const url = new URL(address), host = url.hostname.replace(/^www\./, '');
    if (['youtube.com', 'm.youtube.com', 'youtu.be'].includes(host)) return youtube(document, url, fetcher);
    if (['bilibili.com', 'm.bilibili.com'].includes(host)) fail('Unverified original language: Bilibili exposes subtitle language but no verified original-audio association');
    if (['vimeo.com', 'dailymotion.com', 'twitch.tv', 'tiktok.com', 'nicovideo.jp', 'niconico.jp', 'youku.com', 'douyin.com'].some(h => host === h || host.endsWith('.' + h)) || document.querySelector('meta[property="og:type"][content^="video"]')) fail('Unsupported video transcript extraction');
    if (['x.com', 'twitter.com', 'threads.net', 'threads.com', 'medium.com'].includes(host) || host.endsWith('.medium.com')) return social(document, url);
    const result = new Defuddle(document, { markdown: true }).parse();
    if (!result?.content?.trim()) fail('Ordinary article extraction failed: empty content');
    return result.content;
}

export async function extractArticle(document: Document = globalThis.document, address: string = globalThis.location.href, fetcher: typeof fetch = globalThis.fetch): Promise<string> {
    try {
        return await extractContent(document, address, fetcher);
    } catch (error) {
        // Preserve challenge identity for the existing headed_on_block policy.
        // Classify only explicit challenge UI after failure, never article text.
        const title = document.title.trim();
        let challenge = '';
        if (/^(?:attention required!?|just a moment(?:\.\.\.)?|security verification)\s*[|\-–]\s*cloudflare$/i.test(title)
            || document.querySelector('#cf-challenge-running, #cf-challenge-error')) challenge = 'Cloudflare security challenge';
        else if (/^(?:verify (?:you are|you're) (?:a )?human|captcha(?: security challenge)?)\s*[.!]?$/i.test(title)
            && document.querySelector('form#captcha, iframe[src*="hcaptcha.com/"], iframe[src*="recaptcha/"]')) challenge = 'CAPTCHA security challenge';
        if (challenge) fail(`${challenge}: ${error instanceof Error ? error.message : 'extraction failed'}`);
        throw error;
    }
}
