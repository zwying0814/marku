import { create } from 'zustand'

type Theme = 'light' | 'dark'

const STORAGE_KEY = 'marku-admin-theme'

function readTheme(): Theme {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored === 'light' || stored === 'dark') return stored
  } catch {
    // fall through to system preference
  }
  return window.matchMedia?.('(prefers-color-scheme: dark)').matches
    ? 'dark'
    : 'light'
}

function applyTheme(theme: Theme) {
  document.documentElement.classList.toggle('dark', theme === 'dark')
  document.documentElement.style.colorScheme = theme
}

const initialTheme = readTheme()
applyTheme(initialTheme)

type ThemeState = {
  theme: Theme
  setTheme: (theme: Theme) => void
  toggleTheme: () => void
}

export const useThemeStore = create<ThemeState>()((set, get) => ({
  theme: initialTheme,
  setTheme(theme) {
    applyTheme(theme)
    try {
      localStorage.setItem(STORAGE_KEY, theme)
    } catch {
      // ignore unavailable storage
    }
    set({ theme })
  },
  toggleTheme() {
    get().setTheme(get().theme === 'dark' ? 'light' : 'dark')
  },
}))
