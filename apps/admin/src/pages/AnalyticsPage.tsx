import { useQuery } from '@tanstack/react-query'
import {
  BarChart3,
  CheckCircle2,
  Eye,
  Heart,
  MessageSquare,
  MousePointerClick,
  TrendingUp,
} from 'lucide-react'
import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'

import { PageHeader } from '@/components/PageHeader'
import { StatCard } from '@/components/StatCard'
import { TrendChart } from '@/components/TrendChart'
import { Badge } from '@/components/ui/badge'
import { EmptyState } from '@/components/ui/empty-state'
import { ErrorState } from '@/components/ui/error-state'
import { Select } from '@/components/ui/form'
import { RangeSwitcher } from '@/components/ui/range-switcher'
import { Skeleton } from '@/components/ui/skeleton'
import { adminApi } from '@/lib/admin-api'
import { queryKeys } from '@/lib/query-keys'
import { cn, formatCompact, formatNumber } from '@/lib/utils'
import type { RangeKey } from '@/types/api'

type TopTab = 'views' | 'likes'

export function AnalyticsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [range, setRange] = useState<RangeKey>('7d')
  const [topTab, setTopTab] = useState<TopTab>('views')
  const requestedSiteId = Number(searchParams.get('siteId'))

  const sitesQuery = useQuery({
    queryKey: queryKeys.sites({ page: 1, pageSize: 100 }),
    queryFn: () => adminApi.listSites({ page: 1, pageSize: 100 }),
  })
  const sites = sitesQuery.data?.items ?? []
  const selectedSite =
    sites.find((site) => site.id === requestedSiteId) ?? sites[0]
  const siteId = selectedSite?.id

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

  const statsQuery = useQuery({
    queryKey: queryKeys.siteStats(siteId ?? 0, range),
    queryFn: () => adminApi.siteStats(siteId ?? 0, range, 20),
    enabled: Boolean(siteId),
  })
  const likesQuery = useQuery({
    queryKey: queryKeys.siteLikes(siteId ?? 0, range),
    queryFn: () => adminApi.siteLikes(siteId ?? 0, range, 20),
    enabled: Boolean(siteId),
  })
  const commentsQuery = useQuery({
    queryKey: queryKeys.siteCommentStats(siteId ?? 0, range),
    queryFn: () => adminApi.siteCommentStats(siteId ?? 0, range),
    enabled: Boolean(siteId),
  })

  const stats = statsQuery.data
  const likes = likesQuery.data
  const commentStats = commentsQuery.data
  const likeByDate = new Map(likes?.trend.map((point) => [point.date, point]))
  const trendData =
    stats?.trend.map((point) => ({
      date: point.date,
      views: Number(point.views),
      uniqueViews: Number(point.uniqueViews),
      likes: Number(likeByDate.get(point.date)?.likes ?? 0),
    })) ?? []
  const rangeLikeDelta = (likes?.trend ?? []).reduce(
    (total, point) => total + Number(point.net),
    0,
  )
  const commentTrendData =
    commentStats?.trend.map((point) => ({
      date: point.date,
      comments: Number(point.comments),
      pending: Number(point.pending),
    })) ?? []
  const topItems =
    topTab === 'views'
      ? (stats?.topPages ?? []).map((item) => ({
          path: item.path,
          value: item.views,
          secondary: item.uniqueViews,
        }))
      : (likes?.topPages ?? []).map((item) => ({
          path: item.path,
          value: item.likes,
          secondary: 0,
        }))

  if (sitesQuery.isPending) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-24 rounded-2xl" />
        <Skeleton className="h-96 rounded-2xl" />
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
        eyebrow="Analytics"
        title="数据分析"
        description="按站点查看浏览量、独立访客、点赞与热门页面排行。"
        actions={
          <>
            <Select
              aria-label="选择站点"
              value={siteId ?? ''}
              onChange={(event) => {
                setSearchParams({ siteId: event.target.value })
              }}
              className="w-full sm:w-52"
            >
              {sites.map((site) => (
                <option key={site.id} value={site.id}>
                  {site.name}
                </option>
              ))}
            </Select>
            <RangeSwitcher value={range} onChange={setRange} />
          </>
        }
      />

      {!selectedSite ? (
        <section className="rounded-2xl border border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900">
          <EmptyState
            icon={BarChart3}
            title="暂无可分析的站点"
            description="先创建一个站点并接入浏览量上报，数据会在这里汇总展示。"
          />
        </section>
      ) : (
        <>
          {statsQuery.isError || likesQuery.isError ? (
            <section className="rounded-2xl border border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900">
              <ErrorState
                error={statsQuery.error ?? likesQuery.error}
                onRetry={() => {
                  statsQuery.refetch()
                  likesQuery.refetch()
                }}
              />
            </section>
          ) : (
            <>
              <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
                {statsQuery.isPending ? (
                  Array.from({ length: 4 }).map((_, index) => (
                    <Skeleton key={index} className="h-36 rounded-2xl" />
                  ))
                ) : (
                  <>
                    <StatCard
                      label="累计浏览量"
                      value={formatCompact(stats?.totalViews)}
                      hint="包含重复访问"
                      icon={Eye}
                      accent="sky"
                    />
                    <StatCard
                      label="累计独立访客"
                      value={formatCompact(stats?.totalUniqueViews)}
                      hint="按访客与日期去重"
                      icon={MousePointerClick}
                      accent="indigo"
                    />
                    <StatCard
                      label="累计点赞"
                      value={formatCompact(likes?.totalLikes)}
                      hint="当前所有页面"
                      icon={Heart}
                      accent="rose"
                    />
                    <StatCard
                      label={`${range === '7d' ? '近 7 天' : '近 30 天'}净增点赞`}
                      value={
                        rangeLikeDelta > 0
                          ? `+${formatNumber(rangeLikeDelta)}`
                          : formatNumber(rangeLikeDelta)
                      }
                      hint="点赞减去取消点赞"
                      icon={TrendingUp}
                      accent={rangeLikeDelta >= 0 ? 'emerald' : 'rose'}
                    />
                  </>
                )}
              </section>

              <section className="grid gap-6 xl:grid-cols-[1.4fr_0.6fr]">
                <div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm dark:border-slate-800 dark:bg-slate-900">
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <div>
                      <h2 className="text-sm font-semibold text-slate-950 dark:text-white">
                        {selectedSite.name} · 趋势
                      </h2>
                      <p className="mt-0.5 text-xs text-slate-500 dark:text-slate-400">
                        浏览量、独立访客与点赞的每日变化。
                      </p>
                    </div>
                    <Badge variant="info">
                      {range === '7d' ? '近 7 天' : '近 30 天'}
                    </Badge>
                  </div>
                  {statsQuery.isPending ? (
                    <Skeleton className="mt-5 h-72" />
                  ) : (
                    <div className="mt-4">
                      <TrendChart
                        height={300}
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

                <div className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm dark:border-slate-800 dark:bg-slate-900">
                  <div className="border-b border-slate-100 px-5 py-4 dark:border-slate-800">
                    <h2 className="text-sm font-semibold text-slate-950 dark:text-white">
                      热门页面
                    </h2>
                    <div className="mt-3 inline-flex rounded-lg bg-slate-100 p-0.5 dark:bg-slate-800">
                      {(
                        [
                          ['views', '按浏览'],
                          ['likes', '按点赞'],
                        ] as const
                      ).map(([value, label]) => (
                        <button
                          key={value}
                          type="button"
                          onClick={() => setTopTab(value)}
                          className={cn(
                            'rounded-md px-2.5 py-1.5 text-[11px] font-semibold transition',
                            topTab === value
                              ? 'bg-white text-slate-900 shadow-sm dark:bg-slate-900 dark:text-white'
                              : 'text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-white',
                          )}
                        >
                          {label}
                        </button>
                      ))}
                    </div>
                  </div>
                  {statsQuery.isPending || likesQuery.isPending ? (
                    <div className="space-y-3 p-5">
                      {Array.from({ length: 6 }).map((_, index) => (
                        <Skeleton key={index} className="h-8" />
                      ))}
                    </div>
                  ) : topItems.length === 0 ? (
                    <EmptyState
                      icon={BarChart3}
                      title="暂无排行数据"
                      description="接入统计接口后，热门页面会出现在这里。"
                      className="py-10"
                    />
                  ) : (
                    <ol className="divide-y divide-slate-100 dark:divide-slate-800">
                      {topItems.slice(0, 8).map((item, index) => (
                        <li
                          key={item.path}
                          className="flex items-center gap-3 px-5 py-3"
                        >
                          <span
                            className={cn(
                              'flex size-6 shrink-0 items-center justify-center rounded-lg text-[10px] font-bold',
                              index < 3
                                ? 'bg-indigo-50 text-indigo-600 dark:bg-indigo-500/10 dark:text-indigo-400'
                                : 'bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-400',
                            )}
                          >
                            {index + 1}
                          </span>
                          <span
                            className="min-w-0 flex-1 truncate font-mono text-[11px] text-slate-600 dark:text-slate-300"
                            title={item.path}
                          >
                            {item.path}
                          </span>
                          <span className="shrink-0 text-right">
                            <span className="block text-xs font-semibold text-slate-900 dark:text-white">
                              {formatNumber(item.value)}
                            </span>
                            {topTab === 'views' ? (
                              <span className="block text-[10px] text-slate-400">
                                独立 {formatNumber(item.secondary)}
                              </span>
                            ) : null}
                          </span>
                        </li>
                      ))}
                    </ol>
                  )}
                </div>
              </section>

              <section className="grid gap-6 lg:grid-cols-[1.4fr_0.6fr]">
                <div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm dark:border-slate-800 dark:bg-slate-900">
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <div>
                      <h2 className="text-sm font-semibold text-slate-950 dark:text-white">
                        评论趋势
                      </h2>
                      <p className="mt-0.5 text-xs text-slate-500 dark:text-slate-400">
                        每日新增评论与其中待审核的数量。
                      </p>
                    </div>
                    <MessageSquare className="size-4 text-slate-400" />
                  </div>
                  {commentsQuery.isPending ? (
                    <Skeleton className="mt-5 h-56" />
                  ) : commentsQuery.isError ? (
                    <ErrorState
                      error={commentsQuery.error}
                      onRetry={() => commentsQuery.refetch()}
                    />
                  ) : (
                    <div className="mt-4">
                      <TrendChart
                        height={230}
                        data={commentTrendData}
                        series={[
                          {
                            key: 'comments',
                            label: '新增评论',
                            color: '#f59e0b',
                            type: 'area',
                          },
                          {
                            key: 'pending',
                            label: '待审核',
                            color: '#f43f5e',
                            type: 'line',
                          },
                        ]}
                      />
                    </div>
                  )}
                </div>

                <div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm dark:border-slate-800 dark:bg-slate-900">
                  <div className="flex items-center justify-between">
                    <h2 className="text-sm font-semibold text-slate-950 dark:text-white">
                      审核概况
                    </h2>
                    <CheckCircle2 className="size-4 text-emerald-500" />
                  </div>
                  {commentsQuery.isPending ? (
                    <div className="mt-5 space-y-3">
                      {Array.from({ length: 4 }).map((_, index) => (
                        <Skeleton key={index} className="h-8" />
                      ))}
                    </div>
                  ) : (
                    <>
                      <dl className="mt-4 space-y-3 text-xs">
                        <div className="flex items-center justify-between">
                          <dt className="text-slate-500 dark:text-slate-400">
                            累计评论
                          </dt>
                          <dd className="font-semibold text-slate-900 dark:text-white">
                            {formatNumber(commentStats?.total)}
                          </dd>
                        </div>
                        <div className="flex items-center justify-between">
                          <dt className="text-slate-500 dark:text-slate-400">
                            已通过
                          </dt>
                          <dd className="font-semibold text-emerald-600 dark:text-emerald-400">
                            {formatNumber(commentStats?.approved)}
                          </dd>
                        </div>
                        <div className="flex items-center justify-between">
                          <dt className="text-slate-500 dark:text-slate-400">
                            待审核
                          </dt>
                          <dd className="font-semibold text-amber-600 dark:text-amber-400">
                            {formatNumber(commentStats?.pending)}
                          </dd>
                        </div>
                      </dl>
                      <div className="mt-5">
                        <div className="h-2 overflow-hidden rounded-full bg-slate-100 dark:bg-slate-800">
                          <div
                            className="h-full rounded-full bg-emerald-500"
                            style={{
                              width: `${
                                commentStats?.total
                                  ? Math.round(
                                      (commentStats.approved /
                                        commentStats.total) *
                                        100,
                                    )
                                  : 0
                              }%`,
                            }}
                          />
                        </div>
                        <p className="mt-2 text-[11px] text-slate-500 dark:text-slate-400">
                          公开率{' '}
                          {commentStats?.total
                            ? Math.round(
                                (commentStats.approved / commentStats.total) *
                                  100,
                              )
                            : 0}
                          %
                        </p>
                      </div>
                      <Link
                        to={`/comments?siteId=${selectedSite.id}&status=pending`}
                        className="mt-5 inline-flex items-center gap-1.5 text-xs font-semibold text-indigo-600 hover:text-indigo-500 dark:text-indigo-400"
                      >
                        处理待审核评论
                        <MessageSquare className="size-3.5" />
                      </Link>
                    </>
                  )}
                </div>
              </section>
            </>
          )}
        </>
      )}
    </div>
  )
}
