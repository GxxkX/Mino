'use client';

import { useRef, useState } from 'react';
import { Badge } from '@/components/ui/badge';
import { formatDate } from '@/lib/utils';
import { useTranslation } from '@/lib/i18n';
import type { Task } from '@/types';
import { Check, Circle, Clock, X, ChevronDown } from 'lucide-react';
import { useAppStore } from '@/lib/store';
import { tasksApi } from '@/lib/api';

interface TaskListProps {
  tasks: Task[];
}

const priorityVariant = {
  low: 'default' as const,
  medium: 'amber' as const,
  high: 'rose' as const,
};

const statusIcons: Record<Task['status'], React.ElementType> = {
  pending: Circle,
  in_progress: Clock,
  completed: Check,
  cancelled: X,
};

const statusColors: Record<Task['status'], string> = {
  pending: 'text-text-muted',
  in_progress: 'text-accent-amber',
  completed: 'text-cta',
  cancelled: 'text-text-muted',
};

const statusOrder: Task['status'][] = ['pending', 'in_progress', 'completed', 'cancelled'];

function StatusPicker({ task }: { task: Task }) {
  const [open, setOpen] = useState(false);
  const { updateTaskStatus } = useAppStore();
  const t = useTranslation();
  const ref = useRef<HTMLDivElement>(null);

  const statusLabels: Record<Task['status'], string> = {
    pending: t.tasks.pending,
    in_progress: t.tasks.inProgress,
    completed: t.tasks.completed,
    cancelled: t.tasks.cancelled,
  };

  const color = statusColors[task.status];
  const Icon = statusIcons[task.status];

  const handleSelect = async (status: Task['status']) => {
    setOpen(false);
    updateTaskStatus(task.id, status);
    try {
      await tasksApi.updateTask(task.id, { status });
    } catch {
      updateTaskStatus(task.id, task.status);
    }
  };

  return (
    <div className="relative flex-shrink-0" ref={ref}>
      <button
        onClick={(e) => { e.stopPropagation(); setOpen(o => !o); }}
        className={`group/btn flex items-center gap-1 mt-0.5 rounded px-1 py-0.5 -ml-1 transition-colors duration-150 hover:bg-surface-hover ${color}`}
      >
        <Icon className="w-4 h-4" strokeWidth={task.status === 'completed' ? 2.5 : 1.5} />
        <ChevronDown className="w-2.5 h-2.5 opacity-0 group-hover/btn:opacity-60 transition-opacity duration-150" />
      </button>

      {open && (
        <>
          <div className="fixed inset-0 z-10" onClick={() => setOpen(false)} />
          <div className="absolute left-0 top-full mt-1 z-20 w-28 rounded-lg border border-border bg-surface shadow-lg py-1 overflow-hidden">
            {statusOrder.map((s) => {
              const SIcon = statusIcons[s];
              const sColor = statusColors[s];
              return (
                <button
                  key={s}
                  onClick={(e) => { e.stopPropagation(); handleSelect(s); }}
                  className={`w-full flex items-center gap-2 px-3 py-1.5 text-xs transition-colors duration-100 hover:bg-surface-hover ${
                    task.status === s ? sColor + ' font-medium' : 'text-text-secondary'
                  }`}
                >
                  <SIcon className="w-3.5 h-3.5 flex-shrink-0" strokeWidth={s === 'completed' ? 2.5 : 1.5} />
                  {statusLabels[s]}
                </button>
              );
            })}
          </div>
        </>
      )}
    </div>
  );
}

export function TaskList({ tasks }: TaskListProps) {
  const t = useTranslation();

  const priorityLabels: Record<string, string> = {
    low: t.tasks.priorityLow,
    medium: t.tasks.priorityMedium,
    high: t.tasks.priorityHigh,
  };

  if (tasks.length === 0) {
    return (
      <div className="text-center py-16 text-text-muted text-sm">{t.tasks.empty}</div>
    );
  }

  return (
    <div>
      {tasks.map((task) => (
        <div
          key={task.id}
          className="flex items-start gap-3 py-3 px-1 border-b border-border-subtle last:border-0 hover:bg-surface/50 transition-colors duration-150 -mx-1 rounded"
        >
          <StatusPicker task={task} />
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-0.5">
              <span className={`text-sm font-medium ${task.status === 'completed' ? 'line-through text-text-muted' : 'text-text'}`}>
                {task.title}
              </span>
              <Badge variant={priorityVariant[task.priority]}>
                {priorityLabels[task.priority]}
              </Badge>
            </div>
            {task.description && (
              <p className="text-xs text-text-muted mb-1">{task.description}</p>
            )}
            <div className="flex items-center gap-3 text-[11px] text-text-muted">
              {task.dueDate && <span>{t.tasks.dueDate(formatDate(task.dueDate))}</span>}
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
