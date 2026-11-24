import { useQuery } from '@tanstack/react-query'
import {
  BarChart3,
  ChevronDown,
  Globe2,
  LayoutDashboard,
  LogOut,
  Menu,
  MessageSquare,
  Settings,
  X,
} from 'lucide-react'
import { useEffect, useState } from 'react'
import { NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom'

import { ThemeToggle } from '@/components/ThemeToggle'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { adminApi } from '@/lib/admin-api'
import { queryKeys } from '@/lib/query-keys'
import { cn, getInitials } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth'

const navigation = [
  { name: '概览', to: '/', icon: LayoutDashboard, end: true },
  { name: '站点', to: '/sites', icon: Globe2 },
  { name: '评论', to: '/comments', icon: MessageSquare },
  { name: '数据', to: '/analytics', icon: BarChart3 },
  { name: '设置', to: '/settings', icon: Settings },
]

const pageTitles: Record<string, string> = {
  '/': '概览',
  '/sites': '站点',
  '/comments': '评论',
  '/analytics': '数据',
  '/settings': '设置',
}

function usePageTitle() {
  const { pathname } = useLocation()
  if (pageTitles[pathname]) return pageTitles[pathname]
  if (pathname.startsWith('/sites/')) return '站点详情'
  return '工作台'
}

export function AdminLayout() {
  const [mobileOpen, setMobileOpen] = useState(false)
  const [userMenuOpen, setUserMenuOpen] = useState(false)
  const location = useLocation()
  const navigate = useNavigate()
  const title = usePageTitle()
  const admin = useAuthStore((state) => state.admin)
  const setAdmin = useAuthStore((state) => state.setAdmin)
  const logout = useAuthStore((state) => state.logout)

  const overviewQuery = useQuery({
    queryKey: queryKeys.overview('7d'),
    queryFn: () => adminApi.overview('7d'),
  })
  const meQuery = useQuery({
    queryKey: queryKeys.me,
    queryFn: adminApi.me,
    staleTime: 5 * 60_000,
  })

  useEffect(() => {
    if (meQuery.data) setAdmin(meQuery.data)
  }, [meQuery.data, setAdmin])

  useEffect(() => {
    setMobileOpen(false)
    setUserMenuOpen(false)
  }, [location.pathname])

  function handleLogout() {
    logout()
    navigate('/login', { replace: true })
  }

  const pendingComments = overviewQuery.data?.pendingComments ?? 0
  const displayName = admin?.nickname || admin?.email || '管理员'

  return (
    <div className="min-h-screen bg-slate-50 text-slate-900 dark:bg-slate-950 dark:text-slate-100">
      {mobileOpen ? (
        <button
          type="button"
          aria-label="关闭导航"
          className="fixed inset-0 z-40 bg-slate-950/45 backdrop-blur-sm lg:hidden"
          onClick={() => setMobileOpen(false)}
        />
      ) : null}

      <aside
        className={cn(
          'fixed inset-y-0 left-0 z-50 flex w-64 flex-col border-r border-slate-200 bg-white px-4 py-5 transition-transform duration-200 lg:translate-x-0 dark:border-slate-800 dark:bg-slate-900',
          mobileOpen ? 'translate-x-0' : '-translate-x-full',
        )}
      >
        <div className="flex items-center justify-between px-2">
          <NavLink to="/" className="flex items-center gap-3">
            <span className="flex size-9 items-center justify-center rounded-xl bg-gradient-to-br from-indigo-500 to-violet-600 text-sm font-bold text-white shadow-lg shadow-indigo-500/20">
              M
            </span>
            <span>
              <span className="block text-sm font-semibold tracking-tight text-slate-950 dark:text-white">
                Marku
              </span>
              <span className="block text-[11px] text-slate-500 dark:text-slate-400">
                Headless blog API
              </span>
            </span>
          </NavLink>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="lg:hidden"
            aria-label="关闭导航"
            onClick={() => setMobileOpen(false)}
          >
            <X className="size-4" />
          </Button>
        </div>

        <nav className="mt-8 space-y-1" aria-label="主导航">
          {navigation.map((item) => {
            const Icon = item.icon
            return (
              <NavLink
                key={item.to}
                to={item.to}
                end={item.end}
                className={({ isActive }) =>
                  cn(
                    'group flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition',
                    isActive
                      ? 'bg-slate-900 text-white shadow-sm dark:bg-white dark:text-slate-900'
                      : 'text-slate-600 hover:bg-slate-100 hover:text-slate-950 dark:text-slate-400 dark:hover:bg-slate-800 dark:hover:text-white',
                  )
                }
              >
                <Icon className="size-4" />
                <span className="flex-1">{item.name}</span>
                {item.to === '/comments' && pendingComments > 0 ? (
                  <span className="rounded-full bg-amber-400 px-1.5 py-0.5 text-[10px] font-bold text-amber-950">
                    {pendingComments > 99 ? '99+' : pendingComments}
                  </span>
                ) : null}
              </NavLink>
            )
          })}
        </nav>

        <div className="mt-auto space-y-4">
          <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4 dark:border-slate-800 dark:bg-slate-950/60">
            <div className="flex items-center gap-2 text-xs font-semibold text-slate-700 dark:text-slate-200">
              <span
                className={cn(
                  'size-2 rounded-full',
                  overviewQuery.isError ? 'bg-rose-500' : 'bg-emerald-500',
                )}
              />
              API {overviewQuery.isError ? '连接异常' : '运行中'}
            </div>
            <p className="mt-2 text-[11px] leading-5 text-slate-500 dark:text-slate-400">
              评论、浏览量、点赞数据按站点完全隔离。
            </p>
          </div>
          <p className="px-2 text-[11px] text-slate-400 dark:text-slate-500">
            Marku Admin v0.1.0
          </p>
        </div>
      </aside>

      <div className="flex min-h-screen min-w-0 flex-col lg:pl-64">
        <header className="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-slate-200 bg-white/85 px-4 backdrop-blur-xl lg:px-8 dark:border-slate-800 dark:bg-slate-950/80">
          <div className="flex min-w-0 items-center gap-3">
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="lg:hidden"
              aria-label="打开导航"
              onClick={() => setMobileOpen(true)}
            >
              <Menu className="size-5" />
            </Button>
            <div className="min-w-0">
              <p className="text-[10px] font-semibold uppercase tracking-[0.22em] text-slate-400 dark:text-slate-500">
                Workspace
              </p>
              <h1 className="truncate text-base font-semibold text-slate-950 dark:text-white">
                {title}
              </h1>
            </div>
            {pendingComments > 0 ? (
              <NavLink
                to="/comments?status=pending"
                className="hidden sm:block"
              >
                <Badge variant="warning">
                  <MessageSquare className="size-3" />
                  {pendingComments} 条待审核
                </Badge>
              </NavLink>
            ) : null}
          </div>

          <div className="flex items-center gap-1.5">
            <ThemeToggle />
            <div className="relative">
              <button
                type="button"
                className="flex items-center gap-2 rounded-xl px-1.5 py-1.5 transition hover:bg-slate-100 dark:hover:bg-slate-800"
                aria-haspopup="menu"
                aria-expanded={userMenuOpen}
                onClick={() => setUserMenuOpen((open) => !open)}
              >
                <span className="flex size-8 items-center justify-center rounded-full bg-gradient-to-br from-slate-800 to-slate-950 text-[11px] font-semibold text-white dark:from-indigo-500 dark:to-violet-600">
                  {getInitials(displayName)}
                </span>
                <span className="hidden max-w-32 truncate text-xs font-medium text-slate-700 sm:block dark:text-slate-200">
                  {displayName}
                </span>
                <ChevronDown className="hidden size-3.5 text-slate-400 sm:block" />
              </button>

              {userMenuOpen ? (
                <>
                  <button
                    type="button"
                    className="fixed inset-0 z-40 cursor-default"
                    aria-label="关闭用户菜单"
                    onClick={() => setUserMenuOpen(false)}
                  />
                  <div
                    role="menu"
                    className="absolute right-0 top-full z-50 mt-2 w-56 rounded-2xl border border-slate-200 bg-white p-1.5 shadow-xl dark:border-slate-800 dark:bg-slate-900"
                  >
                    <div className="border-b border-slate-100 px-3 py-2.5 dark:border-slate-800">
                      <p className="truncate text-sm font-medium text-slate-900 dark:text-white">
                        {displayName}
                      </p>
                      <p className="mt-0.5 truncate text-[11px] text-slate-500 dark:text-slate-400">
                        {admin?.email}
                      </p>
                    </div>
                    <NavLink
                      to="/settings"
                      role="menuitem"
                      className="mt-1 flex items-center gap-2.5 rounded-xl px-3 py-2 text-sm text-slate-600 transition hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800"
                    >
                      <Settings className="size-4" />
                      账号设置
                    </NavLink>
                    <button
                      type="button"
                      role="menuitem"
                      onClick={handleLogout}
                      className="flex w-full items-center gap-2.5 rounded-xl px-3 py-2 text-sm text-rose-600 transition hover:bg-rose-50 dark:text-rose-400 dark:hover:bg-rose-500/10"
                    >
                      <LogOut className="size-4" />
                      退出登录
                    </button>
                  </div>
                </>
              ) : null}
            </div>
          </div>
        </header>

        <main className="flex-1 px-4 py-6 lg:px-8 lg:py-8">
          <div className="mx-auto max-w-7xl">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  )
}
