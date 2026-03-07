import { request, PaginatedResponse } from './client';
import type { Conversation } from '@/types';

export async function listConversations(page = 1, limit = 20): Promise<PaginatedResponse<Conversation>> {
  return request<PaginatedResponse<Conversation>>(`/conversations?page=${page}&limit=${limit}`);
}

export async function getConversation(id: string): Promise<Conversation> {
  const res = await request<{ code: number; data: Conversation }>(`/conversations/${id}`);
  return res.data;
}

export async function deleteConversation(id: string, deleteAudio = false): Promise<void> {
  const qs = deleteAudio ? '?delete_audio=true' : '';
  await request(`/conversations/${id}${qs}`, { method: 'DELETE' });
}
