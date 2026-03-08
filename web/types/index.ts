export interface User {
  id: string;
  username: string;
  displayName: string;
  avatarUrl?: string;
  email?: string;
  role: 'admin' | 'user';
  createdAt: string;
}

export interface Conversation {
  id: string;
  userId: string;
  title: string;
  summary: string;
  transcript: string;
  audioUrl?: string;
  audioDuration?: number;
  language: string;
  status: 'recording' | 'processing' | 'completed' | 'failed';
  recordedAt: string;
  createdAt: string;
  tags?: Tag[];
}

export interface Memory {
  id: string;
  userId: string;
  conversationId?: string;
  content: string;
  category: 'insight' | 'fact' | 'preference' | 'event';
  importance: number;
  createdAt: string;
  updatedAt: string;
}

export interface Task {
  id: string;
  userId: string;
  conversationId?: string;
  title: string;
  description?: string;
  status: 'pending' | 'in_progress' | 'completed' | 'cancelled';
  priority: 'low' | 'medium' | 'high';
  dueDate?: string;
  completedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface Tag {
  id: string;
  userId: string;
  name: string;
  color: string;
  createdAt: string;
}

export interface ChatSession {
  id: string;
  userId: string;
  title: string;
  createdAt: string;
  updatedAt: string;
}

export interface ChatMessage {
  id: string;
  sessionId: string;
  role: 'user' | 'assistant';
  content: string;
  sources?: ChatSource[];
  createdAt: string;
}

export interface ChatSource {
  conversationId: string;
  title: string;
  excerpt: string;
}

export interface Extension {
  id: string;
  userId: string;
  name: string;
  description: string;
  icon: string;
  enabled: boolean;
  config?: string;
  createdAt: string;
  updatedAt: string;
}

export interface AppSettings {
  theme: 'dark' | 'light';
  language: string;
  sttProvider: 'zhipu' | 'whisper';
  llmProvider: 'openai' | 'anthropic' | 'ollama';
  llmModel: string;
  cloudSync: boolean;
  mcpEnabled: boolean;
  /** Recording input gain multiplier: 0.0 – 3.0 (1.0 = normal) */
  recordingGain: number;
}

export interface SearchResultItem {
  type: 'conversation' | 'memory';
  id: string;
  title: string;
  snippet: string;
  category?: string;
  importance?: number;
  createdAt: string;
  highlights?: Record<string, unknown>;
}

export interface SearchResponse {
  conversations: SearchResultItem[];
  memories: SearchResultItem[];
  totalFound: number;
}
