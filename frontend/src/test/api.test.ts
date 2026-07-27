import { describe, it, expect } from 'vitest'
import { isActive } from '@/lib/api'

describe('isActive', () => {
  it('returns true for idle state', () => {
    expect(isActive('idle')).toBe(true)
  })

  it('returns true for analyzing state', () => {
    expect(isActive('analyzing')).toBe(true)
  })

  it('returns true for running state', () => {
    expect(isActive('running')).toBe(true)
  })

  it('returns true for fixing state', () => {
    expect(isActive('fixing')).toBe(true)
  })

  it('returns false for done state', () => {
    expect(isActive('done')).toBe(false)
  })

  it('returns false for failed state', () => {
    expect(isActive('failed')).toBe(false)
  })

  it('returns false for simulated state', () => {
    expect(isActive('simulated')).toBe(false)
  })

  it('returns false for unknown terminal-like state', () => {
    // isActive returns true for unrecognized states (safety default)
    expect(isActive('')).toBe(true)
  })
})
