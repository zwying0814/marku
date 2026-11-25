import { MarkuComments } from './comments'
import { MarkuLike } from './like'
import { MarkuStats } from './stats'

export { MARKU_ATTRIBUTES, MarkuElement } from './base'
export { MarkuComments } from './comments'
export { MarkuLike } from './like'
export { MarkuStats } from './stats'
export { baseStyles, commentsStyles, likeStyles, statsStyles } from './styles'

export function registerMarkuElements(
  registry: CustomElementRegistry | undefined = typeof customElements === 'undefined'
    ? undefined
    : customElements,
): void {
  if (!registry) return
  const elements: Array<[string, CustomElementConstructor]> = [
    ['marku-comments', MarkuComments],
    ['marku-like', MarkuLike],
    ['marku-stats', MarkuStats],
  ]
  for (const [name, constructor] of elements) {
    if (!registry.get(name)) registry.define(name, constructor)
  }
}

registerMarkuElements()
