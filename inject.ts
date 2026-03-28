import { Defuddle } from 'defuddle';

(window as any).extractArticle = async (): Promise<string> => {
    try {
        const html = document.documentElement.outerHTML;
        const df = new Defuddle(html);
        const result = await df.parse();
        return result?.content ?? "";
    } catch (e: any) {
        return "";
    }
};
