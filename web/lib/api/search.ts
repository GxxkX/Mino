import { request, type ApiResponse } from './client';
import type { SearchResponse } from '@/types';

export async function search(query: string, limit = 10): Promise<SearchResponse> {
  const res = await request<ApiResponse<SearchResponse>>(
    `/search?q=${encodeURIComponent(query)}&limit=${limit}`,
  );
  return res.data;
}

export async function reindex(): Promise<{ indexed: number }> {
  const res = await request<ApiResponse<{ indexed: number }>>('/search/reindex', {
    method: 'POST',
  });
  return res.data;
}
