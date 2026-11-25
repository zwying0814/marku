import { describe, expect, it, vi } from 'vitest'

import { MarkuClient } from './client'
import { MarkuError, normalizePath } from './support'

interface FakeResponse {
  ok: boolean
  status: number
  statusText: string
  headers: { get(name: string): string | null }
  text(): Promise<string>
}

function jsonResponse(
  data: unknown,
  init: { status?: number; headers?: Record<string, string> } = {},
) {
  const status = init.status ?? 200
  const headers = init.headers ?? {}
  return {
    ok: status >= 200 && status < 300,
    status,
    statusText: status === 200 ? 'OK' : 'Error',
    headers: { get: (name: string) => headers[name] ?? null },
    text: async () => JSON.stringify({ code: 0, message: 'ok', data }),
  } satisfies FakeResponse
}

function createClient(
  fetchImpl: (url: string, init?: RequestInit) => Promise<FakeResponse>,
) {
  return new MarkuClient({
    siteKey: 'mk_test_key',
    baseUrl: 'https://api.example.com/api/v1/',
    visitorId: 'visitor-1',
    fetch: fetchImpl as unknown as typeof fetch,
  })
}

describe('normalizePath', () => {
  it('adds a leading slash and strips query, hash and trailing slash', () => {
    expect(normalizePath('posts/hello/?utm_source=x#comments')).toBe('/posts/hello')
  })

  it('falls back to root for empty input', () => {
    expect(normalizePath('  ')).toBe('/')
  })
})

describe('MarkuClient', () => {
  it('sends site key and unwraps the envelope', async () => {
    const fetchImpl = vi.fn(async () => jsonResponse({ name: 'Blog' }))
    const client = createClient(fetchImpl)

    const site = await client.site()

    expect(site).toEqual({ name: 'Blog' })
    const [url, init] = fetchImpl.mock.calls[0] as unknown as [string, RequestInit]
    expect(url).toBe('https://api.example.com/api/v1/public/site')
    expect(init.headers).toMatchObject({
      'X-Site-Key': 'mk_test_key',
      Accept: 'application/json',
    })
  })

  it('batches stats paths and visitor id', async () => {
    const fetchImpl = vi.fn(async () =>
      jsonResponse({ items: [{ path: '/a', views: 3 }, { path: '/b' }] }),
    )
    const client = createClient(fetchImpl)

    const stats = await client.stats(['a', '/b/'])

    expect(stats).toHaveLength(2)
    const [requestUrl] = fetchImpl.mock.calls[0] as unknown as [string, RequestInit]
    const url = new URL(requestUrl)
    expect(url.searchParams.getAll('paths')).toEqual(['/a', '/b'])
    expect(url.searchParams.get('visitorId')).toBe('visitor-1')
  })

  it('normalizes paths on comment creation', async () => {
    const fetchImpl = vi.fn(async () => jsonResponse({ id: 7, status: 'pending' }))
    const client = createClient(fetchImpl)

    const result = await client.comments.create({
      path: 'posts/hello/',
      authorName: 'Ada',
      content: '你好',
    })

    expect(result).toEqual({ id: 7, status: 'pending' })
    const [, init] = fetchImpl.mock.calls[0] as unknown as [string, RequestInit]
    expect(init.method).toBe('POST')
    expect(JSON.parse(String(init.body))).toMatchObject({
      path: '/posts/hello',
      authorName: 'Ada',
      content: '你好',
    })
    expect((init.headers as Record<string, string>)['Content-Type']).toBe(
      'application/json',
    )
  })

  it('maps API errors with status, code and retry-after', async () => {
    const fetchImpl = vi.fn(async () => ({
      ok: false,
      status: 429,
      statusText: 'Too Many Requests',
      headers: { get: (name: string) => (name === 'Retry-After' ? '30' : null) },
      text: async () =>
        JSON.stringify({ code: 42900, message: 'too many requests', data: null }),
    }))
    const client = createClient(fetchImpl)

    const error = await client.comments
      .create({ path: '/a', authorName: 'Ada', content: 'hi' })
      .catch((value: unknown) => value)

    expect(error).toBeInstanceOf(MarkuError)
    expect(error).toMatchObject({
      status: 429,
      code: 42900,
      retryAfter: 30,
      message: 'too many requests',
    })
  })

  it('rejects non-JSON responses', async () => {
    const client = createClient(async () => ({
      ok: true,
      status: 200,
      statusText: 'OK',
      headers: { get: () => null },
      text: async () => '<html>oops</html>',
    }))

    await expect(client.site()).rejects.toThrow('non-JSON')
  })

  it('requires a visitor id for likes when disabled', async () => {
    const client = new MarkuClient({
      siteKey: 'mk_test_key',
      visitorId: null,
      fetch: (async () => jsonResponse({})) as unknown as typeof fetch,
    })

    await expect(client.likes.toggle('/a')).rejects.toThrow('visitorId')
  })

  it('scopes requests with forPath', async () => {
    const fetchImpl = vi.fn(async () => jsonResponse({ items: [], total: 0 }))
    const client = createClient(fetchImpl)

    await client.forPath('/posts/hello').comments.list({ page: 2 })

    const [requestUrl] = fetchImpl.mock.calls[0] as unknown as [string, RequestInit]
    const url = new URL(requestUrl)
    expect(url.pathname).toBe('/api/v1/public/comments')
    expect(url.searchParams.get('path')).toBe('/posts/hello')
    expect(url.searchParams.get('page')).toBe('2')
  })

  it('aborts after the configured timeout', async () => {
    const client = new MarkuClient({
      siteKey: 'mk_test_key',
      timeoutMs: 10,
      fetch: ((_url: string, init?: RequestInit) =>
        new Promise((_resolve, reject) => {
          init?.signal?.addEventListener('abort', () => {
            reject(Object.assign(new Error('aborted'), { name: 'AbortError' }))
          })
        })) as unknown as typeof fetch,
    })

    await expect(client.site()).rejects.toThrow('timed out')
  })
})
