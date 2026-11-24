import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import {
  Check,
  ChevronRight,
  Eye,
  MessageSquare,
  RotateCcw,
  Search,
  ShieldX,
  Trash2,
  X,
} from 'lucide-react'
import { useDeferredValue, useEffect, useMemo, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { toast } from 'sonner'

import { PageHeader } from '@/components/PageHeader'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ConfirmDialog } from '@/components/ui/confirm-dialog'
import { EmptyState } from '@/components/ui/empty-state'
import { ErrorState } from '@/components/ui/error-state'
import { Checkbox, Input, Select } from '@/components/ui/form'
import { Pagination } from '@/components/ui/pagination'
import { TableSkeleton } from '@/components/ui/skeleton'
import { CommentDetailDrawer } from '@/features/comments/CommentDetailDrawer'
import { adminApi } from '@/lib/admin-api'
import { commentStatusMeta } from '@/lib/constants'
import { queryKeys } from '@/lib/query-keys'
import {
  cn,
  extractErrorMessage,
  formatRelativeTime,
  truncate,
} from '@/lib/utils'
import type { AdminComment, CommentFilters, CommentStatus } from '@/types/api'

const PAGE_SIZE = 20
const statusValues: CommentStatus[] = [
  'pending',
  'approved',
  'rejected',
  'spam',
]

function readStatus(value: string | null): CommentStatus | '' {
  return statusValues.includes(value as CommentStatus)
    ? (value as CommentStatus)
    : ''
}

export function CommentsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const queryClient = useQueryClient()
  const [q, setQ] = useState('')
  const [path, setPath] = useState('')
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [page, setPage] = useState(1)
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set())
  const [detail, setDetail] = useState<AdminComment | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<AdminComment | null>(null)
  const [bulkDeleteOpen, setBulkDeleteOpen] = useState(false)
  const deferredQ = useDeferredValue(q.trim())
  const deferredPath = useDeferredValue(path.trim())
  const status = readStatus(searchParams.get('status'))
  const requestedSiteId = Number(searchParams.get('siteId')) || 0

  const sitesQuery = useQuery({
    queryKey: queryKeys.sites({ page: 1, pageSize: 100 }),
    queryFn: () => adminApi.listSites({ page: 1, pageSize: 100 }),
  })
  const overviewQuery = useQuery({
    queryKey: queryKeys.overview('7d'),
    queryFn: () => adminApi.overview('7d'),
  })

  const sites = sitesQuery.data?.items ?? []
  const selectedSite =
    sites.find((site) => site.id === requestedSiteId) ?? sites[0]
  const siteId = selectedSite?.id ?? 0

  useEffect(() => {
    if (siteId && requestedSiteId !== siteId) {
      setSearchParams(
        (current) => {
          const next = new URLSearchParams(current)
          next.set('siteId', String(siteId))
          return next
        },
        { replace: true },
      )
    }
  }, [siteId, requestedSiteId, setSearchParams])

  const filters = useMemo<CommentFilters>(
    () => ({
      status,
      path: deferredPath,
      q: deferredQ,
      from,
      to,
      page,
      pageSize: PAGE_SIZE,
      sort: 'desc',
    }),
    [status, deferredPath, deferredQ, from, to, page],
  )

  const commentsQuery = useQuery({
    queryKey: queryKeys.comments(siteId, filters),
    queryFn: () => adminApi.listComments(siteId, filters),
    enabled: Boolean(siteId),
    placeholderData: keepPreviousData,
  })

  useEffect(() => {
    setPage(1)
  }, [siteId, status, deferredQ, deferredPath, from, to])

  useEffect(() => {
    setSelectedIds(new Set())
  }, [siteId, status, deferredQ, deferredPath, from, to, page])

  useEffect(() => {
    setDetail(null)
  }, [siteId])

  const comments = commentsQuery.data?.items ?? []
  const total = commentsQuery.data?.total ?? 0
  const hasFilters = Boolean(status || deferredPath || deferredQ || from || to)
  const allSelected =
    comments.length > 0 && comments.every((item) => selectedIds.has(item.id))

  async function invalidateComments() {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ['admin', 'comments'] }),
      queryClient.invalidateQueries({ queryKey: ['admin', 'overview'] }),
    ])
  }

  const quickUpdateMutation = useMutation({
    mutationFn: ({
      id,
      nextStatus,
    }: {
      id: number
      nextStatus: CommentStatus
    }) => adminApi.updateComment(id, nextStatus),
    onSuccess: async (updated) => {
      await invalidateComments()
      setDetail((current) => (current?.id === updated.id ? updated : current))
      toast.success(`评论已${commentStatusMeta[updated.status].label}`)
    },
    onError: (error) => toast.error(extractErrorMessage(error)),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => adminApi.deleteComment(id),
    onSuccess: async () => {
      await invalidateComments()
      if (detail && deleteTarget?.id === detail.id) setDetail(null)
      setDeleteTarget(null)
      toast.success('评论已删除')
    },
    onError: (error) => toast.error(extractErrorMessage(error)),
  })

  const bulkMutation = useMutation({
    mutationFn: (action: 'approve' | 'reject' | 'spam' | 'delete') =>
      adminApi.bulkComments(siteId, Array.from(selectedIds), action),
    onSuccess: async (result, action) => {
      await invalidateComments()
      setSelectedIds(new Set())
      setBulkDeleteOpen(false)
      toast.success(
        action === 'delete'
          ? `已删除 ${result.updated} 条评论`
          : `已处理 ${result.updated} 条评论`,
      )
    },
    onError: (error) => toast.error(extractErrorMessage(error)),
  })

  function setStatusFilter(nextStatus: CommentStatus | '') {
    setSearchParams((current) => {
      const next = new URLSearchParams(current)
      if (nextStatus) next.set('status', nextStatus)
      else next.delete('status')
      if (siteId) next.set('siteId', String(siteId))
      return next
    })
    setPage(1)
  }

  function setSiteFilter(nextSiteId: number) {
    setSearchParams((current) => {
      const next = new URLSearchParams(current)
      next.set('siteId', String(nextSiteId))
      if (status) next.set('status', status)
      return next
    })
    setPage(1)
  }

  function clearFilters() {
    setQ('')
    setPath('')
    setFrom('')
    setTo('')
    setPage(1)
    setSearchParams((current) => {
      const next = new URLSearchParams(current)
      next.delete('status')
      return next
    })
  }

  function toggleSelected(id: number) {
    setSelectedIds((current) => {
      const next = new Set(current)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  function toggleAll() {
    setSelectedIds((current) => {
      const next = new Set(current)
      if (allSelected) {
        for (const item of comments) next.delete(item.id)
      } else {
        for (const item of comments) next.add(item.id)
      }
      return next
    })
  }

  if (sitesQuery.isPending) {
    return (
      <div className="space-y-6">
        <div className="h-24 animate-pulse rounded-2xl bg-slate-200 dark:bg-slate-800" />
        <div className="rounded-2xl border border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900">
          <TableSkeleton rows={8} />
        </div>
      </div>
    )
  }

  if (sitesQuery.isError) {
    return (
      <div className="rounded-2xl border border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900">
        <ErrorState
          error={sitesQuery.error}
          onRetry={() => sitesQuery.refetch()}
        />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow="Comments"
        title="评论管理"
        description="审核、回复与批量处理各站点的读者评论。"
        actions={
          <>
            {(overviewQuery.data?.pendingComments ?? 0) > 0 ? (
              <button
                type="button"
                onClick={() => setStatusFilter('pending')}
                className="inline-flex h-10 items-center gap-2 rounded-xl border border-amber-200 bg-amber-50 px-4 text-sm font-medium text-amber-800 transition hover:bg-amber-100 dark:border-amber-500/20 dark:bg-amber-500/10 dark:text-amber-200"
              >
                <MessageSquare className="size-4" />
                {overviewQuery.data?.pendingComments} 条待审核
              </button>
            ) : null}
          </>
        }
      />

      {!selectedSite ? (
        <section className="rounded-2xl border border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900">
          <EmptyState
            icon={MessageSquare}
            title="还没有可管理的站点"
            description="创建站点并接入评论后，评论会汇总在这里。"
            action={
              <Link
                to="/sites?create=1"
                className="inline-flex h-10 items-center rounded-xl bg-indigo-600 px-4 text-sm font-medium text-white hover:bg-indigo-500"
              >
                创建站点
              </Link>
            }
          />
        </section>
      ) : (
        <>
          <section className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm dark:border-slate-800 dark:bg-slate-900">
            <div className="grid gap-3 lg:grid-cols-12">
              <div className="lg:col-span-3">
                <label
                  htmlFor="comment-site"
                  className="mb-1.5 block text-[11px] font-semibold uppercase tracking-wider text-slate-400"
                >
                  站点
                </label>
                <Select
                  id="comment-site"
                  value={siteId}
                  onChange={(event) =>
                    setSiteFilter(Number(event.target.value))
                  }
                >
                  {sites.map((site) => (
                    <option key={site.id} value={site.id}>
                      {site.name}
                    </option>
                  ))}
                </Select>
              </div>
              <div className="lg:col-span-2">
                <label
                  htmlFor="comment-status"
                  className="mb-1.5 block text-[11px] font-semibold uppercase tracking-wider text-slate-400"
                >
                  状态
                </label>
                <Select
                  id="comment-status"
                  value={status}
                  onChange={(event) =>
                    setStatusFilter(event.target.value as CommentStatus | '')
                  }
                >
                  <option value="">全部状态</option>
                  {statusValues.map((value) => (
                    <option key={value} value={value}>
                      {commentStatusMeta[value].label}
                    </option>
                  ))}
                </Select>
              </div>
              <div className="lg:col-span-3">
                <label
                  htmlFor="comment-path"
                  className="mb-1.5 block text-[11px] font-semibold uppercase tracking-wider text-slate-400"
                >
                  页面路径
                </label>
                <Input
                  id="comment-path"
                  value={path}
                  onChange={(event) => setPath(event.target.value)}
                  placeholder="/posts/hello"
                  className="font-mono text-xs"
                />
              </div>
              <div className="lg:col-span-4">
                <label
                  htmlFor="comment-search"
                  className="mb-1.5 block text-[11px] font-semibold uppercase tracking-wider text-slate-400"
                >
                  关键词
                </label>
                <div className="relative">
                  <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-slate-400" />
                  <Input
                    id="comment-search"
                    value={q}
                    onChange={(event) => setQ(event.target.value)}
                    placeholder="作者、邮箱或内容"
                    className="pl-9"
                  />
                </div>
              </div>

              <div className="lg:col-span-3">
                <label
                  htmlFor="comment-from"
                  className="mb-1.5 block text-[11px] font-semibold uppercase tracking-wider text-slate-400"
                >
                  开始日期
                </label>
                <Input
                  id="comment-from"
                  type="date"
                  value={from}
                  max={to || undefined}
                  onChange={(event) => setFrom(event.target.value)}
                />
              </div>
              <div className="lg:col-span-3">
                <label
                  htmlFor="comment-to"
                  className="mb-1.5 block text-[11px] font-semibold uppercase tracking-wider text-slate-400"
                >
                  结束日期
                </label>
                <Input
                  id="comment-to"
                  type="date"
                  value={to}
                  min={from || undefined}
                  onChange={(event) => setTo(event.target.value)}
                />
              </div>
              <div className="flex items-end lg:col-span-3">
                <Button
                  type="button"
                  variant="outline"
                  onClick={clearFilters}
                  disabled={!hasFilters}
                  className="w-full"
                >
                  <RotateCcw className="size-4" />
                  重置筛选
                </Button>
              </div>
            </div>
          </section>

          {selectedIds.size > 0 ? (
            <div className="flex flex-wrap items-center gap-2 rounded-2xl border border-indigo-200 bg-indigo-50 px-4 py-3 dark:border-indigo-500/20 dark:bg-indigo-500/10">
              <span className="mr-2 text-xs font-semibold text-indigo-900 dark:text-indigo-100">
                已选择 {selectedIds.size} 条
              </span>
              <Button
                type="button"
                size="sm"
                variant="outline"
                loading={
                  bulkMutation.isPending && bulkMutation.variables === 'approve'
                }
                onClick={() => bulkMutation.mutate('approve')}
              >
                <Check className="size-3.5" />
                批量通过
              </Button>
              <Button
                type="button"
                size="sm"
                variant="outline"
                loading={
                  bulkMutation.isPending && bulkMutation.variables === 'reject'
                }
                onClick={() => bulkMutation.mutate('reject')}
              >
                <X className="size-3.5" />
                批量拒绝
              </Button>
              <Button
                type="button"
                size="sm"
                variant="outline"
                loading={
                  bulkMutation.isPending && bulkMutation.variables === 'spam'
                }
                onClick={() => bulkMutation.mutate('spam')}
              >
                <ShieldX className="size-3.5" />
                标记垃圾
              </Button>
              <Button
                type="button"
                size="sm"
                variant="danger"
                className="ml-auto"
                onClick={() => setBulkDeleteOpen(true)}
              >
                <Trash2 className="size-3.5" />
                批量删除
              </Button>
            </div>
          ) : null}

          <section className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm dark:border-slate-800 dark:bg-slate-900">
            <div className="flex items-center justify-between gap-3 border-b border-slate-100 px-5 py-3.5 dark:border-slate-800">
              <p className="text-xs text-slate-500 dark:text-slate-400">
                共 {total} 条评论
              </p>
              <p className="hidden text-[11px] text-slate-400 sm:block">
                点击行查看完整内容与回复
              </p>
            </div>

            {commentsQuery.isPending ? (
              <TableSkeleton rows={8} />
            ) : commentsQuery.isError ? (
              <ErrorState
                error={commentsQuery.error}
                onRetry={() => commentsQuery.refetch()}
              />
            ) : comments.length === 0 ? (
              <EmptyState
                icon={hasFilters ? Search : MessageSquare}
                title={hasFilters ? '没有匹配的评论' : '暂无评论'}
                description={
                  hasFilters
                    ? '尝试调整站点、状态或时间范围。'
                    : '博客接入评论接口后，新评论会显示在这里。'
                }
                action={
                  hasFilters ? (
                    <Button
                      type="button"
                      variant="outline"
                      onClick={clearFilters}
                    >
                      清空筛选
                    </Button>
                  ) : null
                }
              />
            ) : (
              <>
                <div className="hidden overflow-x-auto md:block">
                  <table className="w-full text-left text-sm">
                    <thead>
                      <tr className="border-b border-slate-100 text-[11px] uppercase tracking-wider text-slate-400 dark:border-slate-800 dark:text-slate-500">
                        <th className="w-12 px-5 py-3">
                          <Checkbox
                            checked={allSelected}
                            onChange={toggleAll}
                            aria-label="选择当前页全部评论"
                          />
                        </th>
                        <th className="px-3 py-3 font-semibold">评论</th>
                        <th className="px-3 py-3 font-semibold">页面</th>
                        <th className="px-3 py-3 font-semibold">状态</th>
                        <th className="px-3 py-3 font-semibold">时间</th>
                        <th className="px-5 py-3 text-right font-semibold">
                          操作
                        </th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
                      {comments.map((comment) => (
                        <tr
                          key={comment.id}
                          tabIndex={0}
                          onClick={() => setDetail(comment)}
                          onKeyDown={(event) => {
                            if (event.key === 'Enter' || event.key === ' ') {
                              event.preventDefault()
                              setDetail(comment)
                            }
                          }}
                          className={cn(
                            'cursor-pointer transition hover:bg-slate-50/80 dark:hover:bg-slate-800/40',
                            selectedIds.has(comment.id) &&
                              'bg-indigo-50/50 dark:bg-indigo-500/5',
                          )}
                        >
                          <td
                            className="px-5 py-4"
                            onClick={(event) => event.stopPropagation()}
                          >
                            <Checkbox
                              checked={selectedIds.has(comment.id)}
                              onChange={() => toggleSelected(comment.id)}
                              aria-label={`选择 ${comment.authorName} 的评论`}
                            />
                          </td>
                          <td className="max-w-sm px-3 py-4">
                            <div className="flex items-start gap-3">
                              <span className="flex size-8 shrink-0 items-center justify-center rounded-full bg-slate-100 text-[10px] font-bold uppercase text-slate-500 dark:bg-slate-800 dark:text-slate-300">
                                {comment.authorName.slice(0, 2)}
                              </span>
                              <div className="min-w-0">
                                <div className="flex items-center gap-1.5">
                                  <span className="truncate text-xs font-semibold text-slate-900 dark:text-white">
                                    {comment.authorName}
                                  </span>
                                  {comment.isAdmin ? (
                                    <Badge variant="info">管理员</Badge>
                                  ) : null}
                                </div>
                                <p className="mt-1 line-clamp-2 text-xs leading-5 text-slate-500 dark:text-slate-400">
                                  {truncate(comment.content, 100)}
                                </p>
                              </div>
                            </div>
                          </td>
                          <td className="max-w-48 px-3 py-4">
                            <span
                              className="block truncate font-mono text-[11px] text-slate-500 dark:text-slate-400"
                              title={comment.path}
                            >
                              {comment.path}
                            </span>
                          </td>
                          <td className="px-3 py-4">
                            <Badge
                              variant={
                                commentStatusMeta[comment.status].variant
                              }
                            >
                              {commentStatusMeta[comment.status].label}
                            </Badge>
                          </td>
                          <td className="whitespace-nowrap px-3 py-4 text-xs text-slate-500 dark:text-slate-400">
                            {formatRelativeTime(comment.createdAt)}
                          </td>
                          <td
                            className="px-5 py-4"
                            onClick={(event) => event.stopPropagation()}
                          >
                            <div className="flex items-center justify-end gap-1">
                              {comment.status !== 'approved' ? (
                                <button
                                  type="button"
                                  aria-label="通过评论"
                                  title="通过"
                                  disabled={quickUpdateMutation.isPending}
                                  onClick={() =>
                                    quickUpdateMutation.mutate({
                                      id: comment.id,
                                      nextStatus: 'approved',
                                    })
                                  }
                                  className="flex size-8 items-center justify-center rounded-lg text-slate-400 transition hover:bg-emerald-50 hover:text-emerald-600 dark:hover:bg-emerald-500/10 dark:hover:text-emerald-400"
                                >
                                  <Check className="size-4" />
                                </button>
                              ) : null}
                              <button
                                type="button"
                                aria-label="查看评论详情"
                                title="查看详情"
                                onClick={() => setDetail(comment)}
                                className="flex size-8 items-center justify-center rounded-lg text-slate-400 transition hover:bg-slate-100 hover:text-slate-700 dark:hover:bg-slate-800 dark:hover:text-white"
                              >
                                <Eye className="size-4" />
                              </button>
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>

                <div className="divide-y divide-slate-100 md:hidden dark:divide-slate-800">
                  {comments.map((comment) => (
                    <div
                      key={comment.id}
                      className={cn(
                        'p-4',
                        selectedIds.has(comment.id) &&
                          'bg-indigo-50/50 dark:bg-indigo-500/5',
                      )}
                    >
                      <div className="flex items-start gap-3">
                        <Checkbox
                          checked={selectedIds.has(comment.id)}
                          onChange={() => toggleSelected(comment.id)}
                          aria-label={`选择 ${comment.authorName} 的评论`}
                          className="mt-1"
                        />
                        <button
                          type="button"
                          className="min-w-0 flex-1 text-left"
                          onClick={() => setDetail(comment)}
                        >
                          <div className="flex items-center justify-between gap-2">
                            <span className="truncate text-xs font-semibold text-slate-900 dark:text-white">
                              {comment.authorName}
                            </span>
                            <Badge
                              variant={
                                commentStatusMeta[comment.status].variant
                              }
                            >
                              {commentStatusMeta[comment.status].label}
                            </Badge>
                          </div>
                          <p className="mt-1.5 line-clamp-2 text-xs leading-5 text-slate-500 dark:text-slate-400">
                            {truncate(comment.content, 120)}
                          </p>
                          <p className="mt-2 truncate font-mono text-[10px] text-slate-400">
                            {comment.path}
                          </p>
                        </button>
                        <ChevronRight className="mt-1 size-4 shrink-0 text-slate-300" />
                      </div>
                    </div>
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
        </>
      )}

      <CommentDetailDrawer
        comment={detail}
        onClose={() => setDetail(null)}
        onUpdated={setDetail}
        onDelete={(comment) => setDeleteTarget(comment)}
      />
      <ConfirmDialog
        open={Boolean(deleteTarget)}
        onClose={() => setDeleteTarget(null)}
        onConfirm={() => {
          if (deleteTarget) deleteMutation.mutate(deleteTarget.id)
        }}
        loading={deleteMutation.isPending}
        title="删除这条评论？"
        description={`将永久删除 ${deleteTarget?.authorName ?? ''} 的评论及其在列表中的记录。此操作不可撤销。`}
        confirmLabel="删除评论"
      />
      <ConfirmDialog
        open={bulkDeleteOpen}
        onClose={() => setBulkDeleteOpen(false)}
        onConfirm={() => bulkMutation.mutate('delete')}
        loading={bulkMutation.isPending && bulkMutation.variables === 'delete'}
        title={`批量删除 ${selectedIds.size} 条评论？`}
        description="所选评论将被永久删除，此操作不可撤销。"
        confirmLabel="批量删除"
      />
    </div>
  )
}
