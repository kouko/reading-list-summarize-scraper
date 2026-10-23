import Defuddle from 'defuddle/full';

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

function transcript(document: Document, xml: string): string {
    // Reject partial XML rather than letting the HTML parser repair a cut response.
    if (!/^\s*(?:<\?xml[^>]*>\s*)?<(transcript|timedtext)\b[^>]*>[\s\S]*<\/\1>\s*$/.test(xml)
        || /(?:truncated|incomplete)\s*=\s*["'](?:true|1)["']/i.test(xml)) fail('Unusable subtitles: incomplete payload');
    const holder = document.createElement('div');
    holder.innerHTML = xml;
    const cues = Array.from(holder.querySelectorAll('text, p'));
    const opening = (xml.match(/<(?:text|p)\b/g) || []).length;
    const closing = (xml.match(/<\/(?:text|p)>/g) || []).length;
    if (opening !== closing || cues.length !== opening) fail('Unusable subtitles: malformed cues');
    let previous = -1;
    const lines: string[] = [];
    for (const cue of cues) {
        const startKey = cue.tagName.toLowerCase() === 'p' ? 't' : 'start';
        const durationKey = startKey === 't' ? 'd' : 'dur';
        const rawStart = cue.getAttribute(startKey), rawDuration = cue.getAttribute(durationKey);
        const start = rawStart === null || rawStart.trim() === '' ? NaN : Number(rawStart);
        const duration = rawDuration === null || rawDuration.trim() === '' ? NaN : Number(rawDuration);
        if (!Number.isFinite(start) || !Number.isFinite(duration) || start < 0 || duration <= 0 || start < previous) fail('Unusable subtitles: invalid cue timing');
        previous = start;
        const text = (cue.textContent || '').replace(/\s+/g, ' ').trim();
        if (substantive(text)) lines.push(text);
    }
    if (!lines.length) fail('Unusable subtitles: empty or timestamp-only cues');
    return lines.join('\n\n');
}

async function youtube(document: Document, url: URL, fetcher: typeof fetch): Promise<string> {
    const id = url.hostname === 'youtu.be' ? url.pathname.slice(1) : url.searchParams.get('v') || url.pathname.split('/')[2];
    if (!id) fail('Unavailable subtitles: missing video identity');
    const data = playerData(document);
    if (data?.videoDetails?.videoId !== id) fail('Unavailable subtitles: wrong target player');
    const renderer = data?.captions?.playerCaptionsTracklistRenderer;
    const tracks = renderer?.captionTracks;
    if (!Array.isArray(tracks) || !tracks.length) fail('Unavailable subtitles');
    // The default audio/caption index is a user preference, not original-audio evidence.
    const originals = (renderer.audioTracks || []).filter((a: any) => {
        const label = typeof a.displayName === 'string' ? a.displayName : a.displayName?.simpleText || a.displayName?.runs?.map((r: any) => r.text).join('');
        return /\(original\)/i.test(label || '');
    });
    if (originals.length !== 1) fail('Unverified original language: no unambiguous original audio association');
    const audio = originals[0];
    const associated = (audio.captionTrackIndices || []).map((i: number) => tracks[i]).filter(Boolean);
    const asr = associated.filter((t: any) => t.kind === 'asr' && typeof t.languageCode === 'string');
    const languages = new Set(asr.map((t: any) => family(t.languageCode)));
    if (languages.size !== 1) fail('Unverified original language: missing or conflicting original-audio automatic captions');
    const originalLanguage = asr[0].languageCode;
    // A caption explicitly translated by URL is never eligible.
    const candidates = associated.filter((t: any) => typeof t.languageCode === 'string' && family(t.languageCode) === family(originalLanguage));
    const errors: string[] = [];
    for (const track of candidates) {
        try {
            const endpoint = new URL(track.baseUrl);
            if (endpoint.protocol !== 'https:' || !['www.youtube.com', 'youtube.com'].includes(endpoint.hostname)
                || endpoint.pathname !== '/api/timedtext' || endpoint.searchParams.has('tlang') || endpoint.searchParams.get('v') !== id) fail('Unusable subtitles: disallowed or translated caption endpoint');
            endpoint.searchParams.set('fmt', 'srv3');
            const response = await fetcher(endpoint.href, {credentials: 'include', redirect: 'error', signal: AbortSignal.timeout(10000)});
            if (!response.ok) fail(`Unusable subtitles: HTTP ${response.status}`);
            const body = transcript(document, await response.text());
            const source = track.kind === 'asr' ? 'automatic' : 'human';
            const provenance = {platform: 'youtube', target: id, language: track.languageCode, source, evidence: 'original-audio-asr-association'};
            return `<!-- rlss-video-v1 ${JSON.stringify(provenance)} -->\n\nSubtitle language: ${track.languageCode}\nSubtitle source: ${source}\n\n${body}`;
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

function social(document: Document, url: URL): string {
    const host = url.hostname.replace(/^www\./, '');
    let text = '';
    if (host === 'x.com' || host === 'twitter.com') {
        const id = url.pathname.match(/^\/[^/]+\/status\/(\d+)(?:\/article)?\/?$/)?.[1];
        if (!id) fail('Inaccessible social target: invalid post URL');
        for (const article of Array.from(document.querySelectorAll('article[data-testid="tweet"]'))) {
            const ownsTarget = Array.from(article.querySelectorAll('a')).some(a => a.querySelector('time') && sameTarget(a.getAttribute('href'), new URL(url.origin + url.pathname.replace(/\/article\/?$/, ''))));
            if (!ownsTarget) continue;
            text = (article.querySelector('[data-testid="twitterArticleRichTextView"]') || article.querySelector('[data-testid="tweetText"]'))?.textContent || '';
            break;
        }
    } else if (host === 'threads.net' || host === 'threads.com') {
        if (!/^\/@[^/]+\/post\/[^/]+\/?$/.test(url.pathname)) fail('Inaccessible social target: invalid post URL');
        for (const script of Array.from(document.querySelectorAll('script[type="application/ld+json"]'))) {
            let parsed;
            try { parsed = JSON.parse(script.textContent || ''); } catch { continue; }
            const objects = Array.isArray(parsed) ? parsed : parsed['@graph'] || [parsed];
            for (const item of objects) if (item['@type'] === 'SocialMediaPosting' && sameTarget(item.url, url) && typeof item.articleBody === 'string') text = item.articleBody;
        }
    } else {
        if (!sameTarget(document.querySelector('link[rel="canonical"]')?.getAttribute('href') || null, url)) fail('Inaccessible social target: canonical identity missing or different');
        if (document.querySelector('[data-testid="paywall"], [data-testid="paywall-content"], [data-member-only="true"]')
            || /member.only story|read the rest.*member|sign up to read|purchase.*membership/i.test(document.body.textContent || '')
            || Array.from(document.querySelectorAll('script[type="application/ld+json"]')).some(s => /"isAccessibleForFree"\s*:\s*(?:false|"false")/.test(s.textContent || ''))) fail('Inaccessible social target: subscription or preview');
        const article = document.querySelector('article');
        if (article) text = Array.from(article.querySelectorAll('h1, h2, h3, p, blockquote, pre, li')).filter(el => !el.closest('nav, aside, footer')).map(el => el.textContent?.trim()).filter(Boolean).join('\n\n');
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
