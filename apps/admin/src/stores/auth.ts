import { create } from 'zustand'

import { configureApiClient } from '@/lib/api'
import { queryClient } from '@/lib/query-client'
import type { Admin, LoginResult } from '@/types/api'

const STORAGE_KEY = 'marku-admin-session'

type AuthSession = {
  token: string
  expiresAt: string
  admin: Admin | null
}

type AuthState = AuthSession & {
  isAuthenticated: boolean
  login: (result: LoginResult) => void
  logout: () => void
  setAdmin: (admin: Admin) => void
}

const emptySession: AuthSession = {
  token: '',
  expiresAt: '',
  admin: null,
}

function isUsableSession(session: AuthSession) {
  if (!session.token) return false
  if (!session.expiresAt) return true
  const expiresAt = new Date(session.expiresAt).getTime()
  return Number.isNaN(expiresAt) || expiresAt > Date.now()
}

function readSession(): AuthSession {
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY)
    if (!raw) return emptySession
    const parsed = JSON.parse(raw) as Partial<AuthSession>
    const session: AuthSession = {
      token: typeof parsed.token === 'string' ? parsed.token : '',
      expiresAt: typeof parsed.expiresAt === 'string' ? parsed.expiresAt : '',
      admin: parsed.admin ?? null,
    }
    if (!isUsableSession(session)) {
      sessionStorage.removeItem(STORAGE_KEY)
      return emptySession
    }
    return session
  } catch {
    return emptySession
  }
}

function writeSession(session: AuthSession) {
  try {
    if (!session.token) sessionStorage.removeItem(STORAGE_KEY)
    else sessionStorage.setItem(STORAGE_KEY, JSON.stringify(session))
  } catch {
    // sessionStorage can be unavailable in hardened browser contexts.
  }
}

const initialSession = readSession()

export const useAuthStore = create<AuthState>()((set, get) => ({
  ...initialSession,
  isAuthenticated: Boolean(initialSession.token),

  login(result) {
    const session: AuthSession = {
      token: result.token,
      expiresAt: result.expiresAt,
      admin: result.admin,
    }
    writeSession(session)
    set({ ...session, isAuthenticated: true })
  },

  logout() {
    queryClient.clear()
    writeSession(emptySession)
    set({ ...emptySession, isAuthenticated: false })
  },

  setAdmin(admin) {
    const current = get()
    const session = {
      token: current.token,
      expiresAt: current.expiresAt,
      admin,
    }
    writeSession(session)
    set({ admin })
  },
}))

configureApiClient({
  getToken: () => useAuthStore.getState().token,
  onUnauthorized: () => {
    const hadSession = Boolean(useAuthStore.getState().token)
    useAuthStore.getState().logout()
    if (hadSession && window.location.pathname !== '/login') {
      const next = `${window.location.pathname}${window.location.search}`
      window.location.assign(`/login?next=${encodeURIComponent(next)}`)
    }
  },
})
