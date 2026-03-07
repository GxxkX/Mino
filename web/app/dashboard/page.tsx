'use client';

import { useEffect } from 'react';
import { Header } from '@/components/layout/header';
import { MemoryCard } from '@/components/features/memory-card';
import { TaskList } from '@/components/features/task-list';
import { ConversationCard } from '@/components/features/conversation-card';
import { useAppStore } from '@/lib/store';
import { useTranslation } from '@/lib/i18n';
import { memoriesApi, tasksApi, conversationsApi } from '@/lib/api';
import { Mic, Brain, CheckSquare, TrendingUp } from 'lucide-react';

export default function DashboardPage() {
  const { memories, tasks, conversations, setMemories, setTasks, setConversations } = useAppStore();
  const t = useTranslation();

  useEffect(() => {
    memoriesApi.listMemories(1, 20).then(res => setMemories(res.data ?? [])).catch(console.error);
    tasksApi.listTasks(1, 20).then(res => setTasks(res.data ?? [])).catch(console.error);
    conversationsApi.listConversations(1, 20).then(res => setConversations(res.data ?? [])).catch(console.error);
  }, [setMemories, setTasks, setConversations]);

  const recentMemories = memories.slice(0, 3);
  const pendingTasks = tasks.filter(tk => tk.status !== 'completed').slice(0, 3);
  const recentConversations = conversations.slice(0, 3);

  const stats = [
    { label: t.dashboard.weekConversations, value: conversations.length.toString(), icon: Mic },
    { label: t.dashboard.memoryPoints, value: memories.length.toString(), icon: Brain },
    { label: t.dashboard.pendingTasks, value: tasks.filter(tk => tk.status === 'pending').length.toString(), icon: CheckSquare },
    { label: t.dashboard.completed, value: tasks.filter(tk => tk.status === 'completed').length.toString(), icon: TrendingUp },
  ];

  return (
    <div className="min-h-screen">
      <Header title={t.dashboard.title} description={t.dashboard.description} />

      <div className="px-8 pb-8 space-y-8">
        <div className="flex items-center gap-8 py-4 border-b border-border">
          {stats.map((stat) => (
            <div key={stat.label} className="flex items-center gap-3">
              <stat.icon className="w-4 h-4 text-text-muted" />
              <div>
                <span className="text-lg font-semibold text-text tabular-nums">{stat.value}</span>
                <span className="text-xs text-text-muted ml-2">{stat.label}</span>
              </div>
            </div>
          ))}
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-5 gap-8">
          <div className="lg:col-span-3 space-y-8">
            <section>
              <div className="flex items-center justify-between mb-3">
                <h2 className="section-label">{t.dashboard.recentMemories}</h2>
                <span className="text-[11px] text-text-muted cursor-pointer hover:text-text-secondary transition-colors">{t.dashboard.viewAll}</span>
              </div>
              <div>
                {recentMemories.map((memory) => (
                  <MemoryCard key={memory.id} memory={memory} />
                ))}
              </div>
            </section>

            <section>
              <div className="flex items-center justify-between mb-3">
                <h2 className="section-label">{t.dashboard.recentConversations}</h2>
                <span className="text-[11px] text-text-muted cursor-pointer hover:text-text-secondary transition-colors">{t.dashboard.viewAll}</span>
              </div>
              <div>
                {recentConversations.map((conversation) => (
                  <ConversationCard key={conversation.id} conversation={conversation} />
                ))}
              </div>
            </section>
          </div>

          <div className="lg:col-span-2">
            <section>
              <div className="flex items-center justify-between mb-3">
                <h2 className="section-label">{t.dashboard.pendingTasksSection}</h2>
                <span className="text-[11px] text-text-muted cursor-pointer hover:text-text-secondary transition-colors">{t.dashboard.viewAll}</span>
              </div>
              <TaskList tasks={pendingTasks} />
            </section>
          </div>
        </div>
      </div>
    </div>
  );
}
