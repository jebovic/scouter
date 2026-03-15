import { useState, useCallback, useRef } from 'react'

export interface SwipeGestureState {
  swipeX: number
  isSwiping: boolean
  handlers: {
    onTouchStart: (e: React.TouchEvent) => void
    onTouchMove: (e: React.TouchEvent) => void
    onTouchEnd: () => void
  }
}

interface UseSwipeGestureOptions {
  /** Minimum pixels to travel before it counts as a swipe */
  threshold?: number
  /** Maximum pixels allowed in the perpendicular axis (prevents diagonal swipes) */
  maxVerticalDrift?: number
  onSwipeLeft?: () => void
  onSwipeRight?: () => void
}

export function useSwipeGesture({
  threshold = 80,
  maxVerticalDrift = 50,
  onSwipeLeft,
  onSwipeRight,
}: UseSwipeGestureOptions = {}): SwipeGestureState {
  const [swipeX, setSwipeX] = useState(0)
  const [isSwiping, setIsSwiping] = useState(false)
  const startX = useRef(0)
  const startY = useRef(0)
  const isHorizontal = useRef<boolean | null>(null)

  const onTouchStart = useCallback((e: React.TouchEvent) => {
    const touch = e.touches[0]
    startX.current = touch.clientX
    startY.current = touch.clientY
    isHorizontal.current = null
    setIsSwiping(true)
  }, [])

  const onTouchMove = useCallback((e: React.TouchEvent) => {
    const touch = e.touches[0]
    const dx = touch.clientX - startX.current
    const dy = touch.clientY - startY.current

    // Determine scroll axis on first significant move
    if (isHorizontal.current === null && (Math.abs(dx) > 5 || Math.abs(dy) > 5)) {
      isHorizontal.current = Math.abs(dx) > Math.abs(dy)
    }

    if (!isHorizontal.current) return

    // Block if vertical drift is too large
    if (Math.abs(dy) > maxVerticalDrift) {
      setSwipeX(0)
      return
    }

    setSwipeX(dx)
  }, [maxVerticalDrift])

  const onTouchEnd = useCallback(() => {
    setIsSwiping(false)

    if (swipeX < -threshold && onSwipeLeft) {
      onSwipeLeft()
    } else if (swipeX > threshold && onSwipeRight) {
      onSwipeRight()
    }

    setSwipeX(0)
    isHorizontal.current = null
  }, [swipeX, threshold, onSwipeLeft, onSwipeRight])

  return {
    swipeX,
    isSwiping,
    handlers: { onTouchStart, onTouchMove, onTouchEnd },
  }
}
