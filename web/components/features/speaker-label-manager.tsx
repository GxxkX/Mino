'use client';

import { useEffect, useState } from 'react';
import { Plus, Pencil, Trash2, Users } from 'lucide-react';
import { useAppStore } from '@/lib/store';
import { speakersApi } from '@/lib/api';
import { useTranslation } from '@/lib/i18n';
import { cn } from '@/lib/utils';

interface SpeakerLabelManagerProps {
  /** When true, skip the initial fetch (parent already loaded data into store). */
  initialLoaded?: boolean;
}

export function SpeakerLabelManager({ initialLoaded = false }: SpeakerLabelManagerProps) {
  const { speakers, setSpeakers, addSpeaker, removeSpeaker, updateSpeakerLabel } = useAppStore();
  const t = useTranslation();
  const [newLabel, setNewLabel] = useState('');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editLabel, setEditLabel] = useState('');
  const [loading, setLoading] = useState(!initialLoaded);

  useEffect(() => {
    if (initialLoaded) { setLoading(false); return; }
    speakersApi.listSpeakers()
      .then((res) => { if (res.data) setSpeakers(res.data); })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [setSpeakers, initialLoaded]);

  const handleAdd = async () => {
    const label = newLabel.trim();
    if (!label) return;
    try {
      const res = await speakersApi.createSpeaker(label);
      if (res.data) addSpeaker(res.data);
      setNewLabel('');
    } catch {}
  };

  const handleRename = async (id: string) => {
    const label = editLabel.trim();
    if (!label) { setEditingId(null); return; }
    try {
      await speakersApi.updateSpeaker(id, label);
      updateSpeakerLabel(id, label);
    } catch {}
    setEditingId(null);
  };

  const handleDelete = async (id: string) => {
    try {
      await speakersApi.deleteSpeaker(id);
      removeSpeaker(id);
    } catch {}
  };

  // Speaker color based on index for visual distinction
  const colors = [
    'bg-accent-blue/10 text-accent-blue',
    'bg-accent-violet/10 text-accent-violet',
    'bg-cta/10 text-cta',
    'bg-accent-rose/10 text-accent-rose',
    'bg-accent-amber/10 text-accent-amber',
  ];

  if (loading) {
    return (
      <div className="text-center py-16 text-text-muted text-sm">{t.loading}</div>
    );
  }

  return (
    <div className="space-y-5">
      {/* Add new speaker input */}
      <div className="flex items-center gap-2">
        <input
          value={newLabel}
          onChange={(e) => setNewLabel(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleAdd()}
          placeholder={t.speakers.labelPlaceholder}
          className={cn(
            'flex-1 px-3 py-2 rounded-lg bg-surface border border-border',
            'text-sm text-text placeholder:text-text-muted outline-none',
            'focus:border-cta/50 focus:shadow-[0_0_0_3px_rgba(34,197,94,0.1)] transition-all duration-200',
          )}
        />
        <button
          onClick={handleAdd}
          disabled={!newLabel.trim()}
          className={cn(
            'inline-flex items-center gap-1.5 px-3 py-2 rounded-lg text-xs font-medium transition-all duration-200 cursor-pointer',
            newLabel.trim()
              ? 'bg-cta text-white hover:opacity-90 hover:-translate-y-px'
              : 'bg-surface text-text-muted cursor-not-allowed',
          )}
        >
          <Plus className="w-3.5 h-3.5" />
          <span>{t.speakers.addSpeaker}</span>
        </button>
      </div>

      {/* Speaker list */}
      {speakers.length === 0 ? (
        <div className="text-center py-16 text-text-muted text-sm">
          <Users className="w-8 h-8 mx-auto mb-3 opacity-30" />
          <p>{t.speakers.empty}</p>
        </div>
      ) : (
        <ul className="space-y-2" role="list">
          {speakers.map((s, i) => (
            <li
              key={s.id}
              className="flex items-center gap-3 px-4 py-3 rounded-xl bg-surface border border-border transition-all duration-200 hover:shadow-md hover:-translate-y-px"
            >
              {/* Avatar circle */}
              <div className={cn(
                'w-8 h-8 rounded-full flex items-center justify-center shrink-0 text-xs font-semibold',
                colors[i % colors.length],
              )}>
                {s.label.charAt(0).toUpperCase()}
              </div>

              {/* Name / edit */}
              {editingId === s.id ? (
                <input
                  autoFocus
                  value={editLabel}
                  onChange={(e) => setEditLabel(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') handleRename(s.id);
                    if (e.key === 'Escape') setEditingId(null);
                  }}
                  onBlur={() => handleRename(s.id)}
                  className="flex-1 bg-transparent border-b border-cta/40 text-text text-sm outline-none py-0.5"
                />
              ) : (
                <span className="flex-1 truncate text-sm font-medium text-text">{s.label}</span>
              )}

              {/* Sample count badge */}
              <span className="text-[10px] text-text-muted bg-surface-hover rounded-full px-2 py-0.5 whitespace-nowrap">
                {t.speakers.samples(s.sampleCount)}
              </span>

              {/* Actions */}
              <button
                onClick={() => { setEditingId(s.id); setEditLabel(s.label); }}
                className="p-1.5 rounded-md hover:bg-surface-hover text-text-muted hover:text-text transition-colors duration-200 cursor-pointer focus:outline-none focus:ring-2 focus:ring-cta/40"
                aria-label={t.speakers.rename}
              >
                <Pencil className="w-3.5 h-3.5" />
              </button>
              <button
                onClick={() => handleDelete(s.id)}
                className="p-1.5 rounded-md hover:bg-accent-rose/10 text-text-muted hover:text-accent-rose transition-colors duration-200 cursor-pointer focus:outline-none focus:ring-2 focus:ring-accent-rose/40"
                aria-label={t.speakers.deleteSpeaker}
              >
                <Trash2 className="w-3.5 h-3.5" />
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
