import { Braces, Code2, FileCode2, Layers3 } from 'lucide-react'
import { useState } from 'react'

import { CopyButton } from '@/components/ui/copy-button'
import { API_BASE_URL } from '@/lib/api'
import { cn } from '@/lib/utils'
import type { Site } from '@/types/api'

type SnippetTab = 'nextjs' | 'html' | 'embed'

const tabs: Array<{
  id: SnippetTab
  label: string
  icon: typeof Code2
}> = [
  { id: 'nextjs', label: 'Next.js', icon: Code2 },
  { id: 'html', label: 'HTML / JS', icon: FileCode2 },
  { id: 'embed', label: '@marku/embed', icon: Layers3 },
]

function buildSnippets(site: Site) {
  const siteKey = 'YOUR_SITE_KEY'
  const apiBase = API_BASE_URL

  return {
    nextjs: `// app/blog/[slug]/marku-comments.tsx
const MARKU_API = '${apiBase}'
const SITE_KEY = '${siteKey}'

export async function getComments(path: string) {
  const response = await fetch(
    \`\${MARKU_API}/public/comments?path=\${encodeURIComponent(path)}&sort=asc\`,
    { headers: { 'X-Site-Key': SITE_KEY } },
  )
  if (!response.ok) throw new Error('Failed to load comments')
  const { data } = await response.json()
  return data.items
}

export async function getPageStats(path: string) {
  const response = await fetch(
    \`\${MARKU_API}/public/stats?paths=\${encodeURIComponent(path)}\`,
    { headers: { 'X-Site-Key': SITE_KEY } },
  )
  const { data } = await response.json()
  return data.items[0]
}`,
    html: `<script>
  const MARKU_API = '${apiBase}'
  const SITE_KEY = '${siteKey}'

  async function marku(path, options = {}) {
    const response = await fetch(MARKU_API + path, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        'X-Site-Key': SITE_KEY,
        ...options.headers,
      },
    })
    const body = await response.json()
    if (!response.ok) throw new Error(body.message)
    return body.data
  }

  // 评论列表
  marku('/public/comments?path=' + encodeURIComponent(location.pathname))

  // 记录浏览量
  marku('/public/views', {
    method: 'POST',
    body: JSON.stringify({ path: location.pathname }),
  })
</script>`,
    embed: `<!-- 可选集成包，尚未纳入 v0.1 核心验收 -->
<script type="module" src="/marku-embed.js"></script>

<marku-comments
  api-base="${apiBase}"
  site-key="${siteKey}"
  path="${site.domain ? `/${site.slug}` : '/posts/hello'}"
></marku-comments>

<style>
  marku-comments::part(container) {
    --marku-accent: #6366f1;
    --marku-radius: 12px;
    font-family: inherit;
  }
</style>`,
  }
}

export function IntegrationSnippets({ site }: { site: Site }) {
  const [tab, setTab] = useState<SnippetTab>('nextjs')
  const snippets = buildSnippets(site)
  const code = snippets[tab]

  return (
    <section className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm dark:border-slate-800 dark:bg-slate-900">
      <div className="flex flex-col gap-3 border-b border-slate-100 px-5 py-4 sm:flex-row sm:items-center sm:justify-between dark:border-slate-800">
        <div>
          <h2 className="text-sm font-semibold text-slate-950 dark:text-white">
            接入代码
          </h2>
          <p className="mt-0.5 text-xs text-slate-500 dark:text-slate-400">
            使用站点专属 Site Key 调用公共 API。
          </p>
        </div>
        <CopyButton value={code} label="复制代码" />
      </div>

      <div className="flex gap-1 overflow-x-auto border-b border-slate-100 px-4 py-2 dark:border-slate-800">
        {tabs.map((item) => (
          <button
            key={item.id}
            type="button"
            onClick={() => setTab(item.id)}
            className={cn(
              'inline-flex shrink-0 items-center gap-1.5 rounded-lg px-3 py-2 text-xs font-medium transition',
              tab === item.id
                ? 'bg-slate-900 text-white dark:bg-white dark:text-slate-900'
                : 'text-slate-500 hover:bg-slate-100 hover:text-slate-900 dark:text-slate-400 dark:hover:bg-slate-800 dark:hover:text-white',
            )}
          >
            <item.icon className="size-3.5" />
            {item.label}
          </button>
        ))}
      </div>

      <div className="relative bg-slate-950 p-5">
        <div className="mb-3 flex items-center gap-2 text-[11px] text-slate-500">
          <Braces className="size-3.5" />
          {tab === 'nextjs'
            ? 'Server Component / Route Handler'
            : tab === 'html'
              ? '纯静态页面'
              : '可选集成（v0.2）'}
        </div>
        <pre className="overflow-x-auto font-mono text-[11px] leading-6 text-slate-300">
          <code>{code}</code>
        </pre>
      </div>

      <div className="border-t border-slate-100 px-5 py-3 text-[11px] leading-5 text-slate-500 dark:border-slate-800 dark:text-slate-400">
        当前 Key 前缀：
        <code className="ml-1 font-mono text-slate-700 dark:text-slate-200">
          {site.siteKeyPrefix || '未配置'}
        </code>
        。完整 Key 只在创建或轮换时展示一次。
      </div>
    </section>
  )
}
