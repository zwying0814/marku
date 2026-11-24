import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Check,
  Copy,
  ExternalLink,
  Mail,
  MessageSquareReply,
  ShieldX,
  Trash2,
  X,
} from 'lucide-react'
import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'
import { z } from 'zod'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Drawer } from '@/components/ui/drawer'
import { FieldError, Textarea } from '@/components/ui/form'
import { adminApi } from '@/lib/admin-api'
import { commentStatusMeta } from '@/lib/constants'
import { extractErrorMessage, formatDate } from '@/lib/utils'
import type { AdminComment, CommentStatus } from '@/types/api'

const replySchema = z.object({
  content: z
    .string()
    .trim()
    .min(1, '请输入回复内容')
    .max(2000, '回复最多 2000 个字符'),
})

type ReplyForm = z.infer<typeof replySchema>

export function CommentDetailDrawer({
  comment,
  onClose,
  onUpdated,
  onDelete,
}: {
  comment: AdminComment | null
  onClose: () => void
  onUpdated: (comment: AdminComment) => void
  onDelete: (comment: AdminComment) => void
}) {
  const queryClient = useQueryClient()
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<ReplyForm>({
    resolver: zodResolver(replySchema),
    defaultValues: { content: '' },
  })

  useEffect(() => {
    reset({ content: '' })
  }, [comment?.id, reset])

  async function invalidateCommentData() {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ['admin', 'comments'] }),
      queryClient.invalidateQueries({ queryKey: ['admin', 'overview'] }),
    ])
  }

  const statusMutation = useMutation({
    mutationFn: (status: CommentStatus) => {
      if (!comment) throw new Error('评论不存在')
      return adminApi.updateComment(comment.id, status)
    },
    onSuccess: async (updated) => {
      onUpdated(updated)
      await invalidateCommentData()
      toast.success(`评论已标记为${commentStatusMeta[updated.status].label}`)
    },
    onError: (error) => toast.error(extractErrorMessage(error)),
  })

  const replyMutation = useMutation({
    mutationFn: (content: string) => {
      if (!comment) throw new Error('评论不存在')
      return adminApi.replyComment(comment.id, content)
    },
    onSuccess: async () => {
      reset({ content: '' })
      await invalidateCommentData()
      toast.success('回复已发布')
    },
    onError: (error) => toast.error(extractErrorMessage(error)),
  })

  const busy = statusMutation.isPending || replyMutation.isPending

  return (
    <Drawer
      open={Boolean(comment)}
      onClose={onClose}
      title="评论详情"
      description={comment ? `#${comment.id} · ${comment.path}` : undefined}
      footer={
        comment ? (
          <div className="flex flex-wrap gap-2">
            <Button
              type="button"
              size="sm"
              variant="outline"
              disabled={busy || comment.status === 'approved'}
              onClick={() => statusMutation.mutate('approved')}
            >
              <Check className="size-3.5" />
              通过
            </Button>
            <Button
              type="button"
              size="sm"
              variant="outline"
              disabled={busy || comment.status === 'rejected'}
              onClick={() => statusMutation.mutate('rejected')}
            >
              <X className="size-3.5" />
              拒绝
            </Button>
            <Button
              type="button"
              size="sm"
              variant="outline"
              disabled={busy || comment.status === 'spam'}
              onClick={() => statusMutation.mutate('spam')}
            >
              <ShieldX className="size-3.5" />
              垃圾
            </Button>
            <Button
              type="button"
              size="sm"
              variant="danger"
              className="ml-auto"
              disabled={busy}
              onClick={() => onDelete(comment)}
            >
              <Trash2 className="size-3.5" />
              删除
            </Button>
          </div>
        ) : null
      }
    >
      {comment ? (
        <div className="space-y-6">
          <div className="flex flex-wrap items-center gap-2">
            <Badge variant={commentStatusMeta[comment.status].variant}>
              {commentStatusMeta[comment.status].label}
            </Badge>
            {comment.isAdmin ? <Badge variant="info">管理员回复</Badge> : null}
            {comment.parentId ? (
              <Badge variant="outline">回复 #{comment.parentId}</Badge>
            ) : null}
          </div>

          <div className="rounded-2xl border border-slate-200 bg-slate-50/70 p-5 dark:border-slate-800 dark:bg-slate-950/50">
            <p className="whitespace-pre-wrap break-words text-sm leading-7 text-slate-800 dark:text-slate-100">
              {comment.content}
            </p>
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="rounded-2xl border border-slate-200 p-4 dark:border-slate-800">
              <p className="text-[11px] font-semibold uppercase tracking-wider text-slate-400">
                作者
              </p>
              <p className="mt-2 text-sm font-medium text-slate-900 dark:text-white">
                {comment.authorName}
              </p>
              {comment.authorEmail ? (
                <p className="mt-1 flex items-center gap-1.5 break-all text-xs text-slate-500 dark:text-slate-400">
                  <Mail className="size-3 shrink-0" />
                  {comment.authorEmail}
                </p>
              ) : null}
              {comment.authorWebsite ? (
                <a
                  href={comment.authorWebsite}
                  target="_blank"
                  rel="noreferrer"
                  className="mt-1.5 inline-flex items-center gap-1 text-xs font-medium text-indigo-600 hover:text-indigo-500 dark:text-indigo-400"
                >
                  访问网站
                  <ExternalLink className="size-3" />
                </a>
              ) : null}
            </div>

            <div className="rounded-2xl border border-slate-200 p-4 dark:border-slate-800">
              <p className="text-[11px] font-semibold uppercase tracking-wider text-slate-400">
                来源
              </p>
              <p className="mt-2 break-all font-mono text-xs text-slate-700 dark:text-slate-300">
                {comment.path}
              </p>
              <p className="mt-2 text-xs text-slate-500 dark:text-slate-400">
                {formatDate(comment.createdAt, true)}
              </p>
            </div>
          </div>

          <dl className="space-y-3 rounded-2xl border border-slate-200 p-4 text-xs dark:border-slate-800">
            <div className="flex items-start justify-between gap-4">
              <dt className="shrink-0 text-slate-500 dark:text-slate-400">
                IP 哈希
              </dt>
              <dd className="flex min-w-0 items-center gap-2 font-mono text-slate-700 dark:text-slate-300">
                <span className="truncate">{comment.ipHash || '未记录'}</span>
                {comment.ipHash ? (
                  <button
                    type="button"
                    aria-label="复制 IP 哈希"
                    onClick={() => {
                      navigator.clipboard
                        .writeText(comment.ipHash)
                        .then(() => toast.success('已复制'))
                        .catch(() => toast.error('复制失败'))
                    }}
                    className="text-slate-400 hover:text-slate-700 dark:hover:text-white"
                  >
                    <Copy className="size-3.5" />
                  </button>
                ) : null}
              </dd>
            </div>
            <div className="flex items-start justify-between gap-4">
              <dt className="shrink-0 text-slate-500 dark:text-slate-400">
                更新时间
              </dt>
              <dd className="text-right text-slate-700 dark:text-slate-300">
                {formatDate(comment.updatedAt, true)}
              </dd>
            </div>
            <div className="flex items-start justify-between gap-4">
              <dt className="shrink-0 text-slate-500 dark:text-slate-400">
                User Agent
              </dt>
              <dd
                className="max-w-[70%] break-all text-right font-mono text-[11px] leading-5 text-slate-600 dark:text-slate-400"
                title={comment.userAgent}
              >
                {comment.userAgent || '未记录'}
              </dd>
            </div>
          </dl>

          <form
            onSubmit={handleSubmit(({ content }) =>
              replyMutation.mutate(content),
            )}
          >
            <div className="rounded-2xl border border-indigo-100 bg-indigo-50/50 p-4 dark:border-indigo-500/20 dark:bg-indigo-500/5">
              <div className="mb-3 flex items-center gap-2">
                <MessageSquareReply className="size-4 text-indigo-600 dark:text-indigo-400" />
                <p className="text-sm font-semibold text-slate-900 dark:text-white">
                  管理员回复
                </p>
              </div>
              <Textarea
                rows={3}
                placeholder="以管理员身份回复这条评论…"
                maxLength={2000}
                {...register('content')}
              />
              <FieldError message={errors.content?.message} />
              <div className="mt-3 flex justify-end">
                <Button
                  type="submit"
                  size="sm"
                  loading={replyMutation.isPending}
                >
                  发布回复
                </Button>
              </div>
            </div>
          </form>
        </div>
      ) : null}
    </Drawer>
  )
}
