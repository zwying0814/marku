import { useQuery } from '@tanstack/react-query'
import {
  ArrowRight,
  ArrowUpRight,
  Eye,
  Globe2,
  Heart,
  MessageSquare,
  Plus,
  Sparkles,
} from 'lucide-react'
import { useState } from 'react'
import { Link } from 'react-router-dom'

import { PageHeader } from '@/components/PageHeader'
import { StatCard } from '@/components/StatCard'
import { TrendChart } from '@/components/TrendChart'
import { Badge } from '@/components/ui/badge'
import { EmptyState } from '@/components/ui/empty-state'
import { ErrorState } from '@/components/ui/error-state'
import { RangeSwitcher } from '@/components/ui/range-switcher'
import { Skeleton, TableSkeleton } from '@/components/ui/skeleton'
import { adminApi } from '@/lib/admin-api'
import { moderationMeta, siteStatusMeta } from '@/lib/constants'
import { queryKeys } from '@/lib/query-keys'
import { formatCompact, formatDate, formatNumber } from '@/lib/utils'
import type { RangeKey } from '@/types/api'

export function DashboardPage() {
  const [range, setRange] = useState<RangeKey>('7d')

  const overviewQuery = useQuery({
    queryKey: queryKeys.overview(range),
    queryFn: () => adminApi.overview(range),
  })
  const sitesQuery = useQuery({
    queryKey: queryKeys.sites({ page: 1, pageSize: 5 }),
    queryFn: () => adminApi.listSites({ page: 1, pageSize: 5 }),
  })

  const overview = overviewQuery.data
  const sites = sitesQuery.data?.items ?? []
  const trendData =
    overview?.trend.map((point) => ({
      date: point.date,
      views: Number(point.views),
      likes: Number(point.likes),
    })) ?? []

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow="Overview"
        title="工作台概览"
        description="查看所有站点的互动表现，并快速进入待处理内容。"
        actions={
          <>
            <RangeSwitcher value={range} onChange={setRange} />
            <Link
              to="/sites?create=1"
              className="inline-flex h-10 items-center justify-center gap-2 rounded-xl bg-indigo-600 px-4 text-sm font-medium text-white shadow-sm transition hover:bg-indigo-500"
            >
              <Plus className="size-4" />
              新建站点
            </Link>
          </>
        }
      />

      {overviewQuery.isPending ? (
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
          {Array.from({ length: 4 }).map((_, index) => (
            <Skeleton key={index} className="h-36 rounded-2xl" />
          ))}
        </div>
      ) : overviewQuery.isError ? (
        <section className="rounded-2xl border border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900">
          <ErrorState
            error={overviewQuery.error}
            onRetry={() => overviewQuery.refetch()}
          />
        </section>
      ) : (
        <>
          <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
            <StatCard
              label="站点"
              value={formatNumber(overview?.sites)}
              hint={`${sites.filter((site) => site.status === 'active').length} 个运行中`}
              icon={Globe2}
              accent="indigo"
            />
            <StatCard
              label="评论"
              value={formatNumber(overview?.comments)}
              hint={`${formatNumber(overview?.pendingComments)} 条待审核`}
              icon={MessageSquare}
              accent="amber"
            />
            <StatCard
              label="浏览量"
              value={formatCompact(overview?.views)}
              hint="所有站点累计"
              icon={Eye}
              accent="sky"
            />
            <StatCard
              label="点赞"
              value={formatCompact(overview?.likes)}
              hint="所有站点累计"
              icon={Heart}
              accent="rose"
            />
          </section>

          {(overview?.pendingComments ?? 0) > 0 ? (
            <Link
              to="/comments?status=pending"
              className="group flex items-center justify-between gap-4 rounded-2xl border border-amber-200 bg-amber-50 px-5 py-4 transition hover:border-amber-300 hover:bg-amber-100/70 dark:border-amber-500/20 dark:bg-amber-500/10 dark:hover:bg-amber-500/15"
            >
              <div className="flex items-center gap-3">
                <span className="flex size-9 items-center justify-center rounded-xl bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-300">
                  <MessageSquare className="size-4" />
                </span>
                <div>
                  <p className="text-sm font-semibold text-amber-900 dark:text-amber-100">
                    {overview?.pendingComments} 条评论等待审核
                  </p>
                  <p className="mt-0.5 text-xs text-amber-700/80 dark:text-amber-300/70">
                    及时处理可以让读者看到最新讨论。
                  </p>
                </div>
              </div>
              <ArrowRight className="size-4 text-amber-600 transition group-hover:translate-x-0.5 dark:text-amber-400" />
            </Link>
          ) : null}

          <section className="grid gap-6 xl:grid-cols-[1.45fr_0.55fr]">
            <div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm dark:border-slate-800 dark:bg-slate-900">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <h2 className="text-sm font-semibold text-slate-950 dark:text-white">
                    全局互动趋势
                  </h2>
                  <p className="mt-1 text-xs text-slate-500 dark:text-slate-400">
                    浏览量来自左侧坐标轴，点赞数来自右侧。
                  </p>
                </div>
                <Badge variant="info">
                  {range === '7d' ? '近 7 天' : '近 30 天'}
                </Badge>
              </div>
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
                      key: 'likes',
                      label: '点赞',
                      color: '#f43f5e',
                      type: 'line',
                      yAxis: 'right',
                    },
                  ]}
                />
              </div>
            </div>

            <div className="flex flex-col gap-4">
              <div className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-indigo-600 to-violet-700 p-5 text-white shadow-lg shadow-indigo-500/10">
                <div className="absolute -right-8 -top-8 size-28 rounded-full bg-white/10 blur-2xl" />
                <Sparkles className="size-5 text-indigo-200" />
                <h3 className="mt-4 text-base font-semibold">接入一个新博客</h3>
                <p className="mt-1.5 text-xs leading-6 text-indigo-100/80">
                  创建站点、复制 Site Key，然后用公共 API
                  读取评论、浏览量和点赞。
                </p>
                <Link
                  to="/sites?create=1"
                  className="mt-5 inline-flex items-center gap-1.5 rounded-xl bg-white/15 px-3.5 py-2 text-xs font-semibold ring-1 ring-white/20 transition hover:bg-white/25"
                >
                  开始创建
                  <ArrowUpRight className="size-3.5" />
                </Link>
              </div>

              <div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm dark:border-slate-800 dark:bg-slate-900">
                <h3 className="text-sm font-semibold text-slate-950 dark:text-white">
                  数据健康度
                </h3>
                <dl className="mt-4 space-y-3 text-xs">
                  <div className="flex items-center justify-between">
                    <dt className="text-slate-500 dark:text-slate-400">
                      运行中站点
                    </dt>
                    <dd className="font-semibold text-slate-900 dark:text-white">
                      {overview?.sites ?? 0}
                    </dd>
                  </div>
                  <div className="flex items-center justify-between">
                    <dt className="text-slate-500 dark:text-slate-400">
                      待审核占比
                    </dt>
                    <dd className="font-semibold text-slate-900 dark:text-white">
                      {overview?.comments
                        ? Math.round(
                            (overview.pendingComments / overview.comments) *
                              100,
                          )
                        : 0}
                      %
                    </dd>
                  </div>
                  <div className="flex items-center justify-between">
                    <dt className="text-slate-500 dark:text-slate-400">
                      每千次浏览点赞
                    </dt>
                    <dd className="font-semibold text-slate-900 dark:text-white">
                      {overview?.views
                        ? ((overview.likes / overview.views) * 1000).toFixed(1)
                        : '0.0'}
                    </dd>
                  </div>
                </dl>
              </div>
            </div>
          </section>
        </>
      )}

      <section className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm dark:border-slate-800 dark:bg-slate-900">
        <div className="flex items-center justify-between gap-3 border-b border-slate-100 px-5 py-4 dark:border-slate-800">
          <div>
            <h2 className="text-sm font-semibold text-slate-950 dark:text-white">
              最近站点
            </h2>
            <p className="mt-0.5 text-xs text-slate-500 dark:text-slate-400">
              快速查看各站点状态与接入信息。
            </p>
          </div>
          <Link
            to="/sites"
            className="text-xs font-semibold text-indigo-600 hover:text-indigo-500 dark:text-indigo-400"
          >
            查看全部
          </Link>
        </div>

        {sitesQuery.isPending ? (
          <TableSkeleton rows={4} />
        ) : sitesQuery.isError ? (
          <ErrorState
            error={sitesQuery.error}
            onRetry={() => sitesQuery.refetch()}
          />
        ) : sites.length === 0 ? (
          <EmptyState
            icon={Globe2}
            title="从第一个站点开始"
            description="创建站点后，这里会展示站点状态、评论入口和最近数据。"
            action={
              <Link
                to="/sites?create=1"
                className="inline-flex h-10 items-center justify-center gap-2 rounded-xl bg-indigo-600 px-4 text-sm font-medium text-white transition hover:bg-indigo-500"
              >
                <Plus className="size-4" />
                创建站点
              </Link>
            }
          />
        ) : (
          <div className="divide-y divide-slate-100 dark:divide-slate-800">
            {sites.map((site) => (
              <Link
                key={site.id}
                to={`/sites/${site.id}`}
                className="group flex items-center gap-4 px-5 py-4 transition hover:bg-slate-50 dark:hover:bg-slate-800/40"
              >
                <span className="flex size-9 shrink-0 items-center justify-center rounded-xl bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-400">
                  <Globe2 className="size-4" />
                </span>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium text-slate-900 group-hover:text-indigo-600 dark:text-white dark:group-hover:text-indigo-400">
                    {site.name}
                  </p>
                  <p className="mt-0.5 truncate font-mono text-[11px] text-slate-400">
                    {site.domain || `/${site.slug}`}
                  </p>
                </div>
                <div className="hidden text-right sm:block">
                  <p className="text-[11px] text-slate-400">
                    {moderationMeta[site.commentModeration].label}
                  </p>
                  <p className="mt-0.5 text-[11px] text-slate-500 dark:text-slate-400">
                    {formatDate(site.createdAt)}
                  </p>
                </div>
                <Badge variant={siteStatusMeta[site.status].variant}>
                  {siteStatusMeta[site.status].label}
                </Badge>
                <ArrowUpRight className="size-4 text-slate-300 transition group-hover:text-slate-600 dark:text-slate-600 dark:group-hover:text-slate-300" />
              </Link>
            ))}
          </div>
        )}
      </section>
    </div>
  )
}
