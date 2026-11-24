import type { BadgeVariant } from '@/components/ui/badge'
import type { CommentModeration, CommentStatus, SiteStatus } from '@/types/api'

export const siteStatusMeta: Record<
  SiteStatus,
  { label: string; variant: BadgeVariant }
> = {
  active: { label: '运行中', variant: 'success' },
  disabled: { label: '已停用', variant: 'neutral' },
}

export const moderationMeta: Record<
  CommentModeration,
  { label: string; description: string }
> = {
  pending: {
    label: '先审后发',
    description: '新评论进入待审核队列，管理员通过后公开展示。',
  },
  approved: {
    label: '自动公开',
    description: '新评论立即通过，适合信任度高的站点。',
  },
}

export const commentStatusMeta: Record<
  CommentStatus,
  { label: string; variant: BadgeVariant }
> = {
  pending: { label: '待审核', variant: 'warning' },
  approved: { label: '已通过', variant: 'success' },
  rejected: { label: '已拒绝', variant: 'neutral' },
  spam: { label: '垃圾评论', variant: 'danger' },
}
