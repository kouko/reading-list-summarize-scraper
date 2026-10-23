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
    clone.querySelectorAll('[data-testid="quoteTweet"], [data-testid="simpleTweet"], article[data-testid="tweet"]').forEach(el => el.remove());
    // Upstream TwitterExtractor identifies quoted cards by their labelled
    // container and nested author header. They cannot identify the outer post.
    clone.querySelectorAll('[aria-labelledby*="id__"]').forEach(el => {
        if (el.querySelector('[data-testid="User-Name"]')) el.remove();
    });
    return clone;
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
        if (richText && sameTarget(canonical, url)) text = bodyMarkdown(richText, url);
        for (const article of Array.from(document.querySelectorAll('article[data-testid="tweet"]'))) {
            if (text) break;
            const own = ownTweet(article);
            const permalinks = Array.from(own.querySelectorAll('a')).filter(a => a.querySelector('time') && !a.closest('[data-testid="tweetText"], [data-testid="twitterArticleRichTextView"]'));
            const target = new URL(url.origin + url.pathname.replace(/\/article\/?$/, ''));
            const ownsTarget = permalinks.length > 0 && permalinks.every(a => sameTarget(a.getAttribute('href'), target));
            if (!ownsTarget) continue;
            const body = own.querySelector('[data-testid="twitterArticleRichTextView"]') || own.querySelector('[data-testid="tweetText"]');
            if (body) text = bodyMarkdown(body, url);
            break;
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

export async function extractArticle(document: Document = globalThis.document, address: string = globalThis.location.href, fetcher: typeof fetch = globalThis.fetch): Promise<string> {
    const url = new URL(address), host = url.hostname.replace(/^www\./, '');
    if (['youtube.com', 'm.youtube.com', 'youtu.be'].includes(host)) return youtube(document, url, fetcher);
    if (['bilibili.com', 'm.bilibili.com'].includes(host)) fail('Unverified original language: Bilibili exposes subtitle language but no verified original-audio association');
    if (['vimeo.com', 'dailymotion.com', 'twitch.tv', 'tiktok.com', 'nicovideo.jp', 'niconico.jp', 'youku.com', 'douyin.com'].some(h => host === h || host.endsWith('.' + h)) || document.querySelector('meta[property="og:type"][content^="video"]')) fail('Unsupported video transcript extraction');
    if (['x.com', 'twitter.com', 'threads.net', 'threads.com', 'medium.com'].includes(host) || host.endsWith('.medium.com')) return social(document, url);
    const result = new Defuddle(document, { markdown: true }).parse();
    if (!result?.content?.trim()) fail('Ordinary article extraction failed: empty content');
    return result.content;
}
