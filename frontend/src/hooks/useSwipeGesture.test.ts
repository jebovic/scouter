import { describe, it, expect, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useSwipeGesture } from './useSwipeGesture'

function makeTouchEvent(clientX: number, clientY: number): React.TouchEvent {
  return {
    touches: [{ clientX, clientY }],
  } as unknown as React.TouchEvent
}

describe('useSwipeGesture', () => {
  it('returns initial state with swipeX=0 and isSwiping=false', () => {
    const { result } = renderHook(() => useSwipeGesture())
    expect(result.current.swipeX).toBe(0)
    expect(result.current.isSwiping).toBe(false)
    expect(result.current.handlers).toBeDefined()
  })

  it('sets isSwiping true on touch start', () => {
    const { result } = renderHook(() => useSwipeGesture())
    act(() => {
      result.current.handlers.onTouchStart(makeTouchEvent(100, 200))
    })
    expect(result.current.isSwiping).toBe(true)
  })

  it('updates swipeX on horizontal touch move', () => {
    const { result } = renderHook(() => useSwipeGesture())
    act(() => {
      result.current.handlers.onTouchStart(makeTouchEvent(100, 200))
    })
    act(() => {
      // Move significantly in X to establish horizontal direction
      result.current.handlers.onTouchMove(makeTouchEvent(160, 202))
    })
    expect(result.current.swipeX).toBe(60)
  })

  it('does not update swipeX on vertical move', () => {
    const { result } = renderHook(() => useSwipeGesture())
    act(() => {
      result.current.handlers.onTouchStart(makeTouchEvent(100, 200))
    })
    act(() => {
      // Move significantly in Y — should be treated as scroll
      result.current.handlers.onTouchMove(makeTouchEvent(102, 260))
    })
    expect(result.current.swipeX).toBe(0)
  })

  it('resets swipeX and isSwiping on touch end', () => {
    const { result } = renderHook(() => useSwipeGesture())
    act(() => {
      result.current.handlers.onTouchStart(makeTouchEvent(100, 200))
    })
    act(() => {
      result.current.handlers.onTouchMove(makeTouchEvent(130, 202))
    })
    act(() => {
      result.current.handlers.onTouchEnd()
    })
    expect(result.current.swipeX).toBe(0)
    expect(result.current.isSwiping).toBe(false)
  })

  it('calls onSwipeLeft when swiped left past threshold', () => {
    const onSwipeLeft = vi.fn()
    const { result } = renderHook(() =>
      useSwipeGesture({ threshold: 80, onSwipeLeft })
    )
    act(() => {
      result.current.handlers.onTouchStart(makeTouchEvent(300, 200))
    })
    act(() => {
      result.current.handlers.onTouchMove(makeTouchEvent(210, 202))
    })
    act(() => {
      result.current.handlers.onTouchEnd()
    })
    expect(onSwipeLeft).toHaveBeenCalledTimes(1)
  })

  it('calls onSwipeRight when swiped right past threshold', () => {
    const onSwipeRight = vi.fn()
    const { result } = renderHook(() =>
      useSwipeGesture({ threshold: 80, onSwipeRight })
    )
    act(() => {
      result.current.handlers.onTouchStart(makeTouchEvent(100, 200))
    })
    act(() => {
      result.current.handlers.onTouchMove(makeTouchEvent(190, 202))
    })
    act(() => {
      result.current.handlers.onTouchEnd()
    })
    expect(onSwipeRight).toHaveBeenCalledTimes(1)
  })

  it('does not call callbacks when swipe is below threshold', () => {
    const onSwipeLeft = vi.fn()
    const onSwipeRight = vi.fn()
    const { result } = renderHook(() =>
      useSwipeGesture({ threshold: 80, onSwipeLeft, onSwipeRight })
    )
    act(() => {
      result.current.handlers.onTouchStart(makeTouchEvent(100, 200))
    })
    act(() => {
      // Only 40px — below threshold of 80
      result.current.handlers.onTouchMove(makeTouchEvent(140, 202))
    })
    act(() => {
      result.current.handlers.onTouchEnd()
    })
    expect(onSwipeLeft).not.toHaveBeenCalled()
    expect(onSwipeRight).not.toHaveBeenCalled()
  })

  it('does not call onSwipeLeft when no callback provided (no crash)', () => {
    const { result } = renderHook(() => useSwipeGesture({ threshold: 80 }))
    act(() => {
      result.current.handlers.onTouchStart(makeTouchEvent(300, 200))
    })
    act(() => {
      result.current.handlers.onTouchMove(makeTouchEvent(200, 202))
    })
    expect(() => {
      act(() => {
        result.current.handlers.onTouchEnd()
      })
    }).not.toThrow()
  })

  it('ignores move with excessive vertical drift', () => {
    const { result } = renderHook(() =>
      useSwipeGesture({ maxVerticalDrift: 50 })
    )
    act(() => {
      result.current.handlers.onTouchStart(makeTouchEvent(100, 200))
    })
    act(() => {
      // First move establishes horizontal direction
      result.current.handlers.onTouchMove(makeTouchEvent(130, 202))
    })
    const swipeAfterFirst = result.current.swipeX
    act(() => {
      // Second move has too much vertical drift — swipeX should reset to 0
      result.current.handlers.onTouchMove(makeTouchEvent(150, 260))
    })
    expect(swipeAfterFirst).toBe(30)
    expect(result.current.swipeX).toBe(0)
  })

  it('handles multiple swipe sequences independently', () => {
    const onSwipeLeft = vi.fn()
    const { result } = renderHook(() =>
      useSwipeGesture({ threshold: 80, onSwipeLeft })
    )

    // First swipe — left
    act(() => { result.current.handlers.onTouchStart(makeTouchEvent(300, 200)) })
    act(() => { result.current.handlers.onTouchMove(makeTouchEvent(210, 202)) })
    act(() => { result.current.handlers.onTouchEnd() })

    // Second swipe — left again
    act(() => { result.current.handlers.onTouchStart(makeTouchEvent(300, 200)) })
    act(() => { result.current.handlers.onTouchMove(makeTouchEvent(210, 202)) })
    act(() => { result.current.handlers.onTouchEnd() })

    expect(onSwipeLeft).toHaveBeenCalledTimes(2)
  })
})
