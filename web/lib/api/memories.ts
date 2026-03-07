import { request, PaginatedResponse } from './client';
import type { Memory } from '@/types';

export async function listMemories(page = 1, limit = 20): Promise<PaginatedResponse<Memory>> {
  return request<PaginatedResponse<Memory>>(`/memories?page=${page}&limit=${limit}`);
}

export async function getMemory(id: string): Promise<Memory> {
  const res = await request<{ code: number; data: Memory }>(`/memories/${id}`);
  return res.data;
}

export interface UpdateMemoryPayload {
  content?: string;
  category?: string;
  importance?: number;
}

export async function updateMemory(id: string, payload: UpdateMemoryPayload): Promise<Memory> {
  const res = await request<{ code: number; data: Memory }>(`/memories/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  });
  return res.data;
}

export async function deleteMemory(id: string): Promise<void> {
  await request(`/memories/${id}`, { method: 'DELETE' });
}
