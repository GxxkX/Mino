'use client';

import { useEffect, useRef, useCallback } from 'react';
import { cn } from '@/lib/utils';
import { Button } from './button';
import { X } from 'lucide-react';
import { useTranslation } from '@/lib/i18n';

interface ConfirmDialogProps {
  open: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title: string;
  description?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  /** When true the confirm button uses a destructive (rose) style */
  destructive?: boolean;
  /** Optional slot rendered between description and action buttons */
  children?: React.ReactNode;
}

export function ConfirmDialog({
  open,
  onClose,
  onConfirm,
  title,
  description,
  confirmLabel,
  cancelLabel,
  destructive = false,
  children,
}: ConfirmDialogProps) {
  const t = useTranslation();
  const resolvedConfirmLabel = confirmLabel ?? t.confirmDialog.confirm;
  const resolvedCancelLabel = cancelLabel ?? t.confirmDialog.cancel;
  const overlayRef = useRef<HTMLDivElement>(null);

  // Close on Escape
  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [open, onClose]);

  // Close on overlay click
  const handleOverlayClick = useCallback(
    (e: React.MouseEvent) => {
      if (e.target === overlayRef.current) onClose();
    },
    [onClose],
  );

  if (!open) return null;

  return (
    <div
      ref={overlayRef}
      onClick={handleOverlayClick}
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-[4px] transition-opacity duration-150"
      role="dialog"
      aria-modal="true"
      aria-label={title}
    >
      <div className="relative w-[90%] max-w-[420px] rounded-2xl bg-surface border border-border p-6 shadow-[0_20px_25px_rgba(0,0,0,0.15)]">
        {/* Close button */}
        <button
          onClick={onClose}
          className="absolute top-4 right-4 text-text-muted hover:text-text transition-colors duration-150 cursor-pointer"
          aria-label={t.confirmDialog.close}
        >
          <X className="w-4 h-4" />
        </button>

        {/* Title */}
        <h3 className="text-base font-semibold text-text pr-6">{title}</h3>

        {/* Description */}
        {description && (
          <p className="mt-2 text-sm text-text-muted leading-relaxed">{description}</p>
        )}

        {/* Optional extra content (e.g. checkboxes) */}
        {children && <div className="mt-4">{children}</div>}

        {/* Actions */}
        <div className="mt-6 flex items-center justify-end gap-3">
          <Button variant="secondary" size="sm" onClick={onClose}>
            {resolvedCancelLabel}
          </Button>
          <Button
            variant="primary"
            size="sm"
            onClick={onConfirm}
            className={cn(
              destructive && 'bg-accent-rose hover:bg-accent-rose/90 text-white',
            )}
          >
            {resolvedConfirmLabel}
          </Button>
        </div>
      </div>
    </div>
  );
}
