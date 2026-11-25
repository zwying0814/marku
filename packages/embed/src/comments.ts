import type { MarkuClient, PublicComment } from '@marku/client'
import { isMarkuError } from '@marku/client'

import { MARKU_ATTRIBUTES, MarkuElement } from './base'
import { errorMessage, formatDateTime, formatNumber, safeHttpUrl } from './format'
import { h } from './html'
import { commentsStyles } from './styles'

interface StatusMessage {
  message: string
  tone: 'info' | 'success' | 'error'
}

interface Draft {
  authorName: string
  authorEmail: string
  authorWebsite: string
  content: string
}

interface State {
  loading: boolean
  loadingMore: boolean
  error: string | null
  items: PublicComment[]
  total: number
  page: number
  submitting: boolean
  replyTo: PublicComment | null
  status: StatusMessage | null
  cooldownUntil: number
}

const DRAFT_STORAGE_KEY = 'marku:author'
const MAX_REPLY_DEPTH = 4

export class MarkuComments extends MarkuElement {
  static styles = commentsStyles

  static get observedAttributes(): string[] {
    return [...MARKU_ATTRIBUTES, 'heading', 'page-size', 'sort', 'show-form']
  }

  private state: State = {
    loading: true,
    loadingMore: false,
    error: null,
    items: [],
    total: 0,
    page: 1,
    submitting: false,
    replyTo: null,
    status: null,
    cooldownUntil: 0,
  }

  private draft: Draft = loadDraft()
  private formStartedAt = 0
  private cooldownTimer: ReturnType<typeof setInterval> | null = null

  disconnectedCallback(): void {
    super.disconnectedCallback()
    this.stopCooldownTimer()
  }

  protected async update(): Promise<void> {
    const client = this.clientOrNull()
    const missing = this.configError()
    if (!client) {
      this.render(missing ?? h('div', { part: 'container' }))
      return
    }

    await this.reload(client)
  }

  private async reload(client: MarkuClient): Promise<void> {
    this.state.loading = true
    this.state.error = null
    this.renderView()

    try {
      const [stats, page] = await Promise.all([
        client.pageStats(this.resolvePath(), { signal: this.signal }),
        client.comments.list(
          {
            path: this.resolvePath(),
            page: 1,
            pageSize: this.pageSize(),
            sort: this.sort(),
          },
          { signal: this.signal },
        ),
      ])
      this.state.items = page.items ?? []
      this.state.total = Number(page.total ?? stats.comments ?? this.state.items.length)
      this.state.page = Number(page.page ?? 1)
    } catch (error) {
      this.state.error = errorMessage(error, this.lang)
      this.emitError(error)
    } finally {
      this.state.loading = false
      this.renderView()
    }
  }

  private async loadMore(): Promise<void> {
    const client = this.clientOrNull()
    if (!client || this.state.loadingMore) return

    this.state.loadingMore = true
    this.renderView()
    try {
      const nextPage = this.state.page + 1
      const page = await client.comments.list({
        path: this.resolvePath(),
        page: nextPage,
        pageSize: this.pageSize(),
        sort: this.sort(),
      })
      this.state.items = [...this.state.items, ...(page.items ?? [])]
      this.state.page = Number(page.page ?? nextPage)
      this.state.total = Number(page.total ?? this.state.total)
    } catch (error) {
      this.state.status = {
        message: errorMessage(error, this.lang),
        tone: 'error',
      }
      this.emitError(error)
    } finally {
      this.state.loadingMore = false
      this.renderView()
    }
  }

  private async submit(): Promise<void> {
    const client = this.clientOrNull()
    if (!client || this.state.submitting) return

    const authorName = this.draft.authorName.trim()
    const content = this.draft.content.trim()
    const zh = this.lang.startsWith('zh')
    if (!authorName || !content) {
      this.state.status = {
        message: zh ? '请填写昵称和评论内容。' : 'Please fill in your name and comment.',
        tone: 'error',
      }
      this.renderView()
      return
    }

    this.state.submitting = true
    this.state.status = null
    this.renderView()

    try {
      const result = await client.comments.create(
        {
          path: this.resolvePath(),
          parentId: this.state.replyTo?.id ?? null,
          authorName,
          authorEmail: this.draft.authorEmail.trim() || undefined,
          authorWebsite: this.draft.authorWebsite.trim() || undefined,
          content,
          honeypot: this.honeypotValue(),
          formStartedAt: this.formStartedAt || undefined,
        },
        { signal: this.signal },
      )

      saveDraft(this.draft)
      this.draft.content = ''
      this.formStartedAt = Date.now()
      this.state.replyTo = null
      this.state.status = {
        message:
          result.status === 'approved'
            ? zh
              ? '评论已发布。'
              : 'Your comment is published.'
            : zh
              ? '评论已提交，审核通过后显示。'
              : 'Comment submitted. It will appear after moderation.',
        tone: 'success',
      }

      this.dispatchEvent(
        new CustomEvent('marku:comment-submitted', {
          detail: result,
          bubbles: true,
          composed: true,
        }),
      )

      if (result.status === 'approved') {
        await this.refreshStats(client)
        await this.reload(client)
        this.state.status = {
          message: zh ? '评论已发布。' : 'Your comment is published.',
          tone: 'success',
        }
      } else {
        await this.refreshStats(client)
      }
    } catch (error) {
      if (isMarkuError(error) && error.status === 429) {
        this.startCooldown(error.retryAfter || 30)
      }
      this.state.status = { message: errorMessage(error, this.lang), tone: 'error' }
      this.emitError(error)
    } finally {
      this.state.submitting = false
      this.renderView()
    }
  }

  private async refreshStats(client: MarkuClient): Promise<void> {
    try {
      const stats = await client.pageStats(this.resolvePath())
      this.state.total = Number(stats.comments ?? this.state.total)
    } catch {
      // Counts are best-effort; the comment itself already succeeded.
    }
  }

  private renderView(): void {
    const zh = this.lang.startsWith('zh')
    const heading =
      this.getAttribute('heading') || (zh ? '评论' : 'Comments')
    const showForm = this.getAttribute('show-form') !== 'false'

    const container = h('div', {
      class: 'marku-container marku-comments',
      part: 'container',
    })

    container.append(
      h(
        'div',
        { class: 'marku-comments__header', part: 'header' },
        h('h2', {
          class: 'marku-comments__heading',
          part: 'heading',
          text: heading,
        }),
        h('span', {
          class: 'marku-comments__count',
          part: 'count',
          text: formatNumber(this.totalCount(), this.lang),
        }),
      ),
    )

    container.append(this.renderList())
    if (showForm && !this.state.error) container.append(this.renderForm())

    this.render(container)

    if (this.state.replyTo) {
      queueMicrotask(() => {
        const textarea = this.querySelectorTarget<HTMLTextAreaElement>('textarea')
        textarea?.focus()
      })
    }
  }

  private totalCount(): number {
    if (this.state.loading) return this.state.total
    return Math.max(this.state.total, this.state.items.length)
  }

  private renderList(): HTMLElement {
    const list = h('div', { class: 'marku-comments__list', part: 'list' })
    const zh = this.lang.startsWith('zh')

    if (this.state.loading) {
      list.append(
        h(
          'div',
          { class: 'marku-skeleton', part: 'loading', 'aria-busy': 'true' },
          h('span'),
          h('span'),
          h('span'),
        ),
      )
      return list
    }

    if (this.state.error) {
      list.append(
        h(
          'div',
          { class: 'marku-error', part: 'error', role: 'alert' },
          h('span', { text: this.state.error }),
          h(
            'button',
            {
              class: 'marku-button',
              type: 'button',
              part: 'retry',
              onclick: () => {
                const client = this.clientOrNull()
                if (client) void this.reload(client)
              },
            },
            zh ? '重试' : 'Retry',
          ),
        ),
      )
      return list
    }

    if (this.state.items.length === 0) {
      list.append(
        h('p', {
          class: 'marku-comments__empty',
          part: 'empty',
          text:
            this.getAttribute('empty-text') ||
            (zh ? '还没有评论，来写第一条吧。' : 'No comments yet. Start the conversation.'),
        }),
      )
      return list
    }

    for (const comment of this.state.items) {
      list.append(this.renderComment(comment, 0))
    }

    if (this.state.items.length < this.state.total) {
      list.append(
        h(
          'button',
          {
            class: 'marku-button marku-comments__more',
            type: 'button',
            part: 'load-more',
            disabled: this.state.loadingMore,
            onclick: () => {
              void this.loadMore()
            },
          },
          this.state.loadingMore
            ? zh
              ? '加载中…'
              : 'Loading…'
            : zh
              ? `加载更多（${this.state.items.length}/${this.state.total}）`
              : `Load more (${this.state.items.length}/${this.state.total})`,
        ),
      )
    }

    return list
  }

  private renderComment(comment: PublicComment, depth: number): HTMLElement {
    const article = h('article', { class: 'marku-comment', part: 'comment' })
    const website = safeHttpUrl(comment.authorWebsite)
    const author = h('span', {
      class: 'marku-comment__author',
      part: 'author',
      text: comment.authorName || '匿名',
    })

    const head = h('div', { class: 'marku-comment__head', part: 'comment-header' })
    if (website) {
      head.append(
        h(
          'a',
          {
            href: website,
            target: '_blank',
            rel: 'nofollow noopener noreferrer ugc',
            part: 'author-link',
          },
          author,
        ),
      )
    } else {
      head.append(author)
    }

    if (comment.isAdmin) {
      head.append(
        h('span', {
          class: 'marku-comment__badge',
          part: 'badge',
          text: this.lang.startsWith('zh') ? '站长' : 'Author',
        }),
      )
    }

    const time = formatDateTime(comment.createdAt, this.lang)
    if (time) {
      head.append(
        h('time', {
          class: 'marku-comment__time',
          part: 'time',
          datetime: comment.createdAt,
          text: time,
        }),
      )
    }

    article.append(head)
    article.append(
      h('p', {
        class: 'marku-comment__content',
        part: 'content',
        text: comment.content ?? '',
      }),
    )

    if (depth < MAX_REPLY_DEPTH && comment.id) {
      article.append(
        h(
          'div',
          { class: 'marku-comments__actions', part: 'actions' },
          h(
            'button',
            {
              class: 'marku-link-button',
              type: 'button',
              part: 'reply',
              onclick: () => {
                this.state.replyTo = comment
                this.state.status = null
                this.renderView()
              },
            },
            this.lang.startsWith('zh') ? '回复' : 'Reply',
          ),
        ),
      )
    }

    const replies = comment.replies ?? []
    if (replies.length > 0) {
      const wrapper = h('div', {
        class: 'marku-comment__replies',
        part: 'replies',
      })
      for (const reply of replies) {
        wrapper.append(this.renderComment(reply, depth + 1))
      }
      article.append(wrapper)
    }

    return article
  }

  private renderForm(): HTMLElement {
    const zh = this.lang.startsWith('zh')
    const coolingDown = this.state.cooldownUntil > Date.now()
    const submitLabel =
      this.getAttribute('submit-label') ||
      (zh ? '发表评论' : 'Post comment')

    const form = h('form', {
      class: 'marku-comments__form',
      part: 'form',
      onsubmit: (event: Event) => {
        event.preventDefault()
        void this.submit()
      },
      onfocusin: () => {
        if (!this.formStartedAt) this.formStartedAt = Date.now()
      },
    })

    if (this.state.replyTo) {
      form.append(
        h(
          'div',
          { class: 'marku-comments__reply-banner', part: 'reply-banner' },
          h('span', {
            text: `${zh ? '回复' : 'Replying to'} @${
              this.state.replyTo.authorName || '匿名'
            }`,
          }),
          h(
            'button',
            {
              class: 'marku-link-button',
              type: 'button',
              part: 'cancel-reply',
              onclick: () => {
                this.state.replyTo = null
                this.renderView()
              },
            },
            zh ? '取消' : 'Cancel',
          ),
        ),
      )
    }

    form.append(
      h(
        'div',
        { class: 'marku-comments__form-row' },
        this.field(
          'name',
          'authorName',
          zh ? '昵称 *' : 'Name *',
          this.draft.authorName,
          'text',
          true,
        ),
        this.field(
          'email',
          'authorEmail',
          zh ? '邮箱（不公开）' : 'Email (private)',
          this.draft.authorEmail,
          'email',
          false,
        ),
        this.field(
          'url',
          'authorWebsite',
          zh ? '网站（可选）' : 'Website (optional)',
          this.draft.authorWebsite,
          'url',
          false,
        ),
      ),
    )

    form.append(
      h(
        'div',
        { class: 'marku-field' },
        h('label', {
          for: 'marku-comment-content',
          text: zh ? '评论内容 *' : 'Comment *',
        }),
        h('textarea', {
          id: 'marku-comment-content',
          name: 'content',
          part: 'textarea',
          required: true,
          maxlength: '2000',
          placeholder:
            this.getAttribute('placeholder') ||
            (zh ? '说点什么…' : 'Say something…'),
          value: this.draft.content,
          oninput: (event: Event) => {
            this.draft.content = (event.target as HTMLTextAreaElement).value
          },
        }),
      ),
    )

    form.append(
      h('input', {
        class: 'marku-honeypot',
        name: 'honeypot',
        type: 'text',
        tabindex: '-1',
        autocomplete: 'off',
        'aria-hidden': 'true',
      }),
    )

    const actions = h('div', { class: 'marku-comments__actions', part: 'actions' })
    actions.append(
      h(
        'button',
        {
          class: 'marku-button marku-button--primary',
          type: 'submit',
          part: 'submit',
          disabled: this.state.submitting || coolingDown,
        },
        this.state.submitting
          ? zh
            ? '提交中…'
            : 'Submitting…'
          : coolingDown
            ? zh
              ? `请等待 ${this.cooldownSeconds()}s`
              : `Wait ${this.cooldownSeconds()}s`
            : submitLabel,
      ),
    )

    if (this.state.status) {
      actions.append(
        h('p', {
          class: 'marku-status',
          part: 'status',
          'aria-live': 'polite',
          dataset: { tone: this.state.status.tone },
          text: this.state.status.message,
        }),
      )
    }

    form.append(actions)
    return form
  }

  private field(
    type: string,
    field: keyof Draft,
    label: string,
    value: string,
    inputType: string,
    required: boolean,
  ): HTMLElement {
    return h(
      'div',
      { class: 'marku-field' },
      h('label', { for: `marku-${field}`, text: label }),
      h('input', {
        id: `marku-${field}`,
        name: field,
        type: inputType,
        part: 'input',
        required,
        autocomplete: type === 'email' ? 'email' : type === 'url' ? 'url' : 'nickname',
        value,
        oninput: (event: Event) => {
          this.draft[field] = (event.target as HTMLInputElement).value
        },
      }),
    )
  }

  private honeypotValue(): string {
    return this.querySelectorTarget<HTMLInputElement>('input[name="honeypot"]')?.value ?? ''
  }

  private pageSize(): number {
    const raw = Number(this.getAttribute('page-size') ?? 20)
    if (!Number.isFinite(raw)) return 20
    return Math.min(100, Math.max(1, Math.trunc(raw)))
  }

  private sort(): 'asc' | 'desc' {
    return this.getAttribute('sort') === 'desc' ? 'desc' : 'asc'
  }

  private cooldownSeconds(): number {
    return Math.max(1, Math.ceil((this.state.cooldownUntil - Date.now()) / 1000))
  }

  private startCooldown(seconds: number): void {
    this.state.cooldownUntil = Date.now() + Math.max(1, seconds) * 1000
    this.stopCooldownTimer()
    this.cooldownTimer = setInterval(() => {
      if (this.state.cooldownUntil <= Date.now()) this.stopCooldownTimer()
      this.renderView()
    }, 1000)
  }

  private stopCooldownTimer(): void {
    if (!this.cooldownTimer) return
    clearInterval(this.cooldownTimer)
    this.cooldownTimer = null
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

  private querySelectorTarget<T extends Element>(selector: string): T | null {
    if (this.shadowRoot) return this.shadowRoot.querySelector<T>(selector)
    return this.querySelector<T>(selector)
  }
}

function loadDraft(): Draft {
  try {
    const raw = globalThis.localStorage?.getItem(DRAFT_STORAGE_KEY)
    if (!raw) return emptyDraft()
    const parsed = JSON.parse(raw) as Partial<Draft>
    return {
      authorName: String(parsed.authorName ?? ''),
      authorEmail: String(parsed.authorEmail ?? ''),
      authorWebsite: String(parsed.authorWebsite ?? ''),
      content: '',
    }
  } catch {
    return emptyDraft()
  }
}

function emptyDraft(): Draft {
  return { authorName: '', authorEmail: '', authorWebsite: '', content: '' }
}

function saveDraft(draft: Draft): void {
  try {
    globalThis.localStorage?.setItem(
      DRAFT_STORAGE_KEY,
      JSON.stringify({
        authorName: draft.authorName,
        authorEmail: draft.authorEmail,
        authorWebsite: draft.authorWebsite,
      }),
    )
  } catch {
    // Storage can be unavailable in private mode; ignore.
  }
}
