import { beforeEach, describe, expect, it, vi } from 'vitest'

import './index'

interface Route {
  method: string
  path: string
  reply: (init?: RequestInit) => unknown
}

let routes: Route[] = []

function respond(data: unknown, status = 200) {
  return {
    ok: status >= 200 && status < 300,
    status,
    statusText: status === 200 ? 'OK' : 'Error',
    headers: { get: () => null },
    text: async () => JSON.stringify({ code: 0, message: 'ok', data }),
  }
}

function mount(tag: string, attributes: Record<string, string> = {}) {
  const element = document.createElement(tag)
  element.setAttribute('api', 'https://api.test/api/v1')
  element.setAttribute('site-key', 'mk_test_key')
  element.setAttribute('path', '/posts/hello')
  for (const [name, value] of Object.entries(attributes)) {
    element.setAttribute(name, value)
  }
  document.body.append(element)
  return element
}

function shadowText(element: Element): string {
  return element.shadowRoot?.textContent ?? element.textContent ?? ''
}

async function waitFor(check: () => boolean, timeout = 1000): Promise<void> {
  const started = Date.now()
  while (!check()) {
    if (Date.now() - started > timeout) throw new Error('waitFor timed out')
    await new Promise((resolve) => setTimeout(resolve, 5))
  }
}

beforeEach(() => {
  routes = []
  localStorage.clear()
  sessionStorage.clear()
  document.head.querySelectorAll('style[data-marku]').forEach((node) => node.remove())
  document.body.replaceChildren()

  vi.stubGlobal('fetch', async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = new URL(String(input))
    const method = (init?.method ?? 'GET').toUpperCase()
    const route = routes.find(
      (candidate) => candidate.method === method && candidate.path === url.pathname,
    )
    if (!route) return respond({ message: `unhandled ${method} ${url.pathname}` }, 500)
    return route.reply(init)
  })
})

describe('marku-comments', () => {
  it('renders comments and nested replies', async () => {
    routes = [
      {
        method: 'GET',
        path: '/api/v1/public/stats',
        reply: () => respond({ items: [{ path: '/posts/hello', comments: 2 }] }),
      },
      {
        method: 'GET',
        path: '/api/v1/public/comments',
        reply: () =>
          respond({
            items: [
              {
                id: 1,
                authorName: 'Ada',
                content: '你好，Marku。',
                createdAt: '2026-01-01T10:00:00Z',
                replies: [
                  { id: 2, parentId: 1, authorName: 'Bob', content: '收到。' },
                ],
              },
            ],
            total: 2,
          }),
      },
    ]

    const element = mount('marku-comments')
    await (element as unknown as { ready: Promise<void> }).ready

    const text = shadowText(element)
    expect(text).toContain('Ada')
    expect(text).toContain('你好，Marku。')
    expect(text).toContain('Bob')
    expect(text).toContain('收到。')
    expect(text).toContain('2')
  })

  it('submits a comment and reports moderation', async () => {
    const captured: { body: Record<string, unknown> | null } = { body: null }
    routes = [
      {
        method: 'GET',
        path: '/api/v1/public/stats',
        reply: () => respond({ items: [] }),
      },
      {
        method: 'GET',
        path: '/api/v1/public/comments',
        reply: () => respond({ items: [], total: 0 }),
      },
      {
        method: 'POST',
        path: '/api/v1/public/comments',
        reply: (init) => {
          captured.body = JSON.parse(String(init?.body))
          return respond({ id: 9, status: 'pending' }, 201)
        },
      },
    ]

    const element = mount('marku-comments')
    const root = element.shadowRoot!
    await (element as unknown as { ready: Promise<void> }).ready

    const name = root.querySelector<HTMLInputElement>('input[name="authorName"]')!
    name.value = 'Ada'
    name.dispatchEvent(new Event('input', { bubbles: true }))

    const content = root.querySelector<HTMLTextAreaElement>('textarea')!
    content.value = '第一条评论'
    content.dispatchEvent(new Event('input', { bubbles: true }))

    root.querySelector('form')!.dispatchEvent(
      new Event('submit', { bubbles: true, cancelable: true }),
    )

    await waitFor(() => shadowText(element).includes('审核'))

    expect(captured.body).toMatchObject({
      path: '/posts/hello',
      authorName: 'Ada',
      content: '第一条评论',
    })
    expect(captured.body?.parentId).toBeNull()
  })
})

describe('marku-like', () => {
  it('toggles the like state', async () => {
    let likes = 3
    routes = [
      {
        method: 'GET',
        path: '/api/v1/public/likes',
        reply: () => respond({ likes, liked: false }),
      },
      {
        method: 'POST',
        path: '/api/v1/public/likes',
        reply: () => {
          likes += 1
          return respond({ likes, liked: true })
        },
      },
    ]

    const element = mount('marku-like')
    await (element as unknown as { ready: Promise<void> }).ready

    const button = element.shadowRoot!.querySelector('button')!
    expect(button.getAttribute('aria-pressed')).toBe('false')

    button.click()
    await waitFor(
      () =>
        element.shadowRoot!.querySelector('button')!.getAttribute('aria-pressed') ===
        'true',
    )

    expect(shadowText(element)).toContain('4')
    expect(element.shadowRoot!.querySelector('button')!.textContent).toContain('已点赞')
  })
})

describe('marku-stats', () => {
  it('records a view once per session', async () => {
    let viewPosts = 0
    routes = [
      {
        method: 'GET',
        path: '/api/v1/public/stats',
        reply: () =>
          respond({ items: [{ path: '/posts/hello', views: 10, comments: 2, likes: 1 }] }),
      },
      {
        method: 'POST',
        path: '/api/v1/public/views',
        reply: () => {
          viewPosts += 1
          return respond({ views: 11, uniqueViews: 7, recorded: true })
        },
      },
    ]

    const first = mount('marku-stats', { 'record-view': '' })
    await (first as unknown as { ready: Promise<void> }).ready
    await waitFor(() => shadowText(first).includes('11'))

    const second = mount('marku-stats', { 'record-view': '' })
    await (second as unknown as { ready: Promise<void> }).ready
    await new Promise((resolve) => setTimeout(resolve, 10))

    expect(viewPosts).toBe(1)
  })
})

describe('light DOM mode', () => {
  it('renders into the element and injects a scoped stylesheet', async () => {
    routes = [
      {
        method: 'GET',
        path: '/api/v1/public/likes',
        reply: () => respond({ likes: 1, liked: false }),
      },
    ]

    const element = mount('marku-like', { 'no-shadow': '' })
    await (element as unknown as { ready: Promise<void> }).ready

    expect(element.shadowRoot).toBeNull()
    expect(element.querySelector('button')).not.toBeNull()
    expect(document.head.querySelector('style[data-marku="marku-like"]')).not.toBeNull()
  })
})

describe('shadow styles', () => {
  it('keeps the injected stylesheet across re-renders', async () => {
    routes = [
      {
        method: 'GET',
        path: '/api/v1/public/likes',
        reply: () => respond({ likes: 2, liked: false }),
      },
      {
        method: 'POST',
        path: '/api/v1/public/likes',
        reply: () => respond({ likes: 3, liked: true }),
      },
    ]

    const element = mount('marku-like')
    await (element as unknown as { ready: Promise<void> }).ready
    element.shadowRoot!.querySelector('button')!.click()
    await waitFor(() => shadowText(element).includes('3'))

    const style = element.shadowRoot!.querySelector('style')
    expect(style?.textContent).toContain('--marku-accent')
  })
})
