import { useEffect, type ReactNode } from 'react'
import { Navigate, useLocation } from 'react-router-dom'

import { useAuthStore } from '@/stores/auth'

export function ProtectedRoute({ children }: { children: ReactNode }) {
  const location = useLocation()
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const expiresAt = useAuthStore((state) => state.expiresAt)
  const logout = useAuthStore((state) => state.logout)
  const expired = Boolean(
    expiresAt && new Date(expiresAt).getTime() <= Date.now(),
  )

  useEffect(() => {
    if (expired) logout()
  }, [expired, logout])

  if (!isAuthenticated || expired) {
    return (
      <Navigate
        to="/login"
        replace
        state={{ from: `${location.pathname}${location.search}` }}
      />
    )
  }
  return children
}
