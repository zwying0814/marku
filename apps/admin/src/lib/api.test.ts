import { beforeEach, describe, expect, it, vi } from 'vitest'

import { ApiError, apiRequest, configureApiClient } from '@/lib/api'

function jsonResponse(payload: unknown, status = 200) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

describe('apiRequest', () => {
  beforeEach(() => {
    configureApiClient({ getToken: () => null })
  })

  it('unwraps the response envelope and appends query parameters', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      jsonResponse({
        code: 0,
        message: 'ok',
        data: { items: [1, 2] },
      }),
    )
    vi.stubGlobal('fetch', fetchMock)

    const result = await apiRequest<{ items: number[] }>('/admin/comments', {
      query: { page: 2, q: undefined, status: '' },
    })

    expect(result).toEqual({ items: [1, 2] })
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit]
    const parsed = new URL(url)
    expect(parsed.pathname).toBe('/api/v1/admin/comments')
    expect(parsed.searchParams.get('page')).toBe('2')
    expect(parsed.searchParams.has('q')).toBe(false)
    expect(parsed.searchParams.has('status')).toBe(false)
    expect(init.method).toBeUndefined()
  })

  it('attaches the bearer token and JSON body', async () => {
    configureApiClient({ getToken: () => 'jwt-token' })
    const fetchMock = vi
      .fn()
      .mockResolvedValue(
        jsonResponse({ code: 0, message: 'ok', data: { id: 1 } }),
      )
    vi.stubGlobal('fetch', fetchMock)

    await apiRequest('/admin/sites', {
      method: 'POST',
      body: { name: 'Test' },
    })

    const [, init] = fetchMock.mock.calls[0] as [string, RequestInit]
    const headers = init.headers as Headers
    expect(headers.get('Authorization')).toBe('Bearer jwt-token')
    expect(headers.get('Content-Type')).toBe('application/json')
    expect(init.body).toBe(JSON.stringify({ name: 'Test' }))
  })

  it('normalizes API errors', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        jsonResponse(
          {
            code: 40900,
            message: 'site slug already exists',
            data: null,
          },
          409,
        ),
      ),
    )

    await expect(apiRequest('/admin/sites')).rejects.toMatchObject({
      name: 'ApiError',
      status: 409,
      code: 40900,
      message: 'site slug already exists',
    })
  })

  it('calls the unauthorized handler on a protected 401', async () => {
    const onUnauthorized = vi.fn()
    configureApiClient({ getToken: () => 'expired', onUnauthorized })
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValue(
          jsonResponse(
            { code: 40100, message: 'invalid or expired token', data: null },
            401,
          ),
        ),
    )

    await expect(apiRequest('/admin/overview')).rejects.toBeInstanceOf(ApiError)
    expect(onUnauthorized).toHaveBeenCalledOnce()
  })

  it('returns undefined for 204 responses', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(new Response(null, { status: 204 })),
    )

    await expect(
      apiRequest<void>('/admin/comments/1', { method: 'DELETE' }),
    ).resolves.toBeUndefined()
  })
})
