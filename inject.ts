import { extractArticle, socialSnapshot } from './extraction';

(window as any).extractArticle = (): Promise<string> => extractArticle();
(window as any).socialSnapshot = () => socialSnapshot();
