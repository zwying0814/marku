import type { CommentFilters, RangeKey } from '@/types/api'

export const queryKeys = {
  me: ['admin', 'me'] as const,
  overview: (range: RangeKey) => ['admin', 'overview', range] as const,
  sites: (params: { page?: number; pageSize?: number; q?: string }) =>
    ['admin', 'sites', params] as const,
  siteList: ['admin', 'sites'] as const,
  site: (id: number) => ['admin', 'site', id] as const,
  comments: (siteId: number, filters: CommentFilters) =>
    ['admin', 'comments', siteId, filters] as const,
  siteStats: (siteId: number, range: RangeKey) =>
    ['admin', 'site-stats', siteId, range] as const,
  siteLikes: (siteId: number, range: RangeKey) =>
    ['admin', 'site-likes', siteId, range] as const,
  siteCommentStats: (siteId: number, range: RangeKey) =>
    ['admin', 'site-comment-stats', siteId, range] as const,
}
