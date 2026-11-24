import { CloudOff, RotateCw } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { extractErrorMessage } from '@/lib/utils'

export function ErrorState({
  error,
  onRetry,
  title = '加载失败',
}: {
  error: unknown
  onRetry?: () => void
  title?: string
}) {
  return (
    <div className="flex flex-col items-center justify-center px-6 py-14 text-center">
      <span className="flex size-14 items-center justify-center rounded-2xl bg-rose-50 text-rose-600 dark:bg-rose-500/10 dark:text-rose-400">
        <CloudOff className="size-6" />
      </span>
      <h3 className="mt-4 text-sm font-semibold text-slate-950 dark:text-white">
        {title}
      </h3>
      <p className="mt-1.5 max-w-md text-sm leading-6 text-slate-500 dark:text-slate-400">
        {extractErrorMessage(error)}
      </p>
      {onRetry ? (
        <Button variant="outline" size="sm" className="mt-5" onClick={onRetry}>
          <RotateCw className="size-3.5" />
          重试
        </Button>
      ) : null}
    </div>
  )
}
