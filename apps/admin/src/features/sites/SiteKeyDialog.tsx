import { KeyRound, ShieldAlert } from 'lucide-react'

import { CopyButton } from '@/components/ui/copy-button'
import { Modal } from '@/components/ui/modal'
import type { SiteWithKey } from '@/types/api'

export function SiteKeyDialog({
  site,
  onClose,
}: {
  site: SiteWithKey | null
  onClose: () => void
}) {
  return (
    <Modal
      open={Boolean(site)}
      onClose={onClose}
      title="Site Key 已生成"
      description="这是唯一一次完整展示该 Key，请立即复制并保存到安全位置。"
      size="lg"
      footer={
        <>
          <CopyButton
            value={site?.siteKey ?? ''}
            label="复制 Site Key"
            size="md"
            variant="primary"
          />
          <button
            type="button"
            onClick={onClose}
            className="inline-flex h-10 items-center rounded-xl border border-slate-200 bg-white px-4 text-sm font-medium text-slate-700 transition hover:bg-slate-50 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-200 dark:hover:bg-slate-800"
          >
            我已保存
          </button>
        </>
      }
    >
      <div className="space-y-5">
        <div className="flex items-center gap-3 rounded-2xl border border-indigo-100 bg-indigo-50/70 p-4 dark:border-indigo-500/20 dark:bg-indigo-500/10">
          <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-white text-indigo-600 shadow-sm dark:bg-slate-900 dark:text-indigo-400">
            <KeyRound className="size-5" />
          </span>
          <div>
            <p className="text-sm font-semibold text-slate-900 dark:text-white">
              {site?.name}
            </p>
            <p className="mt-0.5 text-xs text-slate-500 dark:text-slate-400">
              Slug: {site?.slug}
            </p>
          </div>
        </div>

        <div>
          <p className="mb-2 text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">
            X-Site-Key
          </p>
          <div className="flex items-start gap-2 rounded-2xl border border-slate-200 bg-slate-950 p-4 dark:border-slate-700">
            <code className="min-w-0 flex-1 break-all font-mono text-xs leading-6 text-emerald-300">
              {site?.siteKey}
            </code>
            <CopyButton
              value={site?.siteKey ?? ''}
              showLabel={false}
              variant="ghost"
              className="shrink-0 text-slate-400 hover:bg-white/10 hover:text-white"
            />
          </div>
        </div>

        <div className="flex gap-3 rounded-2xl border border-amber-200 bg-amber-50 p-4 text-amber-800 dark:border-amber-500/20 dark:bg-amber-500/10 dark:text-amber-200">
          <ShieldAlert className="mt-0.5 size-4 shrink-0" />
          <p className="text-xs leading-6">
            服务端只保存 Key 的哈希与前后缀。关闭此窗口后无法再次查看完整
            Key；如遗失，请轮换生成新 Key。
          </p>
        </div>
      </div>
    </Modal>
  )
}
