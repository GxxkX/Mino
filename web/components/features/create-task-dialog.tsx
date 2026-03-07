'use client';

import { useEffect, useRef, useCallback, useState } from 'react';
import { X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { tasksApi } from '@/lib/api';
import { useAppStore } from '@/lib/store';
import { useTranslation } from '@/lib/i18n';

interface CreateTaskDialogProps {
  open: boolean;
  onClose: () => void;
}

type Priority = 'low' | 'medium' | 'high';

export function CreateTaskDialog({ open, onClose }: CreateTaskDialogProps) {
  const { addTask } = useAppStore();
  const t = useTranslation();
  const overlayRef = useRef<HTMLDivElement>(null);
  const titleRef = useRef<HTMLInputElement>(null);

  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [priority, setPriority] = useState<Priority>('medium');
  const [dueDate, setDueDate] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  const PRIORITIES: { value: Priority; label: string; activeClass: string }[] = [
    {
      value: 'low',
      label: t.tasks.priorityLow,
      activeClass: 'border-text-muted text-text bg-surface-hover',
    },
    {
      value: 'medium',
      label: t.tasks.priorityMedium,
      activeClass: 'border-accent-amber text-accent-amber bg-accent-amber/10',
    },
    {
      value: 'high',
      label: t.tasks.priorityHigh,
      activeClass: 'border-accent-rose text-accent-rose bg-accent-rose/10',
    },
  ];

  useEffect(() => {
    if (open) {
      setTitle('');
      setDescription('');
      setPriority('medium');
      setDueDate('');
      setError('');
      setTimeout(() => titleRef.current?.focus(), 50);
    }
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [open, onClose]);

  const handleOverlayClick = useCallback(
    (e: React.MouseEvent) => {
      if (e.target === overlayRef.current) onClose();
    },
    [onClose],
  );

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = title.trim();
    if (!trimmed) {
      setError(t.createTask.titleRequired);
      titleRef.current?.focus();
      return;
    }
    setSubmitting(true);
    setError('');
    try {
      const payload: Parameters<typeof tasksApi.createTask>[0] = {
        title: trimmed,
        priority,
      };
      if (description.trim()) payload.description = description.trim();
      if (dueDate) payload.dueDate = new Date(dueDate).toISOString();

      const task = await tasksApi.createTask(payload);
      addTask(task);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : t.createTask.createError);
    } finally {
      setSubmitting(false);
    }
  };

  if (!open) return null;

  return (
    <div
      ref={overlayRef}
      onClick={handleOverlayClick}
      className="fixed inset-0 z-50 flex items-end justify-center sm:items-center bg-black/60 backdrop-blur-[4px]"
      role="dialog"
      aria-modal="true"
      aria-label={t.createTask.title}
    >
      <div className="relative w-full sm:w-[90%] sm:max-w-[480px] rounded-t-2xl sm:rounded-2xl bg-surface border border-border shadow-[0_20px_25px_rgba(0,0,0,0.4)] overflow-hidden">

        <div className="flex items-center justify-between px-6 pt-6 pb-4 border-b border-border-subtle">
          <h3 className="text-sm font-semibold text-text tracking-tight">{t.createTask.title}</h3>
          <button
            type="button"
            onClick={onClose}
            className="flex items-center justify-center w-6 h-6 rounded-md text-text-muted hover:text-text hover:bg-surface-hover transition-colors duration-150 cursor-pointer focus-visible:outline focus-visible:outline-2 focus-visible:outline-cta"
            aria-label={t.close}
          >
            <X className="w-3.5 h-3.5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} noValidate>
          <div className="px-6 py-5 space-y-5">

            <div className="space-y-1.5">
              <label className="section-label" htmlFor="task-title">
                {t.createTask.titleLabel}
              </label>
              <input
                ref={titleRef}
                id="task-title"
                type="text"
                value={title}
                onChange={(e) => { setTitle(e.target.value); setError(''); }}
                placeholder={t.createTask.titlePlaceholder}
                className="w-full bg-background border border-border rounded-lg px-3 py-2.5 text-sm text-text placeholder:text-text-muted/50 focus:outline-none focus:border-text-muted focus-visible:ring-1 focus-visible:ring-cta/40 transition-colors duration-150"
              />
              {error && (
                <p className="text-[11px] text-accent-rose mt-1">{error}</p>
              )}
            </div>

            <div className="space-y-1.5">
              <label className="section-label" htmlFor="task-desc">
                {t.createTask.descLabel} <span className="normal-case font-normal text-text-muted/60">{t.createTask.descOptional}</span>
              </label>
              <textarea
                id="task-desc"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder={t.createTask.descPlaceholder}
                rows={3}
                className="w-full bg-background border border-border rounded-lg px-3 py-2.5 text-sm text-text placeholder:text-text-muted/50 focus:outline-none focus:border-text-muted focus-visible:ring-1 focus-visible:ring-cta/40 transition-colors duration-150 resize-none"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">

              <div className="space-y-1.5">
                <span className="section-label">{t.createTask.priority}</span>
                <div className="flex gap-1.5" role="group" aria-label={t.createTask.priority}>
                  {PRIORITIES.map((p) => (
                    <button
                      key={p.value}
                      type="button"
                      onClick={() => setPriority(p.value)}
                      aria-pressed={priority === p.value}
                      className={`flex-1 py-2 rounded-lg text-xs font-medium border transition-colors duration-150 cursor-pointer focus-visible:outline focus-visible:outline-2 focus-visible:outline-cta ${
                        priority === p.value
                          ? p.activeClass
                          : 'border-border text-text-muted bg-transparent hover:border-text-muted/40 hover:text-text-secondary'
                      }`}
                    >
                      {p.label}
                    </button>
                  ))}
                </div>
              </div>

              <div className="space-y-1.5">
                <label className="section-label" htmlFor="task-due">
                  {t.createTask.dueDateLabel} <span className="normal-case font-normal text-text-muted/60">{t.createTask.dueDateOptional}</span>
                </label>
                <input
                  id="task-due"
                  type="date"
                  value={dueDate}
                  onChange={(e) => setDueDate(e.target.value)}
                  className="w-full bg-background border border-border rounded-lg px-3 py-2.5 text-sm text-text focus:outline-none focus:border-text-muted focus-visible:ring-1 focus-visible:ring-cta/40 transition-colors duration-150 cursor-pointer [color-scheme:dark]"
                />
              </div>
            </div>
          </div>

          <div className="flex items-center justify-end gap-2.5 px-6 py-4 border-t border-border-subtle">
            <Button type="button" variant="secondary" size="sm" onClick={onClose}>
              {t.cancel}
            </Button>
            <Button type="submit" size="sm" disabled={submitting}>
              {submitting ? t.createTask.creating : t.createTask.create}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
