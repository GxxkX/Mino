import { request, type ApiResponse } from './client';

export interface LLMConfig {
  provider: string;
  api_key: string;
  base_url: string;
  model: string;
  embedding_model: string;
}

export async function getLLMConfig(provider?: string): Promise<LLMConfig> {
  const query = provider ? `?provider=${encodeURIComponent(provider)}` : '';
  const res = await request<ApiResponse<LLMConfig>>(`/settings/llm${query}`);
  return res.data;
}

export async function updateLLMConfig(config: Partial<LLMConfig>): Promise<LLMConfig> {
  const res = await request<ApiResponse<LLMConfig>>('/settings/llm', {
    method: 'PUT',
    body: JSON.stringify(config),
  });
  return res.data;
}

export interface CloudConfig {
  // MinIO
  minio_endpoint: string;
  minio_access_key: string;
  minio_secret_key: string;
  minio_secure: boolean;
  minio_region: string;
  minio_public_url: string;
  // PostgreSQL
  db_host: string;
  db_port: string;
  db_name: string;
  db_user: string;
  db_password: string;
  db_ssl_mode: string;
  // Redis
  redis_host: string;
  redis_port: string;
  redis_password: string;
  redis_db: number;
  // Milvus
  milvus_host: string;
  milvus_port: string;
  milvus_user: string;
  milvus_password: string;
  milvus_db_name: string;
  // Typesense
  typesense_host: string;
  typesense_port: string;
  typesense_api_key: string;
}

export async function getCloudConfig(): Promise<CloudConfig> {
  const res = await request<ApiResponse<CloudConfig>>('/settings/cloud');
  return res.data;
}

export async function updateCloudConfig(config: Partial<CloudConfig>): Promise<CloudConfig> {
  const res = await request<ApiResponse<CloudConfig>>('/settings/cloud', {
    method: 'PUT',
    body: JSON.stringify(config),
  });
  return res.data;
}
