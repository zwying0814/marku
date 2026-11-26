import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation } from '@tanstack/react-query'
import {
  ArrowRight,
  Eye,
  EyeOff,
  Globe2,
  Heart,
  MessageSquare,
  ShieldCheck,
} from 'lucide-react'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Navigate, useNavigate, useSearchParams } from 'react-router-dom'
import { toast } from 'sonner'
import { z } from 'zod'

import { ThemeToggle } from '@/components/ThemeToggle'
import { Button } from '@/components/ui/button'
import { FieldError, Input, Label } from '@/components/ui/form'
import { adminApi } from '@/lib/admin-api'
import { API_BASE_URL } from '@/lib/api'
import { extractErrorMessage } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth'

const loginSchema = z.object({
  email: z.string().trim().min(1, '请输入账号').max(64, '账号最多 64 个字符'),
  password: z.string().min(1, '请输入密码'),
})

type LoginForm = z.infer<typeof loginSchema>

function safeNextPath(value: string | null) {
  if (!value || !value.startsWith('/') || value.startsWith('//')) return '/'
  if (value.startsWith('/login')) return '/'
  return value
}

export function LoginPage() {
  const [showPassword, setShowPassword] = useState(false)
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const login = useAuthStore((state) => state.login)
  const nextPath = safeNextPath(searchParams.get('next'))

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: '', password: '' },
  })

  const mutation = useMutation({
    mutationFn: (values: LoginForm) =>
      adminApi.login(values.email, values.password),
    onSuccess: (result) => {
      login(result)
      toast.success(`欢迎回来，${result.admin.nickname || result.admin.email}`)
      navigate(nextPath, { replace: true })
    },
    onError: (error) => {
      const message = extractErrorMessage(error)
      setError('root', { message })
      toast.error(message)
    },
  })

  if (isAuthenticated) return <Navigate to={nextPath} replace />

  return (
    <div className="grid min-h-screen bg-white lg:grid-cols-[1.05fr_0.95fr] dark:bg-slate-950">
      <section className="relative hidden overflow-hidden bg-slate-950 p-12 text-white lg:flex lg:flex-col">
        <div className="absolute -left-24 top-1/4 size-80 rounded-full bg-indigo-600/30 blur-3xl" />
        <div className="absolute -bottom-20 right-0 size-96 rounded-full bg-violet-600/20 blur-3xl" />
        <div className="relative flex items-center gap-3">
          <span className="flex size-10 items-center justify-center rounded-xl bg-white/10 text-sm font-bold ring-1 ring-white/15 backdrop-blur">
            M
          </span>
          <div>
            <p className="text-sm font-semibold">Marku</p>
            <p className="text-xs text-slate-400">Headless blog API</p>
          </div>
        </div>

        <div className="relative my-auto max-w-lg">
          <p className="text-[11px] font-semibold uppercase tracking-[0.28em] text-indigo-300">
            Dynamic content for static blogs
          </p>
          <h1 className="mt-5 text-4xl font-semibold leading-[1.15] tracking-tight xl:text-5xl">
            静态博客，
            <br />
            也能拥有鲜活的互动。
          </h1>
          <p className="mt-5 max-w-md text-sm leading-7 text-slate-400">
            一个 API 管理多个站点。评论审核、浏览量统计与点赞数据各自独立，
            不改变现有静态发布流程。
          </p>

          <div className="mt-10 grid gap-3 sm:grid-cols-3">
            {[
              { icon: MessageSquare, label: '评论', detail: '审核与回复' },
              { icon: Globe2, label: '多站点', detail: '数据隔离' },
              { icon: Heart, label: '点赞', detail: '实时互动' },
            ].map((item) => (
              <div
                key={item.label}
                className="rounded-2xl border border-white/10 bg-white/5 p-4 backdrop-blur"
              >
                <item.icon className="size-4 text-indigo-300" />
                <p className="mt-3 text-sm font-medium">{item.label}</p>
                <p className="mt-1 text-[11px] text-slate-500">{item.detail}</p>
              </div>
            ))}
          </div>
        </div>

        <div className="relative flex items-center gap-2 text-xs text-slate-500">
          <ShieldCheck className="size-4" />
          JWT 会话仅保存在当前浏览器标签页
        </div>
      </section>

      <section className="relative flex items-center justify-center px-5 py-12 sm:px-10">
        <div className="absolute right-5 top-5">
          <ThemeToggle />
        </div>
        <div className="w-full max-w-sm animate-fade-in">
          <div className="mb-10 flex items-center gap-3 lg:hidden">
            <span className="flex size-10 items-center justify-center rounded-xl bg-slate-900 text-sm font-bold text-white dark:bg-white dark:text-slate-900">
              M
            </span>
            <div>
              <p className="text-sm font-semibold text-slate-950 dark:text-white">
                Marku
              </p>
              <p className="text-[11px] text-slate-500">Headless blog API</p>
            </div>
          </div>

          <p className="text-[11px] font-semibold uppercase tracking-[0.22em] text-indigo-600 dark:text-indigo-400">
            Admin console
          </p>
          <h2 className="mt-2 text-3xl font-semibold tracking-tight text-slate-950 dark:text-white">
            登录管理后台
          </h2>
          <p className="mt-2 text-sm leading-6 text-slate-500 dark:text-slate-400">
            使用管理员账号继续管理站点与内容。
          </p>

          <form
            className="mt-8 space-y-5"
            onSubmit={handleSubmit((values) => mutation.mutate(values))}
            noValidate
          >
            <div>
              <Label htmlFor="email">账号</Label>
              <Input
                id="email"
                type="text"
                autoComplete="username"
                placeholder="admin"
                aria-invalid={Boolean(errors.email)}
                {...register('email')}
              />
              <FieldError message={errors.email?.message} />
            </div>

            <div>
              <Label htmlFor="password">密码</Label>
              <div className="relative">
                <Input
                  id="password"
                  type={showPassword ? 'text' : 'password'}
                  autoComplete="current-password"
                  placeholder="输入管理员密码"
                  className="pr-11"
                  aria-invalid={Boolean(errors.password)}
                  {...register('password')}
                />
                <button
                  type="button"
                  className="absolute right-1 top-1 flex size-8 items-center justify-center rounded-lg text-slate-400 transition hover:bg-slate-100 hover:text-slate-700 dark:hover:bg-slate-800 dark:hover:text-slate-200"
                  aria-label={showPassword ? '隐藏密码' : '显示密码'}
                  onClick={() => setShowPassword((visible) => !visible)}
                >
                  {showPassword ? (
                    <EyeOff className="size-4" />
                  ) : (
                    <Eye className="size-4" />
                  )}
                </button>
              </div>
              <FieldError message={errors.password?.message} />
            </div>

            {errors.root?.message ? (
              <div
                role="alert"
                className="rounded-xl border border-rose-200 bg-rose-50 px-3.5 py-3 text-xs font-medium text-rose-700 dark:border-rose-500/20 dark:bg-rose-500/10 dark:text-rose-300"
              >
                {errors.root.message}
              </div>
            ) : null}

            <Button
              type="submit"
              size="lg"
              className="w-full"
              loading={mutation.isPending}
            >
              登录
              <ArrowRight className="size-4" />
            </Button>
          </form>

          <p className="mt-6 rounded-xl border border-amber-200 bg-amber-50 px-3.5 py-3 text-xs leading-5 text-amber-800 dark:border-amber-500/20 dark:bg-amber-500/10 dark:text-amber-200">
            首次部署默认账号为 <code className="font-mono">admin</code>，默认密码为{' '}
            <code className="font-mono">123456</code>。登录后请立即在「账号与安全」中修改。
          </p>

          <div className="mt-8 rounded-2xl border border-slate-200 bg-slate-50 p-4 dark:border-slate-800 dark:bg-slate-900">
            <p className="text-[11px] font-semibold uppercase tracking-wider text-slate-400">
              API endpoint
            </p>
            <p className="mt-1.5 break-all font-mono text-[11px] text-slate-600 dark:text-slate-300">
              {API_BASE_URL}
            </p>
          </div>
        </div>
      </section>
    </div>
  )
}
