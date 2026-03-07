'use client';

import { Pause, Play, Square } from 'lucide-react';
import { useAppStore } from '@/lib/store';
import { useAudioRecorder } from '@/hooks/use-audio-recorder';
import { AudioVisualizer } from '@/components/features/audio-visualizer';
import { formatDuration } from '@/lib/utils';
import { cn } from '@/lib/utils';
import { useTranslation } from '@/lib/i18n';

export function RecordingBanner() {
  const { isRecording: storeRecording, currentTranscript } = useAppStore();
  const recorder = useAudioRecorder();
  const t = useTranslation();

  const active = recorder.isRecording || storeRecording;
  if (!active) return null;

  return (
    <div className="mx-8 mt-6 border border-accent-rose/30 rounded-lg p-5 space-y-4">
      {/* Header row: status + duration + controls */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <span className="relative flex h-2 w-2">
            <span
              className={cn(
                'absolute inline-flex h-full w-full rounded-full bg-accent-rose',
                !recorder.isPaused && 'animate-ping opacity-75',
              )}
            />
            <span className="relative inline-flex rounded-full h-2 w-2 bg-accent-rose" />
          </span>
          <span className="text-xs font-medium text-accent-rose">
            {recorder.isPaused ? t.recordingBanner.paused : t.recordingBanner.recording}
          </span>
          <span className="text-xs text-text-muted font-mono tabular-nums">
            {formatDuration(recorder.duration)}
          </span>
        </div>

        {/* Controls */}
        <div className="flex items-center gap-1.5">
          {recorder.isPaused ? (
            <button
              onClick={recorder.resume}
              className="p-1.5 rounded-md hover:bg-surface-hover text-text-muted hover:text-text transition-colors cursor-pointer"
              aria-label={t.recordingBanner.resume}
            >
              <Play className="w-3.5 h-3.5" />
            </button>
          ) : (
            <button
              onClick={recorder.pause}
              className="p-1.5 rounded-md hover:bg-surface-hover text-text-muted hover:text-text transition-colors cursor-pointer"
              aria-label={t.recordingBanner.pause}
            >
              <Pause className="w-3.5 h-3.5" />
            </button>
          )}
          <button
            onClick={recorder.stop}
            className="p-1.5 rounded-md hover:bg-accent-rose/10 text-accent-rose transition-colors cursor-pointer"
            aria-label={t.recordingBanner.stop}
          >
            <Square className="w-3.5 h-3.5" />
          </button>
        </div>
      </div>

      {/* Frequency visualizer */}
      <AudioVisualizer
        getFrequencyData={recorder.getFrequencyData}
        isActive={recorder.isRecording}
        isPaused={recorder.isPaused}
        barColor="#fb7185"
        barCount={50}
        barMaxHeight={40}
        barWidth={2}
        barGap={1.5}
      />

      {/* Live transcript */}
      <div className="bg-surface rounded-md p-4 min-h-[60px]">
        <p className="text-sm text-text-muted leading-relaxed">
          {currentTranscript || t.recordingBanner.waiting}
        </p>
      </div>

      {/* Error display */}
      {recorder.error && (
        <p className="text-xs text-accent-rose">{recorder.error}</p>
      )}
    </div>
  );
}
