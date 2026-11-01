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

const validLanguage = (value: any): value is string => typeof value === 'string' && /^[a-z]{2,3}(?:[-_][a-z0-9]{2,8})*$/i.test(value);
function captionEndpoint(track: any, id: string): {endpoint: URL; language: string; translated: boolean} {
    const endpoint = new URL(track.baseUrl);
    const targets = endpoint.searchParams.getAll('tlang');
    if (endpoint.protocol !== 'https:' || !['www.youtube.com', 'youtube.com'].includes(endpoint.hostname)
        || endpoint.username || endpoint.password || endpoint.port || endpoint.pathname !== '/api/timedtext'
        || endpoint.searchParams.getAll('v').length !== 1 || endpoint.searchParams.get('v') !== id
        || !validLanguage(track.languageCode) || targets.length > 1
        || (targets.length === 1 && (!validLanguage(targets[0]) || targets[0].toLowerCase().replace(/_/g, '-') === track.languageCode.toLowerCase().replace(/_/g, '-')))
        || (track.translationLanguage?.languageCode && track.translationLanguage.languageCode !== targets[0])) fail('Unusable subtitles: disallowed caption endpoint or translation target');
    return {endpoint, language: targets[0] || track.languageCode, translated: targets.length === 1};
}
function youtubeCandidates(data: any, id: string): {originalLanguage?: string; evidence?: string; candidates: any[]; fallback: any[]} {
    if (data?.videoDetails?.videoId !== id) fail('Unavailable subtitles: wrong target player');
    const renderer = data?.captions?.playerCaptionsTracklistRenderer;
    const tracks = renderer?.captionTracks;
    if (!Array.isArray(tracks) || !tracks.length) return {candidates: [], fallback: []};
    const audioTracks = renderer.audioTracks;
    const formats = [...(data.streamingData?.adaptiveFormats || []), ...(data.streamingData?.formats || [])];
    const audios = formats.map((f: any) => f.audioTrack).filter(Boolean);
    // Recognize complete affirmative labels, not an "original" substring in
    // negated or qualified prose. Unknown declarations cannot supply weaker proof.
    const affirmativeOriginal = (name: string) => /^(?:English(?: \([A-Z]{2,3}\))?|Japanese) (?:original|\(original\))$/i.test(name);
    const originalClaims = audios.filter((a: any) => /\boriginal\b/i.test(a.displayName || ''));
    const originals = originalClaims.filter((a: any) => affirmativeOriginal(a.displayName));
    const originalIDs = new Set(originals.map((a: any) => a.id));
    const indices = (audio: any): number[] | undefined => Array.isArray(audio?.captionTrackIndices) && audio.captionTrackIndices.length
        && audio.captionTrackIndices.every((i: any) => Number.isInteger(i) && i >= 0 && i < tracks.length)
        && new Set(audio.captionTrackIndices).size === audio.captionTrackIndices.length ? audio.captionTrackIndices : undefined;
    let associated: any[] = [], originalLanguage: string | undefined, evidence: string | undefined;
    if (originalClaims.length) {
        const originalID = originals[0]?.id;
        const language = typeof originalID === 'string' ? originalID.match(/^([a-z]{2,3}(?:[-_][a-z0-9]{2,8})*)\.\d+$/i)?.[1] : undefined;
        const matches = Array.isArray(audioTracks) ? audioTracks.filter((a: any) => a.audioTrackId === originalID) : [];
        if (originals.length === originalClaims.length && originalIDs.size === 1 && validLanguage(language) && matches.length === 1 && indices(matches[0])
            && !audios.some((a: any) => a.id === originalID && /\bdubbed\b|\bdubbing\b/i.test(a.displayName || ''))) {
            associated = indices(matches[0])!.map(i => tracks[i]);
            originalLanguage = language; evidence = 'original-audio-metadata';
        }
    } else if (Array.isArray(audioTracks) && audioTracks.length === 1 && indices(audioTracks[0])
        && new Set(audios.map((a: any) => a.id)).size <= 1
        && !audios.some((a: any) => /\bdubbed\b|\bdubbing\b/i.test(a.displayName || ''))) {
        associated = indices(audioTracks[0])!.map(i => tracks[i]);
        const asr = associated.filter((t: any) => t.kind === 'asr');
        if (asr.length === 1 && validLanguage(asr[0].languageCode)) {
            originalLanguage = asr[0].languageCode; evidence = 'original-audio-asr-association';
        }
    }
    if (originalLanguage && associated.some((t: any) => t.kind === 'asr' && (!validLanguage(t.languageCode)
        || family(t.languageCode) !== family(originalLanguage!) || (() => { try { return captionEndpoint(t, id).translated; } catch { return true; } })()))) originalLanguage = undefined;
    const defaults = (Array.isArray(audioTracks) ? audioTracks : []).flatMap((a: any) => a.hasDefaultTrack && Number.isInteger(a.defaultCaptionTrackIndex) ? [tracks[a.defaultCaptionTrackIndex]] : []).filter(Boolean);
    return {originalLanguage, evidence, candidates: originalLanguage ? associated.filter((t: any) => validLanguage(t.languageCode) && family(t.languageCode) === family(originalLanguage!)) : [], fallback: [...new Set([...defaults, ...tracks])]};
}

async function youtube(document: Document, url: URL, fetcher: typeof fetch): Promise<string> {
    const id = url.hostname === 'youtu.be' ? url.pathname.slice(1) : url.searchParams.get('v') || url.pathname.split('/')[2];
    if (!id) fail('Unavailable subtitles: missing video identity');
    const initial = youtubeCandidates(playerData(document), id);
    const errors: string[] = [];
    const selections = [initial];
    const verified = new Set<string>();
    if (initial.originalLanguage) verified.add(family(initial.originalLanguage));
    const deliver = async (track: any, selection: typeof initial, fallback: boolean): Promise<string> => {
        const {endpoint, language, translated} = captionEndpoint(track, id);
        if (!fallback && translated) fail('Unusable subtitles: translated original candidate');
        endpoint.searchParams.set('fmt', 'srv3');
        const response = await fetcher(endpoint.href, {credentials: 'include', redirect: 'error', signal: AbortSignal.timeout(10000)});
        if (!response.ok) fail(`Unusable subtitles: HTTP ${response.status}`);
        const body = transcript(await response.text());
        const source = track.kind === 'asr' ? 'automatic' : 'human';
        const languageStatus = translated || (verified.size === 1 && !verified.has(family(language))) ? 'known-translation'
            : !fallback ? 'verified-original' : 'original-unconfirmed';
        const provenance = {platform: 'youtube', target: id, language, source, languageStatus,
            evidence: !fallback ? selection.evidence : 'validated-platform-caption', sourceLanguage: track.languageCode,
            ...(verified.size === 1 ? {originalLanguage: [...verified][0]} : {}), platformTranslation: translated};
        const label = {'verified-original': 'verified original', 'original-unconfirmed': 'original language unconfirmed', 'known-translation': 'known translation'}[languageStatus];
        return `<!-- rlss-video-v1 ${JSON.stringify(provenance)} -->\n\nSubtitle language: ${language}\nSubtitle source: ${source}\nSubtitle language status: ${label}\n${translated ? `Subtitle source language: ${track.languageCode}\n` : ''}\n${body}`;
    };
    // One inline attempt followed by at most two official native-player requests.
    for (const client of [null, {clientName: 'IOS', clientVersion: '20.10.3'}, {clientName: 'ANDROID', clientVersion: '20.10.38'}]) {
        try {
            let selection = initial;
            if (client) {
                const response = await fetcher('https://www.youtube.com/youtubei/v1/player?prettyPrint=false', {
                    method: 'POST', headers: {'Content-Type': 'application/json'}, credentials: 'include', redirect: 'error',
                    signal: AbortSignal.timeout(10000), body: JSON.stringify({context: {client}, videoId: id})
                });
                if (!response.ok) fail(`Unavailable subtitles: native player HTTP ${response.status}`);
                const data = await response.json();
                if (data?.playabilityStatus?.status !== 'OK') fail('Unavailable subtitles: native player not playable');
                const refreshed = youtubeCandidates(data, id);
                selections.push(refreshed);
                if (refreshed.originalLanguage) verified.add(family(refreshed.originalLanguage));
                selection = refreshed;
            }
            for (const track of verified.size <= 1 ? selection.candidates : []) {
                try {
                    return await deliver(track, selection, false);
                } catch (error) { errors.push(error instanceof Error ? error.message : String(error)); }
            }
        } catch (error) { errors.push(error instanceof Error ? error.message : String(error)); }
    }
    for (const selection of selections) for (const track of selection.fallback) {
        try { return await deliver(track, selection, true); }
        catch (error) { errors.push(error instanceof Error ? error.message : String(error)); }
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
    // The native long-article read view is itself an article owned by the
    // enclosing tweet. Known quote containers still remove any read view inside.
    clone.querySelectorAll('[data-testid="quoteTweet"], [data-testid="simpleTweet"], article:not([data-testid="twitterArticleReadView"]), [data-slot="embedded-tweet-frame"]').forEach(el => el.remove());
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
    const rich = own.querySelectorAll('[data-article-part="body"] .x-article-body, [data-testid="twitterArticleReadView"] [data-testid="twitterArticleRichTextView"]');
    if (rich.length === 1) {
        const title = own.querySelector('[data-article-part="title"], [data-testid="twitter-article-title"]');
        const body = bodyMarkdown(rich[0],url);
        return substantive(body) ? [title ? '# ' + bodyMarkdown(title,url) : '',body].filter(Boolean).join('\n\n') : '';
    }
    if (own.querySelector('[data-article-part], [data-testid="twitterArticleReadView"], [data-testid="twitterArticleRichTextView"]')) return '';
    const plain = own.querySelectorAll('[data-testid="tweetText"], [dir="auto"].whitespace-pre-wrap');
    if (plain.length !== 1) return '';
    const clone = plain[0].cloneNode(true) as Element;
    const breaks = (node: Node) => {
        for (const child of Array.from(node.childNodes)) {
            if (child.nodeType === 3 && child.textContent?.includes('\n')) {
                const fragment = own.ownerDocument.createDocumentFragment();
                child.textContent.split('\n').forEach((line,i) => {
                    if (i) fragment.appendChild(own.ownerDocument.createElement('br'));
                    fragment.appendChild(own.ownerDocument.createTextNode(line));
                });
                node.replaceChild(fragment,child);
            } else breaks(child);
        }
    };
    breaks(clone);
    return bodyMarkdown(clone,url);
}

function xPlainSource(value: any, own: Element, url: URL) {
    // Rich articles have their own structured body and completeness guards.
    if (own.querySelector('[data-article-part], [data-testid="twitterArticleReadView"], [data-testid="twitterArticleRichTextView"]')) return;
    const rawNote = value.note_tweet?.note_tweet_results?.result;
    const renderedNote = value.note_tweet;
    if (typeof rawNote?.text === 'string' && typeof renderedNote?.text === 'string' && rawNote.text !== renderedNote.text) fail('Inaccessible social target: conflicting source body');
    const note = typeof rawNote?.text === 'string' ? rawNote : renderedNote;
    const isNote = typeof note?.text === 'string';
    if (value.note_tweet != null && !isNote) fail('Inaccessible social target: unknown note body source');
    if (!isNote && typeof value.full_text === 'string' && typeof value.legacy?.full_text === 'string' && value.full_text !== value.legacy.full_text) fail('Inaccessible social target: conflicting source body');
    const text = isNote ? note.text : value.full_text ?? value.legacy?.full_text;
    if (typeof text !== 'string') return;
    const links = new Map<string, string>();
    const presentationURLs = new Set<string>();
    for (const entities of [note?.entity_set,value.entities,value.legacy?.entities]) for (const entity of entities?.urls || []) {
        if (typeof entity.url !== 'string' || !entity.url || typeof entity.expanded_url !== 'string' || !text.includes(entity.url)) continue;
        let expanded: URL; try { expanded = new URL(entity.expanded_url); } catch { fail('Inaccessible social target: invalid source reference URL'); }
        if (!['https:','http:'].includes(expanded.protocol)) fail('Inaccessible social target: invalid source reference URL');
        if (links.has(entity.url) && links.get(entity.url) !== expanded.href) fail('Inaccessible social target: conflicting source body URL');
        links.set(entity.url,expanded.href);
        for (const token of [entity.url,entity.expanded_url,expanded.href,entity.display_url]) if (typeof token === 'string' && token) presentationURLs.add(token);
        if (links.size > 5000) fail('Inaccessible social target: source snapshot limit');
    }
    // Ordinary full_text can still be a legacy note preview without a
    // truncated flag. Compare authored captions, ignoring known URL display
    // forms, before allowing it to replace the owned rendered body.
    const caption = (text: string) => {
        for (const token of presentationURLs) text=text.split(token).join('');
        return text.replace(/\s+/g,' ').trim();
    };
    const plain = own.querySelectorAll('[data-testid="tweetText"], [dir="auto"].whitespace-pre-wrap');
    const agreesRendered = isNote || (plain.length === 1 && caption(text) === caption(plain[0].textContent || ''));
    const container = own.ownerDocument.createElement('div');
    const append = (text: string) => text.split('\n').forEach((line,i) => {
        if (i) container.appendChild(own.ownerDocument.createElement('br'));
        container.appendChild(own.ownerDocument.createTextNode(line));
    });
    let cursor = 0;
    while (cursor < text.length) {
        let next = text.length, source = '';
        for (const candidate of links.keys()) {
            const index = text.indexOf(candidate,cursor);
            if (index >= 0 && index < next) { next=index; source=candidate; }
        }
        append(text.slice(cursor,next));
        if (!source) break;
        const anchor = own.ownerDocument.createElement('a');
        anchor.setAttribute('href',links.get(source)!); anchor.textContent=links.get(source)!;
        container.appendChild(anchor); cursor=next+source.length;
    }
    const body = bodyMarkdown(container,url);
    return {body,note:isNote,agreesRendered,complete:substantive(body) && (isNote || !(value.truncated === true || value.legacy?.truncated === true))};
}

type SocialPost = {id: string; url: string; authorID: string; body: string; bodyComplete: boolean; bodySource?: 'native'; parentID?: string | null; rootID?: string; isReply?: boolean; replyCount?: number; conversationStart?: boolean; conversationEnd?: boolean};
type SocialSequence = {rootID: string; ids: string[]; end: boolean; cursor: string | null; missingIDs: string[]};
type ThreadsReplyPage = {postID: string; after: string; sortOrder: 'TOP'; data: any};
type SocialObservations = {document: Document; address: string; posts: SocialPost[]; counts: [string, string][]};

// A snapshot is evidence for the browser walker, never a complete-article claim.
// Unknown parent/extent is deliberately distinct from an explicit null/end.
export function socialSnapshot(document: Document = globalThis.document, address: string = globalThis.location.href, pages: ThreadsReplyPage[] = [], previous?: SocialObservations) {
    const url = new URL(address), host = url.hostname.replace(/^www\./, '');
    const isX = host === 'x.com' || host === 'twitter.com';
    const posts = new Map<string, SocialPost>();
    const observedCounts = new Map<string, string>();
    const sequences: SocialSequence[] = [], branches: {rootID: string; ids: string[]; end: boolean}[] = [];
    const replyPages: {rootID: string; end: boolean; cursor: string | null}[] = [];
    const add = (post: SocialPost) => {
        const prior = posts.get(post.id);
        if (prior && (prior.authorID !== post.authorID || prior.url !== post.url || (prior.bodyComplete && post.bodyComplete && prior.body !== post.body)
            || (prior.parentID !== undefined && post.parentID !== undefined && prior.parentID !== post.parentID))) fail('Inaccessible social target: conflicting source identity or body');
        if (isX && prior?.bodyComplete && !post.bodyComplete) posts.set(post.id, {...post,...prior});
        else if (!prior || post.bodyComplete || !prior.bodyComplete) posts.set(post.id, {...prior,...post});
        if (isX && posts.size > 2000) fail('Inaccessible social target: source snapshot limit');
    };
    if (host === 'threads.com' || host === 'threads.net') {
        const rawPosts = new Map<string, any>(), collections: any[] = [];
        const ancestry = new Map<string, string>();
        let remaining = 250000;
        const visit = (value: any, depth = 0) => {
            if (!value || typeof value !== 'object') return;
            if (--remaining < 0 || depth > 100) fail('Inaccessible social target: source snapshot limit');
            // Relay may deliver the containing-thread relation separately from
            // the media/caption fragment. Join only by its stable info identity.
            if (typeof value.id === 'string' && value.id.startsWith('XDTTextPostAppMediaInfo:')) {
                const ancestors = value.containing_thread?.posts?.edges || [];
                for (const edge of ancestors) if (edge.node?.pk && edge.node?.text_post_app_info?.is_reply === false) {
                    const root = String(edge.node.pk), prior = ancestry.get(value.id);
                    if (prior && prior !== root) fail('Inaccessible social target: conflicting source ancestor');
                    ancestry.set(value.id,root);
                }
            }
            if (value.pk && value.code && value.user?.id && value.text_post_app_info?.id) {
                const prior = rawPosts.get(String(value.pk));
                if (prior && (prior.code !== value.code || String(prior.user.id) !== String(value.user.id)
                    || prior.user.username !== value.user.username || prior.text_post_app_info.id !== value.text_post_app_info.id
                    || (prior.text_post_app_info.is_reply !== undefined && value.text_post_app_info.is_reply !== undefined && prior.text_post_app_info.is_reply !== value.text_post_app_info.is_reply)
                    || (substantive(prior.caption?.text || '') && substantive(value.caption?.text || '') && prior.caption.text.trim() !== value.caption.text.trim()))) fail('Inaccessible social target: conflicting source identity or body');
                rawPosts.set(String(value.pk), {...prior,...value,
                    caption:substantive(value.caption?.text || '') ? value.caption : prior?.caption || value.caption,
                    text_post_app_info:{...prior?.text_post_app_info,...value.text_post_app_info}});
            }
            if (value.id && value.self_thread?.posts) collections.push(value);
            for (const child of Object.values(value)) visit(child, depth + 1);
        };
        for (const script of Array.from(document.querySelectorAll('script[type="application/json"]'))) {
            const text = script.textContent || '';
            if (text.length > 16000000) fail('Inaccessible social target: source snapshot limit');
            let parsed; try { parsed = JSON.parse(text); } catch { continue; }
            visit(parsed);
        }
        if (!Array.isArray(pages) || pages.length > 32) fail('Inaccessible social target: source page limit');
        const cursors = new Map<string, Set<string>>();
        for (const page of pages) {
            if (JSON.stringify(page).length > 16000000) fail('Inaccessible social target: source snapshot limit');
            const root = rawPosts.get(page?.postID), info = page?.data?.data?.media?.text_post_app_info;
            const owned = collections.filter(c => c.id === root?.text_post_app_info.id);
            const distinct = [...new Set(owned.map(c => JSON.stringify(c)))];
            const connection = owned[0]?.direct_replies, incoming = info?.direct_replies;
            const seen = cursors.get(page?.postID) || new Set<string>();
            if (!root || !info || info.id !== root.text_post_app_info.id || distinct.length !== 1
                || page.sortOrder !== 'TOP' || typeof page.after !== 'string' || !page.after
                || connection?.page_info?.has_next_page !== true || connection.page_info.end_cursor !== page.after
                || seen.has(page.after) || !Array.isArray(connection.edges) || !Array.isArray(incoming?.edges)
                || typeof incoming.page_info?.has_next_page !== 'boolean'
                || (incoming.page_info.has_next_page ? typeof incoming.page_info.end_cursor !== 'string' || !incoming.page_info.end_cursor
                    || incoming.page_info.end_cursor === page.after || seen.has(incoming.page_info.end_cursor) : incoming.page_info.end_cursor !== null)) fail('Inaccessible social target: invalid source page owner or cursor');
            seen.add(page.after); cursors.set(page.postID,seen);
            // Decode public fragments without treating a response's absent
            // self_thread as evidence that the initial sequence disappeared.
            const count = collections.length;
            visit(info);
            collections.length = count;
            for (const c of owned) c.direct_replies = {edges:[...connection.edges,...incoming.edges],page_info:incoming.page_info};
        }
        for (const [id, p] of rawPosts) {
            const caption = typeof p.caption?.text === 'string' ? p.caption.text.trim() : '';
            const preview = (p.text_post_app_info.text_fragments?.fragments || []).map((f: any) => f.plaintext || '').join('');
            const rootID = ancestry.get(p.text_post_app_info.id);
            add({id, url:`https://${host}/@${p.user.username}/post/${p.code}`, authorID:String(p.user.id), body:caption || preview,
                bodyComplete:substantive(caption), isReply:p.text_post_app_info.is_reply,
                ...(Number.isInteger(p.text_post_app_info.direct_reply_count) ? {replyCount:p.text_post_app_info.direct_reply_count} : {}),
                ...(rootID ? {rootID} : {})});
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
            const replyPage = {rootID,end:c.direct_replies?.page_info?.has_next_page === false,cursor:c.direct_replies?.page_info?.end_cursor || null};
            if (!replyPages.some(p => JSON.stringify(p) === JSON.stringify(replyPage))) replyPages.push(replyPage);
            for (const edge of c.direct_replies?.edges || []) {
                const replies = edge.node?.posts;
                const branch = {rootID, ids:ids(replies), end:replies?.page_info?.has_next_page === false};
                if (!branches.some(b => JSON.stringify(b) === JSON.stringify(branch))) branches.push(branch);
            }
        }
    } else if (host === 'x.com' || host === 'twitter.com') {
        const replyCounts = new Map<string, unknown>();
        if (previous?.document === document && previous.address === url.href) {
            for (const post of previous.posts) add({...post});
            for (const [id, count] of previous.counts) observedCounts.set(id,count);
        }
        // Only inspect bounded props belonging to rendered articles. Never return
        // React/app objects, account state, fetch tokens or arbitrary script data.
        for (const article of Array.from(document.querySelectorAll('article'))) {
            if (article.parentElement?.closest('article, [data-slot="embedded-tweet-frame"]')) continue;
            const key = Object.keys(article).find(k => k.startsWith('__reactFiber$'));
            let fiber = key ? (article as any)[key] : null;
            let budget = 5000;
            let position: any;
            const evidence = new Map<string, Partial<SocialPost>>();
            const sourceBodies = new Map<string, NonNullable<ReturnType<typeof xPlainSource>>>();
            const permalinks = new Map<string, string>();
            const own = ownTweet(article);
            const links = Array.from(own.querySelectorAll('a[href]'));
            const seen = new Set<any>();
            const visit = (value: any, depth = 0) => {
                if (!value || typeof value !== 'object' || seen.has(value) || depth > 14 || --budget < 0) return;
                seen.add(value);
                if (typeof value.conversationPosition?.isStart === 'boolean' && typeof value.conversationPosition?.isEnd === 'boolean') position = value.conversationPosition;
                const userResult = value.core?.user_results?.result;
                // Rendered tweet props normalize the API identity and legacy
                // fields. Analytics IDs alone are never public post evidence.
                const normalized = typeof value.id_str === 'string' && typeof value.user?.id_str === 'string'
                    && (typeof value.full_text === 'string' || typeof value.text === 'string');
                if (typeof value.rest_id === 'string' || normalized) {
                    if (typeof value.rest_id === 'string' && typeof value.id_str === 'string' && value.rest_id !== value.id_str) fail('Inaccessible social target: conflicting source identity');
                    const id = typeof value.rest_id === 'string' ? value.rest_id : value.id_str;
                    const permalink = links.find(a => new URL(a.getAttribute('href')!, url).pathname.endsWith('/status/' + id));
                    if (!permalink) return;
                    permalinks.set(id,new URL(permalink.getAttribute('href')!,url).href);
                    const sourceBody = xPlainSource(value,own,url), priorBody = sourceBodies.get(id);
                    if (sourceBody) {
                        if (priorBody && priorBody.note === sourceBody.note && priorBody.complete && sourceBody.complete && priorBody.body !== sourceBody.body) fail('Inaccessible social target: conflicting source body');
                        if (!priorBody || (sourceBody.note && !priorBody.note) || (sourceBody.note === priorBody.note && (sourceBody.complete || !priorBody.complete))) sourceBodies.set(id,sourceBody);
                    }
                    const fields: Partial<SocialPost> = {authorID:posts.get(id)?.authorID,...evidence.get(id)};
                    for (const author of [userResult?.user?.rest_id,userResult?.rest_id,normalized ? value.user.id_str : undefined]) if (typeof author === 'string') {
                        if (fields.authorID !== undefined && fields.authorID !== author) fail('Inaccessible social target: conflicting source identity');
                        fields.authorID=author;
                    }
                    for (const count of [value.counts?.reply_count,value.legacy?.reply_count,normalized ? value.reply_count : undefined]) if (count !== undefined) {
                        if (replyCounts.has(id) && replyCounts.get(id) !== count) fail('Inaccessible social target: conflicting source reply count');
                        replyCounts.set(id,count);
                        // Retain scalar public count evidence, never arbitrary
                        // source objects. Invalid scalar counts prove no extent.
                        if (previous && count !== null && !['number','string','boolean'].includes(typeof count)) fail('Inaccessible social target: invalid source reply count');
                        const signature = JSON.stringify([typeof count,typeof count === 'number' ? String(count) : ['string','boolean'].includes(typeof count) ? count : null]);
                        if (observedCounts.has(id) && observedCounts.get(id) !== signature) fail('Inaccessible social target: conflicting source reply count');
                        observedCounts.set(id,signature);
                        if (Number.isInteger(count) && count >= 0) fields.replyCount=count;
                    }
                    if (position) { fields.conversationStart=position.isStart; fields.conversationEnd=position.isEnd; }
                    const parents: (string | null)[] = [];
                    if (value.reply_to_results === null) parents.push(null);
                    else if (typeof value.reply_to_results?.rest_id === 'string') parents.push(value.reply_to_results.rest_id);
                    if (typeof value.legacy?.in_reply_to_status_id_str === 'string') parents.push(value.legacy.in_reply_to_status_id_str);
                    if (value.legacy?.conversation_id_str === id) parents.push(null);
                    if (normalized && typeof value.in_reply_to_status_id_str === 'string') parents.push(value.in_reply_to_status_id_str);
                    if (normalized && value.conversation_id_str === id) parents.push(null);
                    for (const parent of parents) {
                        if (fields.parentID !== undefined && fields.parentID !== parent) fail('Inaccessible social target: conflicting source parent relationship');
                        fields.parentID=parent;
                    }
                    evidence.set(id,fields);
                }
                // Props only; avoid React owner/DOM back references and functions.
                for (const [key, child] of Object.entries(value)) if (!['children','_owner','ref','quoted_status_result','quoted_status'].includes(key)) visit(child,depth+1);
            };
            for (let i=0; fiber && i<30; i++, fiber=fiber.return) {
                visit(fiber.memoizedProps || fiber.pendingProps);
                // useSyncExternalStore keeps public post query snapshots in hook
                // state; inspect only that named data and React's memo cache.
                let hook=fiber.memoizedState;
                for (let j=0; hook && j<20; j++, hook=hook.next) visit(hook.memoizedState?.snapshot?.data);
                visit(fiber.updateQueue?.memoCache?.data);
            }
            // Decide the owned body once after joining metadata fragments; a
            // later identity-only fragment cannot reassert flattened DOM text.
            for (const [id,fields] of evidence) if (fields.authorID) {
                const candidate = sourceBodies.get(id), source = candidate?.agreesRendered ? candidate : undefined, prior = posts.get(id);
                const retained = prior?.bodySource === 'native' && prior.bodyComplete
                    && !own.querySelector('[data-article-part], [data-testid="twitterArticleReadView"], [data-testid="twitterArticleRichTextView"]') ? prior : undefined;
                const body = source?.body ?? retained?.body ?? modernXBody(own,url);
                add({id,url:permalinks.get(id)!,body,bodySource:source || retained ? 'native' : undefined,
                    bodyComplete:(source?.complete ?? (candidate && !candidate.complete ? false : retained?.bodyComplete ?? substantive(body))) && !postPreview(own),...fields} as SocialPost);
            }
        }
        for (const post of posts.values()) {
            const count = post.replyCount;
            const children = [...posts.values()].filter(child => child.parentID === post.id);
            if (Number.isInteger(count) && (count as number) >= 0 && count === children.length) replyPages.push({rootID:post.id,end:true,cursor:null});
        }
    } else fail('Inaccessible social target: unsupported snapshot platform');
    const targets = [...posts.values()].filter(p => sameTarget(p.url,url));
    if (targets.length !== 1) fail('Inaccessible social target: source target identity unavailable or ambiguous');
    const counts = [...observedCounts].filter(([id]) => posts.has(id));
    if (isX && new TextEncoder().encode(JSON.stringify({posts:[...posts.values()],counts})).byteLength > 16000000) fail('Inaccessible social target: source snapshot limit');
    return {version:1, platform:host.startsWith('threads')?'threads':'x', targetID:targets[0].id, posts:[...posts.values()], sequences, branches, replyPages,
        ...(isX ? {observedReplyCounts:counts} : {})};
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
