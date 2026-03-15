import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { PricePredictionBadge } from './PricePredictionBadge'
import * as hooks from '../../hooks/usePricePrediction'
import React, { ReactNode } from 'react'

vi.mock('../../hooks/usePricePrediction')

const makeWrapper =
  (qc: QueryClient) =>
  ({ children }: { children: ReactNode }) =>
    React.createElement(QueryClientProvider, { client: qc }, children)

describe('PricePredictionBadge', () => {
  let qc: QueryClient

  beforeEach(() => {
    qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    vi.clearAllMocks()
  })

  function mockHook(partial: Partial<ReturnType<typeof hooks.usePricePrediction>>) {
    vi.spyOn(hooks, 'usePricePrediction').mockReturnValue({
      data: undefined,
      isLoading: false,
      isSuccess: false,
      isError: false,
      error: null,
      isPending: false,
      isFetching: false,
      isStale: false,
      refetch: vi.fn(),
      ...partial,
    } as ReturnType<typeof hooks.usePricePrediction>)
  }

  it('renders nothing when data is undefined', () => {
    mockHook({ data: undefined })
    const { container } = render(
      <PricePredictionBadge itemId="item-1" />,
      { wrapper: makeWrapper(qc) },
    )
    expect(container.firstChild).toBeNull()
  })

  it('renders rising trend badge', () => {
    mockHook({
      data: {
        itemId: 'item-1',
        currentPrice: 100,
        predictedPrice: 120,
        confidenceLevel: 'high',
        trend: 'rising',
        trendPct: 20,
        dataPoints: 5,
        forecastDays: 30,
      },
    })

    render(<PricePredictionBadge itemId="item-1" />, { wrapper: makeWrapper(qc) })

    expect(screen.getByText(/En hausse/i)).toBeInTheDocument()
  })

  it('renders falling trend badge', () => {
    mockHook({
      data: {
        itemId: 'item-1',
        currentPrice: 100,
        predictedPrice: 80,
        confidenceLevel: 'medium',
        trend: 'falling',
        trendPct: -20,
        dataPoints: 4,
        forecastDays: 30,
      },
    })

    render(<PricePredictionBadge itemId="item-1" />, { wrapper: makeWrapper(qc) })

    expect(screen.getByText(/En baisse/i)).toBeInTheDocument()
  })

  it('renders stable trend badge', () => {
    mockHook({
      data: {
        itemId: 'item-1',
        currentPrice: 100,
        predictedPrice: 100,
        confidenceLevel: 'low',
        trend: 'stable',
        trendPct: 0,
        dataPoints: 2,
        forecastDays: 30,
      },
    })

    render(<PricePredictionBadge itemId="item-1" />, { wrapper: makeWrapper(qc) })

    expect(screen.getByText(/Stable/i)).toBeInTheDocument()
  })

  it('shows high confidence indicator', () => {
    mockHook({
      data: {
        itemId: 'item-1',
        currentPrice: 100,
        predictedPrice: 110,
        confidenceLevel: 'high',
        trend: 'rising',
        trendPct: 10,
        dataPoints: 6,
        forecastDays: 30,
      },
    })

    render(<PricePredictionBadge itemId="item-1" />, { wrapper: makeWrapper(qc) })

    // Green dot for high confidence
    expect(screen.getByTitle(/confiance élevée/i)).toBeInTheDocument()
  })

  it('shows low confidence indicator', () => {
    mockHook({
      data: {
        itemId: 'item-1',
        currentPrice: 100,
        predictedPrice: 100,
        confidenceLevel: 'low',
        trend: 'stable',
        trendPct: 0,
        dataPoints: 1,
        forecastDays: 30,
      },
    })

    render(<PricePredictionBadge itemId="item-1" />, { wrapper: makeWrapper(qc) })

    expect(screen.getByTitle(/confiance faible/i)).toBeInTheDocument()
  })

  it('displays tooltip with data points count', () => {
    mockHook({
      data: {
        itemId: 'item-1',
        currentPrice: 100,
        predictedPrice: 115,
        confidenceLevel: 'high',
        trend: 'rising',
        trendPct: 15,
        dataPoints: 7,
        forecastDays: 30,
      },
    })

    render(<PricePredictionBadge itemId="item-1" />, { wrapper: makeWrapper(qc) })

    // Tooltip text should include number of data points
    expect(screen.getByTitle(/7 relevés/i)).toBeInTheDocument()
  })

  it('does not render when dataPoints is 0', () => {
    mockHook({
      data: {
        itemId: 'item-1',
        currentPrice: 100,
        predictedPrice: 100,
        confidenceLevel: 'low',
        trend: 'stable',
        trendPct: 0,
        dataPoints: 0,
        forecastDays: 30,
      },
    })

    const { container } = render(
      <PricePredictionBadge itemId="item-1" />,
      { wrapper: makeWrapper(qc) },
    )

    expect(container.firstChild).toBeNull()
  })

  it('shows percentage in rising badge', () => {
    mockHook({
      data: {
        itemId: 'item-1',
        currentPrice: 100,
        predictedPrice: 125,
        confidenceLevel: 'high',
        trend: 'rising',
        trendPct: 25,
        dataPoints: 5,
        forecastDays: 30,
      },
    })

    render(<PricePredictionBadge itemId="item-1" />, { wrapper: makeWrapper(qc) })

    expect(screen.getByText(/25/)).toBeInTheDocument()
  })
})
