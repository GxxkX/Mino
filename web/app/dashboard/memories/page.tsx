'use client';

import { useEffect, useState } from 'react';
import { Header } from '@/components/layout/header';
import { MemoryCard } from '@/components/features/memory-card';
import { useAppStore } from '@/lib/store';
import { useTranslation } from '@/lib/i18n';
import { memoriesApi } from '@/lib/api';
import { Input } from '@/components/ui/input';
import { Search } from 'lucide-react';

type Category = 'all' | 'insight' | 'fact' | 'preference' | 'event';

export default function MemoriesPage() {
  const { memories, setMemories } = useAppStore();
  const t = useTranslation();
  const [search, setSearch] = useState('');
  const [category, setCategory] = useState<Category>('all');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    memoriesApi.listMemories(1, 100)
      .then(res => setMemories(res.data ?? []))
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [setMemories]);

  const filtered = memories.filter(m => {
    if (category !== 'all' && m.category !== category) return false;
    if (search && !m.content.toLowerCase().includes(search.toLowerCase())) return false;
    return true;
  });

  const categories: { value: Category; label: string }[] = [
    { value: 'all', label: t.memories.all },
    { value: 'insight', label: t.memories.insight },
    { value: 'fact', label: t.memories.fact },
    { value: 'preference', label: t.memories.preference },
    { value: 'event', label: t.memories.event },
  ];

  return (
    <div className="min-h-screen">
      <Header title={t.memories.title} description={t.memories.countDescription(memories.length)} />

      <div className="px-8 pb-8 pt-5 space-y-5">
        <div className="flex items-center gap-4">
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-text-muted" />
            <Input
              placeholder={t.memories.searchPlaceholder}
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="pl-9"
            />
          </div>
        </div>

        <div className="flex gap-1">
          {categories.map((cat) => (
            <button
              key={cat.value}
              onClick={() => setCategory(cat.value)}
              className={`px-3 py-1.5 rounded-md text-xs font-medium transition-colors duration-150 cursor-pointer ${
                category === cat.value
                  ? 'bg-surface-hover text-text'
                  : 'text-text-muted hover:text-text-secondary'
              }`}
            >
              {cat.label}
            </button>
          ))}
        </div>

        <div>
          {loading ? (
            <div className="text-center py-16 text-text-muted text-sm">{t.loading}</div>
          ) : filtered.length > 0 ? (
            filtered.map((memory) => (
              <MemoryCard key={memory.id} memory={memory} />
            ))
          ) : (
            <div className="text-center py-16 text-text-muted text-sm">
              {t.memories.empty}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
