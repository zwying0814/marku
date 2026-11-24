import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

const numberFormatter = new Intl.NumberFormat('zh-CN')
const compactFormatter = new Intl.NumberFormat('zh-CN', {
  notation: 'compact',
  maximumFractionDigits: 1,
})

export function formatNumber(value: number | null | undefined) {
  return numberFormatter.format(value ?? 0)
}

export function formatCompact(value: number | null | undefined) {
  const number = value ?? 0
  return Math.abs(number) >= 10_000
    ? compactFormatter.format(number)
    : numberFormatter.format(number)
}

export function formatDate(value: string | null | undefined, withTime = false) {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    ...(withTime ? { hour: '2-digit', minute: '2-digit' } : {}),
  }).format(date)
}

export function formatRelativeTime(value: string | null | undefined) {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  const seconds = Math.round((date.getTime() - Date.now()) / 1000)
  const formatter = new Intl.RelativeTimeFormat('zh-CN', { numeric: 'auto' })
  const units: Array<[Intl.RelativeTimeFormatUnit, number]> = [
    ['year', 60 * 60 * 24 * 365],
    ['month', 60 * 60 * 24 * 30],
    ['week', 60 * 60 * 24 * 7],
    ['day', 60 * 60 * 24],
    ['hour', 60 * 60],
    ['minute', 60],
    ['second', 1],
  ]
  for (const [unit, size] of units) {
    if (Math.abs(seconds) >= size || unit === 'second') {
      return formatter.format(Math.round(seconds / size), unit)
    }
  }
  return '刚刚'
}

export function formatShortDate(value: string) {
  const date = new Date(`${value}T00:00:00Z`)
  if (Number.isNaN(date.getTime())) return value
  return new Intl.DateTimeFormat('zh-CN', {
    month: '2-digit',
    day: '2-digit',
  }).format(date)
}

export function slugify(value: string) {
  return value
    .normalize('NFKD')
    .replace(/[\u0300-\u036f]/g, '')
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 64)
}

export function parseOrigins(value: string) {
  return Array.from(
    new Set(
      value
        .split(/[\n,]/)
        .map((item) => item.trim().replace(/\/+$/, ''))
        .filter(Boolean),
    ),
  )
}

export function truncate(value: string, length = 120) {
  if (value.length <= length) return value
  return `${value.slice(0, length).trimEnd()}…`
}

export function extractErrorMessage(error: unknown) {
  if (error instanceof Error) return error.message
  return '发生未知错误，请稍后重试'
}

export function getInitials(value: string) {
  const trimmed = value.trim()
  if (!trimmed) return 'A'
  return trimmed.slice(0, 2).toUpperCase()
}
