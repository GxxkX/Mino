'use client';

import { cn } from '@/lib/utils';
import { formatDate } from '@/lib/utils';
import type { ChatMessage } from '@/types';
import { User, Quote } from 'lucide-react';
import { useTranslation } from '@/lib/i18n';

interface ChatMessageProps {
  message: ChatMessage;
}

export function ChatMessageComponent({ message }: ChatMessageProps) {
  const isUser = message.role === 'user';
  const t = useTranslation();

  return (
    <div className={cn('flex gap-3 max-w-2xl', isUser ? 'ml-auto flex-row-reverse' : '')}>
      <div className={cn(
        'w-7 h-7 rounded-full flex items-center justify-center flex-shrink-0 mt-0.5',
        isUser ? 'bg-surface-hover' : 'bg-cta/10'
      )}>
        {isUser ? (
          <User className="w-3.5 h-3.5 text-text-muted" />
        ) : (
          <span className="text-xs font-semibold text-cta">M</span>
        )}
      </div>
      
      <div className={cn('max-w-[80%]', isUser ? 'text-right' : 'text-left')}>
        <div className={cn(
          'inline-block rounded-lg px-3.5 py-2.5 text-sm leading-relaxed text-left',
          isUser 
            ? 'bg-surface-hover text-text' 
            : 'bg-surface border border-border text-text-secondary'
        )}>
          <p className="whitespace-pre-wrap">{message.content}</p>
        </div>
        
        <p className="text-[11px] text-text-muted mt-1.5 px-1">
          {formatDate(message.createdAt)}
        </p>

        {!isUser && message.sources && message.sources.length > 0 && (
          <div className="mt-2 px-1 space-y-1">
            <p className="text-[11px] text-text-muted">{t.chat.sources}</p>
            {message.sources.map((source, idx) => (
              <div key={idx} className="flex items-start gap-1.5 text-[11px] text-text-muted hover:text-text-secondary cursor-pointer transition-colors">
                <Quote className="w-3 h-3 mt-0.5 flex-shrink-0" />
                <span>{source.title}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
