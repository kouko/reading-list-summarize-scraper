import Defuddle from 'defuddle';

(window as any).extractArticle = async (): Promise<string> => {
    try {
        const df = new Defuddle(document);
        const result = df.parse();
        return result?.contentMarkdown ?? result?.content ?? "";
    } catch (e: any) {
        console.error("defuddle extraction error:", e);
        return "";
    }
};
