import { request, type ApiResponse } from './client';
import type { ChatSession, ChatMessage, ChatSource } from '@/types';

// --------------- Sessions ---------------

export async function createSession(title?: string): Promise<ChatSession> {
  const res = await request<ApiResponse<ChatSession>>('/chat/sessions', {
    method: 'POST',
    body: JSON.stringify({ title: title || '' }),
  });
  return res.data;
}

export async function listSessions(): Promise<ChatSession[]> {
  const res = await request<ApiResponse<ChatSession[]>>('/chat/sessions');
  return res.data ?? [];
}

export async function updateSession(id: string, title: string): Promise<void> {
  await request(`/chat/sessions/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ title }),
  });
}

export async function deleteSession(id: string): Promise<void> {
  await request(`/chat/sessions/${id}`, { method: 'DELETE' });
}

// --------------- Messages ---------------

export async function getMessages(sessionId: string): Promise<ChatMessage[]> {
  const res = await request<ApiResponse<ChatMessage[]>>(`/chat/sessions/${sessionId}/messages`);
  return res.data ?? [];
}

export async function sendMessage(sessionId: string, message: string): Promise<ChatMessage> {
  const res = await request<ApiResponse<ChatMessage>>(`/chat/sessions/${sessionId}/messages`, {
    method: 'POST',
    body: JSON.stringify({ message }),
  });
  return res.data;
}

// SSE event types from the stream endpoint
export type SSEChunkEvent = { type: 'chunk'; content: string };
export type SSESourcesEvent = { type: 'sources'; sources: ChatSource[] };
export type SSEDoneEvent = { type: 'done'; id: string; createdAt: string };
export type SSEErrorEvent = { type: 'error'; message: string };
export type SSEEvent = SSEChunkEvent | SSESourcesEvent | SSEDoneEvent | SSEErrorEvent;

export interface StreamCallbacks {
  onChunk: (chunk: string) => void;
  onSources?: (sources: ChatSource[]) => void;
  onDone: (id: string, createdAt: string) => void;
  onError?: (message: string) => void;
}

/**
 * sendMessageStream posts a message and consumes the SSE stream.
 * Calls the appropriate callback for each event type.
 * Returns a cleanup function that aborts the request if called.
 */
export function sendMessageStream(
  sessionId: string,
  message: string,
  callbacks: StreamCallbacks,
): () => void {
  const controller = new AbortController();

  // Direct connection to backend, bypassing any proxy that would buffer SSE.
  const STREAM_BASE = process.env.NEXT_PUBLIC_API_URL || '/api/v1';
  const token = typeof window !== 'undefined' ? localStorage.getItem('access_token') : null;

  fetch(`${STREAM_BASE}/chat/sessions/${sessionId}/messages/stream`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify({ message }),
    signal: controller.signal,
  })
    .then(async (res) => {
      if (!res.ok || !res.body) {
        callbacks.onError?.(`HTTP ${res.status}`);
        return;
      }

      const reader = res.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });

        // SSE lines are separated by double newlines
        const parts = buffer.split('\n\n');
        buffer = parts.pop() ?? '';

        for (const part of parts) {
          const line = part.trim();
          if (!line.startsWith('data:')) continue;
          const jsonStr = line.slice('data:'.length).trim();
          try {
            const event: SSEEvent = JSON.parse(jsonStr);
            if (event.type === 'chunk') callbacks.onChunk(event.content);
            else if (event.type === 'sources') callbacks.onSources?.(event.sources);
            else if (event.type === 'done') callbacks.onDone(event.id, event.createdAt);
            else if (event.type === 'error') callbacks.onError?.(event.message);
          } catch {
            // malformed JSON — skip
          }
        }
      }
    })
    .catch((err) => {
      if (err.name !== 'AbortError') {
        callbacks.onError?.(err.message ?? 'stream error');
      }
    });

  return () => controller.abort();
}
