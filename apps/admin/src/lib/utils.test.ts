import { describe, expect, it } from 'vitest'

import {
  cn,
  extractErrorMessage,
  formatNumber,
  formatRelativeTime,
  getInitials,
  parseOrigins,
  slugify,
  truncate,
} from '@/lib/utils'

describe('utils', () => {
  it('merges conditional class names', () => {
    expect(cn('px-2', undefined, 'px-4')).toBe('px-4')
  })

  it('formats numbers for the admin UI', () => {
    expect(formatNumber(1200)).toBe('1,200')
    expect(formatNumber(null)).toBe('0')
  })

  it('formats relative dates', () => {
    const twoHoursAgo = new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString()
    expect(formatRelativeTime(twoHoursAgo)).toContain('2')
  })

  it('slugifies latin site names', () => {
    expect(slugify('My Tech Blog!')).toBe('my-tech-blog')
    expect(slugify('Café Notes')).toBe('cafe-notes')
  })

  it('parses and deduplicates origins', () => {
    expect(
      parseOrigins(
        'https://blog.example.com/\nhttp://localhost:36583, https://blog.example.com',
      ),
    ).toEqual(['https://blog.example.com', 'http://localhost:36583'])
  })

  it('truncates long content at the requested limit', () => {
    expect(truncate('hello world', 8)).toBe('hello wo…')
    expect(truncate('short', 8)).toBe('short')
  })

  it('normalizes unknown errors', () => {
    expect(extractErrorMessage(new Error('network down'))).toBe('network down')
    expect(extractErrorMessage('bad')).toBe('发生未知错误，请稍后重试')
  })

  it('builds avatar initials', () => {
    expect(getInitials('Admin')).toBe('AD')
    expect(getInitials('')).toBe('A')
  })
})
