import { cn } from '@/lib/utils';
import { InputHTMLAttributes, forwardRef } from 'react';

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, label, ...props }, ref) => {
    return (
      <div className="w-full">
        {label && (
          <label className="block text-xs font-medium text-text-muted mb-1.5">
            {label}
          </label>
        )}
        <input
          ref={ref}
          className={cn(
            'w-full bg-surface border border-border rounded-md px-3 py-2 text-sm text-text placeholder:text-text-muted/60 focus:outline-none focus:border-text-muted transition-colors duration-150',
            className
          )}
          {...props}
        />
      </div>
    );
  }
);

Input.displayName = 'Input';
