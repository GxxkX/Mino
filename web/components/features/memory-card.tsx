'use client';

import { useState } from 'react';
import { Badge } from '@/components/ui/badge';
import { formatDate } from '@/lib/utils';
import type { Memory } from '@/types';
import { Lightbulb, BookOpen, Heart, CalendarDays, Play, Pause } from 'lucide-react';
import { demoConversations } from '@/lib/demo-data';
import { useTranslation } from '@/lib/i18n';

interface MemoryCardProps {
  memory: Memory;
}

const categoryIcons = {
  insight: Lightbulb,
  fact: BookOpen,
  preference: Heart,
  event: CalendarDays,
};

const categoryVariants = {
  insight: 'lime' as const,
  fact: 'blue' as const,
  preference: 'amber' as const,
  event: 'rose' as const,
};

export function MemoryCard({ memory }: MemoryCardProps) {
  const [playing, setPlaying] = useState(false);
  const t = useTranslation();

  const categoryLabels: Record<string, string> = {
    insight: t.memories.insight,
    fact: t.memories.fact,
    preference: t.memories.preference,
    event: t.memories.event,
  };

  const Icon = categoryIcons[memory.category];
  const config = { label: categoryLabels[memory.category], variant: categoryVariants[memory.category] };

  const sourceConversation = memory.conversationId
    ? demoConversations.find(c => c.id === memory.conversationId)
    : null;

  return (
    <div className="group flex items-start gap-3 py-3 px-1 border-b border-border-subtle last:border-0 cursor-pointer hover:bg-surface/50 transition-colors duration-150 -mx-1 rounded">
      <div className="mt-0.5 w-7 h-7 rounded-md bg-surface-hover flex items-center justify-center flex-shrink-0">
        <Icon className="w-3.5 h-3.5 text-text-muted" />
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-1">
          <Badge variant={config.variant}>{config.label}</Badge>
          <span className="text-[11px] text-text-muted">{formatDate(memory.createdAt)}</span>
        </div>
        <p className="text-sm text-text-secondary leading-relaxed">{memory.content}</p>

        {sourceConversation?.audioDuration && (
          <button
            onClick={(e) => { e.stopPropagation(); setPlaying(p => !p); }}
            className={`mt-2 flex items-center gap-1.5 transition-all duration-150 ${
              playing
                ? 'text-cta'
                : 'text-text-muted hover:text-cta'
            }`}
          >
            <span className={`w-5 h-5 rounded-full flex items-center justify-center flex-shrink-0 transition-all duration-150 ${
              playing
                ? 'bg-cta text-bg shadow-[0_0_8px_rgba(163,230,53,0.35)]'
                : 'bg-surface-hover border border-border hover:border-cta/40'
            }`}>
              {playing
                ? <Pause className="w-2.5 h-2.5 fill-current" />
                : <Play className="w-2.5 h-2.5 fill-current ml-px" />
              }
            </span>
            <span className="text-[11px]">{sourceConversation.title}</span>
          </button>
        )}
      </div>
      <div className="flex items-center gap-px mt-1.5 flex-shrink-0">
        {Array.from({ length: 5 }).map((_, i) => (
          <div
            key={i}
            className={`w-1 h-2.5 rounded-full ${
              i < Math.ceil(memory.importance / 2) ? 'bg-cta/60' : 'bg-border'
            }`}
          />
        ))}
      </div>
    </div>
  );
}
