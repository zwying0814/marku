import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Check,
  KeyRound,
  Laptop,
  Moon,
  ShieldCheck,
  Sun,
  UserRound,
} from 'lucide-react'
import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'
import { z } from 'zod'

import { PageHeader } from '@/components/PageHeader'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { FieldError, Input, Label } from '@/components/ui/form'
import { adminApi } from '@/lib/admin-api'
import { API_BASE_URL } from '@/lib/api'
import { queryKeys } from '@/lib/query-keys'
import { cn, extractErrorMessage, formatDate, getInitials } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth'
import { useThemeStore } from '@/stores/theme'

const profileSchema = z
  .object({
    username: z.string().trim().min(1, '请输入账号').max(64, '账号最多 64 个字符'),
    currentPassword: z.string().min(1, '请输入当前密码'),
    newPassword: z.string().max(128, '新密码最多 128 个字符'),
    confirmPassword: z.string(),
  })
  .refine((values) => values.newPassword === values.confirmPassword, {
    path: ['confirmPassword'],
    message: '两次输入的新密码不一致',
  })
  .refine((values) => !values.newPassword || values.newPassword.length >= 8, {
    path: ['newPassword'],
    message: '新密码至少 8 个字符',
  })

type ProfileForm = z.infer<typeof profileSchema>

export function SettingsPage() {
  const queryClient = useQueryClient()
  const admin = useAuthStore((state) => state.admin)
  const setAdmin = useAuthStore((state) => state.setAdmin)
  const theme = useThemeStore((state) => state.theme)
  const setTheme = useThemeStore((state) => state.setTheme)

  const meQuery = useQuery({
    queryKey: queryKeys.me,
    queryFn: adminApi.me,
    staleTime: 5 * 60_000,
  })

  const {
    register,
    handleSubmit,
    reset,
    setError,
    formState: { errors },
  } = useForm<ProfileForm>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      username: '',
      currentPassword: '',
      newPassword: '',
      confirmPassword: '',
    },
  })

  useEffect(() => {
    if (!meQuery.data) return
    setAdmin(meQuery.data)
    reset({
      username: meQuery.data.email,
      currentPassword: '',
      newPassword: '',
      confirmPassword: '',
    })
  }, [meQuery.data, reset, setAdmin])

  const profileMutation = useMutation({
    mutationFn: (values: ProfileForm) =>
      adminApi.updateProfile({
        currentPassword: values.currentPassword,
        username: values.username,
        newPassword: values.newPassword || undefined,
      }),
    onSuccess: (updated) => {
      setAdmin(updated)
      queryClient.setQueryData(queryKeys.me, updated)
      reset({
        username: updated.email,
        currentPassword: '',
        newPassword: '',
        confirmPassword: '',
      })
      toast.success('账号信息已更新')
    },
    onError: (error) => {
      const message = extractErrorMessage(error)
      setError('root', { message })
      toast.error(message)
    },
  })

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow="Settings"
        title="账号与安全"
        description="查看管理员信息、切换界面主题并更新登录账号与密码。"
      />

      <div className="grid gap-6 lg:grid-cols-[0.9fr_1.1fr]">
        <div className="space-y-6">
          <section className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm dark:border-slate-800 dark:bg-slate-900">
            <div className="flex items-center gap-2">
              <UserRound className="size-4 text-slate-400" />
              <h2 className="text-sm font-semibold text-slate-950 dark:text-white">
                管理员账号
              </h2>
            </div>
            <div className="mt-5 flex items-center gap-4">
              <span className="flex size-14 items-center justify-center rounded-2xl bg-gradient-to-br from-indigo-500 to-violet-600 text-lg font-bold text-white shadow-lg shadow-indigo-500/20">
                {getInitials(admin?.nickname || admin?.email || 'Admin')}
              </span>
              <div className="min-w-0">
                <p className="truncate text-base font-semibold text-slate-950 dark:text-white">
                  {admin?.nickname || '管理员'}
                </p>
                <p className="truncate text-xs text-slate-500 dark:text-slate-400">
                  {admin?.email}
                </p>
              </div>
              <Badge variant="success" className="ml-auto">
                {admin?.status === 'active' ? '正常' : '已停用'}
              </Badge>
            </div>
            <dl className="mt-5 space-y-3 border-t border-slate-100 pt-4 text-xs dark:border-slate-800">
              <div className="flex items-center justify-between gap-3">
                <dt className="text-slate-500 dark:text-slate-400">账号 ID</dt>
                <dd className="font-mono text-slate-800 dark:text-slate-200">
                  #{admin?.id}
                </dd>
              </div>
              <div className="flex items-center justify-between gap-3">
                <dt className="text-slate-500 dark:text-slate-400">最近登录</dt>
                <dd className="text-slate-800 dark:text-slate-200">
                  {formatDate(admin?.lastLoginAt, true)}
                </dd>
              </div>
              <div className="flex items-center justify-between gap-3">
                <dt className="text-slate-500 dark:text-slate-400">创建时间</dt>
                <dd className="text-slate-800 dark:text-slate-200">
                  {formatDate(admin?.createdAt)}
                </dd>
              </div>
            </dl>
          </section>

          <section className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm dark:border-slate-800 dark:bg-slate-900">
            <div className="flex items-center gap-2">
              <Laptop className="size-4 text-slate-400" />
              <h2 className="text-sm font-semibold text-slate-950 dark:text-white">
                界面主题
              </h2>
            </div>
            <div className="mt-4 grid grid-cols-2 gap-3">
              {(
                [
                  {
                    value: 'light',
                    label: '浅色',
                    icon: Sun,
                  },
                  {
                    value: 'dark',
                    label: '深色',
                    icon: Moon,
                  },
                ] as const
              ).map((option) => (
                <button
                  key={option.value}
                  type="button"
                  onClick={() => setTheme(option.value)}
                  className={cn(
                    'flex items-center gap-3 rounded-xl border p-3.5 text-left transition',
                    theme === option.value
                      ? 'border-indigo-300 bg-indigo-50/70 text-indigo-700 ring-4 ring-indigo-500/10 dark:border-indigo-500/40 dark:bg-indigo-500/10 dark:text-indigo-300'
                      : 'border-slate-200 text-slate-600 hover:border-slate-300 dark:border-slate-700 dark:text-slate-300 dark:hover:border-slate-600',
                  )}
                >
                  <option.icon className="size-4" />
                  <span className="flex-1 text-sm font-medium">
                    {option.label}
                  </span>
                  {theme === option.value ? <Check className="size-4" /> : null}
                </button>
              ))}
            </div>
          </section>
        </div>

        <div className="space-y-6">
          <section className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm dark:border-slate-800 dark:bg-slate-900">
            <div className="flex items-center gap-2">
              <KeyRound className="size-4 text-slate-400" />
              <h2 className="text-sm font-semibold text-slate-950 dark:text-white">
                修改账号与密码
              </h2>
            </div>
            <p className="mt-1 text-xs leading-5 text-slate-500 dark:text-slate-400">
              修改登录账号或密码，均需要验证当前密码。新密码留空表示不修改。
            </p>

            <form
              className="mt-5 space-y-4"
              onSubmit={handleSubmit((values) =>
                profileMutation.mutate(values),
              )}
              noValidate
            >
              <div>
                <Label htmlFor="username">登录账号</Label>
                <Input
                  id="username"
                  type="text"
                  autoComplete="username"
                  {...register('username')}
                />
                <FieldError message={errors.username?.message} />
              </div>
              <div>
                <Label htmlFor="current-password">当前密码</Label>
                <Input
                  id="current-password"
                  type="password"
                  autoComplete="current-password"
                  {...register('currentPassword')}
                />
                <FieldError message={errors.currentPassword?.message} />
              </div>
              <div className="grid gap-4 sm:grid-cols-2">
                <div>
                  <Label htmlFor="new-password">新密码（可选）</Label>
                  <Input
                    id="new-password"
                    type="password"
                    autoComplete="new-password"
                    {...register('newPassword')}
                  />
                  <FieldError message={errors.newPassword?.message} />
                </div>
                <div>
                  <Label htmlFor="confirm-password">确认新密码（可选）</Label>
                  <Input
                    id="confirm-password"
                    type="password"
                    autoComplete="new-password"
                    {...register('confirmPassword')}
                  />
                  <FieldError message={errors.confirmPassword?.message} />
                </div>
              </div>
              {errors.root?.message ? (
                <div className="rounded-xl border border-rose-200 bg-rose-50 px-3.5 py-3 text-xs font-medium text-rose-700 dark:border-rose-500/20 dark:bg-rose-500/10 dark:text-rose-300">
                  {errors.root.message}
                </div>
              ) : null}
              <div className="flex justify-end">
                <Button type="submit" loading={profileMutation.isPending}>
                  保存修改
                </Button>
              </div>
            </form>
          </section>

          <section className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm dark:border-slate-800 dark:bg-slate-900">
            <div className="flex items-center gap-2">
              <ShieldCheck className="size-4 text-emerald-500" />
              <h2 className="text-sm font-semibold text-slate-950 dark:text-white">
                会话与 API
              </h2>
            </div>
            <div className="mt-4 space-y-4 text-xs leading-6 text-slate-500 dark:text-slate-400">
              <div>
                <p className="font-semibold text-slate-700 dark:text-slate-200">
                  Token 存储
                </p>
                <p>
                  JWT 存在当前标签页的 <code>sessionStorage</code>
                  ，关闭标签页或退出登录后清除。API 返回 <code>401</code>
                  时会自动清理会话并跳转登录。
                </p>
              </div>
              <div>
                <p className="font-semibold text-slate-700 dark:text-slate-200">
                  API 地址
                </p>
                <code className="mt-1 block break-all rounded-xl bg-slate-100 px-3 py-2 font-mono text-[11px] text-slate-700 dark:bg-slate-950 dark:text-slate-300">
                  {API_BASE_URL}
                </code>
              </div>
              <div>
                <p className="font-semibold text-slate-700 dark:text-slate-200">
                  生产建议
                </p>
                <p>生产环境应使用 HTTPS，并通过反向代理限制管理端访问范围。</p>
              </div>
            </div>
          </section>
        </div>
      </div>
    </div>
  )
}
