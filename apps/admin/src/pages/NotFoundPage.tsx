import { Compass } from 'lucide-react'
import { Link } from 'react-router-dom'

export function NotFoundPage() {
  return (
    <div className="flex min-h-[60vh] flex-col items-center justify-center text-center">
      <span className="flex size-16 items-center justify-center rounded-2xl bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-400">
        <Compass className="size-7" />
      </span>
      <p className="mt-6 text-sm font-semibold text-indigo-600 dark:text-indigo-400">
        404
      </p>
      <h1 className="mt-2 text-2xl font-semibold tracking-tight text-slate-950 dark:text-white">
        页面不存在
      </h1>
      <p className="mt-2 text-sm text-slate-500 dark:text-slate-400">
        你访问的地址可能已移动或被删除。
      </p>
      <Link
        to="/"
        className="mt-6 rounded-xl bg-slate-900 px-4 py-2.5 text-sm font-medium text-white transition hover:bg-slate-700 dark:bg-white dark:text-slate-900 dark:hover:bg-slate-200"
      >
        返回概览
      </Link>
    </div>
  )
}
