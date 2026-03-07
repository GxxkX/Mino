import { request, ApiResponse } from './client';
import type { Extension } from '@/types';

export async function listExtensions(): Promise<Extension[]> {
  const res = await request<ApiResponse<Extension[]>>('/extensions');
  return res.data ?? [];
}

export async function getExtension(id: string): Promise<Extension> {
  const res = await request<ApiResponse<Extension>>(`/extensions/${id}`);
  return res.data;
}

export interface CreateExtensionPayload {
  name: string;
  description?: string;
  icon?: string;
  config?: string;
}

export async function createExtension(payload: CreateExtensionPayload): Promise<Extension> {
  const res = await request<ApiResponse<Extension>>('/extensions', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
  return res.data;
}

export interface UpdateExtensionPayload {
  name?: string;
  description?: string;
  icon?: string;
  enabled?: boolean;
  config?: string;
}

export async function updateExtension(id: string, payload: UpdateExtensionPayload): Promise<Extension> {
  const res = await request<ApiResponse<Extension>>(`/extensions/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  });
  return res.data;
}

export async function deleteExtension(id: string): Promise<void> {
  await request(`/extensions/${id}`, { method: 'DELETE' });
}
