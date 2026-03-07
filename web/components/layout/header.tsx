'use client';

import { useState, useEffect, useRef, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { Search, Mic, User, X, FileText, Brain, Loader2, ArrowLeft } from 'lucide-react';
import { useAppStore } from '@/lib/store';
import { useAudioRecorder } from '@/hooks/use-audio-recorder';
import { cn } from '@/lib/utils';
import { useTranslation } from '@/lib/i18n';
import { RecordingBanner } from '@/components/features/recording-banner';
import { searchApi } from '@/lib/api';
import type { SearchResponse, SearchResultItem } from '@/types';

interface HeaderProps {
  title: string;
  description?: string;
  backHref?: string;
}

export function Header({ title, description, backHref }: HeaderProps) {
  const { user, isRecording } = useAppStore();
  const recorder = useAudioRecorder();
  const router = useRouter();
  const t = useTranslation();

  const [query, setQuery] = useState('');
  const [results, setResults] = useState<SearchResponse | null>(null);
  const [isOpen, setIsOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout>>();

  // Debounced search
  const doSearch = useCallback(async (q: string) => {
    if (q.trim().length < 2) {
      setResults(null);
      setIsOpen(false);
      return;
    }
    setIsLoading(true);
    try {
      const res = await searchApi.search(q, 5);
      setResults(res);
      setIsOpen(true);
    } catch {
      setResults(null);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => doSearch(query), 300);
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
  }, [query, doSearch]);

  // Close dropdown on outside click
  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, []);

  // Keyboard shortcut: Cmd/Ctrl+K to focus search
  useEffect(() => {
    function handleKey(e: KeyboardEvent) {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        inputRef.current?.focus();
      }
      if (e.key === 'Escape') {
        setIsOpen(false);
        inputRef.current?.blur();
      }
    }
    document.addEventListener('keydown', handleKey);
    return () => document.removeEventListener('keydown', handleKey);
  }, []);

  function handleSelect(item: SearchResultItem) {
    setIsOpen(false);
    setQuery('');
    if (item.type === 'conversation') {
      router.push(`/dashboard/audio`);
    } else {
      router.push(`/dashboard/memories`);
    }
  }

  const hasResults = results && (results.conversations?.length > 0 || results.memories?.length > 0);

  return (
    <>
      <header className="flex items-center justify-between px-8 py-6">
        <div className="flex items-center gap-3">
          {backHref && (
            <button
              onClick={() => router.push(backHref)}
              className="w-8 h-8 rounded-lg flex items-center justify-center text-text-muted hover:text-text hover:bg-surface-hover transition-colors duration-200 cursor-pointer focus:outline-none focus:ring-2 focus:ring-cta/40"
              aria-label={t.header.back}
            >
              <ArrowLeft className="w-4 h-4" />
            </button>
          )}
          <div>
            <h1 className="text-xl font-semibold text-text tracking-tight">{title}</h1>
            {description && (
              <p className="text-sm text-text-muted mt-0.5">{description}</p>
            )}
          </div>
        </div>

        <div className="flex items-center gap-3">
          {/* Search bar */}
          <div ref={containerRef} className="relative">
            <div className="relative flex items-center">
              <Search className="absolute left-2.5 w-3.5 h-3.5 text-text-muted pointer-events-none" />
              <input
                ref={inputRef}
                type="text"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                onFocus={() => { if (results && query.trim().length >= 2) setIsOpen(true); }}
                placeholder={t.header.search}
                className="w-48 pl-8 pr-7 py-1.5 rounded-md bg-surface text-xs text-text placeholder:text-text-muted border border-transparent focus:border-border focus:bg-surface-hover outline-none transition-all duration-150"
              />
              {query && (
                <button
                  onClick={() => { setQuery(''); setResults(null); setIsOpen(false); }}
                  className="absolute right-2 text-text-muted hover:text-text cursor-pointer"
                >
                  <X className="w-3 h-3" />
                </button>
              )}
              {isLoading && (
                <Loader2 className="absolute right-2 w-3 h-3 text-text-muted animate-spin" />
              )}
            </div>

            {/* Dropdown results */}
            {isOpen && (
              <div className="absolute right-0 top-full mt-1.5 w-80 max-h-96 overflow-y-auto rounded-lg bg-surface border border-border shadow-lg z-50">
                {hasResults ? (
                  <div className="py-1.5">
                    {results!.conversations?.length > 0 && (
                      <div>
                        <p className="section-label px-3 py-1.5 text-[10px] uppercase tracking-wider text-text-muted">
                          {t.header.conversations}
                        </p>
                        {results!.conversations.map((item) => (
                          <button
                            key={item.id}
                            onClick={() => handleSelect(item)}
                            className="w-full flex items-start gap-2.5 px-3 py-2 hover:bg-surface-hover transition-colors text-left cursor-pointer"
                          >
                            <FileText className="w-3.5 h-3.5 mt-0.5 text-accent-blue shrink-0" />
                            <div className="min-w-0">
                              <p className="text-xs font-medium text-text truncate">{item.title || t.untitled}</p>
                              {item.snippet && (
                                <p className="text-[11px] text-text-muted truncate mt-0.5">{item.snippet}</p>
                              )}
                            </div>
                          </button>
                        ))}
                      </div>
                    )}
                    {results!.memories?.length > 0 && (
                      <div>
                        <p className="section-label px-3 py-1.5 text-[10px] uppercase tracking-wider text-text-muted">
                          {t.header.memoriesLabel}
                        </p>
                        {results!.memories.map((item) => (
                          <button
                            key={item.id}
                            onClick={() => handleSelect(item)}
                            className="w-full flex items-start gap-2.5 px-3 py-2 hover:bg-surface-hover transition-colors text-left cursor-pointer"
                          >
                            <Brain className="w-3.5 h-3.5 mt-0.5 text-accent-violet shrink-0" />
                            <div className="min-w-0">
                              <p className="text-xs font-medium text-text truncate">{item.title}</p>
                              {item.category && (
                                <span className="inline-block text-[10px] text-text-muted bg-surface-hover rounded px-1 mt-0.5">
                                  {item.category}
                                </span>
                              )}
                            </div>
                          </button>
                        ))}
                      </div>
                    )}
                  </div>
                ) : (
                  <div className="px-3 py-6 text-center text-xs text-text-muted">
                    {t.noResults}
                  </div>
                )}
              </div>
            )}
          </div>

          <div className="w-px h-5 bg-border" />

          <button
            onClick={() => {
              if (isRecording || recorder.isRecording) {
                recorder.stop();
              } else {
                recorder.start();
              }
            }}
            className={cn(
              'inline-flex items-center gap-2 px-3 py-1.5 rounded-md text-xs font-medium transition-colors duration-150 cursor-pointer',
              (isRecording || recorder.isRecording)
                ? 'bg-accent-rose/10 text-accent-rose'
                : 'bg-surface text-text-muted hover:text-text'
            )}
          >
            {(isRecording || recorder.isRecording) ? (
              <>
                <span className="relative flex h-2 w-2">
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent-rose opacity-75" />
                  <span className="relative inline-flex rounded-full h-2 w-2 bg-accent-rose" />
                </span>
                <span>{t.header.recording}</span>
              </>
            ) : (
              <>
                <Mic className="w-3.5 h-3.5" />
                <span>{t.header.record}</span>
              </>
            )}
          </button>

          <div className="w-px h-5 bg-border" />

          <div className="flex items-center gap-2 cursor-pointer">
            <div className="w-7 h-7 rounded-full bg-surface-hover flex items-center justify-center">
              <User className="w-3.5 h-3.5 text-text-muted" />
            </div>
            <span className="text-xs text-text-secondary">{user?.displayName}</span>
          </div>
        </div>
      </header>
      <RecordingBanner />
    </>
  );
}
