import { request, PaginatedResponse } from './client';
import type { Task } from '@/types';

export async function listTasks(page = 1, limit = 20): Promise<PaginatedResponse<Task>> {
  return request<PaginatedResponse<Task>>(`/tasks?page=${page}&limit=${limit}`);
}

export interface CreateTaskPayload {
  title: string;
  description?: string;
  priority?: 'low' | 'medium' | 'high';
  dueDate?: string;
}

export async function createTask(payload: CreateTaskPayload): Promise<Task> {
  const res = await request<{ code: number; data: Task }>('/tasks', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
  return res.data;
}

export interface UpdateTaskPayload {
  title?: string;
  description?: string;
  status?: Task['status'];
  priority?: Task['priority'];
  dueDate?: string;
}

export async function updateTask(id: string, payload: UpdateTaskPayload): Promise<Task> {
  const res = await request<{ code: number; data: Task }>(`/tasks/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  });
  return res.data;
}

export async function deleteTask(id: string): Promise<void> {
  await request(`/tasks/${id}`, { method: 'DELETE' });
}
