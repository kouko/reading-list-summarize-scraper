import Defuddle from 'defuddle/full';

(window as any).extractArticle = async (): Promise<string> => {
    try {
        const df = new Defuddle(document, { markdown: true });
        const result = df.parse();
        return result?.content ?? "";
    } catch (e: any) {
        console.error("defuddle extraction error:", e);
        return "";
    }
};
