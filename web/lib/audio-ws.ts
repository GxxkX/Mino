/**
 * WebSocket client for real-time audio streaming to backend.
 * Protocol:
 *   - Control messages (start/stop/pause/resume) are sent as JSON text frames.
 *   - Audio data is sent as raw binary frames (opus/webm blobs).
 *   - Server responses are always JSON text frames.
 */

import type { DiarizedSegment, SpeakerMatch } from '@/types';

type ServerMessageType = 'status' | 'transcript' | 'completed' | 'error' | 'diarized';

export interface WSServerMessage {
  type: ServerMessageType;
  text?: string;
  is_final?: boolean;
  timestamp?: number;
  conversation_id?: string;
  title?: string;
  summary?: string;
  action_items?: string[];
  memories?: unknown;
  error?: string;
  // Speaker diarization fields
  diarized_segments?: DiarizedSegment[];
  speakers?: Record<string, SpeakerMatch>;
}

interface WSControlMessage {
  type: 'control';
  action: 'start' | 'stop' | 'pause' | 'resume';
  timestamp: number;
}

export type AudioWSListener = (msg: WSServerMessage) => void;

export class AudioWebSocket {
  private ws: WebSocket | null = null;
  private listeners: Set<AudioWSListener> = new Set();

  get connected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      const token = localStorage.getItem('access_token');
      if (!token) {
        reject(new Error('No auth token'));
        return;
      }

      const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = `${proto}//${window.location.host}/api/v1/ws/audio?token=${token}`;

      this.ws = new WebSocket(wsUrl);
      this.ws.binaryType = 'arraybuffer';

      this.ws.onopen = () => resolve();
      this.ws.onerror = () => reject(new Error('WebSocket connection failed'));

      this.ws.onmessage = (event) => {
        // Server always sends JSON text frames
        if (typeof event.data === 'string') {
          try {
            const msg: WSServerMessage = JSON.parse(event.data);
            this.listeners.forEach((fn) => fn(msg));
          } catch {
            // ignore malformed messages
          }
        }
      };

      this.ws.onclose = () => {
        this.ws = null;
      };
    });
  }

  sendControl(action: WSControlMessage['action']) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type: 'control', action, timestamp: Date.now() }));
    }
  }

  /** Send an opus/webm audio blob as a binary WebSocket frame */
  sendAudioBinary(data: ArrayBuffer | Blob) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(data);
    }
  }

  onMessage(listener: AudioWSListener) {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.listeners.clear();
  }
}

// Singleton instance
export const audioWS = new AudioWebSocket();
