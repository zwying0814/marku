import {
  MarkuError,
  createVisitorId,
  getVisitorId,
  normalizePath,
} from './support'
import type {
  CommentListQuery,
  CommentSubmitResult,
  CreateCommentInput,
  LikeState,
  PageStats,
  PageViewItem,
  PublicCommentPage,
  PublicSite,
  RecordViewResult,
} from './types'

export type MarkuFetch = (
  input: RequestInfo | URL,
  init?: RequestInit,
) => Promise<Response>

export interface MarkuClientOptions {
  siteKey: string
  baseUrl?: string
  fetch?: MarkuFetch
  visitorId?: string | null | (() => string | null | undefined)
  timeoutMs?: number
  headers?: Record<string, string>
}

export interface RequestOptions {
  signal?: AbortSignal
  timeoutMs?: number
}

type QueryValue = string | number | boolean | null | undefined
type Query = Record<string, QueryValue | QueryValue[]>

interface RequestInput extends RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  query?: Query
  body?: unknown
}

const DEFAULT_BASE_URL = '/api/v1'
const DEFAULT_TIMEOUT_MS = 15_000

export class MarkuClient {
  readonly siteKey: string
  readonly baseUrl: string
  private readonly fetchImpl: MarkuFetch
  private readonly timeoutMs: number
  private readonly headers: Record<string, string>
  private readonly visitorOption: MarkuClientOptions['visitorId']
  private explicitVisitorId: string | null | undefined

  constructor(options: MarkuClientOptions) {
    const siteKey = String(options?.siteKey ?? '').trim()
    if (!siteKey) {
      throw new TypeError('MarkuClient requires a siteKey')
    }
    if (!siteKey.startsWith('mk_')) {
      throw new TypeError('MarkuClient siteKey should start with "mk_"')
    }

    const fetchImpl = options.fetch ?? globalThis.fetch
    if (typeof fetchImpl !== 'function') {
      throw new TypeError('MarkuClient requires a fetch implementation')
    }

    this.siteKey = siteKey
    this.baseUrl = (options.baseUrl ?? DEFAULT_BASE_URL).trim().replace(/\/+$/, '')
    this.fetchImpl = fetchImpl.bind(globalThis)
    this.timeoutMs = options.timeoutMs ?? DEFAULT_TIMEOUT_MS
    this.headers = { ...options.headers }
    this.visitorOption = options.visitorId
  }

  setVisitorId(visitorId: string | null): void {
    this.explicitVisitorId = visitorId
  }

  resolveVisitorId(): string {
    if (this.explicitVisitorId !== undefined) {
      if (this.explicitVisitorId) return this.explicitVisitorId
      throw new MarkuError('visitorId is required for this request')
    }
    if (typeof this.visitorOption === 'function') {
      const value = this.visitorOption()
      if (value) return value
      throw new MarkuError('visitorId is required for this request')
    }
    if (this.visitorOption === null) {
      throw new MarkuError('visitorId is required for this request')
    }
    return this.visitorOption || getVisitorId()
  }

  async site(options: RequestOptions = {}): Promise<PublicSite> {
    return this.request<PublicSite>('/public/site', options)
  }

  async stats(
    paths: string | string[],
    options: RequestOptions & { visitorId?: string } = {},
  ): Promise<PageStats[]> {
    const list = (Array.isArray(paths) ? paths : [paths]).map(normalizePath)
    if (list.length === 0) return []

    const query: Query = { paths: list }
    const visitorId = options.visitorId ?? this.optionalVisitorId()
    if (visitorId) query.visitorId = visitorId

    const data = await this.request<{ items?: PageStats[] }>('/public/stats', {
      ...options,
      query,
    })
    return data?.items ?? []
  }

  async pageStats(path: string, options: RequestOptions = {}): Promise<PageStats> {
    const [stats] = await this.stats(path, options)
    return stats ?? { path: normalizePath(path) }
  }

  readonly comments = {
    list: async (
      input: CommentListQuery,
      options: RequestOptions = {},
    ): Promise<PublicCommentPage> => {
      const data = await this.request<PublicCommentPage>('/public/comments', {
        ...options,
        query: {
          path: normalizePath(input.path),
          page: input.page,
          pageSize: input.pageSize,
          sort: input.sort,
        },
      })
      return data ?? {}
    },

    count: async (path: string, options: RequestOptions = {}): Promise<number> => {
      const data = await this.request<{ count?: number }>('/public/comments/count', {
        ...options,
        query: { path: normalizePath(path) },
      })
      return data?.count ?? 0
    },

    create: async (
      input: CreateCommentInput,
      options: RequestOptions = {},
    ): Promise<CommentSubmitResult> => {
      const data = await this.request<CommentSubmitResult>('/public/comments', {
        ...options,
        method: 'POST',
        body: { ...input, path: normalizePath(input.path) },
      })
      return data ?? {}
    },
  }

  readonly views = {
    record: async (
      path: string,
      options: RequestOptions & { visitorId?: string } = {},
    ): Promise<RecordViewResult> => {
      const visitorId = options.visitorId ?? this.optionalVisitorId()
      const data = await this.request<RecordViewResult>('/public/views', {
        ...options,
        method: 'POST',
        body: { path: normalizePath(path), visitorId },
      })
      return data ?? {}
    },

    list: async (
      paths: string | string[],
      options: RequestOptions = {},
    ): Promise<PageViewItem[]> => {
      const list = (Array.isArray(paths) ? paths : [paths]).map(normalizePath)
      if (list.length === 0) return []
      const data = await this.request<{ items?: PageViewItem[] }>('/public/views', {
        ...options,
        query: { paths: list },
      })
      return data?.items ?? []
    },
  }

  readonly likes = {
    get: async (
      path: string,
      options: RequestOptions = {},
    ): Promise<LikeState> => {
      const data = await this.request<LikeState>('/public/likes', {
        ...options,
        query: {
          path: normalizePath(path),
          visitorId: this.optionalVisitorId(),
        },
      })
      return data ?? {}
    },

    toggle: async (
      path: string,
      options: RequestOptions = {},
    ): Promise<LikeState> => {
      const data = await this.request<LikeState>('/public/likes', {
        ...options,
        method: 'POST',
        body: { path: normalizePath(path), visitorId: this.requireVisitorId() },
      })
      return data ?? {}
    },

    set: async (
      path: string,
      liked: boolean,
      options: RequestOptions = {},
    ): Promise<LikeState> => {
      const normalized = normalizePath(path)
      const visitorId = this.requireVisitorId()
      const data = await this.request<LikeState>('/public/likes', {
        ...options,
        method: liked ? 'PUT' : 'DELETE',
        query: liked ? undefined : { path: normalized, visitorId },
        body: liked ? { path: normalized, visitorId } : undefined,
      })
      return data ?? {}
    },
  }

  forPath(path: string) {
    const normalized = normalizePath(path)
    return {
      path: normalized,
      stats: (options?: RequestOptions) => this.pageStats(normalized, options),
      comments: {
        list: (
          input: Omit<CommentListQuery, 'path'> = {},
          options?: RequestOptions,
        ) => this.comments.list({ ...input, path: normalized }, options),
        count: (options?: RequestOptions) => this.comments.count(normalized, options),
        create: (
          input: Omit<CreateCommentInput, 'path'>,
          options?: RequestOptions,
        ) => this.comments.create({ ...input, path: normalized }, options),
      },
      views: {
        record: (options?: RequestOptions) => this.views.record(normalized, options),
        list: (options?: RequestOptions) => this.views.list(normalized, options),
      },
      likes: {
        get: (options?: RequestOptions) => this.likes.get(normalized, options),
        toggle: (options?: RequestOptions) => this.likes.toggle(normalized, options),
        set: (liked: boolean, options?: RequestOptions) =>
          this.likes.set(normalized, liked, options),
      },
    }
  }

  private optionalVisitorId(): string | undefined {
    try {
      return this.resolveVisitorId()
    } catch {
      return undefined
    }
  }

  private requireVisitorId(): string {
    return this.resolveVisitorId()
  }

  private async request<T>(path: string, input: RequestInput = {}): Promise<T> {
    const url = this.buildUrl(path, input.query)
    const controller = new AbortController()
    const timeoutMs = input.timeoutMs ?? this.timeoutMs
    const timeout = setTimeout(() => controller.abort(), Math.max(0, timeoutMs))
    const signal = mergeSignals(controller.signal, input.signal)

    const headers: Record<string, string> = {
      Accept: 'application/json',
      'X-Site-Key': this.siteKey,
      ...this.headers,
    }
    const init: RequestInit = {
      method: input.method ?? 'GET',
      headers,
      signal,
    }
    if (input.body !== undefined) {
      headers['Content-Type'] = 'application/json'
      init.body = JSON.stringify(input.body)
    }

    let response: Response
    try {
      response = await this.fetchImpl(url, init)
    } catch (error) {
      const aborted = isAbortError(error)
      throw new MarkuError(
        aborted
          ? 'Marku request timed out'
          : 'Unable to reach the Marku API. Check api URL, allowed origins and network.',
        { url, cause: error },
      )
    } finally {
      clearTimeout(timeout)
    }

    const text = await response.text()
    let payload: { code?: number; message?: string; data?: unknown } | null = null
    if (text) {
      try {
        payload = JSON.parse(text)
      } catch {
        throw new MarkuError('Marku API returned a non-JSON response', {
          status: response.status,
          url,
        })
      }
    }

    if (!response.ok || (payload?.code ?? 0) !== 0) {
      throw new MarkuError(
        payload?.message || response.statusText || 'Marku request failed',
        {
          status: response.status,
          code: payload?.code ?? 0,
          retryAfter: Number(response.headers.get('Retry-After') ?? 0) || 0,
          data: payload?.data,
          url,
        },
      )
    }

    return (payload?.data ?? null) as T
  }

  private buildUrl(path: string, query?: Query): string {
    const root =
      typeof location !== 'undefined' && location.origin
        ? location.origin
        : 'http://localhost'
    const url = new URL(`${this.baseUrl}${path}`, root)
    for (const [key, value] of Object.entries(query ?? {})) {
      if (value === undefined || value === null || value === '') continue
      if (Array.isArray(value)) {
        for (const item of value) {
          if (item === undefined || item === null || item === '') continue
          url.searchParams.append(key, String(item))
        }
        continue
      }
      url.searchParams.set(key, String(value))
    }
    return url.toString()
  }
}

export function createMarkuClient(options: MarkuClientOptions): MarkuClient {
  return new MarkuClient(options)
}

export function createMarkuVisitorId(): string {
  return createVisitorId()
}

function mergeSignals(signal: AbortSignal, other?: AbortSignal): AbortSignal {
  if (!other) return signal
  if (other.aborted) return other

  const controller = new AbortController()
  const abort = () => controller.abort()
  signal.addEventListener('abort', abort, { once: true })
  other.addEventListener('abort', abort, { once: true })
  return controller.signal
}

function isAbortError(error: unknown): boolean {
  return (
    typeof error === 'object' &&
    error !== null &&
    'name' in error &&
    (error as { name?: string }).name === 'AbortError'
  )
}
