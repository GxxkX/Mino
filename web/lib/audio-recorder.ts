/**
 * Singleton audio recorder service.
 * Uses AudioWorklet to capture raw PCM (16-bit, 16kHz, mono) for real-time
 * streaming to the backend for live STT.
 *
 * Audio graph:
 *   mic → GainNode → AnalyserNode  (visualization, 60fps rAF)
 *                  → AudioWorkletNode (pcm-processor) → WS binary (Int16 PCM)
 */

import { audioWS } from '@/lib/audio-ws';
import type { DiarizedSegment, SpeakerMatch } from '@/types';

const FFT_SIZE = 256;

/** Scalar state — safe for React reconciliation */
export interface RecorderState {
  isRecording: boolean;
  isPaused: boolean;
  duration: number;
  error: string | null;
}

type Listener = (state: RecorderState) => void;

class AudioRecorderService {
  private mediaStream: MediaStream | null = null;
  private audioContext: AudioContext | null = null;
  private analyser: AnalyserNode | null = null;
  private source: MediaStreamAudioSourceNode | null = null;
  private gainNode: GainNode | null = null;
  private workletNode: AudioWorkletNode | null = null;
  private raf = 0;
  private durationTimer: ReturnType<typeof setInterval> | undefined;
  private startTime = 0;
  private wsUnsub: (() => void) | null = null;
  private transcriptCallback: ((text: string, isFinal: boolean) => void) | null = null;
  private completedCallback: (() => void) | null = null;
  private diarizedCallback: ((segments: DiarizedSegment[], speakers: Record<string, SpeakerMatch>) => void) | null = null;
  private diarizingCallback: ((isDiarizing: boolean) => void) | null = null;

  /**
   * Frequency data buffer — mutated in-place at 60fps.
   * The visualizer reads it directly via a ref, bypassing React renders.
   */
  private _frequencyData = new Uint8Array(FFT_SIZE / 2);

  private state: RecorderState = {
    isRecording: false,
    isPaused: false,
    duration: 0,
    error: null,
  };

  private listeners = new Set<Listener>();

  subscribe(fn: Listener) {
    this.listeners.add(fn);
    return () => this.listeners.delete(fn);
  }

  onTranscript(cb: (text: string, isFinal?: boolean) => void) {
    this.transcriptCallback = cb;
  }

  onCompleted(cb: () => void) {
    this.completedCallback = cb;
  }

  onDiarized(cb: (segments: DiarizedSegment[], speakers: Record<string, SpeakerMatch>) => void) {
    this.diarizedCallback = cb;
  }

  onDiarizing(cb: (isDiarizing: boolean) => void) {
    this.diarizingCallback = cb;
  }

  getState(): RecorderState {
    return this.state;
  }

  getFrequencyData(): Uint8Array {
    return this._frequencyData;
  }

  /** Update the recording gain in real-time (safe to call while recording) */
  setGain(value: number) {
    if (this.gainNode) {
      this.gainNode.gain.value = value;
    }
  }

  private emit() {
    this.state = { ...this.state };
    this.listeners.forEach((fn) => fn(this.state));
  }

  /** rAF loop — writes into the shared _frequencyData buffer */
  private updateFrequency = () => {
    if (this.analyser) {
      this.analyser.getByteFrequencyData(this._frequencyData);
    }
    this.raf = requestAnimationFrame(this.updateFrequency);
  };

  async start(initialGain = 1.0) {
    if (this.state.isRecording) return;

    try {
      this.state = { ...this.state, error: null };

      if (!navigator.mediaDevices?.getUserMedia) {
        throw new Error(
          window.isSecureContext === false
            ? '录音功能需要 HTTPS 或 localhost 环境'
            : '当前浏览器不支持录音功能',
        );
      }

      const stream = await navigator.mediaDevices.getUserMedia({
        audio: { echoCancellation: true, noiseSuppression: true },
      });
      this.mediaStream = stream;

      // 16kHz mono — required for real-time STT
      const ctx = new AudioContext({ sampleRate: 16000 });
      this.audioContext = ctx;

      // Load the PCM worklet processor
      await ctx.audioWorklet.addModule('/pcm-processor.js');

      const source = ctx.createMediaStreamSource(stream);
      this.source = source;

      // GainNode — applies the user's recording gain setting
      const gain = ctx.createGain();
      gain.gain.value = initialGain;
      this.gainNode = gain;
      source.connect(gain);

      // AnalyserNode — for visualization
      const analyser = ctx.createAnalyser();
      analyser.fftSize = FFT_SIZE;
      analyser.smoothingTimeConstant = 0.8;
      this.analyser = analyser;
      gain.connect(analyser);

      if (this._frequencyData.length !== analyser.frequencyBinCount) {
        this._frequencyData = new Uint8Array(analyser.frequencyBinCount);
      } else {
        this._frequencyData.fill(0);
      }

      // AudioWorkletNode — captures PCM and sends Int16 frames to main thread
      const worklet = new AudioWorkletNode(ctx, 'pcm-processor', {
        numberOfInputs: 1,
        numberOfOutputs: 0,
        channelCount: 1,
        channelCountMode: 'explicit',
        channelInterpretation: 'discrete',
      });
      this.workletNode = worklet;
      gain.connect(worklet);

      // Connect WebSocket
      try {
        await audioWS.connect();
        audioWS.sendControl('start');
      } catch {
        console.warn('WebSocket unavailable, recording locally only');
      }

      this.wsUnsub = audioWS.onMessage((msg) => {
        if (msg.type === 'transcript' && msg.text) {
          this.transcriptCallback?.(msg.text, msg.is_final ?? false);
        }
        if (msg.type === 'status' && msg.text === 'diarizing') {
          this.diarizingCallback?.(true);
        }
        if (msg.type === 'diarized') {
          this.diarizingCallback?.(false);
          this.diarizedCallback?.(msg.diarized_segments ?? [], msg.speakers ?? {});
        }
        if (msg.type === 'completed') {
          this.completedCallback?.();
        }
      });

      // Forward PCM frames from worklet to WebSocket
      worklet.port.onmessage = (e: MessageEvent<ArrayBuffer>) => {
        if (this.state.isRecording && !this.state.isPaused) {
          audioWS.sendAudioBinary(new Blob([e.data]));
        }
      };

      // Start the frequency visualization rAF loop
      this.raf = requestAnimationFrame(this.updateFrequency);

      this.startTime = Date.now();
      this.durationTimer = setInterval(() => {
        this.state = { ...this.state, duration: Math.floor((Date.now() - this.startTime) / 1000) };
        this.emit();
      }, 200);

      this.state = { ...this.state, isRecording: true, isPaused: false, duration: 0 };
      this.emit();
    } catch (err) {
      const message = err instanceof Error ? err.message : '无法访问麦克风';
      this.state = { ...this.state, error: message };
      this.emit();
    }
  }

  stop() {
    cancelAnimationFrame(this.raf);
    if (this.durationTimer) clearInterval(this.durationTimer);

    // Disconnect worklet first to stop PCM frames
    this.workletNode?.disconnect();
    this.workletNode = null;

    // Send stop control to backend
    audioWS.sendControl('stop');

    // Clean up local audio resources
    this.gainNode?.disconnect();
    this.source?.disconnect();
    this.analyser?.disconnect();
    this.audioContext?.close();
    this.mediaStream?.getTracks().forEach((t) => t.stop());

    this.mediaStream = null;
    this.audioContext = null;
    this.analyser = null;
    this.source = null;
    this.gainNode = null;

    this._frequencyData.fill(0);
    this.state = { isRecording: false, isPaused: false, duration: 0, error: null };
    this.emit();

    // Keep the WS connection alive to receive the "completed" message
    const cleanupTimeout = setTimeout(() => {
      this.wsUnsub?.();
      this.wsUnsub = null;
      audioWS.disconnect();
    }, 120_000);

    const origUnsub = this.wsUnsub;
    this.wsUnsub = null;
    const completedUnsub = audioWS.onMessage((msg) => {
      if (msg.type === 'completed') {
        this.completedCallback?.();
        clearTimeout(cleanupTimeout);
        completedUnsub();
        origUnsub?.();
        audioWS.disconnect();
      }
      if (msg.type === 'error') {
        clearTimeout(cleanupTimeout);
        completedUnsub();
        origUnsub?.();
        audioWS.disconnect();
      }
    });
  }

  pause() {
    if (!this.state.isRecording || this.state.isPaused) return;
    audioWS.sendControl('pause');
    this.state = { ...this.state, isPaused: true };
    this.emit();
  }

  resume() {
    if (!this.state.isRecording || !this.state.isPaused) return;
    audioWS.sendControl('resume');
    this.state = { ...this.state, isPaused: false };
    this.emit();
  }
}

/** Global singleton */
export const audioRecorder = new AudioRecorderService();
