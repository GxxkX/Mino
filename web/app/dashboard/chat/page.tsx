'use client';

import { useEffect, useState, useRef, useCallback } from 'react';
import { Header } from '@/components/layout/header';
import { ChatMessageComponent } from '@/components/features/chat-message';
import { useAppStore } from '@/lib/store';
import { useTranslation } from '@/lib/i18n';
import { chatApi } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Send, Plus, Trash2, MessageCircle } from 'lucide-react';
import { cn } from '@/lib/utils';
import type { ChatSession } from '@/types';

export default function ChatPage() {
  const {
    chatSessions, setChatSessions, addChatSession, removeChatSession,
    activeSessionId, setActiveSessionId,
    chatMessages, setChatMessages, addChatMessage, appendToLastMessage,
  } = useAppStore();
  const t = useTranslation();

  const [input, setInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const abortStreamRef = useRef<(() => void) | null>(null);

  useEffect(() => {
    chatApi.listSessions()
      .then(setChatSessions)
      .catch(console.error);
  }, [setChatSessions]);

  useEffect(() => {
    if (!activeSessionId) {
      setChatMessages([]);
      return;
    }
    chatApi.getMessages(activeSessionId)
      .then(setChatMessages)
      .catch(console.error);
  }, [activeSessionId, setChatMessages]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [chatMessages]);

  useEffect(() => {
    return () => { abortStreamRef.current?.(); };
  }, []);

  const handleNewSession = useCallback(async () => {
    try {
      const sess = await chatApi.createSession();
      addChatSession(sess);
      setActiveSessionId(sess.id);
    } catch (err) {
      console.error(err);
    }
  }, [addChatSession, setActiveSessionId]);

  const handleDeleteSession = useCallback(async (e: React.MouseEvent, id: string) => {
    e.stopPropagation();
    try {
      await chatApi.deleteSession(id);
      removeChatSession(id);
    } catch (err) {
      console.error(err);
    }
  }, [removeChatSession]);

  const handleSelectSession = useCallback((sess: ChatSession) => {
    setActiveSessionId(sess.id);
  }, [setActiveSessionId]);

  const handleSend = async () => {
    if (!input.trim() || isLoading || !activeSessionId) return;

    const userMessage = {
      id: Date.now().toString(),
      sessionId: activeSessionId,
      role: 'user' as const,
      content: input.trim(),
      createdAt: new Date().toISOString(),
    };

    addChatMessage(userMessage);
    setInput('');
    setIsLoading(true);

    const placeholderId = (Date.now() + 1).toString();
    addChatMessage({
      id: placeholderId,
      sessionId: activeSessionId,
      role: 'assistant',
      content: '',
      createdAt: new Date().toISOString(),
    });

    const abort = chatApi.sendMessageStream(activeSessionId, userMessage.content, {
      onChunk: (chunk) => {
        appendToLastMessage(chunk);
      },
      onSources: (sources) => {
        useAppStore.setState((state) => {
          if (state.chatMessages.length === 0) return state;
          const messages = [...state.chatMessages];
          const last = messages[messages.length - 1];
          messages[messages.length - 1] = { ...last, sources };
          return { chatMessages: messages };
        });
      },
      onDone: (id, createdAt) => {
        useAppStore.setState((state) => {
          const messages = state.chatMessages.map((m) =>
            m.id === placeholderId ? { ...m, id, createdAt } : m
          );
          return { chatMessages: messages };
        });
        setIsLoading(false);
        abortStreamRef.current = null;
      },
      onError: () => {
        useAppStore.setState((state) => {
          const messages = state.chatMessages.map((m) =>
            m.id === placeholderId
              ? { ...m, content: t.chat.errorReply }
              : m
          );
          return { chatMessages: messages };
        });
        setIsLoading(false);
        abortStreamRef.current = null;
      },
    });

    abortStreamRef.current = abort;
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <div className="h-screen flex flex-col">
      <Header title={t.chat.title} />

      <div className="flex-1 flex min-h-0">
        <div className="w-56 border-r border-border flex flex-col bg-background">
          <div className="p-3">
            <Button onClick={handleNewSession} size="sm" className="w-full justify-start gap-2">
              <Plus className="w-3.5 h-3.5" />
              {t.chat.newChat}
            </Button>
          </div>

          <div className="flex-1 overflow-auto px-2 pb-2 space-y-0.5">
            {chatSessions.map((sess) => (
              <div
                key={sess.id}
                onClick={() => handleSelectSession(sess)}
                className={cn(
                  'group flex items-center gap-2 px-2.5 py-2 rounded-md text-[13px] cursor-pointer transition-colors',
                  activeSessionId === sess.id
                    ? 'bg-surface-hover text-text font-medium'
                    : 'text-text-muted hover:text-text-secondary hover:bg-surface'
                )}
              >
                <MessageCircle className="w-3.5 h-3.5 flex-shrink-0" />
                <span className="flex-1 truncate">{sess.title}</span>
                <button
                  onClick={(e) => handleDeleteSession(e, sess.id)}
                  className="opacity-0 group-hover:opacity-100 p-0.5 rounded hover:bg-red-500/10 hover:text-red-400 transition-all"
                  aria-label={t.chat.deleteSession}
                >
                  <Trash2 className="w-3 h-3" />
                </button>
              </div>
            ))}

            {chatSessions.length === 0 && (
              <p className="text-[12px] text-text-muted text-center py-6">{t.chat.emptySessions}</p>
            )}
          </div>
        </div>

        <div className="flex-1 flex flex-col min-h-0">
          {!activeSessionId ? (
            <div className="flex-1 flex flex-col items-center justify-center text-center px-8">
              <div className="w-10 h-10 rounded-lg bg-surface-hover flex items-center justify-center mb-4">
                <span className="text-sm font-semibold text-cta">M</span>
              </div>
              <h3 className="text-base font-medium text-text mb-1.5">{t.chat.welcomeTitle}</h3>
              <p className="text-sm text-text-muted max-w-sm leading-relaxed mb-4">
                {t.chat.welcomeDesc}
              </p>
              <Button onClick={handleNewSession} size="sm" className="gap-2">
                <Plus className="w-3.5 h-3.5" />
                {t.chat.startChat}
              </Button>
            </div>
          ) : (
            <>
              <div className="flex-1 overflow-auto px-8 py-6 space-y-5">
                {chatMessages.length === 0 && (
                  <div className="flex flex-col items-center justify-center h-full text-center">
                    <p className="text-sm text-text-muted">{t.chat.startMessage}</p>
                  </div>
                )}
                {chatMessages.map((message) => (
                  <ChatMessageComponent key={message.id} message={message} />
                ))}
                {isLoading && chatMessages[chatMessages.length - 1]?.content === '' && (
                  <div className="flex items-center gap-2 text-text-muted text-sm">
                    <span className="animate-pulse">{t.chat.thinking}</span>
                  </div>
                )}
                <div ref={messagesEndRef} />
              </div>

              <div className="px-8 py-4 border-t border-border">
                <div className="max-w-2xl mx-auto flex gap-2">
                  <Input
                    placeholder={t.chat.sendPlaceholder}
                    value={input}
                    onChange={(e) => setInput(e.target.value)}
                    onKeyPress={handleKeyPress}
                    disabled={isLoading}
                    className="flex-1"
                  />
                  <Button onClick={handleSend} disabled={isLoading || !input.trim()} size="md">
                    <Send className="w-3.5 h-3.5" />
                  </Button>
                </div>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
