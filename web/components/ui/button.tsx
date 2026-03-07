import { cn } from '@/lib/utils';
import { ButtonHTMLAttributes, forwardRef } from 'react';

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'ghost';
  size?: 'sm' | 'md' | 'lg';
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'primary', size = 'md', children, ...props }, ref) => {
    return (
      <button
        ref={ref}
        className={cn(
          'rounded-md font-medium transition-colors duration-150 cursor-pointer inline-flex items-center justify-center gap-2',
          {
            'bg-cta text-background hover:bg-cta-hover': variant === 'primary',
            'bg-surface text-text-secondary border border-border hover:bg-surface-hover hover:text-text': variant === 'secondary',
            'bg-transparent text-text-muted hover:text-text hover:bg-surface': variant === 'ghost',
          },
          {
            'px-2.5 py-1.5 text-xs': size === 'sm',
            'px-3.5 py-2 text-sm': size === 'md',
            'px-5 py-2.5 text-sm': size === 'lg',
          },
          'disabled:opacity-40 disabled:cursor-not-allowed',
          className
        )}
        {...props}
      >
        {children}
      </button>
    );
  }
);

Button.displayName = 'Button';
