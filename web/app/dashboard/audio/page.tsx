'use client';

import { useEffect, useState } from 'react';
import { Header } from '@/components/layout/header';
import { ConversationCard } from '@/components/features/conversation-card';
import { useAppStore } from '@/lib/store';
import { useTranslation } from '@/lib/i18n';
import { conversationsApi } from '@/lib/api';
import { Input } from '@/components/ui/input';
import { Search, Mic } from 'lucide-react';

export default function AudioPage() {
  const { conversations, setConversations } = useAppStore();
  const t = useTranslation();
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    conversationsApi.listConversations(1, 50)
      .then(res => setConversations(res.data ?? []))
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [setConversations]);

  const filtered = search
    ? conversations.filter(c =>
        c.title?.toLowerCase().includes(search.toLowerCase()) ||
        c.summary?.toLowerCase().includes(search.toLowerCase())
      )
    : conversations;

  const totalDuration = conversations.reduce((acc, c) => acc + (c.audioDuration || 0), 0);
  const totalMinutes = Math.floor(totalDuration / 60);

  return (
    <div className="min-h-screen">
      <Header title={t.audio.title} description={t.audio.description(conversations.length, totalMinutes)} />

      <div className="px-8 pb-8 space-y-5">
        <div className="flex items-center gap-6 py-3 border-b border-border">
          <div className="flex items-baseline gap-1.5">
            <span className="text-lg font-semibold text-text tabular-nums">{conversations.length}</span>
            <span className="text-xs text-text-muted">{t.audio.totalRecordings}</span>
          </div>
          <div className="w-px h-5 bg-border" />
          <div className="flex items-baseline gap-1.5">
            <span className="text-lg font-medium text-text tabular-nums">{totalMinutes}</span>
            <span className="text-xs text-text-muted">{t.audio.minutes}</span>
          </div>
          <div className="flex items-baseline gap-1.5">
            <span className="text-lg font-medium text-cta tabular-nums">
              {conversations.filter(c => c.status === 'completed').length}
            </span>
            <span className="text-xs text-text-muted">{t.audio.processed}</span>
          </div>
        </div>

        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-text-muted" />
          <Input
            placeholder={t.audio.searchPlaceholder}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9"
          />
        </div>

        <div>
          {loading ? (
            <div className="text-center py-16 text-text-muted text-sm">{t.loading}</div>
          ) : filtered.length > 0 ? (
            filtered.map((conversation) => (
              <ConversationCard key={conversation.id} conversation={conversation} />
            ))
          ) : (
            <div className="text-center py-16 text-text-muted text-sm">
              <Mic className="w-8 h-8 mx-auto mb-3 opacity-30" />
              <p>{t.audio.empty}</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
