import { request, type ApiResponse } from './client';
import type { SpeakerProfile } from '@/types';

export async function listSpeakers() {
  return request<ApiResponse<SpeakerProfile[]>>('/speakers');
}

export async function createSpeaker(label: string) {
  return request<ApiResponse<SpeakerProfile>>('/speakers', {
    method: 'POST',
    body: JSON.stringify({ label }),
  });
}

export async function updateSpeaker(id: string, label: string) {
  return request<ApiResponse<SpeakerProfile>>(`/speakers/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ label }),
  });
}

export async function deleteSpeaker(id: string) {
  return request<void>(`/speakers/${id}`, { method: 'DELETE' });
}
