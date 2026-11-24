import { lazy, Suspense, type ReactNode } from 'react'
import { createBrowserRouter } from 'react-router-dom'

import { PageLoader } from '@/components/PageLoader'
import { AdminLayout } from '@/layouts/AdminLayout'
import { ProtectedRoute } from '@/routes/ProtectedRoute'

const LoginPage = lazy(() =>
  import('@/pages/LoginPage').then((module) => ({
    default: module.LoginPage,
  })),
)
const DashboardPage = lazy(() =>
  import('@/pages/DashboardPage').then((module) => ({
    default: module.DashboardPage,
  })),
)
const SitesPage = lazy(() =>
  import('@/pages/SitesPage').then((module) => ({
    default: module.SitesPage,
  })),
)
const SiteDetailPage = lazy(() =>
  import('@/pages/SiteDetailPage').then((module) => ({
    default: module.SiteDetailPage,
  })),
)
const CommentsPage = lazy(() =>
  import('@/pages/CommentsPage').then((module) => ({
    default: module.CommentsPage,
  })),
)
const AnalyticsPage = lazy(() =>
  import('@/pages/AnalyticsPage').then((module) => ({
    default: module.AnalyticsPage,
  })),
)
const SettingsPage = lazy(() =>
  import('@/pages/SettingsPage').then((module) => ({
    default: module.SettingsPage,
  })),
)
const NotFoundPage = lazy(() =>
  import('@/pages/NotFoundPage').then((module) => ({
    default: module.NotFoundPage,
  })),
)

function page(element: ReactNode) {
  return <Suspense fallback={<PageLoader />}>{element}</Suspense>
}

export const router = createBrowserRouter([
  {
    path: '/login',
    element: page(<LoginPage />),
  },
  {
    path: '/',
    element: (
      <ProtectedRoute>
        <AdminLayout />
      </ProtectedRoute>
    ),
    children: [
      { index: true, element: page(<DashboardPage />) },
      { path: 'sites', element: page(<SitesPage />) },
      { path: 'sites/:siteId', element: page(<SiteDetailPage />) },
      { path: 'comments', element: page(<CommentsPage />) },
      { path: 'analytics', element: page(<AnalyticsPage />) },
      { path: 'settings', element: page(<SettingsPage />) },
      { path: '*', element: page(<NotFoundPage />) },
    ],
  },
])
