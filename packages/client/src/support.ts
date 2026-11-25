export class MarkuError extends Error {
  readonly status: number
  readonly code: number
  readonly retryAfter: number
  readonly data: unknown
  readonly url: string

  constructor(
    message: string,
    options: {
      status?: number
      code?: number
      retryAfter?: number
      data?: unknown
      url?: string
      cause?: unknown
    } = {},
  ) {
    super(message, { cause: options.cause })
    this.name = 'MarkuError'
    this.status = options.status ?? 0
    this.code = options.code ?? 0
    this.retryAfter = options.retryAfter ?? 0
    this.data = options.data
    this.url = options.url ?? ''
  }
}

export function isMarkuError(value: unknown): value is MarkuError {
  return value instanceof MarkuError
}

const MAX_PATH_LENGTH = 512

export function normalizePath(input: string): string {
  let path = String(input ?? '').trim()
  if (!path) return '/'

  const hashIndex = path.indexOf('#')
  if (hashIndex >= 0) path = path.slice(0, hashIndex)
  const queryIndex = path.indexOf('?')
  if (queryIndex >= 0) path = path.slice(0, queryIndex)

  try {
    if (/^[a-z][a-z\d+.-]*:\/\//i.test(path)) {
      path = new URL(path).pathname
    }
  } catch {
    // Keep the original value when the URL cannot be parsed.
  }

  if (!path.startsWith('/')) path = `/${path}`
  if (path.endsWith('/index.html')) {
    path = path.slice(0, -'/index.html'.length)
  }
  while (path.length > 1 && path.endsWith('/')) {
    path = path.slice(0, -1)
  }

  return (path || '/').slice(0, MAX_PATH_LENGTH)
}

const VISITOR_STORAGE_KEY = 'marku:visitor-id'
let memoryVisitorId: string | null = null

export function createVisitorId(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  return `visitor-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`
}

export function getVisitorId(): string {
  if (memoryVisitorId) return memoryVisitorId

  try {
    const stored = globalThis.localStorage?.getItem(VISITOR_STORAGE_KEY)
    if (stored) {
      memoryVisitorId = stored
      return stored
    }
    const created = createVisitorId()
    globalThis.localStorage?.setItem(VISITOR_STORAGE_KEY, created)
    memoryVisitorId = created
    return created
  } catch {
    memoryVisitorId = createVisitorId()
    return memoryVisitorId
  }
}
