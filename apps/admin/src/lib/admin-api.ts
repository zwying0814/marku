import { apiRequest } from '@/lib/api'
import type {
  Admin,
  AdminComment,
  AdminCommentPage,
  AdminOverview,
  CommentFilters,
  CommentStatus,
  CreateSitePayload,
  LoginResult,
  RangeKey,
  Site,
  SiteCommentStats,
  SiteLikeStats,
  SiteList,
  SiteViewStats,
  SiteWithKey,
  UpdateSitePayload,
} from '@/types/api'

export const adminApi = {
  login(email: string, password: string) {
    return apiRequest<LoginResult>('/admin/auth/login', {
      method: 'POST',
      body: { email, password },
      auth: false,
    })
  },

  me() {
    return apiRequest<Admin>('/admin/auth/me')
  },

  changePassword(currentPassword: string, newPassword: string) {
    return apiRequest<void>('/admin/auth/change-password', {
      method: 'POST',
      body: { currentPassword, newPassword },
    })
  },

  updateProfile(input: {
    currentPassword: string
    username?: string
    newPassword?: string
  }) {
    return apiRequest<Admin>('/admin/auth/profile', {
      method: 'PATCH',
      body: input,
    })
  },

  overview(range: RangeKey) {
    return apiRequest<AdminOverview>('/admin/overview', {
      query: { range },
    })
  },

  listSites(params: { page?: number; pageSize?: number; q?: string } = {}) {
    return apiRequest<SiteList>('/admin/sites', { query: params })
  },

  createSite(payload: CreateSitePayload) {
    return apiRequest<SiteWithKey>('/admin/sites', {
      method: 'POST',
      body: payload,
    })
  },

  getSite(id: number) {
    return apiRequest<Site>(`/admin/sites/${id}`)
  },

  updateSite(id: number, payload: UpdateSitePayload) {
    return apiRequest<Site>(`/admin/sites/${id}`, {
      method: 'PATCH',
      body: payload,
    })
  },

  deleteSite(id: number) {
    return apiRequest<void>(`/admin/sites/${id}`, { method: 'DELETE' })
  },

  rotateSiteKey(id: number) {
    return apiRequest<SiteWithKey>(`/admin/sites/${id}/rotate-key`, {
      method: 'POST',
    })
  },

  revokeSiteKey(id: number) {
    return apiRequest<void>(`/admin/sites/${id}/key`, { method: 'DELETE' })
  },

  listComments(siteId: number, filters: CommentFilters) {
    return apiRequest<AdminCommentPage>(`/admin/sites/${siteId}/comments`, {
      query: { ...filters },
    })
  },

  updateComment(id: number, status: CommentStatus) {
    return apiRequest<AdminComment>(`/admin/comments/${id}`, {
      method: 'PATCH',
      body: { status },
    })
  },

  deleteComment(id: number) {
    return apiRequest<void>(`/admin/comments/${id}`, { method: 'DELETE' })
  },

  replyComment(id: number, content: string) {
    return apiRequest<AdminComment>(`/admin/comments/${id}/reply`, {
      method: 'POST',
      body: { content },
    })
  },

  bulkComments(
    siteId: number,
    ids: number[],
    action: 'approve' | 'reject' | 'spam' | 'delete',
  ) {
    return apiRequest<{ updated: number }>(
      `/admin/sites/${siteId}/comments/bulk`,
      {
        method: 'POST',
        body: { ids, action },
      },
    )
  },

  siteStats(siteId: number, range: RangeKey, limit = 10) {
    return apiRequest<SiteViewStats>(`/admin/sites/${siteId}/stats`, {
      query: { range, limit },
    })
  },

  siteLikes(siteId: number, range: RangeKey, limit = 10) {
    return apiRequest<SiteLikeStats>(`/admin/sites/${siteId}/likes`, {
      query: { range, limit },
    })
  },

  siteCommentStats(siteId: number, range: RangeKey) {
    return apiRequest<SiteCommentStats>(
      `/admin/sites/${siteId}/comments/stats`,
      { query: { range } },
    )
  },
}
