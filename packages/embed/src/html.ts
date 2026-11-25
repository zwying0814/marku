export type Child = Node | string | number | null | undefined | false

type Props = Record<string, unknown>

export function h<K extends keyof HTMLElementTagNameMap>(
  tag: K,
  props: Props = {},
  ...children: Child[]
): HTMLElementTagNameMap[K] {
  const element = document.createElement(tag)

  for (const [key, value] of Object.entries(props)) {
    if (value === undefined || value === null || value === false) continue
    if (key === 'class') {
      element.className = String(value)
      continue
    }
    if (key === 'text') {
      element.textContent = String(value)
      continue
    }
    if (key === 'dataset' && typeof value === 'object') {
      Object.assign(element.dataset, value)
      continue
    }
    if (key === 'value' && 'value' in element && typeof value !== 'boolean') {
      ;(element as HTMLInputElement | HTMLTextAreaElement).value = String(value)
      continue
    }
    if (key.startsWith('on') && typeof value === 'function') {
      element.addEventListener(key.slice(2).toLowerCase(), value as EventListener)
      continue
    }
    if (value === true) {
      element.setAttribute(key, '')
      continue
    }
    element.setAttribute(key, String(value))
  }

  append(element, children)
  return element
}

export function append(parent: Node, children: Child[]): void {
  for (const child of children) {
    if (child === null || child === undefined || child === false) continue
    const node =
      typeof child === 'number' || typeof child === 'string'
        ? document.createTextNode(String(child))
        : child
    parent.appendChild(node)
  }
}

export function isAbortError(error: unknown): boolean {
  return (
    typeof error === 'object' &&
    error !== null &&
    'name' in error &&
    (error as { name?: string }).name === 'AbortError'
  )
}
