import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Activity,
  ArrowLeft,
  BarChart3,
  Eye,
  Globe2,
  Heart,
  KeyRound,
  MessageSquare,
  Pencil,
  Power,
  RefreshCw,
  ShieldOff,
  Trash2,
} from 'lucide-react'
import { useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { toast } from 'sonner'

import { StatCard } from '@/components/StatCard'
import { TrendChart } from '@/components/TrendChart'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ConfirmDialog } from '@/components/ui/confirm-dialog'
import { ErrorState } from '@/components/ui/error-state'
import { RangeSwitcher } from '@/components/ui/range-switcher'
import { Skeleton } from '@/components/ui/skeleton'
import { EditSiteDialog } from '@/features/sites/EditSiteDialog'
import { IntegrationSnippets } from '@/features/sites/IntegrationSnippets'
import { SiteKeyDialog } from '@/features/sites/SiteKeyDialog'
import { adminApi } from '@/lib/admin-api'
import { moderationMeta, siteStatusMeta } from '@/lib/constants'
import { queryKeys } from '@/lib/query-keys'
import {
  cn,
  extractErrorMessage,
  formatCompact,
  formatDate,
  formatNumber,
} from '@/lib/utils'
import type { RangeKey, SiteWithKey } from '@/types/api'

export function SiteDetailPage() {
  const params = useParams()
  const siteId = Number(params.siteId)
  const validId = Number.isInteger(siteId) && siteId > 0
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [range, setRange] = useState<RangeKey>('7d')
  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [rotateOpen, setRotateOpen] = useState(false)
  const [revokeOpen, setRevokeOpen] = useState(false)
  const [keySite, setKeySite] = useState<SiteWithKey | null>(null)

  const siteQuery = useQuery({
    queryKey: queryKeys.site(siteId),
    queryFn: () => adminApi.getSite(siteId),
    enabled: validId,
  })
  const statsQuery = useQuery({
    queryKey: queryKeys.siteStats(siteId, range),
    queryFn: () => adminApi.siteStats(siteId, range, 8),
    enabled: validId,
  })
  const likesQuery = useQuery({
    queryKey: queryKeys.siteLikes(siteId, range),
    queryFn: () => adminApi.siteLikes(siteId, range, 8),
    enabled: validId,
  })
  const pendingQuery = useQuery({
    queryKey: queryKeys.comments(siteId, {
      status: 'pending',
      page: 1,
      pageSize: 1,
    }),
    queryFn: () =>
      adminApi.listComments(siteId, {
        status: 'pending',
        page: 1,
        pageSize: 1,
      }),
    enabled: validId,
  })

  const invalidateSite = async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: queryKeys.site(siteId) }),
      queryClient.invalidateQueries({ queryKey: queryKeys.siteList }),
      queryClient.invalidateQueries({ queryKey: ['admin', 'overview'] }),
    ])
  }

  const toggleStatusMutation = useMutation({
    mutationFn: (status: 'active' | 'disabled') =>
      adminApi.updateSite(siteId, { status }),
    onSuccess: async (site) => {
      queryClient.setQueryData(queryKeys.site(siteId), site)
      await invalidateSite()
      toast.success(site.status === 'active' ? '站点已启用' : '站点已停用')
    },
    onError: (error) => toast.error(extractErrorMessage(error)),
  })

  const deleteMutation = useMutation({
    mutationFn: () => adminApi.deleteSite(siteId),
    onSuccess: async () => {
      await invalidateSite()
      toast.success('站点已删除')
      navigate('/sites', { replace: true })
    },
    onError: (error) => toast.error(extractErrorMessage(error)),
  })

  const rotateMutation = useMutation({
    mutationFn: () => adminApi.rotateSiteKey(siteId),
    onSuccess: async (site) => {
      queryClient.setQueryData(queryKeys.site(siteId), site)
      await invalidateSite()
      setRotateOpen(false)
      setKeySite(site)
      toast.success('Site Key 已轮换，旧 Key 已立即失效')
    },
    onError: (error) => toast.error(extractErrorMessage(error)),
  })

  const revokeMutation = useMutation({
    mutationFn: () => adminApi.revokeSiteKey(siteId),
    onSuccess: async () => {
      await invalidateSite()
      setRevokeOpen(false)
      toast.success('Site Key 已吊销')
    },
    onError: (error) => toast.error(extractErrorMessage(error)),
  })

  if (!validId) {
    return (
      <div className="rounded-2xl border border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900">
        <ErrorState error={new Error('无效的站点 ID')} />
      </div>
    )
  }

  if (siteQuery.isPending) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-24 rounded-2xl" />
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
          {Array.from({ length: 4 }).map((_, index) => (
            <Skeleton key={index} className="h-36 rounded-2xl" />
          ))}
        </div>
        <Skeleton className="h-80 rounded-2xl" />
      </div>
    )
  }

  if (siteQuery.isError) {
    return (
      <div className="space-y-4">
        <Link
          to="/sites"
          className="inline-flex items-center gap-1.5 text-sm font-medium text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-white"
        >
          <ArrowLeft className="size-4" />
          返回站点列表
        </Link>
        <div className="rounded-2xl border border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900">
          <ErrorState
            error={siteQuery.error}
            onRetry={() => siteQuery.refetch()}
          />
        </div>
      </div>
    )
  }

  const site = siteQuery.data
  const stats = statsQuery.data
  const likes = likesQuery.data
  const likeByDate = new Map(likes?.trend.map((point) => [point.date, point]))
  const trendData =
    stats?.trend.map((point) => ({
      date: point.date,
      views: Number(point.views),
      uniqueViews: Number(point.uniqueViews),
      likes: Number(likeByDate.get(point.date)?.likes ?? 0),
    })) ?? []

  return (
    <div className="space-y-6">
      <div>
        <Link
          to="/sites"
          className="inline-flex items-center gap-1.5 text-sm font-medium text-slate-500 transition hover:text-slate-900 dark:text-slate-400 dark:hover:text-white"
        >
          <ArrowLeft className="size-4" />
          返回站点列表
        </Link>
      </div>

      <div className="flex flex-col justify-between gap-5 sm:flex-row sm:items-end">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <span className="flex size-10 items-center justify-center rounded-xl bg-indigo-50 text-indigo-600 dark:bg-indigo-500/10 dark:text-indigo-400">
              <Globe2 className="size-5" />
            </span>
            <h1 className="text-2xl font-semibold tracking-tight text-slate-950 sm:text-3xl dark:text-white">
              {site.name}
            </h1>
            <Badge variant={siteStatusMeta[site.status].variant}>
              {siteStatusMeta[site.status].label}
            </Badge>
          </div>
          <p className="mt-2 font-mono text-xs text-slate-500 dark:text-slate-400">
            {site.domain || '未填写主域名'} · /{site.slug}
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Link
            to={`/comments?siteId=${site.id}`}
            className="inline-flex h-10 items-center gap-2 rounded-xl border border-slate-200 bg-white px-4 text-sm font-medium text-slate-700 shadow-sm transition hover:bg-slate-50 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-200 dark:hover:bg-slate-800"
          >
            <MessageSquare className="size-4" />
            审核评论
          </Link>
          <Button type="button" onClick={() => setEditOpen(true)}>
            <Pencil className="size-4" />
            编辑站点
          </Button>
        </div>
      </div>

      {site.status === 'disabled' ? (
        <div className="flex flex-col justify-between gap-3 rounded-2xl border border-rose-200 bg-rose-50 px-5 py-4 sm:flex-row sm:items-center dark:border-rose-500/20 dark:bg-rose-500/10">
          <div className="flex items-center gap-3">
            <ShieldOff className="size-5 text-rose-600 dark:text-rose-400" />
            <div>
              <p className="text-sm font-semibold text-rose-900 dark:text-rose-100">
                站点已停用
              </p>
              <p className="mt-0.5 text-xs text-rose-700 dark:text-rose-300">
                所有公共 API 请求都会被拒绝，重新启用后恢复。
              </p>
            </div>
          </div>
          <Button
            type="button"
            variant="outline"
            size="sm"
            loading={toggleStatusMutation.isPending}
            onClick={() => toggleStatusMutation.mutate('active')}
          >
            <Power className="size-3.5" />
            重新启用
          </Button>
        </div>
      ) : null}

      <section className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-sm font-semibold text-slate-950 dark:text-white">
            核心指标
          </h2>
          <p className="mt-0.5 text-xs text-slate-500 dark:text-slate-400">
            点赞与浏览量为累计值，趋势由右侧时间范围控制。
          </p>
        </div>
        <RangeSwitcher value={range} onChange={setRange} />
      </section>

      <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard
          label="累计浏览量"
          value={formatCompact(stats?.totalViews)}
          hint={`独立访客 ${formatNumber(stats?.totalUniqueViews)}`}
          icon={Eye}
          accent="sky"
        />
        <StatCard
          label="累计点赞"
          value={formatCompact(likes?.totalLikes)}
          hint="所有页面合计"
          icon={Heart}
          accent="rose"
        />
        <StatCard
          label="待审核评论"
          value={formatNumber(pendingQuery.data?.total)}
          hint="需要管理员处理"
          icon={MessageSquare}
          accent="amber"
        />
        <StatCard
          label="评论审核模式"
          value={moderationMeta[site.commentModeration].label}
          hint={
            site.commentModeration === 'pending'
              ? '新评论需人工通过'
              : '新评论自动公开'
          }
          icon={Activity}
          accent="emerald"
        />
      </section>

      <section className="grid gap-6 xl:grid-cols-[1.45fr_0.55fr]">
        <div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm dark:border-slate-800 dark:bg-slate-900">
          <div className="flex items-center justify-between gap-3">
            <div>
              <h2 className="text-sm font-semibold text-slate-950 dark:text-white">
                互动趋势
              </h2>
              <p className="mt-0.5 text-xs text-slate-500 dark:text-slate-400">
                浏览量、独立访客与点赞的每日变化。
              </p>
            </div>
            <BarChart3 className="size-4 text-slate-400" />
          </div>
          {statsQuery.isPending ? (
            <Skeleton className="mt-5 h-64" />
          ) : statsQuery.isError ? (
            <ErrorState
              error={statsQuery.error}
              onRetry={() => statsQuery.refetch()}
            />
          ) : (
            <div className="mt-4">
              <TrendChart
                data={trendData}
                dualAxis
                series={[
                  {
                    key: 'views',
                    label: '浏览量',
                    color: '#6366f1',
                    type: 'area',
                  },
                  {
                    key: 'uniqueViews',
                    label: '独立访客',
                    color: '#0ea5e9',
                    type: 'line',
                  },
                  {
                    key: 'likes',
                    label: '点赞',
                    color: '#f43f5e',
                    type: 'line',
                    yAxis: 'right',
                  },
                ]}
              />
            </div>
          )}
        </div>

        <div className="space-y-4">
          <div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm dark:border-slate-800 dark:bg-slate-900">
            <h2 className="text-sm font-semibold text-slate-950 dark:text-white">
              站点信息
            </h2>
            <dl className="mt-4 space-y-3.5 text-xs">
              <div className="flex items-start justify-between gap-3">
                <dt className="text-slate-500 dark:text-slate-400">Slug</dt>
                <dd className="font-mono text-slate-800 dark:text-slate-200">
                  {site.slug}
                </dd>
              </div>
              <div className="flex items-start justify-between gap-3">
                <dt className="text-slate-500 dark:text-slate-400">Site Key</dt>
                <dd className="font-mono text-slate-800 dark:text-slate-200">
                  {site.siteKeyPrefix ? `${site.siteKeyPrefix}…` : '已吊销'}
                </dd>
              </div>
              <div className="flex items-start justify-between gap-3">
                <dt className="text-slate-500 dark:text-slate-400">创建时间</dt>
                <dd className="text-slate-800 dark:text-slate-200">
                  {formatDate(site.createdAt)}
                </dd>
              </div>
              <div className="flex items-start justify-between gap-3">
                <dt className="text-slate-500 dark:text-slate-400">最近更新</dt>
                <dd className="text-slate-800 dark:text-slate-200">
                  {formatDate(site.updatedAt, true)}
                </dd>
              </div>
            </dl>
            <div className="mt-5 border-t border-slate-100 pt-4 dark:border-slate-800">
              <p className="text-xs font-medium text-slate-500 dark:text-slate-400">
                允许来源
              </p>
              <div className="mt-2 flex flex-wrap gap-1.5">
                {site.allowedOrigins.length === 0 ? (
                  <Badge variant="warning">未限制（建议配置正式域名）</Badge>
                ) : (
                  site.allowedOrigins.map((origin) => (
                    <Badge key={origin} variant="outline" className="font-mono">
                      {origin}
                    </Badge>
                  ))
                )}
              </div>
            </div>
            <Link
              to={`/analytics?siteId=${site.id}`}
              className="mt-5 inline-flex items-center gap-1.5 text-xs font-semibold text-indigo-600 hover:text-indigo-500 dark:text-indigo-400"
            >
              查看完整数据
              <BarChart3 className="size-3.5" />
            </Link>
          </div>

          <div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm dark:border-slate-800 dark:bg-slate-900">
            <h2 className="text-sm font-semibold text-slate-950 dark:text-white">
              热门页面
            </h2>
            <div className="mt-3 space-y-2.5">
              {stats?.topPages.length ? (
                stats.topPages.slice(0, 5).map((page) => (
                  <div
                    key={page.path}
                    className="flex items-center justify-between gap-3 text-xs"
                  >
                    <span
                      className="min-w-0 truncate font-mono text-slate-600 dark:text-slate-300"
                      title={page.path}
                    >
                      {page.path}
                    </span>
                    <span className="shrink-0 font-semibold text-slate-900 dark:text-white">
                      {formatNumber(page.views)}
                    </span>
                  </div>
                ))
              ) : (
                <p className="text-xs text-slate-500 dark:text-slate-400">
                  还没有浏览量数据。
                </p>
              )}
            </div>
          </div>
        </div>
      </section>

      <IntegrationSnippets site={site} />

      <section className="rounded-2xl border border-slate-200 bg-white shadow-sm dark:border-slate-800 dark:bg-slate-900">
        <div className="border-b border-slate-100 px-5 py-4 dark:border-slate-800">
          <h2 className="text-sm font-semibold text-slate-950 dark:text-white">
            站点操作
          </h2>
          <p className="mt-0.5 text-xs text-slate-500 dark:text-slate-400">
            Key 管理与站点生命周期操作。
          </p>
        </div>
        <div className="divide-y divide-slate-100 dark:divide-slate-800">
          <div className="flex flex-col justify-between gap-3 px-5 py-4 sm:flex-row sm:items-center">
            <div>
              <p className="text-sm font-medium text-slate-900 dark:text-white">
                轮换 Site Key
              </p>
              <p className="mt-1 text-xs text-slate-500 dark:text-slate-400">
                生成新的完整 Key，旧 Key 立即失效，适合密钥泄露后处理。
              </p>
            </div>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setRotateOpen(true)}
            >
              <RefreshCw className="size-3.5" />
              轮换 Key
            </Button>
          </div>
          <div className="flex flex-col justify-between gap-3 px-5 py-4 sm:flex-row sm:items-center">
            <div>
              <p className="text-sm font-medium text-slate-900 dark:text-white">
                吊销 Site Key
              </p>
              <p className="mt-1 text-xs text-slate-500 dark:text-slate-400">
                立即阻止该站点的所有公共 API 调用；可稍后重新轮换恢复。
              </p>
            </div>
            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={!site.siteKeyPrefix}
              onClick={() => setRevokeOpen(true)}
            >
              <KeyRound className="size-3.5" />
              吊销 Key
            </Button>
          </div>
          <div className="flex flex-col justify-between gap-3 px-5 py-4 sm:flex-row sm:items-center">
            <div>
              <p className="text-sm font-medium text-slate-900 dark:text-white">
                {site.status === 'active' ? '停用站点' : '启用站点'}
              </p>
              <p className="mt-1 text-xs text-slate-500 dark:text-slate-400">
                停用会保留全部数据，随时可以重新启用。
              </p>
            </div>
            <Button
              type="button"
              variant="outline"
              size="sm"
              loading={toggleStatusMutation.isPending}
              onClick={() =>
                toggleStatusMutation.mutate(
                  site.status === 'active' ? 'disabled' : 'active',
                )
              }
            >
              <Power className="size-3.5" />
              {site.status === 'active' ? '停用' : '启用'}
            </Button>
          </div>
          <div
            className={cn(
              'flex flex-col justify-between gap-3 px-5 py-4 sm:flex-row sm:items-center',
              'bg-rose-50/40 dark:bg-rose-500/5',
            )}
          >
            <div>
              <p className="text-sm font-medium text-rose-700 dark:text-rose-300">
                删除站点
              </p>
              <p className="mt-1 text-xs text-rose-600/80 dark:text-rose-400/80">
                软删除后站点将从管理后台移除，历史数据不会立即物理清除。
              </p>
            </div>
            <Button
              type="button"
              variant="danger"
              size="sm"
              onClick={() => setDeleteOpen(true)}
            >
              <Trash2 className="size-3.5" />
              删除站点
            </Button>
          </div>
        </div>
      </section>

      <EditSiteDialog
        site={site}
        open={editOpen}
        onClose={() => setEditOpen(false)}
        onSaved={() => setEditOpen(false)}
      />
      <ConfirmDialog
        open={rotateOpen}
        onClose={() => setRotateOpen(false)}
        onConfirm={() => rotateMutation.mutate()}
        loading={rotateMutation.isPending}
        destructive={false}
        title="轮换 Site Key？"
        description="新 Key 将仅展示一次。旧 Key 会立即失效，所有已部署的博客必须同步更新。"
        confirmLabel="生成新 Key"
      />
      <ConfirmDialog
        open={revokeOpen}
        onClose={() => setRevokeOpen(false)}
        onConfirm={() => revokeMutation.mutate()}
        loading={revokeMutation.isPending}
        title="吊销 Site Key？"
        description="吊销后该站点的公共 API 请求将全部失败。已有评论、浏览量和点赞数据会保留。"
        confirmLabel="确认吊销"
      />
      <ConfirmDialog
        open={deleteOpen}
        onClose={() => setDeleteOpen(false)}
        onConfirm={() => deleteMutation.mutate()}
        loading={deleteMutation.isPending}
        title="删除站点？"
        description="站点将从管理后台移除，相关数据不再可访问。此操作无法通过界面恢复，请确认已做好备份。"
        confirmLabel="删除站点"
      />
      <SiteKeyDialog site={keySite} onClose={() => setKeySite(null)} />
    </div>
  )
}
