(() => {
  'use strict'

  const PLACEHOLDER_KEY = 'mk_replace_with_your_site_key'
  const config = window.MARKU_EXAMPLE ?? {}
  const apiBaseUrl = String(config.apiBaseUrl ?? '').replace(/\/+$/, '')
  const siteKey = String(config.siteKey ?? '').trim()

  const elements = {
    apiState: document.querySelector('#api-state'),
    apiStateText: document.querySelector('#api-state-text'),
    articlePath: document.querySelector('#article-path'),
    setupAlert: document.querySelector('#setup-alert'),
    viewCount: document.querySelector('#view-count'),
    commentCount: document.querySelector('#comment-count'),
    commentCountInline: document.querySelector('#comment-count-inline'),
    likeCount: document.querySelector('#like-count'),
    likeButton: document.querySelector('#like-button'),
    likeButtonLabel: document.querySelector('#like-button-label'),
    commentLoading: document.querySelector('#comment-loading'),
    commentError: document.querySelector('#comment-error'),
    commentErrorMessage: document.querySelector('#comment-error-message'),
    commentEmpty: document.querySelector('#comment-empty'),
    commentList: document.querySelector('#comment-list'),
    retryComments: document.querySelector('#retry-comments'),
    form: document.querySelector('#comment-form'),
    submitButton: document.querySelector('#submit-comment'),
    submitLabel: document.querySelector('#submit-label'),
    formStatus: document.querySelector('#form-status'),
    replyTarget: document.querySelector('#reply-target'),
    replyTargetName: document.querySelector('#reply-target-name'),
    cancelReply: document.querySelector('#cancel-reply'),
    toast: document.querySelector('#toast'),
  }

  const fields = {
    authorName: elements.form.elements.namedItem('authorName'),
    authorEmail: elements.form.elements.namedItem('authorEmail'),
    authorWebsite: elements.form.elements.namedItem('authorWebsite'),
    content: elements.form.elements.namedItem('content'),
    honeypot: elements.form.elements.namedItem('honeypot'),
  }

  const state = {
    path: resolveCanonicalPath(),
    visitorId: getVisitorId(),
    site: null,
    parentId: null,
    liked: false,
    interactive: false,
    likeBusy: false,
    submitting: false,
    cooldownUntil: 0,
    formStartedAt: Date.now(),
    toastTimer: null,
    cooldownTimer: null,
  }

  class ApiError extends Error {
    constructor(message, status = 0, code = 0, retryAfter = 0) {
      super(message)
      this.name = 'ApiError'
      this.status = status
      this.code = code
      this.retryAfter = retryAfter
    }
  }

  function resolveCanonicalPath() {
    const explicit = document
      .querySelector('meta[name="marku:path"]')
      ?.getAttribute('content')
      ?.trim()
    if (explicit) return normalizePath(explicit)

    const canonical = document
      .querySelector('link[rel="canonical"]')
      ?.getAttribute('href')
    if (canonical) {
      try {
        const pathname = new URL(canonical, window.location.origin).pathname
        if (pathname && pathname !== '/') return normalizePath(pathname)
      } catch {
        // Invalid canonical URLs fall back to the current pathname.
      }
    }

    return normalizePath(window.location.pathname)
  }

  function normalizePath(value) {
    let path = value.trim() || '/'
    if (!path.startsWith('/')) path = `/${path}`
    if (path.endsWith('/index.html')) path = path.slice(0, -'/index.html'.length)
    return path || '/'
  }

  function getVisitorId() {
    const storageKey = 'marku-html-visitor-id'
    try {
      const existing = localStorage.getItem(storageKey)
      if (existing) return existing
      const created =
        typeof crypto.randomUUID === 'function'
          ? crypto.randomUUID()
          : `visitor-${Date.now()}-${Math.random().toString(16).slice(2)}`
      localStorage.setItem(storageKey, created)
      return created
    } catch {
      return `session-${Date.now()}-${Math.random().toString(16).slice(2)}`
    }
  }

  function isConfigured() {
    return Boolean(
      apiBaseUrl &&
        siteKey &&
        siteKey !== PLACEHOLDER_KEY &&
        !/replace_with_your/i.test(siteKey),
    )
  }

  function setApiState(mode, text) {
    elements.apiState.dataset.state = mode
    elements.apiStateText.textContent = text
  }

  function showSetupAlert(show = true) {
    elements.setupAlert.hidden = !show
  }

  function setInteractive(interactive) {
    state.interactive = interactive
    for (const field of Object.values(fields)) {
      if (field instanceof HTMLInputElement || field instanceof HTMLTextAreaElement) {
        field.disabled = !interactive
      }
    }
    elements.likeButton.disabled = !interactive || state.likeBusy
    updateSubmitButton()
  }

  function updateSubmitButton() {
    const coolingDown = Date.now() < state.cooldownUntil
    elements.submitButton.disabled =
      !state.interactive || state.submitting || coolingDown

    if (state.submitting) {
      elements.submitLabel.textContent = '提交中…'
      return
    }
    if (coolingDown) {
      const seconds = Math.max(
        1,
        Math.ceil((state.cooldownUntil - Date.now()) / 1000),
      )
      elements.submitLabel.textContent = `请等待 ${seconds}s`
      return
    }
    elements.submitLabel.textContent = '提交评论'
  }

  function buildUrl(path, query) {
    if (!apiBaseUrl) throw new ApiError('API 地址尚未配置')
    const url = new URL(`${apiBaseUrl}${path}`, window.location.origin)
    for (const [key, value] of Object.entries(query ?? {})) {
      if (value === undefined || value === null || value === '') continue
      url.searchParams.set(key, String(value))
    }
    return url.toString()
  }

  async function api(path, options = {}) {
    const {
      method = 'GET',
      body,
      query,
      timeout = 15000,
    } = options
    const headers = { Accept: 'application/json', 'X-Site-Key': siteKey }
    if (body !== undefined) headers['Content-Type'] = 'application/json'

    const controller = new AbortController()
    const timeoutId = window.setTimeout(() => controller.abort(), timeout)
    let response
    try {
      response = await fetch(buildUrl(path, query), {
        method,
        headers,
        body: body === undefined ? undefined : JSON.stringify(body),
        signal: controller.signal,
      })
    } catch (error) {
      const message =
        error instanceof DOMException && error.name === 'AbortError'
          ? '请求超时，请稍后重试。'
          : '无法连接 Marku API。请检查 API 地址、站点允许来源与全局 CORS 配置。'
      throw new ApiError(message)
    } finally {
      window.clearTimeout(timeoutId)
    }

    const text = await response.text()
    let payload = null
    if (text) {
      try {
        payload = JSON.parse(text)
      } catch {
        throw new ApiError('API 返回了非 JSON 响应。', response.status)
      }
    }

    if (!response.ok || payload?.code !== 0) {
      throw new ApiError(
        payload?.message || response.statusText || '请求失败',
        response.status,
        payload?.code ?? 0,
        Number(response.headers.get('Retry-After') ?? 0),
      )
    }
    return payload?.data
  }

  function describeError(error) {
    if (!(error instanceof ApiError)) {
      return '发生未知错误，请稍后重试。'
    }
    if (error.status === 401) {
      showSetupAlert()
      return 'Site Key 无效或已吊销，请检查 config.js。'
    }
    if (error.status === 403) {
      return '站点已停用，或当前页面来源不在允许来源列表中。'
    }
    if (error.status === 429) {
      const seconds = error.retryAfter || 30
      return `请求过于频繁，请 ${seconds} 秒后再试。`
    }
    if (error.status === 0) return error.message
    return error.message
  }

  function showToast(message, tone = 'info') {
    window.clearTimeout(state.toastTimer)
    elements.toast.textContent = message
    elements.toast.dataset.tone = tone
    elements.toast.classList.add('is-visible')
    state.toastTimer = window.setTimeout(() => {
      elements.toast.classList.remove('is-visible')
    }, 4200)
  }

  function setFormStatus(message, tone = '') {
    elements.formStatus.textContent = message
    elements.formStatus.dataset.tone = tone
  }

  function formatNumber(value) {
    return new Intl.NumberFormat('zh-CN').format(Number(value) || 0)
  }

  function formatCommentTime(value) {
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return ''
    return new Intl.DateTimeFormat('zh-CN', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }).format(date)
  }

  function renderStats(stats) {
    elements.viewCount.textContent = formatNumber(stats?.views)
    elements.commentCount.textContent = formatNumber(stats?.comments)
    elements.commentCountInline.textContent = formatNumber(stats?.comments)
    elements.likeCount.textContent = formatNumber(stats?.likes)
    state.liked = Boolean(stats?.liked)
    updateLikeButton()
  }

  function updateLikeButton() {
    elements.likeButton.classList.toggle('is-liked', state.liked)
    elements.likeButton.setAttribute('aria-pressed', String(state.liked))
    elements.likeButtonLabel.textContent = state.liked
      ? '已点赞'
      : '点赞这篇文章'
    elements.likeButton.disabled =
      !state.interactive || state.likeBusy
  }

  async function loadStats() {
    const data = await api('/public/stats', {
      query: { paths: state.path, visitorId: state.visitorId },
    })
    renderStats(data?.items?.[0])
  }

  async function recordViewIfNeeded() {
    if (!state.site) return
    const key = `marku:viewed:${state.site.id}:${state.path}`
    try {
      if (sessionStorage.getItem(key)) return
    } catch {
      // Session storage can be unavailable; still record the view.
    }

    try {
      await api('/public/views', {
        method: 'POST',
        body: { path: state.path, visitorId: state.visitorId },
      })
      try {
        sessionStorage.setItem(key, '1')
      } catch {
        // Ignore storage failures.
      }
    } catch (error) {
      showToast(describeError(error), 'error')
    }
  }

  function createElement(tag, className, text) {
    const element = document.createElement(tag)
    if (className) element.className = className
    if (text !== undefined) element.textContent = text
    return element
  }

  function renderComment(comment, depth) {
    const article = createElement('article', 'comment')
    const avatar = createElement(
      'span',
      'comment-avatar',
      (comment.authorName || '?').slice(0, 1).toUpperCase(),
    )
    avatar.setAttribute('aria-hidden', 'true')

    const body = createElement('div', 'comment-body')
    const head = createElement('div', 'comment-head')
    const author = createElement('span', 'comment-author')
    if (comment.authorWebsite) {
      const link = createElement('a', '', comment.authorName)
      link.href = comment.authorWebsite
      link.target = '_blank'
      link.rel = 'nofollow ugc noopener noreferrer'
      author.append(link)
    } else {
      author.textContent = comment.authorName
    }
    head.append(author)

    if (comment.isAdmin) {
      head.append(createElement('span', 'comment-badge', '作者回复'))
    }
    head.append(
      createElement('time', 'comment-time', formatCommentTime(comment.createdAt)),
    )

    const content = createElement('p', 'comment-content', comment.content)
    const actions = createElement('div', 'comment-actions')

    if (depth < 3) {
      const replyButton = createElement('button', '', '回复')
      replyButton.type = 'button'
      replyButton.addEventListener('click', () => beginReply(comment))
      actions.append(replyButton)
    }

    body.append(head, content, actions)
    article.append(avatar, body)

    if (Array.isArray(comment.replies) && comment.replies.length) {
      const children = createElement('div', 'comment-children')
      for (const reply of comment.replies) {
        children.append(renderComment(reply, depth + 1))
      }
      body.append(children)
    }
    return article
  }

  function renderComments(comments) {
    elements.commentList.replaceChildren()
    if (!comments.length) {
      elements.commentEmpty.hidden = false
      return
    }
    elements.commentEmpty.hidden = true
    const fragment = document.createDocumentFragment()
    for (const comment of comments) {
      fragment.append(renderComment(comment, 0))
    }
    elements.commentList.append(fragment)
  }

  async function loadComments() {
    elements.commentLoading.hidden = false
    elements.commentError.hidden = true
    elements.commentEmpty.hidden = true
    elements.retryComments.hidden = true
    elements.commentList.replaceChildren()

    try {
      const data = await api('/public/comments', {
        query: {
          path: state.path,
          page: 1,
          pageSize: 50,
          sort: 'asc',
        },
      })
      renderComments(data?.items ?? [])
      elements.commentLoading.hidden = true
    } catch (error) {
      elements.commentLoading.hidden = true
      elements.commentError.hidden = false
      elements.commentErrorMessage.textContent = describeError(error)
      elements.retryComments.hidden = false
    }
  }

  function beginReply(comment) {
    state.parentId = comment.id
    elements.replyTarget.hidden = false
    elements.replyTargetName.textContent = comment.authorName
    elements.form.scrollIntoView({ behavior: 'smooth', block: 'center' })
    window.setTimeout(() => fields.content.focus(), 250)
  }

  function cancelReply() {
    state.parentId = null
    elements.replyTarget.hidden = true
    elements.replyTargetName.textContent = ''
  }

  function startCooldown(seconds) {
    state.cooldownUntil = Date.now() + Math.max(1, seconds) * 1000
    window.clearInterval(state.cooldownTimer)
    state.cooldownTimer = window.setInterval(() => {
      if (Date.now() >= state.cooldownUntil) {
        window.clearInterval(state.cooldownTimer)
      }
      updateSubmitButton()
    }, 500)
    updateSubmitButton()
  }

  function validateForm() {
    const name = fields.authorName.value.trim()
    const email = fields.authorEmail.value.trim()
    const website = fields.authorWebsite.value.trim()
    const content = fields.content.value.trim()

    if (!name) {
      setFormStatus('请填写称呼。', 'error')
      fields.authorName.focus()
      return null
    }
    if (email && !fields.authorEmail.checkValidity()) {
      setFormStatus('邮箱格式不正确。', 'error')
      fields.authorEmail.focus()
      return null
    }
    if (website && !fields.authorWebsite.checkValidity()) {
      setFormStatus('网站地址需要以 http:// 或 https:// 开头。', 'error')
      fields.authorWebsite.focus()
      return null
    }
    if (!content) {
      setFormStatus('请写下评论内容。', 'error')
      fields.content.focus()
      return null
    }
    if (content.length > 2000) {
      setFormStatus('评论最多 2000 个字符。', 'error')
      fields.content.focus()
      return null
    }
    return {
      authorName: name,
      authorEmail: email,
      authorWebsite: website,
      content,
    }
  }

  async function submitComment(event) {
    event.preventDefault()
    if (!state.interactive || state.submitting) return

    const values = validateForm()
    if (!values) return

    const payload = {
      path: state.path,
      authorName: values.authorName,
      authorEmail: values.authorEmail,
      authorWebsite: values.authorWebsite,
      content: values.content,
      honeypot: fields.honeypot.value,
      formStartedAt: state.formStartedAt,
    }
    if (state.parentId) payload.parentId = state.parentId

    state.submitting = true
    setFormStatus('')
    updateSubmitButton()

    try {
      const result = await api('/public/comments', {
        method: 'POST',
        body: payload,
      })
      const pending = result?.status === 'pending'
      setFormStatus(
        pending
          ? '评论已提交，审核通过后会出现在这里。'
          : '评论已发布，谢谢你的参与。',
        'success',
      )
      showToast(pending ? '评论已进入审核队列。' : '评论已发布。')
      fields.content.value = ''
      state.formStartedAt = Date.now()
      cancelReply()
      await Promise.all([loadComments(), loadStats()])
    } catch (error) {
      if (error instanceof ApiError && error.status === 429) {
        startCooldown(error.retryAfter || 60)
      }
      setFormStatus(describeError(error), 'error')
      showToast(describeError(error), 'error')
    } finally {
      state.submitting = false
      updateSubmitButton()
    }
  }

  async function toggleLike() {
    if (!state.interactive || state.likeBusy) return
    state.likeBusy = true
    updateLikeButton()

    try {
      const nextLiked = !state.liked
      const data = nextLiked
        ? await api('/public/likes', {
            method: 'PUT',
            body: { path: state.path, visitorId: state.visitorId },
          })
        : await api('/public/likes', {
            method: 'DELETE',
            query: { path: state.path, visitorId: state.visitorId },
          })
      state.liked = Boolean(data?.liked)
      elements.likeCount.textContent = formatNumber(data?.likes)
      showToast(state.liked ? '已点赞。' : '已取消点赞。')
    } catch (error) {
      showToast(describeError(error), 'error')
    } finally {
      state.likeBusy = false
      updateLikeButton()
    }
  }

  async function initialize() {
    elements.articlePath.textContent = state.path
    elements.formStartedAt = Date.now()
    updateSubmitButton()

    if (!isConfigured()) {
      showSetupAlert()
      setApiState('config', '等待配置')
      elements.commentLoading.hidden = true
      setInteractive(false)
      return
    }

    try {
      state.site = await api('/public/site')
      setApiState('online', `已连接 · ${state.site.name}`)
      setInteractive(true)
      await recordViewIfNeeded()
      await Promise.all([
        loadStats().catch((error) => {
          showToast(describeError(error), 'error')
        }),
        loadComments(),
      ])
    } catch (error) {
      showSetupAlert(error instanceof ApiError && error.status === 401)
      setApiState('error', '连接失败')
      elements.commentLoading.hidden = true
      elements.commentError.hidden = false
      elements.commentErrorMessage.textContent = describeError(error)
      elements.retryComments.hidden = false
      setInteractive(false)
    }
  }

  elements.likeButton.addEventListener('click', toggleLike)
  elements.form.addEventListener('submit', submitComment)
  elements.cancelReply.addEventListener('click', cancelReply)
  elements.retryComments.addEventListener('click', () => {
    loadComments()
  })

  initialize().catch((error) => {
    console.error(error)
    setApiState('error', '初始化失败')
    showToast('示例初始化失败，请查看浏览器控制台。', 'error')
  })
})()
