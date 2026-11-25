import type { MarkuClient } from '@marku/client'

import { MARKU_ATTRIBUTES, MarkuElement } from './base'
import { errorMessage, formatNumber } from './format'
import { h } from './html'
import { likeStyles } from './styles'

interface State {
  likes: number
  liked: boolean
  loading: boolean
  busy: boolean
  error: string
}

export class MarkuLike extends MarkuElement {
  static styles = likeStyles

  static get observedAttributes(): string[] {
    return [...MARKU_ATTRIBUTES, 'label', 'liked-label', 'show-count']
  }

  private state: State = {
    likes: 0,
    liked: false,
    loading: true,
    busy: false,
    error: '',
  }

  protected async update(): Promise<void> {
    const client = this.clientOrNull()
    const missing = this.configError()
    if (!client) {
      this.render(missing ?? h('div'))
      return
    }

    this.state.loading = true
    this.renderView()
    try {
      const state = await client.likes.get(this.resolvePath(), {
        signal: this.signal,
      })
      this.state.likes = Number(state.likes ?? 0)
      this.state.liked = Boolean(state.liked)
      this.state.error = ''
    } catch (error) {
      this.state.error = errorMessage(error, this.lang)
      this.emitError(error)
    } finally {
      this.state.loading = false
      this.renderView()
    }
  }

  private async toggle(): Promise<void> {
    const client = this.clientOrNull()
    if (!client || this.state.busy) return

    this.state.busy = true
    this.state.error = ''
    this.renderView()
    try {
      const state = await client.likes.toggle(this.resolvePath(), {
        signal: this.signal,
      })
      this.state.likes = Number(state.likes ?? 0)
      this.state.liked = Boolean(state.liked)
      this.dispatchEvent(
        new CustomEvent('marku:like-change', {
          detail: { liked: this.state.liked, likes: this.state.likes },
          bubbles: true,
          composed: true,
        }),
      )
    } catch (error) {
      this.state.error = errorMessage(error, this.lang)
      this.emitError(error)
    } finally {
      this.state.busy = false
      this.renderView()
    }
  }

  private renderView(): void {
    const zh = this.lang.startsWith('zh')
    const label = this.getAttribute('label') || (zh ? '点赞' : 'Like')
    const likedLabel = this.getAttribute('liked-label') || (zh ? '已点赞' : 'Liked')
    const showCount = this.getAttribute('show-count') !== 'false'

    const button = h(
      'button',
      {
        class: 'marku-like__button',
        type: 'button',
        part: 'button',
        'aria-pressed': String(this.state.liked),
        disabled: this.state.loading || this.state.busy,
        onclick: () => {
          void this.toggle()
        },
      },
      h('span', {
        class: 'marku-like__icon',
        part: 'icon',
        'aria-hidden': 'true',
        text: this.state.liked ? '♥' : '♡',
      }),
      h('span', {
        class: 'marku-like__label',
        part: 'label',
        text: this.state.liked ? likedLabel : label,
      }),
    )

    if (showCount) {
      button.append(
        h('span', {
          class: 'marku-like__count',
          part: 'count',
          text: formatNumber(this.state.likes, this.lang),
        }),
      )
    }

    const wrapper = h('div', {
      class: 'marku-container marku-like',
      part: 'container',
    })
    wrapper.append(button)
    if (this.state.error) {
      wrapper.append(
        h('span', {
          class: 'marku-status',
          part: 'error',
          role: 'status',
          dataset: { tone: 'error' },
          text: this.state.error,
        }),
      )
    }
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
