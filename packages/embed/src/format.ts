export function formatNumber(value: unknown, lang = 'zh-CN'): string {
  const number = Number(value)
  if (!Number.isFinite(number)) return '0'
  return new Intl.NumberFormat(lang).format(number)
}

export function formatDateTime(value: unknown, lang = 'zh-CN'): string {
  if (!value) return ''
  const date = new Date(String(value))
  if (Number.isNaN(date.getTime())) return ''
  return new Intl.DateTimeFormat(lang, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

export function safeHttpUrl(value: unknown): string | null {
  if (!value) return null
  try {
    const url = new URL(String(value))
    if (url.protocol !== 'http:' && url.protocol !== 'https:') return null
    return url.toString()
  } catch {
    return null
  }
}

export function errorMessage(error: unknown, lang = 'zh-CN'): string {
  const candidate = error as {
    status?: number
    retryAfter?: number
    message?: string
  }
  const status = Number(candidate?.status ?? 0)
  const zh = lang.startsWith('zh')

  if (status === 401) {
    return zh ? 'Site Key 无效或已吊销。' : 'The site key is invalid or revoked.'
  }
  if (status === 403) {
    return zh
      ? '站点已停用，或当前页面来源不在允许来源列表中。'
      : 'The site is disabled, or this origin is not allowed.'
  }
  if (status === 429) {
    const seconds = Number(candidate?.retryAfter ?? 0) || 30
    return zh
      ? `请求过于频繁，请 ${seconds} 秒后再试。`
      : `Too many requests. Try again in ${seconds}s.`
  }
  if (candidate?.message) return candidate.message
  return zh ? '请求失败，请稍后重试。' : 'Request failed. Please try again.'
}
