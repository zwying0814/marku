import { keepPreviousData, useQuery } from '@tanstack/react-query'
import {
  ArrowUpRight,
  Globe2,
  KeyRound,
  MessageSquare,
  Plus,
  Search,
  X,
} from 'lucide-react'
import { useDeferredValue, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'

import { PageHeader } from '@/components/PageHeader'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { EmptyState } from '@/components/ui/empty-state'
import { ErrorState } from '@/components/ui/error-state'
import { Input } from '@/components/ui/form'
import { Pagination } from '@/components/ui/pagination'
import { TableSkeleton } from '@/components/ui/skeleton'
import { CreateSiteDialog } from '@/features/sites/CreateSiteDialog'
import { SiteKeyDialog } from '@/features/sites/SiteKeyDialog'
import { adminApi } from '@/lib/admin-api'
import { moderationMeta, siteStatusMeta } from '@/lib/constants'
import { queryKeys } from '@/lib/query-keys'
import { formatDate } from '@/lib/utils'
import type { SiteWithKey } from '@/types/api'

const PAGE_SIZE = 10

export function SitesPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [search, setSearch] = useState('')
  const deferredSearch = useDeferredValue(search.trim())
  const [page, setPage] = useState(1)
  const [keySite, setKeySite] = useState<SiteWithKey | null>(null)
  const createOpen = searchParams.get('create') === '1'

  const sitesQuery = useQuery({
    queryKey: queryKeys.sites({
      page,
      pageSize: PAGE_SIZE,
      q: deferredSearch,
    }),
    queryFn: () =>
      adminApi.listSites({
        page,
        pageSize: PAGE_SIZE,
        q: deferredSearch,
      }),
    placeholderData: keepPreviousData,
  })

  function setCreateOpen(open: boolean) {
    setSearchParams((current) => {
      const next = new URLSearchParams(current)
      if (open) next.set('create', '1')
      else next.delete('create')
      return next
    })
  }

  const sites = sitesQuery.data?.items ?? []
  const total = sitesQuery.data?.total ?? 0
  const hasFilters = Boolean(deferredSearch)

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow="Sites"
        title="站点管理"
        description="每个站点拥有独立的 Site Key、评论队列与互动数据。"
        actions={
          <Button type="button" onClick={() => setCreateOpen(true)}>
            <Plus className="size-4" />
            新建站点
          </Button>
        }
      />

      <section className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm dark:border-slate-800 dark:bg-slate-900">
        <div className="flex flex-col gap-3 border-b border-slate-100 p-4 sm:flex-row sm:items-center sm:justify-between dark:border-slate-800">
          <div className="relative w-full sm:max-w-sm">
            <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-slate-400" />
            <Input
              value={search}
              onChange={(event) => {
                setSearch(event.target.value)
                setPage(1)
              }}
              placeholder="搜索站点名称、slug 或域名"
              className="pl-9 pr-9"
              aria-label="搜索站点"
            />
            {search ? (
              <button
                type="button"
                aria-label="清空搜索"
                onClick={() => {
                  setSearch('')
                  setPage(1)
                }}
                className="absolute right-2 top-1/2 flex size-6 -translate-y-1/2 items-center justify-center rounded-md text-slate-400 hover:bg-slate-100 hover:text-slate-700 dark:hover:bg-slate-800"
              >
                <X className="size-3.5" />
              </button>
            ) : null}
          </div>
          <p className="text-xs text-slate-500 dark:text-slate-400">
            共 {total} 个站点
          </p>
        </div>

        {sitesQuery.isPending ? (
          <TableSkeleton rows={5} />
        ) : sitesQuery.isError ? (
          <ErrorState
            error={sitesQuery.error}
            onRetry={() => sitesQuery.refetch()}
          />
        ) : sites.length === 0 ? (
          <EmptyState
            icon={hasFilters ? Search : Globe2}
            title={hasFilters ? '没有匹配的站点' : '还没有站点'}
            description={
              hasFilters
                ? '尝试调整关键词，或清空搜索条件。'
                : '创建第一个站点，获取 Site Key 后即可接入博客。'
            }
            action={
              hasFilters ? (
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => {
                    setSearch('')
                    setPage(1)
                  }}
                >
                  清空筛选
                </Button>
              ) : (
                <Button type="button" onClick={() => setCreateOpen(true)}>
                  <Plus className="size-4" />
                  创建站点
                </Button>
              )
            }
          />
        ) : (
          <>
            <div className="hidden overflow-x-auto md:block">
              <table className="w-full text-left text-sm">
                <thead>
                  <tr className="border-b border-slate-100 text-[11px] uppercase tracking-wider text-slate-400 dark:border-slate-800 dark:text-slate-500">
                    <th className="px-5 py-3 font-semibold">站点</th>
                    <th className="px-4 py-3 font-semibold">状态</th>
                    <th className="px-4 py-3 font-semibold">审核</th>
                    <th className="px-4 py-3 font-semibold">Site Key</th>
                    <th className="px-4 py-3 font-semibold">创建时间</th>
                    <th className="px-5 py-3 text-right font-semibold">操作</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
                  {sites.map((site) => (
                    <tr
                      key={site.id}
                      className="group transition hover:bg-slate-50/80 dark:hover:bg-slate-800/40"
                    >
                      <td className="px-5 py-4">
                        <Link
                          to={`/sites/${site.id}`}
                          className="flex items-center gap-3"
                        >
                          <span className="flex size-9 shrink-0 items-center justify-center rounded-xl bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-400">
                            <Globe2 className="size-4" />
                          </span>
                          <span className="min-w-0">
                            <span className="block truncate font-medium text-slate-900 group-hover:text-indigo-600 dark:text-white dark:group-hover:text-indigo-400">
                              {site.name}
                            </span>
                            <span className="mt-0.5 block truncate font-mono text-[11px] text-slate-400">
                              {site.domain || `/${site.slug}`}
                            </span>
                          </span>
                        </Link>
                      </td>
                      <td className="px-4 py-4">
                        <Badge variant={siteStatusMeta[site.status].variant}>
                          {siteStatusMeta[site.status].label}
                        </Badge>
                      </td>
                      <td className="px-4 py-4 text-xs text-slate-600 dark:text-slate-300">
                        {moderationMeta[site.commentModeration].label}
                      </td>
                      <td className="px-4 py-4">
                        {site.siteKeyPrefix ? (
                          <span className="inline-flex items-center gap-1.5 font-mono text-[11px] text-slate-500 dark:text-slate-400">
                            <KeyRound className="size-3" />
                            {site.siteKeyPrefix}…
                          </span>
                        ) : (
                          <Badge variant="danger">已吊销</Badge>
                        )}
                      </td>
                      <td className="px-4 py-4 text-xs text-slate-500 dark:text-slate-400">
                        {formatDate(site.createdAt)}
                      </td>
                      <td className="px-5 py-4">
                        <div className="flex items-center justify-end gap-1">
                          <Link
                            to={`/comments?siteId=${site.id}`}
                            className="flex size-8 items-center justify-center rounded-lg text-slate-400 transition hover:bg-slate-100 hover:text-slate-700 dark:hover:bg-slate-800 dark:hover:text-white"
                            aria-label={`查看 ${site.name} 的评论`}
                            title="查看评论"
                          >
                            <MessageSquare className="size-4" />
                          </Link>
                          <Link
                            to={`/sites/${site.id}`}
                            className="flex size-8 items-center justify-center rounded-lg text-slate-400 transition hover:bg-slate-100 hover:text-slate-700 dark:hover:bg-slate-800 dark:hover:text-white"
                            aria-label={`打开 ${site.name} 详情`}
                            title="站点详情"
                          >
                            <ArrowUpRight className="size-4" />
                          </Link>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div className="divide-y divide-slate-100 md:hidden dark:divide-slate-800">
              {sites.map((site) => (
                <Link
                  key={site.id}
                  to={`/sites/${site.id}`}
                  className="block p-4 transition hover:bg-slate-50 dark:hover:bg-slate-800/40"
                >
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <p className="truncate font-medium text-slate-900 dark:text-white">
                        {site.name}
                      </p>
                      <p className="mt-1 truncate font-mono text-[11px] text-slate-400">
                        {site.domain || `/${site.slug}`}
                      </p>
                    </div>
                    <Badge variant={siteStatusMeta[site.status].variant}>
                      {siteStatusMeta[site.status].label}
                    </Badge>
                  </div>
                  <div className="mt-3 flex items-center justify-between text-[11px] text-slate-500 dark:text-slate-400">
                    <span>{moderationMeta[site.commentModeration].label}</span>
                    <span>{formatDate(site.createdAt)}</span>
                  </div>
                </Link>
              ))}
            </div>

            <Pagination
              page={page}
              pageSize={PAGE_SIZE}
              total={total}
              onPageChange={setPage}
            />
          </>
        )}
      </section>

      <CreateSiteDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onCreated={(site) => {
          setCreateOpen(false)
          setKeySite(site)
        }}
      />
      <SiteKeyDialog site={keySite} onClose={() => setKeySite(null)} />
    </div>
  )
}
