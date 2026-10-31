import { extractArticle, socialSnapshot } from './extraction';

(window as any).extractArticle = async (address = globalThis.location.href): Promise<string> => {
    const requested = new URL(address), delivered = new URL(globalThis.location.href);
    const host = (url: URL) => url.hostname.replace(/^www\./, '');
    try {
        if (requested.protocol === 'https:' && ['facebook.com', 'felixrieseberg.com'].includes(host(requested))) {
            const reject = (): never => { throw Error('Article target unverified: requested document was not delivered'); };
            if (delivered.protocol !== requested.protocol || host(delivered) !== host(requested)
                || delivered.port !== requested.port || requested.username || requested.password || delivered.username || delivered.password) reject();
            if (host(requested) === 'facebook.com') {
                const numeric = (url: URL) => url.pathname.match(/^\/([^/]+)\/posts\/(?:[^/]+\/)?(\d+)\/?$/);
                const target = numeric(requested), current = numeric(delivered);
                if (!target) reject();
                if (!current || current[1] !== target![1] || current[2] !== target![2]) {
                    const aliasPath = delivered.pathname.match(/^\/([^/]+)\/posts\/(pfbid[A-Za-z0-9]+)\/?$/);
                    if (!aliasPath || aliasPath[1] !== target![1]) reject();
                    // A pfbid alias needs an explicit public record joining it to the
                    // requested numeric ID. The adapter still verifies body/media.
                    let matched = false, nodes = 0, bytes = 0;
                    const visit = (value: any, depth = 0) => {
                        if (!value || typeof value !== 'object') return;
                        if (++nodes > 350000 || depth > 70) reject();
                        if (String(value.post_id) === target![2] && typeof value.permalink_url === 'string') {
                            let alias: URL;
                            try { alias = new URL(value.permalink_url, requested); } catch { reject(); }
                            if (alias.protocol === delivered.protocol && host(alias) === host(delivered) && !alias.username && !alias.password
                                && alias.port === delivered.port && alias.pathname.replace(/\/$/, '') === delivered.pathname.replace(/\/$/, '')) matched = true;
                        }
                        for (const [name, child] of Object.entries(value)) {
                            if (!/(?:token|cookie|password|credential|session|viewer|account|feedback|tracking)/i.test(name)) visit(child, depth + 1);
                        }
                    };
                    for (const script of document.querySelectorAll('script[type="application/json"]')) {
                        const text = script.textContent || '';
                        if ((bytes += text.length) > 20_000_000) reject();
                        let value: any;
                        try { value = JSON.parse(text); } catch { continue; }
                        visit(value);
                    }
                    if (!matched) reject();
                }
            } else if (delivered.pathname.replace(/\/$/, '') !== requested.pathname.replace(/\/$/, '') || delivered.search !== requested.search) reject();
        }
    } catch (error) {
        // Delivery can fail before the adapter's challenge classifier runs.
        // Use the same explicit UI evidence so blocked retry policy survives.
        const title = document.title.trim();
        let challenge = '';
        if (/^(?:attention required!?|just a moment(?:\.\.\.)?|security verification)\s*[|\-–]\s*cloudflare$/i.test(title)
            || document.querySelector('#cf-challenge-running, #cf-challenge-error')) challenge = 'Cloudflare security challenge';
        else if (/^(?:verify (?:you are|you're) (?:a )?human|captcha(?: security challenge)?)\s*[.!]?$/i.test(title)
            && document.querySelector('form#captcha, iframe[src*="hcaptcha.com/"], iframe[src*="recaptcha/"]')) challenge = 'CAPTCHA security challenge';
        if (challenge) throw Error(`${challenge}: ${error instanceof Error ? error.message : 'delivery failed'}`);
        throw error;
    }
    return extractArticle(document, address);
};
let previous: Parameters<typeof socialSnapshot>[3];
(window as any).socialSnapshot = (pages = []) => {
    const document = globalThis.document, address = globalThis.location.href;
    if (!previous || previous.document !== document || previous.address !== address) previous = {document,address,posts:[],counts:[]};
    const snapshot = socialSnapshot(document,address,pages,previous);
    previous = snapshot.platform === 'x' ? {document,address,posts:snapshot.posts.map(post => ({...post})),
        counts:(snapshot.observedReplyCounts || []).map(([id,count]): [string,string] => [id,count])} : undefined;
    return snapshot;
};
