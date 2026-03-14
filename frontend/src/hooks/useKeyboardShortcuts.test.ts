import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook } from '@testing-library/react'
import { useKeyboardShortcuts, type ShortcutMap } from './useKeyboardShortcuts'

function fireKey(key: string, options: Partial<KeyboardEventInit> = {}) {
  const event = new KeyboardEvent('keydown', { key, bubbles: true, ...options })
  document.dispatchEvent(event)
}

describe('useKeyboardShortcuts', () => {
  let shortcuts: ShortcutMap

  beforeEach(() => {
    shortcuts = {
      n: vi.fn(),
      r: vi.fn(),
      p: vi.fn(),
      '/': vi.fn(),
    }
  })

  it('calls the correct callback when N is pressed', () => {
    renderHook(() => useKeyboardShortcuts(shortcuts))
    fireKey('n')
    expect(shortcuts.n).toHaveBeenCalledOnce()
  })

  it('calls the correct callback when R is pressed', () => {
    renderHook(() => useKeyboardShortcuts(shortcuts))
    fireKey('r')
    expect(shortcuts.r).toHaveBeenCalledOnce()
  })

  it('calls the correct callback when P is pressed', () => {
    renderHook(() => useKeyboardShortcuts(shortcuts))
    fireKey('p')
    expect(shortcuts.p).toHaveBeenCalledOnce()
  })

  it('calls the correct callback when / is pressed', () => {
    renderHook(() => useKeyboardShortcuts(shortcuts))
    fireKey('/')
    expect(shortcuts['/'] as ReturnType<typeof vi.fn>).toHaveBeenCalledOnce()
  })

  it('does not fire when Ctrl modifier is held', () => {
    renderHook(() => useKeyboardShortcuts(shortcuts))
    fireKey('n', { ctrlKey: true })
    expect(shortcuts.n).not.toHaveBeenCalled()
  })

  it('does not fire when Meta modifier is held', () => {
    renderHook(() => useKeyboardShortcuts(shortcuts))
    fireKey('n', { metaKey: true })
    expect(shortcuts.n).not.toHaveBeenCalled()
  })

  it('does not fire when Alt modifier is held', () => {
    renderHook(() => useKeyboardShortcuts(shortcuts))
    fireKey('n', { altKey: true })
    expect(shortcuts.n).not.toHaveBeenCalled()
  })

  it('does not fire when target is an input element', () => {
    renderHook(() => useKeyboardShortcuts(shortcuts))
    const input = document.createElement('input')
    document.body.appendChild(input)
    const event = new KeyboardEvent('keydown', { key: 'n', bubbles: true })
    Object.defineProperty(event, 'target', { value: input })
    document.dispatchEvent(event)
    expect(shortcuts.n).not.toHaveBeenCalled()
    document.body.removeChild(input)
  })

  it('does not fire when target is a textarea', () => {
    renderHook(() => useKeyboardShortcuts(shortcuts))
    const textarea = document.createElement('textarea')
    document.body.appendChild(textarea)
    const event = new KeyboardEvent('keydown', { key: 'n', bubbles: true })
    Object.defineProperty(event, 'target', { value: textarea })
    document.dispatchEvent(event)
    expect(shortcuts.n).not.toHaveBeenCalled()
    document.body.removeChild(textarea)
  })

  it('removes listener on unmount', () => {
    const { unmount } = renderHook(() => useKeyboardShortcuts(shortcuts))
    unmount()
    fireKey('n')
    expect(shortcuts.n).not.toHaveBeenCalled()
  })

  it('does not fire for unregistered keys', () => {
    renderHook(() => useKeyboardShortcuts(shortcuts))
    fireKey('x')
    Object.values(shortcuts).forEach(fn => expect(fn as ReturnType<typeof vi.fn>).not.toHaveBeenCalled())
  })
})
