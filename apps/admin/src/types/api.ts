import type { components } from '@marku/shared'

type Schemas = components['schemas']

export type Admin = Required<Schemas['Admin']>
export type Site = Required<Schemas['Site']>
export type SiteWithKey = Site & { siteKey: string }
export type SiteList = {
  items: Site[]
  total: number
  page: number
  pageSize: number
}
export type SiteStatus = NonNullable<Site['status']>
export type CommentModeration = NonNullable<Site['commentModeration']>
export type AdminComment = Required<Schemas['AdminComment']>
export type CommentStatus = NonNullable<AdminComment['status']>
export type AdminCommentPage = {
  items: AdminComment[]
  total: number
  page: number
  pageSize: number
}
export type PageViewItem = Required<Schemas['PageViewItem']>
export type ViewTrendPoint = Required<Schemas['ViewTrendPoint']>
export type SiteViewStats = Omit<
  Required<Schemas['SiteViewStats']>,
  'topPages' | 'trend'
> & {
  topPages: PageViewItem[]
  trend: ViewTrendPoint[]
}
export type LikePageItem = Required<Schemas['LikePageItem']>
export type LikeTrendPoint = Required<Schemas['LikeTrendPoint']>
export type SiteLikeStats = Omit<
  Required<Schemas['SiteLikeStats']>,
  'path' | 'topPages' | 'trend'
> & {
  path?: string
  topPages: LikePageItem[]
  trend: LikeTrendPoint[]
}
export type OverviewTrendPoint = Required<Schemas['OverviewTrendPoint']>
export type AdminOverview = Omit<
  Required<Schemas['AdminOverview']>,
  'trend'
> & {
  trend: OverviewTrendPoint[]
}
export type CommentTrendPoint = Required<Schemas['CommentTrendPoint']>
export type SiteCommentStats = Omit<
  Required<Schemas['SiteCommentStats']>,
  'trend'
> & {
  trend: CommentTrendPoint[]
}

export type LoginResult = {
  token: string
  expiresAt: string
  admin: Admin
}

export type CreateSitePayload = {
  name: string
  slug: string
  domain?: string
  allowedOrigins?: string[]
  commentModeration: CommentModeration
}

export type UpdateSitePayload = {
  name?: string
  domain?: string
  status?: SiteStatus
  allowedOrigins?: string[]
  commentModeration?: CommentModeration
}

export type CommentFilters = {
  status?: CommentStatus | ''
  path?: string
  q?: string
  from?: string
  to?: string
  page?: number
  pageSize?: number
  sort?: 'asc' | 'desc'
}

export type RangeKey = '7d' | '30d'
