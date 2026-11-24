import type { LucideIcon } from 'lucide-react'
import type { ReactNode } from 'react'

type StatCardProps = {
  label: string
  value: string
  hint?: ReactNode
  icon: LucideIcon
  accent?: 'indigo' | 'emerald' | 'amber' | 'rose' | 'sky'
}

const accents = {
  indigo:
    'bg-indigo-50 text-indigo-600 dark:bg-indigo-500/10 dark:text-indigo-400',
  emerald:
    'bg-emerald-50 text-emerald-600 dark:bg-emerald-500/10 dark:text-emerald-400',
  amber: 'bg-amber-50 text-amber-600 dark:bg-amber-500/10 dark:text-amber-400',
  rose: 'bg-rose-50 text-rose-600 dark:bg-rose-500/10 dark:text-rose-400',
  sky: 'bg-sky-50 text-sky-600 dark:bg-sky-500/10 dark:text-sky-400',
}

export function StatCard({
  label,
  value,
  hint,
  icon: Icon,
  accent = 'indigo',
}: StatCardProps) {
  return (
    <article className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm transition hover:-translate-y-0.5 hover:shadow-md dark:border-slate-800 dark:bg-slate-900">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-sm font-medium text-slate-500 dark:text-slate-400">
            {label}
          </p>
          <p className="mt-3 text-3xl font-semibold tracking-tight text-slate-950 dark:text-white">
            {value}
          </p>
        </div>
        <span
          className={`flex size-10 items-center justify-center rounded-xl ${accents[accent]}`}
        >
          <Icon className="size-5" />
        </span>
      </div>
      {hint ? (
        <div className="mt-4 text-xs text-slate-500 dark:text-slate-400">
          {hint}
        </div>
      ) : null}
    </article>
  )
}
