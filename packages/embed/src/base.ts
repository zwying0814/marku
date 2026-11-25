import { MarkuClient, MarkuError } from '@marku/client'

import { errorMessage } from './format'
import { h, isAbortError } from './html'
import { baseStyles } from './styles'

const LIGHT_TOKENS: Record<string, string> = {
  '--marku-text': '#1d1d20',
  '--marku-muted': '#6a6a74',
  '--marku-surface': '#ffffff',
  '--marku-border': 'rgba(20, 20, 24, 0.12)',
  '--marku-accent': '#2f6f4e',
  '--marku-accent-text': '#ffffff',
  '--marku-danger': '#b42318',
}

const DARK_TOKENS: Record<string, string> = {
  '--marku-text': '#e8e8ec',
  '--marku-muted': '#a0a0ab',
  '--marku-surface': '#17171b',
  '--marku-border': 'rgba(255, 255, 255, 0.14)',
  '--marku-accent': '#6cc59a',
  '--marku-accent-text': '#0d1a14',
  '--marku-danger': '#ff8a80',
}

export const MARKU_ATTRIBUTES = [
  'api',
  'site-key',
  'path',
  'lang',
  'theme',
  'no-shadow',
] as const

export abstract class MarkuElement extends HTMLElement {
  static styles = ''

  protected target: HTMLElement | ShadowRoot = this
  protected client: MarkuClient | null = null
  get lang(): string {
    return (
      this.getAttribute('lang') ||
      (typeof document !== 'undefined' ? document.documentElement.lang : '') ||
      'zh-CN'
    )
  }

  private signature = ''
  private mounted = false
  private token = 0
  private controller: AbortController | null = null
  private resolveReady?: () => void
  readonly ready: Promise<void>

  constructor() {
    super()
    this.ready = new Promise<void>((resolve) => {
      this.resolveReady = resolve
    })
  }

  connectedCallback(): void {
    if (!this.mounted) {
      this.mounted = true
      this.setupRoot()
    }
    this.applyTheme()
    void this.enqueue()
  }

  disconnectedCallback(): void {
    this.token += 1
    this.controller?.abort()
  }

  attributeChangedCallback(
    name: string,
    oldValue: string | null,
    newValue: string | null,
  ): void {
    if (!this.mounted || oldValue === newValue) return
    if (name === 'theme') this.applyTheme()
    void this.enqueue()
  }

  protected abstract update(): Promise<void> | void

  protected render(...nodes: Node[]): void {
    this.target.replaceChildren(...nodes)
  }

  protected get signal(): AbortSignal | undefined {
    return this.controller?.signal
  }

  protected resolvePath(): string {
    const explicit = this.getAttribute('path')?.trim()
    if (explicit) return explicit

    const meta = document
      .querySelector('meta[name="marku:path"]')
      ?.getAttribute('content')
      ?.trim()
    if (meta) return meta

    const canonical = document
      .querySelector('link[rel="canonical"]')
      ?.getAttribute('href')
    if (canonical) {
      try {
        return new URL(canonical, window.location.origin).pathname
      } catch {
        // Fall back to the current location.
      }
    }
    return window.location.pathname
  }

  protected resolveApi(): string {
    const explicit = this.getAttribute('api')?.trim()
    if (explicit) return explicit
    const meta = document
      .querySelector('meta[name="marku:api"]')
      ?.getAttribute('content')
      ?.trim()
    if (meta) return meta
    return '/api/v1'
  }

  protected clientOrNull(): MarkuClient | null {
    const siteKey = this.getAttribute('site-key')?.trim()
    if (!siteKey) {
      this.client = null
      this.signature = ''
      return null
    }

    const signature = `${this.resolveApi()}|${siteKey}`
    if (!this.client || signature !== this.signature) {
      try {
        this.client = new MarkuClient({
          siteKey,
          baseUrl: this.resolveApi(),
        })
        this.signature = signature
      } catch {
        this.client = null
        this.signature = ''
      }
    }
    return this.client
  }

  protected configError(): Node | null {
    if (this.getAttribute('site-key')?.trim()) return null
    const zh = this.lang.startsWith('zh')
    return h(
      'div',
      { class: 'marku-error', part: 'error' },
      h('span', {
        text: zh
          ? '缺少 site-key 属性。请从 Marku 管理后台复制站点 Key。'
          : 'Missing site-key attribute. Copy a site key from the Marku admin.',
      }),
    )
  }

  protected renderFailure(error: unknown): void {
    const message = errorMessage(error, this.lang)
    this.dispatchEvent(
      new CustomEvent('marku:error', {
        detail: { error },
        bubbles: true,
        composed: true,
      }),
    )
    this.render(
      h(
        'div',
        { class: 'marku-container', part: 'container' },
        h(
          'div',
          { class: 'marku-error', part: 'error', role: 'alert' },
          h('span', { text: message }),
          h(
            'button',
            {
              class: 'marku-button',
              type: 'button',
              part: 'retry',
              onclick: () => {
                void this.enqueue()
              },
            },
            this.lang.startsWith('zh') ? '重试' : 'Retry',
          ),
        ),
      ),
    )
  }

  protected isLive(token: number): boolean {
    return token === this.token && this.isConnected
  }

  protected enqueue(): Promise<void> {
    const token = ++this.token
    this.controller?.abort()
    this.controller = new AbortController()

    return (async () => {
      try {
        await this.update()
      } catch (error) {
        if (!this.isLive(token) || isAbortError(error)) return
        this.renderFailure(error)
      } finally {
        if (token === this.token) this.resolveReady?.()
      }
    })()
  }

  private setupRoot(): void {
    const selector = this.hasAttribute('no-shadow')
      ? this.tagName.toLowerCase()
      : ':host'
    const css = (baseStyles + (this.constructor as typeof MarkuElement).styles)
      .split(':host')
      .join(selector)

    if (this.hasAttribute('no-shadow')) {
      const tag = this.tagName.toLowerCase()
      if (!document.head.querySelector(`style[data-marku="${tag}"]`)) {
        const style = document.createElement('style')
        style.dataset.marku = tag
        style.textContent = css
        document.head.append(style)
      }
      this.target = this
      return
    }

    const root = this.shadowRoot ?? this.attachShadow({ mode: 'open' })
    root.replaceChildren()
    const style = document.createElement('style')
    style.textContent = css
    const host = document.createElement('div')
    host.dataset.markuHost = ''
    root.append(style, host)
    this.target = host
  }

  private applyTheme(): void {
    const theme = this.getAttribute('theme')
    const explicit = theme === 'dark' || theme === 'light' ? theme : null
    if (explicit) {
      const tokens = explicit === 'dark' ? DARK_TOKENS : LIGHT_TOKENS
      for (const [name, value] of Object.entries(tokens)) {
        this.style.setProperty(name, value)
      }
      this.style.colorScheme = explicit
      return
    }

    for (const name of Object.keys(LIGHT_TOKENS)) {
      this.style.removeProperty(name)
    }
    this.style.removeProperty('color-scheme')
  }
}

export function requireSiteKey(element: HTMLElement): string {
  const siteKey = element.getAttribute('site-key')?.trim()
  if (!siteKey) {
    throw new MarkuError('Missing site-key attribute')
  }
  return siteKey
}
