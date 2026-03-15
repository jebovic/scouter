import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { usePerformanceMetrics } from './usePerformanceMetrics'

describe('usePerformanceMetrics', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Navigation Timing API integration', () => {
    it('computes TTFB (Time to First Byte) correctly', () => {
      const mockTiming = {
        navigationStart: 1000,
        responseStart: 1150,
      }
      Object.defineProperty(window, 'performance', {
        value: { timing: mockTiming },
        writable: true,
      })

      const { result } = renderHook(() => usePerformanceMetrics())
      expect(result.current.metrics.ttfb).toBe(150)
    })

    it('computes DOM Content Loaded time correctly', () => {
      const mockTiming = {
        navigationStart: 1000,
        domContentLoadedEventEnd: 2500,
      }
      Object.defineProperty(window, 'performance', {
        value: { timing: mockTiming },
        writable: true,
      })

      const { result } = renderHook(() => usePerformanceMetrics())
      expect(result.current.metrics.domContentLoaded).toBe(1500)
    })

    it('computes full load time correctly', () => {
      const mockTiming = {
        navigationStart: 1000,
        loadEventEnd: 3500,
      }
      Object.defineProperty(window, 'performance', {
        value: { timing: mockTiming },
        writable: true,
      })

      const { result } = renderHook(() => usePerformanceMetrics())
      expect(result.current.metrics.loadTime).toBe(2500)
    })

    it('returns null when timing values are 0 (not available)', () => {
      const mockTiming = {
        navigationStart: 0,
        responseStart: 0,
        domContentLoadedEventEnd: 0,
        loadEventEnd: 0,
      }
      Object.defineProperty(window, 'performance', {
        value: { timing: mockTiming },
        writable: true,
      })

      const { result } = renderHook(() => usePerformanceMetrics())
      expect(result.current.metrics.ttfb).toBeNull()
      expect(result.current.metrics.domContentLoaded).toBeNull()
      expect(result.current.metrics.loadTime).toBeNull()
    })
  })

  describe('Memory API integration (Chrome only)', () => {
    it('reads JS heap size from performance.memory if available', () => {
      const mockMemory = {
        usedJSHeapSize: 25000000,
        totalJSHeapSize: 50000000,
      }
      Object.defineProperty(window, 'performance', {
        value: { memory: mockMemory },
        writable: true,
      })

      const { result } = renderHook(() => usePerformanceMetrics())
      expect(result.current.metrics.usedJSHeapSize).toBe(25000000)
      expect(result.current.metrics.totalJSHeapSize).toBe(50000000)
    })

    it('returns null for memory when performance.memory is not available', () => {
      const originalPerformance = window.performance
      Object.defineProperty(window, 'performance', {
        value: { memory: undefined },
        writable: true,
      })

      const { result } = renderHook(() => usePerformanceMetrics())
      expect(result.current.metrics.usedJSHeapSize).toBeNull()
      expect(result.current.metrics.totalJSHeapSize).toBeNull()

      Object.defineProperty(window, 'performance', {
        value: originalPerformance,
        writable: true,
      })
    })
  })

  describe('Connection API integration', () => {
    it('reads effective connection type', () => {
      const mockConnection = { effectiveType: '4g' }
      Object.defineProperty(navigator, 'connection', {
        value: mockConnection,
        writable: true,
        configurable: true,
      })

      const { result } = renderHook(() => usePerformanceMetrics())
      expect(result.current.metrics.connectionType).toBe('4g')

      delete (navigator as any).connection
    })

    it('returns null for connection type when not available', () => {
      const originalConnection = (navigator as any).connection
      Object.defineProperty(navigator, 'connection', {
        value: undefined,
        writable: true,
        configurable: true,
      })

      const { result } = renderHook(() => usePerformanceMetrics())
      expect(result.current.metrics.connectionType).toBeNull()

      if (originalConnection) {
        Object.defineProperty(navigator, 'connection', {
          value: originalConnection,
          writable: true,
          configurable: true,
        })
      }
    })
  })

  describe('collectedAt timestamp', () => {
    it('sets collectedAt to current date on initial render', () => {
      const beforeRender = new Date()
      const { result } = renderHook(() => usePerformanceMetrics())
      const afterRender = new Date()

      expect(result.current.metrics.collectedAt).toBeInstanceOf(Date)
      expect(result.current.metrics.collectedAt.getTime()).toBeGreaterThanOrEqual(
        beforeRender.getTime()
      )
      expect(result.current.metrics.collectedAt.getTime()).toBeLessThanOrEqual(
        afterRender.getTime()
      )
    })

    it('updates collectedAt when refresh() is called', async () => {
      const { result } = renderHook(() => usePerformanceMetrics())
      const initialCollectedAt = result.current.metrics.collectedAt

      // Wait a bit to ensure time difference
      await new Promise((resolve) => setTimeout(resolve, 10))

      result.current.refresh()

      await waitFor(() => {
        expect(result.current.metrics.collectedAt).not.toBe(initialCollectedAt)
        expect(result.current.metrics.collectedAt.getTime()).toBeGreaterThan(
          initialCollectedAt.getTime()
        )
      })
    })
  })

  describe('refresh() function', () => {
    it('provides a refresh function that re-collects metrics', async () => {
      const mockTiming = {
        navigationStart: 1000,
        responseStart: 1150,
      }
      Object.defineProperty(window, 'performance', {
        value: { timing: mockTiming },
        writable: true,
      })

      const { result } = renderHook(() => usePerformanceMetrics())
      expect(typeof result.current.refresh).toBe('function')

      const initialTtfb = result.current.metrics.ttfb
      result.current.refresh()

      expect(result.current.metrics.ttfb).toBe(initialTtfb)
    })
  })

  describe('metric interface structure', () => {
    it('returns metrics object with all required properties', () => {
      const { result } = renderHook(() => usePerformanceMetrics())
      const metrics = result.current.metrics

      expect(metrics).toHaveProperty('domContentLoaded')
      expect(metrics).toHaveProperty('loadTime')
      expect(metrics).toHaveProperty('ttfb')
      expect(metrics).toHaveProperty('usedJSHeapSize')
      expect(metrics).toHaveProperty('totalJSHeapSize')
      expect(metrics).toHaveProperty('connectionType')
      expect(metrics).toHaveProperty('collectedAt')
    })

    it('returns object with refresh function', () => {
      const { result } = renderHook(() => usePerformanceMetrics())
      expect(result.current).toHaveProperty('metrics')
      expect(result.current).toHaveProperty('refresh')
    })
  })
})
