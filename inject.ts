import { extractArticle, socialSnapshot } from './extraction';

(window as any).extractArticle = (): Promise<string> => extractArticle();
let previous: Parameters<typeof socialSnapshot>[3];
(window as any).socialSnapshot = (pages = []) => {
    const document = globalThis.document, address = globalThis.location.href;
    if (!previous || previous.document !== document || previous.address !== address) previous = {document,address,posts:[],counts:[]};
    const snapshot = socialSnapshot(document,address,pages,previous);
    previous = snapshot.platform === 'x' ? {document,address,posts:snapshot.posts.map(post => ({...post})),
        counts:(snapshot.observedReplyCounts || []).map(([id,count]): [string,string] => [id,count])} : undefined;
    return snapshot;
};
