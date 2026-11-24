const configuredBaseUrl = (
  import.meta.env.VITE_API_BASE_URL || '/api/v1'
).replace(/\/+$/, '')

export const API_BASE_URL = configuredBaseUrl

type ApiEnvelope<T> = {
  code: number
  message: string
  data: T
}

export type QueryParams = Record<
  string,
  string | number | boolean | null | undefined
>

export type ApiRequestOptions = Omit<RequestInit, 'body'> & {
  body?: unknown
  query?: QueryParams
  timeout?: number
  auth?: boolean
}

export class ApiError extends Error {
  status: number
  code: number

  constructor(message: string, status = 0, code = 0) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

type ApiClientConfig = {
  getToken: () => string | null
  onUnauthorized?: () => void
}

let clientConfig: ApiClientConfig = {
  getToken: () => null,
}

export function configureApiClient(config: ApiClientConfig) {
  clientConfig = config
}

function buildUrl(path: string, query?: QueryParams) {
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  const url = new URL(
    `${API_BASE_URL}${normalizedPath}`,
    typeof window === 'undefined' ? 'http://localhost' : window.location.origin,
  )
  if (query) {
    for (const [key, value] of Object.entries(query)) {
      if (value === undefined || value === null || value === '') continue
      url.searchParams.set(key, String(value))
    }
  }
  return url.toString()
}

async function parseEnvelope<T>(response: Response): Promise<T> {
  if (response.status === 204) return undefined as T
  const text = await response.text()
  if (!text) {
    if (response.ok) return undefined as T
    throw new ApiError(response.statusText || '请求失败', response.status)
  }

  let payload: ApiEnvelope<T>
  try {
    payload = JSON.parse(text) as ApiEnvelope<T>
  } catch {
    throw new ApiError(
      response.ok ? '响应格式异常' : response.statusText || '请求失败',
      response.status,
    )
  }

  if (!response.ok || payload.code !== 0) {
    throw new ApiError(
      payload.message || response.statusText || '请求失败',
      response.status,
      payload.code,
    )
  }
  return payload.data
}

export async function apiRequest<T>(
  path: string,
  options: ApiRequestOptions = {},
): Promise<T> {
  const {
    body,
    query,
    timeout = 15_000,
    auth = true,
    headers: rawHeaders,
    signal: externalSignal,
    ...rest
  } = options

  const headers = new Headers(rawHeaders)
  headers.set('Accept', 'application/json')
  if (body !== undefined) headers.set('Content-Type', 'application/json')

  const token = auth ? clientConfig.getToken() : null
  if (token) headers.set('Authorization', `Bearer ${token}`)

  const controller = new AbortController()
  let timedOut = false
  const timeoutId = globalThis.setTimeout(() => {
    timedOut = true
    controller.abort()
  }, timeout)

  const abortFromExternal = () => controller.abort()
  if (externalSignal) {
    if (externalSignal.aborted) controller.abort()
    else
      externalSignal.addEventListener('abort', abortFromExternal, {
        once: true,
      })
  }

  try {
    const response = await fetch(buildUrl(path, query), {
      ...rest,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
      signal: controller.signal,
    })

    if (response.status === 401 && auth) {
      clientConfig.onUnauthorized?.()
    }
    return await parseEnvelope<T>(response)
  } catch (error) {
    if (error instanceof ApiError) throw error
    if (controller.signal.aborted) {
      throw new ApiError(
        timedOut ? '请求超时，请稍后重试' : '请求已取消',
        0,
        timedOut ? 40800 : 49900,
      )
    }
    throw new ApiError(error instanceof Error ? error.message : '网络请求失败')
  } finally {
    globalThis.clearTimeout(timeoutId)
    externalSignal?.removeEventListener('abort', abortFromExternal)
  }
}
