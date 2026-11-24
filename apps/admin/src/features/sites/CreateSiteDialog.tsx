import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  ArrowLeft,
  ArrowRight,
  Check,
  Globe2,
  RefreshCw,
  Sparkles,
} from 'lucide-react'
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'
import { z } from 'zod'

import { Button } from '@/components/ui/button'
import {
  FieldError,
  FieldHint,
  Input,
  Label,
  Textarea,
} from '@/components/ui/form'
import { Modal } from '@/components/ui/modal'
import { adminApi } from '@/lib/admin-api'
import { moderationMeta } from '@/lib/constants'
import { queryKeys } from '@/lib/query-keys'
import { cn, extractErrorMessage, parseOrigins, slugify } from '@/lib/utils'
import type { SiteWithKey } from '@/types/api'

const originPattern = /^https?:\/\/[^\s/]+$/i

const createSiteSchema = z
  .object({
    name: z
      .string()
      .trim()
      .min(1, '请输入站点名称')
      .max(100, '站点名称最多 100 个字符'),
    slug: z
      .string()
      .trim()
      .min(2, 'Slug 至少 2 个字符')
      .max(64, 'Slug 最多 64 个字符')
      .regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/, '仅支持小写字母、数字和单个连字符'),
    domain: z.string().trim().max(255, '域名最多 255 个字符'),
    allowedOrigins: z.string(),
    commentModeration: z.enum(['pending', 'approved']),
  })
  .superRefine((values, context) => {
    const origins = parseOrigins(values.allowedOrigins)
    for (const origin of origins) {
      if (origin === '*') {
        context.addIssue({
          code: 'custom',
          path: ['allowedOrigins'],
          message: '不允许使用通配符 *；留空表示允许所有来源',
        })
        return
      }
      if (!originPattern.test(origin)) {
        context.addIssue({
          code: 'custom',
          path: ['allowedOrigins'],
          message: `来源 ${origin} 无效，请填写完整 origin，例如 https://blog.example.com`,
        })
        return
      }
    }
  })

type CreateSiteForm = z.infer<typeof createSiteSchema>

const defaultValues: CreateSiteForm = {
  name: '',
  slug: '',
  domain: '',
  allowedOrigins: '',
  commentModeration: 'pending',
}

export function CreateSiteDialog({
  open,
  onClose,
  onCreated,
}: {
  open: boolean
  onClose: () => void
  onCreated: (site: SiteWithKey) => void
}) {
  const [step, setStep] = useState<1 | 2>(1)
  const [slugTouched, setSlugTouched] = useState(false)
  const queryClient = useQueryClient()
  const {
    register,
    handleSubmit,
    trigger,
    watch,
    setValue,
    reset,
    formState: { errors },
  } = useForm<CreateSiteForm>({
    resolver: zodResolver(createSiteSchema),
    defaultValues,
  })

  const name = watch('name')
  const moderation = watch('commentModeration')

  useEffect(() => {
    if (!open) {
      setStep(1)
      setSlugTouched(false)
      reset(defaultValues)
    }
  }, [open, reset])

  useEffect(() => {
    if (!slugTouched) {
      const nextSlug = slugify(name)
      if (nextSlug !== watch('slug')) setValue('slug', nextSlug)
    }
  }, [name, slugTouched, setValue, watch])

  const mutation = useMutation({
    mutationFn: (values: CreateSiteForm) =>
      adminApi.createSite({
        name: values.name,
        slug: values.slug,
        domain: values.domain,
        allowedOrigins: parseOrigins(values.allowedOrigins),
        commentModeration: values.commentModeration,
      }),
    onSuccess: async (site) => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.siteList })
      await queryClient.invalidateQueries({ queryKey: ['admin', 'overview'] })
      toast.success('站点创建成功')
      reset(defaultValues)
      setStep(1)
      onCreated(site)
    },
    onError: (error) => toast.error(extractErrorMessage(error)),
  })

  async function goNext() {
    const valid = await trigger(['name', 'slug', 'domain'])
    if (valid) setStep(2)
  }

  function regenerateSlug() {
    const next = slugify(name)
    setSlugTouched(false)
    setValue('slug', next)
    if (!next) {
      toast.info('站点名称为中文时无法自动生成，请手动填写英文 Slug')
    }
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={step === 1 ? '创建站点' : '评论与来源'}
      description={
        step === 1
          ? '每个站点拥有独立的 Site Key，评论、浏览量和点赞互不影响。'
          : '配置允许调用公共 API 的浏览器来源，以及评论审核策略。'
      }
      size="lg"
      dismissible={!mutation.isPending}
      footer={
        step === 1 ? (
          <>
            <Button
              type="button"
              variant="outline"
              onClick={onClose}
              disabled={mutation.isPending}
            >
              取消
            </Button>
            <Button type="button" onClick={goNext}>
              下一步
              <ArrowRight className="size-4" />
            </Button>
          </>
        ) : (
          <>
            <Button
              type="button"
              variant="outline"
              onClick={() => setStep(1)}
              disabled={mutation.isPending}
            >
              <ArrowLeft className="size-4" />
              上一步
            </Button>
            <Button
              type="button"
              loading={mutation.isPending}
              onClick={handleSubmit((values) => mutation.mutate(values))}
            >
              <Sparkles className="size-4" />
              创建并生成 Key
            </Button>
          </>
        )
      }
    >
      <div className="mb-7 flex items-center gap-2">
        {[1, 2].map((value) => (
          <div key={value} className="flex flex-1 items-center gap-2">
            <span
              className={cn(
                'flex size-7 items-center justify-center rounded-full text-xs font-semibold',
                step >= value
                  ? 'bg-indigo-600 text-white'
                  : 'bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-400',
              )}
            >
              {step > value ? <Check className="size-3.5" /> : value}
            </span>
            <span className="text-xs font-medium text-slate-500 dark:text-slate-400">
              {value === 1 ? '基本信息' : '接入配置'}
            </span>
            {value === 1 ? (
              <span className="h-px flex-1 bg-slate-200 dark:bg-slate-700" />
            ) : null}
          </div>
        ))}
      </div>

      {step === 1 ? (
        <div className="space-y-5">
          <div>
            <Label htmlFor="site-name">站点名称</Label>
            <Input
              id="site-name"
              placeholder="例如：我的技术博客"
              {...register('name')}
            />
            <FieldError message={errors.name?.message} />
          </div>

          <div>
            <Label htmlFor="site-slug">Slug</Label>
            <div className="flex gap-2">
              <Input
                id="site-slug"
                placeholder="my-tech-blog"
                className="font-mono"
                {...register('slug', {
                  onChange: () => setSlugTouched(true),
                })}
              />
              <Button
                type="button"
                variant="outline"
                size="icon"
                className="size-10"
                aria-label="根据名称生成 Slug"
                title="根据名称生成 Slug"
                onClick={regenerateSlug}
              >
                <RefreshCw className="size-4" />
              </Button>
            </div>
            <FieldHint>
              2–64 位小写字母、数字或连字符，用于区分站点，创建后不可修改。
            </FieldHint>
            <FieldError message={errors.slug?.message} />
          </div>

          <div>
            <Label htmlFor="site-domain">主域名（可选）</Label>
            <Input
              id="site-domain"
              placeholder="blog.example.com"
              {...register('domain')}
            />
            <FieldHint>仅用于后台识别，不会限制 API 请求来源。</FieldHint>
            <FieldError message={errors.domain?.message} />
          </div>

          <div className="flex items-start gap-3 rounded-2xl border border-slate-200 bg-slate-50 p-4 dark:border-slate-800 dark:bg-slate-950/50">
            <Globe2 className="mt-0.5 size-4 shrink-0 text-slate-400" />
            <p className="text-xs leading-6 text-slate-500 dark:text-slate-400">
              创建后可继续添加管理员回复、轮换 Key、停用站点或删除站点。
            </p>
          </div>
        </div>
      ) : (
        <div className="space-y-6">
          <div>
            <Label htmlFor="site-origins">允许来源</Label>
            <Textarea
              id="site-origins"
              rows={4}
              placeholder={'https://blog.example.com\nhttp://localhost:36583'}
              className="font-mono text-xs"
              {...register('allowedOrigins')}
            />
            <FieldHint>
              每行一个
              origin；留空表示不限制来源。生产环境建议至少填写正式域名。
            </FieldHint>
            <FieldError message={errors.allowedOrigins?.message} />
          </div>

          <fieldset>
            <legend className="mb-2 text-sm font-medium text-slate-700 dark:text-slate-200">
              评论审核模式
            </legend>
            <div className="grid gap-3 sm:grid-cols-2">
              {(['pending', 'approved'] as const).map((value) => (
                <label
                  key={value}
                  className={cn(
                    'cursor-pointer rounded-2xl border p-4 transition',
                    moderation === value
                      ? 'border-indigo-300 bg-indigo-50/70 ring-4 ring-indigo-500/10 dark:border-indigo-500/50 dark:bg-indigo-500/10'
                      : 'border-slate-200 bg-white hover:border-slate-300 dark:border-slate-700 dark:bg-slate-900 dark:hover:border-slate-600',
                  )}
                >
                  <input
                    type="radio"
                    value={value}
                    className="sr-only"
                    {...register('commentModeration')}
                  />
                  <span className="flex items-center justify-between gap-2">
                    <span className="text-sm font-semibold text-slate-900 dark:text-white">
                      {moderationMeta[value].label}
                    </span>
                    {moderation === value ? (
                      <span className="flex size-5 items-center justify-center rounded-full bg-indigo-600 text-white">
                        <Check className="size-3" />
                      </span>
                    ) : (
                      <span className="size-5 rounded-full border border-slate-300 dark:border-slate-600" />
                    )}
                  </span>
                  <span className="mt-2 block text-xs leading-5 text-slate-500 dark:text-slate-400">
                    {moderationMeta[value].description}
                  </span>
                </label>
              ))}
            </div>
            <FieldError message={errors.commentModeration?.message} />
          </fieldset>
        </div>
      )}
    </Modal>
  )
}
