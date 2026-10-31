// Public, already-delivered post data only. Never execute scripts or inspect account state.
export function facebook(document: Document, requested: URL): string {
    const fail = (kind: string, reason: string): never => { throw Error(`Article ${kind}: Facebook ${reason}`); };
    const url = (value: string) => { try { return new URL(value, requested.href); } catch { return null; } };
    const ownURL = (value: string) => {
        const parsed = url(value);
        return parsed?.protocol === 'https:' && !parsed.username && !parsed.password && !parsed.port
            && /^(?:www\.)?facebook\.com$/.test(parsed.hostname) ? parsed : null;
    };
    const identity = requested.pathname.match(/^\/([^/]+)\/posts\/(?:[^/]+\/)?(\d+)\/?$/);
    if (requested.protocol !== 'https:' || !identity) fail('target unverified', 'unsupported requested identity');
    const [, author, postID] = identity!;
    const key = (value: string) => ownURL(value)?.pathname.replace(/\/$/, '');
    const asset = (value: string) => {
        const parsed = url(value);
        return parsed?.protocol === 'https:' && /(?:^|\.)fbcdn\.net$/.test(parsed.hostname) ? parsed.pathname : null;
    };
    const records: any[] = [];
    let nodes = 0, bytes = 0;
    const visit = (value: any, depth = 0) => {
        if (!value || typeof value !== 'object') return;
        if (++nodes > 350000 || depth > 70) fail('incomplete', 'public data limit exceeded');
        if (String(value.post_id) === postID) records.push(value);
        for (const [name, child] of Object.entries(value)) {
            if (!/(?:token|cookie|password|credential|session|viewer|account|feedback|tracking)/i.test(name)) visit(child, depth + 1);
        }
    };
    for (const script of Array.from(document.querySelectorAll('script[type="application/json"]'))) {
        const source = script.textContent || '';
        bytes += source.length;
        if (bytes > 20_000_000) fail('incomplete', 'public data limit exceeded');
        let parsed: any;
        try { parsed = JSON.parse(source); } catch { continue; }
        visit(parsed);
    }
    if (!records.length) fail('target unverified', 'requested post absent');
    const candidates = records.filter(record => record.comet_sections?.content?.story?.message);
    if (!candidates.length) fail('body unavailable', 'original message absent');
    let agreed = '';
    let result = '';
    let permalinkAlias = '';
    const verifyPermalink = (value: string) => {
        const path = ownURL(value)?.pathname.match(/^\/([^/]+)\/posts\/(?:[^/]+\/)?(\d+|pfbid[^/]+)\/?$/);
        if (!path || path[1] !== author || (/^\d+$/.test(path[2]) && path[2] !== postID)) fail('target unverified', 'conflicting public permalink');
        if (path![2].startsWith('pfbid')) {
            if (permalinkAlias && permalinkAlias !== path![2]) fail('target unverified', 'conflicting public permalink');
            permalinkAlias = path![2];
        }
    };
    for (const record of candidates) {
        const story = record.comet_sections.content.story;
        const actors = record.actors;
        if (!record.id || story.id !== record.id || !Array.isArray(actors) || actors.length !== 1 || !actors[0].id
            || key(actors[0].url) !== `/${author}` || !key(record.permalink_url)) fail('target unverified', 'author or story identity');
        const text = story.message.text;
        if (typeof text !== 'string' || !text.trim()) fail('body unavailable', 'empty original message');
        for (const source of [...records, story]) {
            for (const field of ['permalink_url', 'url']) {
                if (source[field] !== undefined) verifyPermalink(source[field]);
            }
            if ((source.message?.text !== undefined && source.message.text !== text)
                || (source.id && source.id !== record.id)
                || source.is_truncated || source.message?.is_truncated) fail('incomplete', 'conflicting target sources');
            if (source.actors !== undefined && (!Array.isArray(source.actors) || source.actors.length !== 1 || !source.actors[0]
                || (source.actors[0].id !== undefined && source.actors[0].id !== actors[0].id)
                || (source.actors[0].url !== undefined && key(source.actors[0].url) !== key(actors[0].url)))) fail('target unverified', 'conflicting public author');
        }
        if (story.message.is_truncated || story.is_truncated || record.is_truncated) fail('incomplete', 'truncated original message');
        const messages = Array.from(document.querySelectorAll('[data-ad-preview="message"], [data-ad-comet-preview="message"], [data-testid="post_message"]'));
        const owners = messages.map(message => message.closest('[role="article"], article')).filter((owner): owner is Element => !!owner && !owner.parentElement?.closest('[role="article"], article'));
        const matching = [...new Set(owners)].filter(owner => Array.from(owner.querySelectorAll('a[href]')).some(link => !link.parentElement?.closest('[data-ad-preview="message"], [role="article"] [role="article"]') && key(link.getAttribute('href') || '') === key(record.permalink_url)));
        if (matching.length !== 1) fail('target unverified', 'ambiguous owned article');
        const owner = matching[0].cloneNode(true) as Element;
        owner.querySelectorAll('[role="article"], article').forEach(el => el.remove());
        const bodies = owner.querySelectorAll('[data-ad-preview="message"], [data-ad-comet-preview="message"], [data-testid="post_message"]');
        if (bodies.length !== 1) fail('target unverified', 'ambiguous owned message');
        const body = bodies[0];
        if (Array.from(body.querySelectorAll('button, [role="button"], a')).some(el => /^(?:see|show|read) more$|^顯示更多$|^查看更多$/i.test((el.textContent || '').trim()))) fail('incomplete', 'message preview');
        const plain = (node: Node): string => {
            if (node.nodeType === 3) return node.textContent || '';
            const el = node as Element;
            if (el.tagName === 'IMG') return /\/images\/emoji.php\//.test(el.getAttribute('src') || '') ? el.getAttribute('alt') || '' : '';
            if (el.tagName === 'BR') return '\n';
            return Array.from(node.childNodes).map(plain).join('') + (/^(?:DIV|P)$/.test(el.tagName) ? '\n' : '');
        };
        const normalize = (value: string) => value.replace(/\s+/gu, '');
        if (normalize(plain(body)) !== normalize(text)) fail('incomplete', 'source and owned message conflict');
        let rendered = '', offset = 0;
        for (const link of Array.from(body.querySelectorAll('a[href]'))) {
            const label = plain(link).trim(), destination = url(link.getAttribute('href') || '');
            const start = text.indexOf(label, offset);
            if (!label || start < 0 || !destination || !/^https?:$/.test(destination.protocol)) fail('incomplete', 'body link correspondence');
            rendered += text.slice(offset, start) + `[${label}](${destination!.href})`;
            offset = start + label.length;
        }
        rendered += text.slice(offset);
        const attachments = record.attachments, contentAttachments = story.attachments;
        if (!Array.isArray(attachments) || !Array.isArray(contentAttachments) || attachments.length !== contentAttachments.length) fail('incomplete', 'attachment extent unavailable');
        const photos: string[] = [], mediaIDs: string[] = [];
        for (let i = 0; i < attachments.length; i++) {
            const attachment = attachments[i], content = contentAttachments[i];
            const media = attachment.styles?.attachment?.media;
            const other = content.styles?.attachment?.media || content.media;
            if (!media || media.__typename !== 'Photo' || !media.id || other?.id !== media.id || media.creation_story?.id !== record.id
                || !Array.isArray(content.all_subattachments?.nodes) || content.all_subattachments.nodes.length
                || content.all_subattachments.page_info?.has_next_page || attachment.all_subattachments?.nodes?.length
                || attachment.all_subattachments?.page_info?.has_next_page
                || !asset(media.photo_image?.uri) || !(media.photo_image.width > 0 && media.photo_image.height > 0)
                || ownURL(media.url)?.searchParams.get('fbid') !== media.id) fail('incomplete', 'unverified attachment extent');
            for (const copy of [attachment.media, content.media, content.styles?.attachment?.media].filter(Boolean)) {
                if ((copy.id !== undefined && copy.id !== media.id)
                    || (copy.__typename !== undefined && copy.__typename !== 'Photo')
                    || (copy.photo_image?.uri !== undefined && asset(copy.photo_image.uri) !== asset(media.photo_image.uri))
                    || (copy.creation_story?.id !== undefined && copy.creation_story.id !== record.id)
                    || (copy.url !== undefined && ownURL(copy.url)?.searchParams.get('fbid') !== media.id)) fail('incomplete', 'conflicting attachment sources');
            }
            const images = Array.from(owner.querySelectorAll('a[href] img')).filter(image => {
                const anchor = image.closest('a[href]');
                return ownURL(anchor?.getAttribute('href') || '')?.searchParams.get('fbid') === media.id && asset(image.getAttribute('src') || '') === asset(media.photo_image.uri);
            });
            if (images.length !== 1 || mediaIDs.includes(media.id)) fail('incomplete', 'owned photo missing or ambiguous');
            mediaIDs.push(media.id);
            photos.push(`![${(images[0].getAttribute('alt') || '').replace(/[\[\]\n]/g, '')}](${media.photo_image.uri})`);
        }
        const ownedPhotoIDs = Array.from(owner.querySelectorAll('a[href]')).map(link => ownURL(link.getAttribute('href') || '')?.searchParams.get('fbid')).filter(Boolean);
        if (ownedPhotoIDs.some(id => !mediaIDs.includes(id!))) fail('incomplete', 'unlisted owned photo');
        const signature = JSON.stringify([actors[0].id, text, mediaIDs, photos]);
        if (agreed && agreed !== signature) fail('incomplete', 'conflicting target sources');
        agreed = signature;
        result = rendered + (photos.length ? '\n\n' + photos.join('\n\n') : '');
    }
    return result;
}
