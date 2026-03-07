import { useAppStore } from '@/lib/store';
import zhCN from './zh-CN';
import en from './en';
import type { Dictionary } from './zh-CN';

const dictionaries: Record<string, Dictionary> = {
  'zh-CN': zhCN,
  en,
};

export function useTranslation(): Dictionary {
  const lang = useAppStore((s) => s.settings.language);
  return dictionaries[lang] ?? zhCN;
}

export type { Dictionary };
