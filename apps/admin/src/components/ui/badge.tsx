import type { HTMLAttributes } from 'react'

import { cn } from '@/lib/utils'

export type BadgeVariant =
  'neutral' | 'success' | 'warning' | 'danger' | 'info' | 'outline'

const variants: Record<BadgeVariant, string> = {
  neutral:
    'bg-slate-100 text-slate-600 ring-slate-200 dark:bg-slate-800 dark:text-slate-300 dark:ring-slate-700',
  success:
    'bg-emerald-50 text-emerald-700 ring-emerald-200 dark:bg-emerald-500/10 dark:text-emerald-300 dark:ring-emerald-500/20',
  warning:
    'bg-amber-50 text-amber-700 ring-amber-200 dark:bg-amber-500/10 dark:text-amber-300 dark:ring-amber-500/20',
  danger:
    'bg-rose-50 text-rose-700 ring-rose-200 dark:bg-rose-500/10 dark:text-rose-300 dark:ring-rose-500/20',
  info: 'bg-indigo-50 text-indigo-700 ring-indigo-200 dark:bg-indigo-500/10 dark:text-indigo-300 dark:ring-indigo-500/20',
  outline:
    'bg-transparent text-slate-600 ring-slate-200 dark:text-slate-300 dark:ring-slate-700',
}

export function Badge({
  variant = 'neutral',
  className,
  children,
  ...props
}: HTMLAttributes<HTMLSpanElement> & { variant?: BadgeVariant }) {
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-[11px] font-semibold ring-1 ring-inset',
        variants[variant],
        className,
      )}
      {...props}
    >
      {children}
    </span>
  )
}
