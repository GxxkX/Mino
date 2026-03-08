import { create } from 'zustand';
import type { User, Conversation, Memory, Task, ChatMessage, ChatSession, AppSettings, Extension } from '@/types';

interface AppState {
  user: User | null;
  conversations: Conversation[];
  memories: Memory[];
  tasks: Task[];
  extensions: Extension[];
  chatSessions: ChatSession[];
  activeSessionId: string | null;
  chatMessages: ChatMessage[];
  settings: AppSettings;
  isRecording: boolean;
  isPaused: boolean;
  recordingDuration: number;
  currentTranscript: string;
  recordingError: string | null;
  /** ID of the conversation currently being played (null = nothing playing) */
  playingConversationId: string | null;
  setUser: (user: User | null) => void;
  setConversations: (conversations: Conversation[]) => void;
  removeConversation: (id: string) => void;
  setMemories: (memories: Memory[]) => void;
  setTasks: (tasks: Task[]) => void;
  addTask: (task: Task) => void;
  updateTaskStatus: (id: string, status: Task['status']) => void;
  setExtensions: (extensions: Extension[]) => void;
  updateExtension: (id: string, patch: Partial<Extension>) => void;
  removeExtension: (id: string) => void;
  setChatSessions: (sessions: ChatSession[]) => void;
  addChatSession: (session: ChatSession) => void;
  removeChatSession: (id: string) => void;
  setActiveSessionId: (id: string | null) => void;
  setChatMessages: (messages: ChatMessage[]) => void;
  addChatMessage: (message: ChatMessage) => void;
  appendToLastMessage: (chunk: string) => void;
  setSettings: (settings: Partial<AppSettings>) => void;
  setIsRecording: (isRecording: boolean) => void;
  setIsPaused: (isPaused: boolean) => void;
  setRecordingDuration: (duration: number) => void;
  setCurrentTranscript: (transcript: string) => void;
  setRecordingError: (error: string | null) => void;
  setPlayingConversationId: (id: string | null) => void;
}

export const useAppStore = create<AppState>((set) => ({
  user: null,
  conversations: [],
  memories: [],
  tasks: [],
  extensions: [],
  chatSessions: [],
  activeSessionId: null,
  chatMessages: [],
  settings: {
    theme: 'dark',
    language: 'zh-CN',
    sttProvider: 'whisper',
    llmProvider: 'openai',
    llmModel: 'gpt-4o',
    cloudSync: true,
    mcpEnabled: false,
    recordingGain: 1.0,
  },
  isRecording: false,
  isPaused: false,
  recordingDuration: 0,
  currentTranscript: '',
  recordingError: null,
  playingConversationId: null,
  setUser: (user) => set({ user }),
  setConversations: (conversations) => set({ conversations }),
  removeConversation: (id) => set((state) => ({
    conversations: state.conversations.filter(c => c.id !== id),
    playingConversationId: state.playingConversationId === id ? null : state.playingConversationId,
  })),
  setMemories: (memories) => set({ memories }),
  setTasks: (tasks) => set({ tasks }),
  addTask: (task) => set((state) => ({ tasks: [task, ...state.tasks] })),
  updateTaskStatus: (id, status) => set((state) => ({
    tasks: state.tasks.map(t => t.id === id ? { ...t, status, updatedAt: new Date().toISOString() } : t)
  })),
  setExtensions: (extensions) => set({ extensions }),
  updateExtension: (id, patch) => set((state) => ({
    extensions: state.extensions.map(e => e.id === id ? { ...e, ...patch, updatedAt: new Date().toISOString() } : e)
  })),
  removeExtension: (id) => set((state) => ({
    extensions: state.extensions.filter(e => e.id !== id)
  })),
  setChatSessions: (chatSessions) => set({ chatSessions }),
  addChatSession: (session) => set((state) => ({
    chatSessions: [session, ...state.chatSessions],
  })),
  removeChatSession: (id) => set((state) => ({
    chatSessions: state.chatSessions.filter(s => s.id !== id),
    activeSessionId: state.activeSessionId === id ? null : state.activeSessionId,
    chatMessages: state.activeSessionId === id ? [] : state.chatMessages,
  })),
  setActiveSessionId: (activeSessionId) => set({ activeSessionId }),
  setChatMessages: (chatMessages) => set({ chatMessages }),
  addChatMessage: (message) => set((state) => ({
    chatMessages: [...state.chatMessages, message],
  })),
  appendToLastMessage: (chunk) => set((state) => {
    if (state.chatMessages.length === 0) return state;
    const messages = [...state.chatMessages];
    const last = messages[messages.length - 1];
    messages[messages.length - 1] = { ...last, content: last.content + chunk };
    return { chatMessages: messages };
  }),
  setSettings: (newSettings) => set((state) => ({ 
    settings: { ...state.settings, ...newSettings } 
  })),
  setIsRecording: (isRecording) => set({ isRecording }),
  setIsPaused: (isPaused) => set({ isPaused }),
  setRecordingDuration: (recordingDuration) => set({ recordingDuration }),
  setCurrentTranscript: (currentTranscript) => set({ currentTranscript }),
  setRecordingError: (recordingError) => set({ recordingError }),
  setPlayingConversationId: (playingConversationId) => set({ playingConversationId }),
}));
