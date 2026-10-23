import { extractArticle } from './extraction';

(window as any).extractArticle = (): Promise<string> => extractArticle();
