import { useState, useCallback } from 'react'

export interface PerformanceMetrics {
  // Navigation Timing
  domContentLoaded: number | null // ms
  loadTime: number | null // ms
  ttfb: number | null // Time to First Byte ms
  // Memory (Chrome only)
  usedJSHeapSize: number | null // bytes
  totalJSHeapSize: number | null // bytes
  // Connection
  connectionType: string | null // 4g, 3g, etc
  // Collected at
  collectedAt: Date
}

interface UsePerformanceMetricsReturn {
  metrics: PerformanceMetrics
  refresh: () => void
}

function collectMetrics(): PerformanceMetrics {
  const collectedAt = new Date()

  // Collect Navigation Timing metrics
  let domContentLoaded: number | null = null
  let loadTime: number | null = null
  let ttfb: number | null = null

  if (typeof window !== 'undefined' && window.performance && window.performance.timing) {
    const timing = window.performance.timing
    const navigationStart = timing.navigationStart || 0

    // TTFB: Time from navigationStart to responseStart
    if (timing.responseStart && timing.responseStart > 0) {
      ttfb = Math.max(0, timing.responseStart - navigationStart)
    }

    // DOM Content Loaded: Time from navigationStart to domContentLoadedEventEnd
    if (timing.domContentLoadedEventEnd && timing.domContentLoadedEventEnd > 0) {
      domContentLoaded = Math.max(0, timing.domContentLoadedEventEnd - navigationStart)
    }

    // Full Load Time: Time from navigationStart to loadEventEnd
    if (timing.loadEventEnd && timing.loadEventEnd > 0) {
      loadTime = Math.max(0, timing.loadEventEnd - navigationStart)
    }
  }

  // Collect Memory metrics (Chrome only)
  let usedJSHeapSize: number | null = null
  let totalJSHeapSize: number | null = null

  if (
    typeof window !== 'undefined' &&
    window.performance &&
    (window.performance as any).memory
  ) {
    const memory = (window.performance as any).memory
    usedJSHeapSize = memory.usedJSHeapSize || null
    totalJSHeapSize = memory.totalJSHeapSize || null
  }

  // Collect Connection info
  let connectionType: string | null = null

  if (typeof navigator !== 'undefined' && (navigator as any).connection) {
    connectionType = (navigator as any).connection.effectiveType || null
  }

  return {
    domContentLoaded,
    loadTime,
    ttfb,
    usedJSHeapSize,
    totalJSHeapSize,
    connectionType,
    collectedAt,
  }
}

export function usePerformanceMetrics(): UsePerformanceMetricsReturn {
  const [metrics, setMetrics] = useState<PerformanceMetrics>(collectMetrics())

  const refresh = useCallback(() => {
    setMetrics(collectMetrics())
  }, [])

  return { metrics, refresh }
}
