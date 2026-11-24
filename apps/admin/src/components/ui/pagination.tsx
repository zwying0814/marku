import { ChevronLeft, ChevronRight } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

type PaginationProps = {
  page: number
  pageSize: number
  total: number
  onPageChange: (page: number) => void
  className?: string
}

export function Pagination({
  page,
  pageSize,
  total,
  onPageChange,
  className,
}: PaginationProps) {
  const totalPages = Math.max(1, Math.ceil(total / pageSize))
  const pages = Array.from(
    new Set(
      [page - 1, page, page + 1].filter(
        (value) => value >= 1 && value <= totalPages,
      ),
    ),
  )

  return (
    <div
      className={cn(
        'flex flex-wrap items-center justify-between gap-3 border-t border-slate-100 px-5 py-3.5 dark:border-slate-800',
        className,
      )}
    >
      <p className="text-xs text-slate-500 dark:text-slate-400">
        共 {total} 条，第 {page} / {totalPages} 页
      </p>
      <div className="flex items-center gap-1">
        <Button
          type="button"
          variant="ghost"
          size="icon"
          aria-label="上一页"
          disabled={page <= 1}
          onClick={() => onPageChange(page - 1)}
        >
          <ChevronLeft className="size-4" />
        </Button>
        {pages.map((value) => (
          <button
            key={value}
            type="button"
            onClick={() => onPageChange(value)}
            className={cn(
              'size-9 rounded-xl text-xs font-semibold transition',
              value === page
                ? 'bg-slate-900 text-white dark:bg-white dark:text-slate-900'
                : 'text-slate-500 hover:bg-slate-100 hover:text-slate-900 dark:text-slate-400 dark:hover:bg-slate-800 dark:hover:text-white',
            )}
          >
            {value}
          </button>
        ))}
        <Button
          type="button"
          variant="ghost"
          size="icon"
          aria-label="下一页"
          disabled={page >= totalPages}
          onClick={() => onPageChange(page + 1)}
        >
          <ChevronRight className="size-4" />
        </Button>
      </div>
    </div>
  )
}
