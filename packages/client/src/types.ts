export interface PublicSite {
  id?: number
  name?: string
  slug?: string
  domain?: string
  commentModeration?: 'pending' | 'approved'
}

export interface PublicComment {
  id?: number
  parentId?: number | null
  authorName?: string
  authorWebsite?: string
  content?: string
  isAdmin?: boolean
  createdAt?: string
  replies?: PublicComment[]
}

export interface PublicCommentPage {
  items?: PublicComment[]
  total?: number
  page?: number
  pageSize?: number
}

export interface CommentListQuery {
  path: string
  page?: number
  pageSize?: number
  sort?: 'asc' | 'desc'
}

export interface CreateCommentInput {
  path: string
  parentId?: number | null
  authorName: string
  authorEmail?: string
  authorWebsite?: string
  content: string
  honeypot?: string
  formStartedAt?: number
}

export interface CommentSubmitResult {
  id?: number
  status?: 'pending' | 'approved'
  createdAt?: string
}

export interface PageStats {
  path?: string
  comments?: number
  views?: number
  uniqueViews?: number
  likes?: number
  liked?: boolean
}

export interface PageViewItem {
  path?: string
  views?: number
  uniqueViews?: number
}

export interface RecordViewResult {
  views?: number
  uniqueViews?: number
  recorded?: boolean
}

export interface LikeState {
  likes?: number
  liked?: boolean
}
