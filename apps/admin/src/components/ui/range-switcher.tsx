import type { RangeKey } from '@/types/api'
import { cn } from '@/lib/utils'

export function RangeSwitcher({
  value,
  onChange,
}: {
  value: RangeKey
  onChange: (value: RangeKey) => void
}) {
  return (
    <div
      role="group"
      aria-label="统计范围"
      className="inline-flex rounded-xl border border-slate-200 bg-white p-1 shadow-sm dark:border-slate-700 dark:bg-slate-900"
    >
      {(['7d', '30d'] as const).map((range) => (
        <button
          key={range}
          type="button"
          aria-pressed={value === range}
          onClick={() => onChange(range)}
          className={cn(
            'rounded-lg px-3 py-1.5 text-xs font-semibold transition',
            value === range
              ? 'bg-slate-900 text-white dark:bg-white dark:text-slate-900'
              : 'text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-white',
          )}
        >
          {range === '7d' ? '近 7 天' : '近 30 天'}
        </button>
      ))}
    </div>
  )
}
