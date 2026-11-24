import { LoaderCircle } from 'lucide-react'

export function PageLoader() {
  return (
    <div className="flex min-h-[55vh] items-center justify-center">
      <div className="flex flex-col items-center gap-3 text-slate-400 dark:text-slate-500">
        <LoaderCircle className="size-6 animate-spin" />
        <p className="text-xs font-medium">加载页面…</p>
      </div>
    </div>
  )
}
