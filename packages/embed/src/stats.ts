import type { MarkuClient } from '@marku/client'

import { MARKU_ATTRIBUTES, MarkuElement } from './base'
import { errorMessage, formatNumber } from './format'
import { h } from './html'
import { statsStyles } from './styles'

type Metric = 'views' | 'comments' | 'likes'

const METRIC_LABELS: Record<Metric, { zh: string; en: string }> = {
  views: { zh: '浏览', en: 'views' },
  comments: { zh: '评论', en: 'comments' },
  likes: { zh: '赞', en: 'likes' },
}

export class MarkuStats extends MarkuElement {
  static styles = statsStyles

  static get observedAttributes(): string[] {
    return [...MARKU_ATTRIBUTES, 'show', 'separator', 'record-view']
  }

  private values: Record<Metric, number> = { views: 0, comments: 0, likes: 0 }
  private loading = true
  private error = ''
  private recorded = false

  protected async update(): Promise<void> {
    const client = this.clientOrNull()
    const missing = this.configError()
    if (!client) {
      this.render(missing ?? h('div'))
      return
    }

    this.loading = true
    this.renderView()
    try {
      const path = this.resolvePath()
      const stats = await client.pageStats(path, { signal: this.signal })
      this.values = {
        views: Number(stats.views ?? 0),
        comments: Number(stats.comments ?? 0),
        likes: Number(stats.likes ?? 0),
      }
      this.error = ''

      if (this.hasAttribute('record-view')) {
        await this.recordView(client, path)
      }
    } catch (error) {
      this.error = errorMessage(error, this.lang)
      this.emitError(error)
    } finally {
      this.loading = false
      this.renderView()
    }
  }

  private async recordView(client: MarkuClient, path: string): Promise<void> {
    if (this.recorded || wasViewed(path)) return

    try {
      const result = await client.views.record(path, { signal: this.signal })
      if (typeof result.views === 'number') this.values.views = result.views
      this.recorded = true
      markViewed(path)
    } catch (error) {
      this.emitError(error)
    }
  }

  private metrics(): Metric[] {
    const raw = this.getAttribute('show') || 'views,comments,likes'
    const list = raw
      .split(',')
      .map((value) => value.trim().toLowerCase())
      .filter((value): value is Metric => value in METRIC_LABELS)
    return list.length > 0 ? [...new Set(list)] : ['views', 'comments', 'likes']
  }

  private renderView(): void {
    const zh = this.lang.startsWith('zh')
    const wrapper = h('div', {
      class: 'marku-container marku-stats',
      part: 'container',
      'aria-busy': String(this.loading),
    })

    if (this.error) {
      wrapper.append(
        h('span', {
          class: 'marku-status',
          part: 'error',
          role: 'status',
          dataset: { tone: 'error' },
          text: this.error,
        }),
      )
      this.render(wrapper)
      return
    }

    const separator = this.getAttribute('separator') ?? '·'
    const metrics = this.metrics()

    metrics.forEach((metric, index) => {
      if (index > 0 && separator) {
        wrapper.append(
          h('span', {
            class: 'marku-stats__separator',
            part: 'separator',
            'aria-hidden': 'true',
            text: separator,
          }),
        )
      }

      const item = h('span', {
        class: 'marku-stats__item',
        part: 'stat',
        dataset: { metric },
      })
      item.append(
        h('span', {
          class: 'marku-stats__value',
          part: 'value',
          text: this.loading && this.values[metric] === 0 ? '–' : formatNumber(this.values[metric], this.lang),
        }),
        h('span', {
          class: 'marku-stats__label',
          part: 'label',
          text: zh ? METRIC_LABELS[metric].zh : METRIC_LABELS[metric].en,
        }),
      )
      wrapper.append(item)
    })

    this.render(wrapper)
  }

  private emitError(error: unknown): void {
    this.dispatchEvent(
      new CustomEvent('marku:error', {
        detail: { error },
        bubbles: true,
        composed: true,
      }),
    )
  }
}

function viewedKey(path: string): string {
  return `marku:viewed:${path}`
}

function wasViewed(path: string): boolean {
  try {
    return globalThis.sessionStorage?.getItem(viewedKey(path)) === '1'
  } catch {
    return false
  }
}

function markViewed(path: string): void {
  try {
    globalThis.sessionStorage?.setItem(viewedKey(path), '1')
  } catch {
    // Ignore storage failures.
  }
}
