import { describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { renderWithProviders } from '../test/utils'
import PerformancePage from './PerformancePage'

// Mock the hook
vi.mock('../hooks/usePerformanceMetrics', () => ({
  usePerformanceMetrics: vi.fn(),
}))

import { usePerformanceMetrics } from '../hooks/usePerformanceMetrics'

describe('PerformancePage', () => {
  it('renders the performance page title', () => {
    const mockMetrics = {
      domContentLoaded: null,
      loadTime: null,
      ttfb: null,
      usedJSHeapSize: null,
      totalJSHeapSize: null,
      connectionType: null,
      collectedAt: new Date(),
    }
    vi.mocked(usePerformanceMetrics).mockReturnValue({
      metrics: mockMetrics,
      refresh: vi.fn(),
    })

    renderWithProviders(<PerformancePage />)
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent(/performance|perfs/i)
  })

  describe('metric card rendering', () => {
    it('renders metric cards for each metric', () => {
      const mockMetrics = {
        domContentLoaded: 1500,
        loadTime: 2500,
        ttfb: 150,
        usedJSHeapSize: 25000000,
        totalJSHeapSize: 50000000,
        connectionType: '4g',
        collectedAt: new Date(),
      }
      vi.mocked(usePerformanceMetrics).mockReturnValue({
        metrics: mockMetrics,
        refresh: vi.fn(),
      })

      renderWithProviders(<PerformancePage />)
      expect(screen.getByText(/TTFB/i)).toBeInTheDocument()
      expect(screen.getByText(/DOM Content Loaded/i)).toBeInTheDocument()
      expect(screen.getByText(/Load Time/i)).toBeInTheDocument()
      expect(screen.getByText(/JS Heap Used/i)).toBeInTheDocument()
    })

    it('displays metric values in correct units', () => {
      const mockMetrics = {
        domContentLoaded: 1500,
        loadTime: 2500,
        ttfb: 150,
        usedJSHeapSize: 26214400, // 25 MB exactly
        totalJSHeapSize: 50000000,
        connectionType: '4g',
        collectedAt: new Date(),
      }
      vi.mocked(usePerformanceMetrics).mockReturnValue({
        metrics: mockMetrics,
        refresh: vi.fn(),
      })

      renderWithProviders(<PerformancePage />)
      // TTFB: 150ms
      expect(screen.getByText('150')).toBeInTheDocument()
      // DOM Content Loaded: 1500ms
      expect(screen.getByText('1500')).toBeInTheDocument()
      // Load Time: 2500ms
      expect(screen.getByText('2500')).toBeInTheDocument()
      // JS Heap: should be in MB (25.0MB)
      expect(screen.getByText('25.0')).toBeInTheDocument()
    })

    it('shows "N/A" for null metrics', () => {
      const mockMetrics = {
        domContentLoaded: null,
        loadTime: null,
        ttfb: null,
        usedJSHeapSize: null,
        totalJSHeapSize: null,
        connectionType: null,
        collectedAt: new Date(),
      }
      vi.mocked(usePerformanceMetrics).mockReturnValue({
        metrics: mockMetrics,
        refresh: vi.fn(),
      })

      renderWithProviders(<PerformancePage />)
      const naElements = screen.getAllByText('N/A')
      expect(naElements.length).toBeGreaterThan(0)
    })
  })

  describe('color coding based on thresholds', () => {
    it('applies green color for TTFB < 200ms', () => {
      const mockMetrics = {
        domContentLoaded: 1500,
        loadTime: 2500,
        ttfb: 100,
        usedJSHeapSize: 25000000,
        totalJSHeapSize: 50000000,
        connectionType: '4g',
        collectedAt: new Date(),
      }
      vi.mocked(usePerformanceMetrics).mockReturnValue({
        metrics: mockMetrics,
        refresh: vi.fn(),
      })

      renderWithProviders(<PerformancePage />)
      // Check for green color indicator (check for metric card with specific class)
      const ttfbCard = screen.getByText(/TTFB/i).closest('div.metricCard')
      expect(ttfbCard).toHaveClass('green')
    })

    it('applies gold color for TTFB 200-500ms', () => {
      const mockMetrics = {
        domContentLoaded: 1500,
        loadTime: 2500,
        ttfb: 300,
        usedJSHeapSize: 25000000,
        totalJSHeapSize: 50000000,
        connectionType: '4g',
        collectedAt: new Date(),
      }
      vi.mocked(usePerformanceMetrics).mockReturnValue({
        metrics: mockMetrics,
        refresh: vi.fn(),
      })

      renderWithProviders(<PerformancePage />)
      const ttfbCard = screen.getByText(/TTFB/i).closest('div.metricCard')
      expect(ttfbCard).toHaveClass('gold')
    })

    it('applies coral color for TTFB >= 500ms', () => {
      const mockMetrics = {
        domContentLoaded: 1500,
        loadTime: 2500,
        ttfb: 600,
        usedJSHeapSize: 25000000,
        totalJSHeapSize: 50000000,
        connectionType: '4g',
        collectedAt: new Date(),
      }
      vi.mocked(usePerformanceMetrics).mockReturnValue({
        metrics: mockMetrics,
        refresh: vi.fn(),
      })

      renderWithProviders(<PerformancePage />)
      const ttfbCard = screen.getByText(/TTFB/i).closest('div.metricCard')
      expect(ttfbCard).toHaveClass('coral')
    })

    it('applies green for load time < 1000ms', () => {
      const mockMetrics = {
        domContentLoaded: 500,
        loadTime: 800,
        ttfb: 150,
        usedJSHeapSize: 25000000,
        totalJSHeapSize: 50000000,
        connectionType: '4g',
        collectedAt: new Date(),
      }
      vi.mocked(usePerformanceMetrics).mockReturnValue({
        metrics: mockMetrics,
        refresh: vi.fn(),
      })

      renderWithProviders(<PerformancePage />)
      const loadTimeCard = screen.getByText(/Load Time/i).closest('div.metricCard')
      expect(loadTimeCard).toHaveClass('green')
    })

    it('applies gold for load time 1000-3000ms', () => {
      const mockMetrics = {
        domContentLoaded: 1500,
        loadTime: 2000,
        ttfb: 150,
        usedJSHeapSize: 25000000,
        totalJSHeapSize: 50000000,
        connectionType: '4g',
        collectedAt: new Date(),
      }
      vi.mocked(usePerformanceMetrics).mockReturnValue({
        metrics: mockMetrics,
        refresh: vi.fn(),
      })

      renderWithProviders(<PerformancePage />)
      const loadTimeCard = screen.getByText(/Load Time/i).closest('div.metricCard')
      expect(loadTimeCard).toHaveClass('gold')
    })

    it('applies coral for load time >= 3000ms', () => {
      const mockMetrics = {
        domContentLoaded: 1500,
        loadTime: 5000,
        ttfb: 150,
        usedJSHeapSize: 25000000,
        totalJSHeapSize: 50000000,
        connectionType: '4g',
        collectedAt: new Date(),
      }
      vi.mocked(usePerformanceMetrics).mockReturnValue({
        metrics: mockMetrics,
        refresh: vi.fn(),
      })

      renderWithProviders(<PerformancePage />)
      const loadTimeCard = screen.getByText(/Load Time/i).closest('div.metricCard')
      expect(loadTimeCard).toHaveClass('coral')
    })

    it('applies green for heap < 50MB', () => {
      const mockMetrics = {
        domContentLoaded: 1500,
        loadTime: 2500,
        ttfb: 150,
        usedJSHeapSize: 30000000,
        totalJSHeapSize: 50000000,
        connectionType: '4g',
        collectedAt: new Date(),
      }
      vi.mocked(usePerformanceMetrics).mockReturnValue({
        metrics: mockMetrics,
        refresh: vi.fn(),
      })

      renderWithProviders(<PerformancePage />)
      const heapCard = screen.getByText(/JS Heap Used/i).closest('div.metricCard')
      expect(heapCard).toHaveClass('green')
    })

    it('applies gold for heap 50-100MB', () => {
      const mockMetrics = {
        domContentLoaded: 1500,
        loadTime: 2500,
        ttfb: 150,
        usedJSHeapSize: 75000000,
        totalJSHeapSize: 150000000,
        connectionType: '4g',
        collectedAt: new Date(),
      }
      vi.mocked(usePerformanceMetrics).mockReturnValue({
        metrics: mockMetrics,
        refresh: vi.fn(),
      })

      renderWithProviders(<PerformancePage />)
      const heapCard = screen.getByText(/JS Heap Used/i).closest('div.metricCard')
      expect(heapCard).toHaveClass('gold')
    })

    it('applies coral for heap >= 100MB', () => {
      const mockMetrics = {
        domContentLoaded: 1500,
        loadTime: 2500,
        ttfb: 150,
        usedJSHeapSize: 120000000,
        totalJSHeapSize: 200000000,
        connectionType: '4g',
        collectedAt: new Date(),
      }
      vi.mocked(usePerformanceMetrics).mockReturnValue({
        metrics: mockMetrics,
        refresh: vi.fn(),
      })

      renderWithProviders(<PerformancePage />)
      const heapCard = screen.getByText(/JS Heap Used/i).closest('div.metricCard')
      expect(heapCard).toHaveClass('coral')
    })
  })

  describe('refresh button', () => {
    it('renders a refresh button', () => {
      const mockMetrics = {
        domContentLoaded: null,
        loadTime: null,
        ttfb: null,
        usedJSHeapSize: null,
        totalJSHeapSize: null,
        connectionType: null,
        collectedAt: new Date(),
      }
      vi.mocked(usePerformanceMetrics).mockReturnValue({
        metrics: mockMetrics,
        refresh: vi.fn(),
      })

      renderWithProviders(<PerformancePage />)
      expect(screen.getByRole('button')).toBeInTheDocument()
    })

    it('calls refresh() when button is clicked', async () => {
      const mockRefresh = vi.fn()
      const mockMetrics = {
        domContentLoaded: null,
        loadTime: null,
        ttfb: null,
        usedJSHeapSize: null,
        totalJSHeapSize: null,
        connectionType: null,
        collectedAt: new Date(),
      }
      vi.mocked(usePerformanceMetrics).mockReturnValue({
        metrics: mockMetrics,
        refresh: mockRefresh,
      })

      renderWithProviders(<PerformancePage />)
      const button = screen.getByRole('button')
      await userEvent.click(button)
      expect(mockRefresh).toHaveBeenCalled()
    })
  })

  describe('disclaimer', () => {
    it('renders disclaimer text about real-time metrics', () => {
      const mockMetrics = {
        domContentLoaded: null,
        loadTime: null,
        ttfb: null,
        usedJSHeapSize: null,
        totalJSHeapSize: null,
        connectionType: null,
        collectedAt: new Date(),
      }
      vi.mocked(usePerformanceMetrics).mockReturnValue({
        metrics: mockMetrics,
        refresh: vi.fn(),
      })

      renderWithProviders(<PerformancePage />)
      // French text: "Métriques mesurées en temps réel"
      expect(screen.getByText(/Métriques mesurées|measured in real/i)).toBeInTheDocument()
    })
  })
})
