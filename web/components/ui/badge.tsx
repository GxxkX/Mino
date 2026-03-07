import { cn } from '@/lib/utils';

interface BadgeProps {
  children: React.ReactNode;
  variant?: 'default' | 'lime' | 'blue' | 'amber' | 'rose' | 'violet';
  className?: string;
}

const variantStyles = {
  default: 'bg-surface-hover text-text-muted',
  lime: 'bg-cta/10 text-cta',
  blue: 'bg-accent-blue/10 text-accent-blue',
  amber: 'bg-accent-amber/10 text-accent-amber',
  rose: 'bg-accent-rose/10 text-accent-rose',
  violet: 'bg-accent-violet/10 text-accent-violet',
};

export function Badge({ children, variant = 'default', className }: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium',
        variantStyles[variant],
        className
      )}
    >
      {children}
    </span>
  );
}
