'use client';

import { useRef, useEffect, useCallback, useState } from 'react';
import { Badge } from '@/components/ui/badge';
import { ConfirmDialog } from '@/components/ui/confirm-dialog';
import { formatDate, formatDuration, truncate } from '@/lib/utils';
import { useAppStore } from '@/lib/store';
import { conversationsApi } from '@/lib/api';
import type { Conversation } from '@/types';
import { Play, Pause, Trash2 } from 'lucide-react';
import { useTranslation } from '@/lib/i18n';

const PREF_KEY = 'mino:delete-audio-with-recording';

interface ConversationCardProps {
  conversation: Conversation;
}

const tagVariantMap: Record<string, 'lime' | 'blue' | 'amber' | 'rose' | 'violet'> = {
  '#22C55E': 'lime',
  '#3B82F6': 'blue',
  '#F59E0B': 'amber',
  '#EC4899': 'rose',
};

function readDeleteAudioPref(): boolean {
  try {
    return localStorage.getItem(PREF_KEY) === 'true';
  } catch {
    return false;
  }
}

function writeDeleteAudioPref(value: boolean) {
  try {
    localStorage.setItem(PREF_KEY, String(value));
  } catch {
    // storage unavailable — silently ignore
  }
}

export function ConversationCard({ conversation }: ConversationCardProps) {
  const audioRef = useRef<HTMLAudioElement | null>(null);
  const [progress, setProgress] = useState(0);
  const { playingConversationId, setPlayingConversationId, removeConversation } = useAppStore();
  const t = useTranslation();

  // Delete confirmation state
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [deleteAudio, setDeleteAudio] = useState(readDeleteAudioPref);
  const [deleting, setDeleting] = useState(false);

  const isPlaying = playingConversationId === conversation.id;
  const hasAudio = !!(conversation.audioDuration && conversation.audioUrl);

  // Proxy MinIO audio through the internal Next.js route to avoid CORS /
  // mixed-content issues and keep MinIO off the public network.
  const proxiedAudioUrl = conversation.audioUrl
    ? `/api/minio?url=${encodeURIComponent(conversation.audioUrl)}`
    : undefined;

  // Attach event listeners via callback ref so they bind as soon as the
  // <audio> element mounts — regardless of render timing.
  const attachRef = useCallback(
    (el: HTMLAudioElement | null) => {
      const prev = audioRef.current;
      if (prev && prev !== el) {
        prev.removeEventListener('timeupdate', onTimeUpdate);
        prev.removeEventListener('ended', onEnded);
      }
      audioRef.current = el;
      if (el) {
        el.addEventListener('timeupdate', onTimeUpdate);
        el.addEventListener('ended', onEnded);
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [],
  );

  function onTimeUpdate(this: HTMLAudioElement) {
    if (this.duration && Number.isFinite(this.duration)) {
      setProgress(this.currentTime / this.duration);
    }
  }

  function onEnded() {
    setProgress(0);
    useAppStore.getState().setPlayingConversationId(null);
  }

  // Cleanup listeners on unmount
  useEffect(() => {
    return () => {
      const el = audioRef.current;
      if (el) {
        el.removeEventListener('timeupdate', onTimeUpdate);
        el.removeEventListener('ended', onEnded);
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Pause this card's audio when another card starts playing
  useEffect(() => {
    const audio = audioRef.current;
    if (!audio) return;
    if (!isPlaying && !audio.paused) {
      audio.pause();
    }
  }, [isPlaying]);

  const togglePlay = useCallback(() => {
    const audio = audioRef.current;
    if (!audio) return;

    if (isPlaying) {
      audio.pause();
      setPlayingConversationId(null);
    } else {
      setPlayingConversationId(conversation.id);
      audio.play().catch(console.error);
    }
  }, [isPlaying, conversation.id, setPlayingConversationId]);

  /** Click on progress bar to seek */
  const handleSeek = useCallback((e: React.MouseEvent<HTMLDivElement>) => {
    const audio = audioRef.current;
    if (!audio || !Number.isFinite(audio.duration)) return;
    const rect = e.currentTarget.getBoundingClientRect();
    const ratio = Math.max(0, Math.min(1, (e.clientX - rect.left) / rect.width));
    audio.currentTime = ratio * audio.duration;
    setProgress(ratio);
  }, []);

  // --- Delete flow ---
  const openDeleteConfirm = useCallback(() => {
    setDeleteAudio(readDeleteAudioPref());
    setConfirmOpen(true);
  }, []);

  const handleDeleteAudioToggle = useCallback((checked: boolean) => {
    setDeleteAudio(checked);
    writeDeleteAudioPref(checked);
  }, []);

  const handleConfirmDelete = useCallback(async () => {
    setDeleting(true);
    try {
      await conversationsApi.deleteConversation(conversation.id, deleteAudio);
      removeConversation(conversation.id);
      setConfirmOpen(false);
    } catch (err) {
      console.error('Failed to delete conversation:', err);
    } finally {
      setDeleting(false);
    }
  }, [conversation.id, deleteAudio, removeConversation]);

  return (
    <>
      <div className="group py-3 px-1 border-b border-border-subtle last:border-0 -mx-1 rounded hover:bg-surface/50 transition-colors duration-150">
        <div className="flex items-start gap-3">
          {/* Play button */}
          {conversation.audioDuration && (
            <button
              onClick={hasAudio ? togglePlay : undefined}
              disabled={!hasAudio}
              className={`mt-0.5 flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center transition-all duration-150 ${
                !hasAudio
                  ? 'bg-border/30 text-text-muted cursor-not-allowed'
                  : isPlaying
                    ? 'bg-cta text-bg shadow-[0_0_12px_rgba(163,230,53,0.4)]'
                    : 'bg-cta/10 border border-cta/30 text-cta hover:bg-cta/20 hover:border-cta/60 hover:shadow-[0_0_8px_rgba(163,230,53,0.2)]'
              }`}
            >
              {isPlaying
                ? <Pause className="w-3.5 h-3.5 fill-current" />
                : <Play className="w-3.5 h-3.5 fill-current ml-0.5" />
              }
            </button>
          )}

          <div className="flex-1 min-w-0 cursor-pointer">
            <div className="flex items-center justify-between mb-1">
              <h3 className="text-sm font-medium text-text">{conversation.title}</h3>
              <div className="flex items-center gap-2 flex-shrink-0 ml-2">
                <span className="text-[11px] text-text-muted">{formatDate(conversation.recordedAt)}</span>
                {/* Delete button — visible on hover */}
                <button
                  onClick={openDeleteConfirm}
                  className="opacity-0 group-hover:opacity-100 p-1 rounded text-text-muted hover:text-accent-rose hover:bg-accent-rose/10 transition-all duration-150 cursor-pointer"
                  aria-label={t.audio.deleteTitle}
                >
                  <Trash2 className="w-3.5 h-3.5" />
                </button>
              </div>
            </div>

            <p className="text-xs text-text-muted leading-relaxed mb-2">
              {truncate(conversation.summary, 100)}
            </p>

            <div className="flex items-center justify-between">
              <div className="flex items-center gap-1.5">
                {conversation.tags?.map((tag) => (
                  <Badge key={tag.id} variant={tagVariantMap[tag.color] || 'default'}>
                    {tag.name}
                  </Badge>
                ))}
              </div>
              {conversation.audioDuration && (
                <span className="text-[11px] text-text-muted tabular-nums">
                  {formatDuration(conversation.audioDuration)}
                </span>
              )}
            </div>
          </div>
        </div>

        {/* Playback progress bar — clickable to seek */}
        {isPlaying && (
          <div className="mt-2.5 ml-11 cursor-pointer" onClick={handleSeek}>
            <div className="h-1 bg-border rounded-full overflow-hidden">
              <div
                className="h-full bg-cta rounded-full"
                style={{ width: `${Math.round(progress * 100)}%` }}
              />
            </div>
          </div>
        )}

        {/* Hidden audio element — routed through internal MinIO proxy */}
        {hasAudio && (
          <audio ref={attachRef} src={proxiedAudioUrl} preload="none" />
        )}
      </div>

      {/* Delete confirmation dialog */}
      <ConfirmDialog
        open={confirmOpen}
        onClose={() => setConfirmOpen(false)}
        onConfirm={handleConfirmDelete}
        title={t.audio.deleteTitle}
        description={t.audio.deleteConfirm(truncate(conversation.title, 30))}
        confirmLabel={deleting ? t.audio.deleting : t.delete}
        destructive
      >
        {hasAudio && (
          <label className="flex items-center gap-2.5 cursor-pointer select-none group/check">
            <span className="relative flex items-center justify-center">
              <input
                type="checkbox"
                checked={deleteAudio}
                onChange={(e) => handleDeleteAudioToggle(e.target.checked)}
                className="peer sr-only"
              />
              <span className="w-4 h-4 rounded border border-border bg-background peer-checked:bg-cta peer-checked:border-cta transition-colors duration-150 flex items-center justify-center">
                {deleteAudio && (
                  <svg className="w-2.5 h-2.5 text-background" viewBox="0 0 12 12" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M2 6l3 3 5-5" />
                  </svg>
                )}
              </span>
            </span>
            <span className="text-sm text-text-secondary group-hover/check:text-text transition-colors duration-150">
              {t.audio.deleteAudioFile}
            </span>
          </label>
        )}
      </ConfirmDialog>
    </>
  );
}
