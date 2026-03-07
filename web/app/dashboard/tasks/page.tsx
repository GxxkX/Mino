'use client';

import { useEffect, useState } from 'react';
import { Header } from '@/components/layout/header';
import { TaskList } from '@/components/features/task-list';
import { CreateTaskDialog } from '@/components/features/create-task-dialog';
import { useAppStore } from '@/lib/store';
import { useTranslation } from '@/lib/i18n';
import { tasksApi } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Plus } from 'lucide-react';

type StatusFilter = 'all' | 'pending' | 'in_progress' | 'completed';

export default function TasksPage() {
  const { tasks, setTasks } = useAppStore();
  const t = useTranslation();
  const [filter, setFilter] = useState<StatusFilter>('all');
  const [loading, setLoading] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);

  useEffect(() => {
    tasksApi.listTasks(1, 100)
      .then(res => setTasks(res.data ?? []))
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [setTasks]);

  const filters: { value: StatusFilter; label: string }[] = [
    { value: 'all', label: t.tasks.all },
    { value: 'pending', label: t.tasks.pending },
    { value: 'in_progress', label: t.tasks.inProgress },
    { value: 'completed', label: t.tasks.completed },
  ];

  const filteredTasks = filter === 'all' ? tasks : tasks.filter(tk => tk.status === filter);

  const stats = {
    total: tasks.length,
    pending: tasks.filter(tk => tk.status === 'pending').length,
    inProgress: tasks.filter(tk => tk.status === 'in_progress').length,
    completed: tasks.filter(tk => tk.status === 'completed').length,
  };

  return (
    <div className="min-h-screen">
      <Header title={t.tasks.title} description={t.tasks.description(stats.pending, stats.inProgress, stats.completed)} />

      <div className="px-8 pb-8 space-y-5">
        <div className="flex items-center gap-6 py-3 border-b border-border">
          <div className="flex items-baseline gap-1.5">
            <span className="text-2xl font-semibold text-text tabular-nums">{stats.total}</span>
            <span className="text-xs text-text-muted">{t.tasks.total}</span>
          </div>
          <div className="w-px h-5 bg-border" />
          <div className="flex items-baseline gap-1.5">
            <span className="text-lg font-medium text-text tabular-nums">{stats.pending}</span>
            <span className="text-xs text-text-muted">{t.tasks.pending}</span>
          </div>
          <div className="flex items-baseline gap-1.5">
            <span className="text-lg font-medium text-accent-amber tabular-nums">{stats.inProgress}</span>
            <span className="text-xs text-text-muted">{t.tasks.inProgress}</span>
          </div>
          <div className="flex items-baseline gap-1.5">
            <span className="text-lg font-medium text-cta tabular-nums">{stats.completed}</span>
            <span className="text-xs text-text-muted">{t.tasks.completed}</span>
          </div>
        </div>

        <div className="flex items-center justify-between">
          <div className="flex gap-1">
            {filters.map((f) => (
              <button
                key={f.value}
                onClick={() => setFilter(f.value)}
                className={`px-3 py-1.5 rounded-md text-xs font-medium transition-colors duration-150 cursor-pointer ${
                  filter === f.value
                    ? 'bg-surface-hover text-text'
                    : 'text-text-muted hover:text-text-secondary'
                }`}
              >
                {f.label}
              </button>
            ))}
          </div>
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <Plus className="w-3.5 h-3.5" />
            {t.tasks.newTask}
          </Button>
        </div>

        {loading ? (
          <div className="text-center py-16 text-text-muted text-sm">{t.loading}</div>
        ) : (
          <TaskList tasks={filteredTasks} />
        )}
      </div>

      <CreateTaskDialog open={createOpen} onClose={() => setCreateOpen(false)} />
    </div>
  );
}
