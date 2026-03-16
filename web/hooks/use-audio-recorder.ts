'use client';

import { useSyncExternalStore, useCallback, useRef, useEffect } from 'react';
import { audioRecorder, type RecorderState } from '@/lib/audio-recorder';
import { useAppStore } from '@/lib/store';
import { conversationsApi } from '@/lib/api';

/**
 * Hook that exposes the global AudioRecorderService as reactive state.
 * All components calling this hook share the exact same recorder instance.
 *
 * frequencyData is NOT part of React state — it's a mutable Uint8Array
 * updated at 60fps by the service's rAF loop. The visualizer reads it
 * directly via a ref, bypassing React's render cycle entirely.
 */
export function useAudioRecorder() {
  const { setIsRecording, setCurrentTranscript, setConversations, setDiarizedSegments, setSpeakerMatches, setIsDiarizing, settings } = useAppStore();

  // Wire callbacks once — use refs to avoid stale closures
  const callbacksWired = useRef(false);
  if (!callbacksWired.current) {
    callbacksWired.current = true;

    audioRecorder.onTranscript((text) => setCurrentTranscript(text));

    audioRecorder.onDiarized((segments, speakers) => {
      setDiarizedSegments(segments);
      setSpeakerMatches(speakers);
    });

    audioRecorder.onDiarizing((v) => setIsDiarizing(v));

    audioRecorder.onCompleted(() => {
      // Re-fetch conversations list when backend finishes processing
      conversationsApi
        .listConversations(1, 50)
        .then((res) => setConversations(res.data ?? []))
        .catch(console.error);
    });
  }

  // Sync recording gain setting to the live GainNode whenever it changes
  useEffect(() => {
    audioRecorder.setGain(settings.recordingGain ?? 1.0);
  }, [settings.recordingGain]);

  const state = useSyncExternalStore<RecorderState>(
    (onStoreChange) => audioRecorder.subscribe(onStoreChange),
    () => audioRecorder.getState(),
    () => audioRecorder.getState(),
  );

  const start = useCallback(async () => {
    const gain = useAppStore.getState().settings.recordingGain ?? 1.0;
    await audioRecorder.start(gain);
    setIsRecording(true);
    setCurrentTranscript('');
  }, [setIsRecording, setCurrentTranscript]);

  const stop = useCallback(() => {
    audioRecorder.stop();
    setIsRecording(false);
  }, [setIsRecording]);

  const pause = useCallback(() => {
    audioRecorder.pause();
  }, []);

  const resume = useCallback(() => {
    audioRecorder.resume();
  }, []);

  return {
    ...state,
    /** Returns the live frequency buffer. Call this to get the current reference. */
    getFrequencyData: audioRecorder.getFrequencyData.bind(audioRecorder),
    start,
    stop,
    pause,
    resume,
  };
}
