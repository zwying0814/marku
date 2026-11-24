import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'
import { z } from 'zod'

import { Button } from '@/components/ui/button'
import {
  FieldError,
  FieldHint,
  Input,
  Label,
  Select,
  Textarea,
} from '@/components/ui/form'
import { Modal } from '@/components/ui/modal'
import { adminApi } from '@/lib/admin-api'
import { queryKeys } from '@/lib/query-keys'
import { extractErrorMessage, parseOrigins } from '@/lib/utils'
import type { Site } from '@/types/api'

const editSiteSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, '请输入站点名称')
    .max(100, '站点名称最多 100 个字符'),
  domain: z.string().trim().max(255, '域名最多 255 个字符'),
  status: z.enum(['active', 'disabled']),
  allowedOrigins: z.string(),
  commentModeration: z.enum(['pending', 'approved']),
})

type EditSiteForm = z.infer<typeof editSiteSchema>

function toFormValues(site: Site): EditSiteForm {
  return {
    name: site.name,
    domain: site.domain ?? '',
    status: site.status,
    allowedOrigins: site.allowedOrigins.join('\n'),
    commentModeration: site.commentModeration,
  }
}

export function EditSiteDialog({
  site,
  open,
  onClose,
  onSaved,
}: {
  site: Site
  open: boolean
  onClose: () => void
  onSaved: (site: Site) => void
}) {
  const queryClient = useQueryClient()
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<EditSiteForm>({
    resolver: zodResolver(editSiteSchema),
    defaultValues: toFormValues(site),
  })

  useEffect(() => {
    if (open) reset(toFormValues(site))
  }, [open, reset, site])

  const mutation = useMutation({
    mutationFn: (values: EditSiteForm) =>
      adminApi.updateSite(site.id, {
        name: values.name,
        domain: values.domain,
        status: values.status,
        allowedOrigins: parseOrigins(values.allowedOrigins),
        commentModeration: values.commentModeration,
      }),
    onSuccess: async (updated) => {
      queryClient.setQueryData(queryKeys.site(site.id), updated)
      await queryClient.invalidateQueries({ queryKey: queryKeys.siteList })
      await queryClient.invalidateQueries({ queryKey: ['admin', 'overview'] })
      toast.success('站点设置已更新')
      onSaved(updated)
    },
    onError: (error) => toast.error(extractErrorMessage(error)),
  })

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="编辑站点"
      description="修改基础信息、运行状态和评论策略。Slug 创建后不可更改。"
      size="lg"
      dismissible={!mutation.isPending}
      footer={
        <>
          <Button
            type="button"
            variant="outline"
            onClick={onClose}
            disabled={mutation.isPending}
          >
            取消
          </Button>
          <Button
            type="button"
            loading={mutation.isPending}
            onClick={handleSubmit((values) => mutation.mutate(values))}
          >
            保存修改
          </Button>
        </>
      }
    >
      <div className="space-y-5">
        <div className="grid gap-5 sm:grid-cols-2">
          <div>
            <Label htmlFor="edit-name">站点名称</Label>
            <Input id="edit-name" {...register('name')} />
            <FieldError message={errors.name?.message} />
          </div>
          <div>
            <Label htmlFor="edit-domain">主域名</Label>
            <Input
              id="edit-domain"
              placeholder="blog.example.com"
              {...register('domain')}
            />
            <FieldError message={errors.domain?.message} />
          </div>
        </div>

        <div className="grid gap-5 sm:grid-cols-2">
          <div>
            <Label htmlFor="edit-status">运行状态</Label>
            <Select id="edit-status" {...register('status')}>
              <option value="active">运行中</option>
              <option value="disabled">已停用</option>
            </Select>
            <FieldHint>停用后公共 API 将拒绝该站点的请求。</FieldHint>
          </div>
          <div>
            <Label htmlFor="edit-moderation">评论审核</Label>
            <Select id="edit-moderation" {...register('commentModeration')}>
              <option value="pending">先审后发</option>
              <option value="approved">自动公开</option>
            </Select>
          </div>
        </div>

        <div>
          <Label htmlFor="edit-origins">允许来源</Label>
          <Textarea
            id="edit-origins"
            rows={4}
            className="font-mono text-xs"
            placeholder={'https://blog.example.com\nhttp://localhost:36583'}
            {...register('allowedOrigins')}
          />
          <FieldHint>每行一个 origin；留空表示不限制来源。</FieldHint>
        </div>
      </div>
    </Modal>
  )
}
