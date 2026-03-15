import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { BudgetHeatmap } from './BudgetHeatmap'

vi.mock('../../hooks/useBudgetVariance', () => ({
  useBudgetVariance: vi.fn(),
}))

vi.mock('../../hooks/useEnvelopes', () => ({
  useEnvelopes: vi.fn(),
}))

import { useBudgetVariance } from '../../hooks/useBudgetVariance'
import { useEnvelopes } from '../../hooks/useEnvelopes'

const mockUseBudgetVariance = useBudgetVariance as any
const mockUseEnvelopes = useEnvelopes as any

describe('BudgetHeatmap', () => {
  afterEach(() => {
    vi.clearAllMocks()
  })

  it('renders empty state when no envelopes exist', () => {
    mockUseEnvelopes.mockReturnValue({
      envelopes: [],
      isLoading: false,
      error: null,
    })
    mockUseBudgetVariance.mockReturnValue({
      months: [],
      categories: [],
      cells: [],
    })

    render(<BudgetHeatmap />)
    expect(
      screen.getByText(/créez des enveloppes/i),
    ).toBeInTheDocument()
  })

  it('renders envelope names as row headers', () => {
    mockUseEnvelopes.mockReturnValue({
      envelopes: [],
      isLoading: false,
      error: null,
    })
    mockUseBudgetVariance.mockReturnValue({
      months: ['2025-01', '2025-02'],
      categories: ['Housing', 'Food'],
      cells: [
        {
          month: '2025-01',
          monthLabel: 'Jan 2025',
          category: 'Housing',
          budgeted: 1000,
          spent: 900,
          variance: -100,
          variancePct: -10,
        },
        {
          month: '2025-02',
          monthLabel: 'Feb 2025',
          category: 'Housing',
          budgeted: 1000,
          spent: 950,
          variance: -50,
          variancePct: -5,
        },
        {
          month: '2025-01',
          monthLabel: 'Jan 2025',
          category: 'Food',
          budgeted: 500,
          spent: 600,
          variance: 100,
          variancePct: 20,
        },
        {
          month: '2025-02',
          monthLabel: 'Feb 2025',
          category: 'Food',
          budgeted: 500,
          spent: 450,
          variance: -50,
          variancePct: -10,
        },
      ],
    })

    render(<BudgetHeatmap />)
    expect(screen.getByText('Housing')).toBeInTheDocument()
    expect(screen.getByText('Food')).toBeInTheDocument()
  })

  it('renders month labels as column headers', () => {
    mockUseEnvelopes.mockReturnValue({
      envelopes: [],
      isLoading: false,
      error: null,
    })
    mockUseBudgetVariance.mockReturnValue({
      months: ['2025-01', '2025-02'],
      categories: ['Housing'],
      cells: [
        {
          month: '2025-01',
          monthLabel: 'Jan 2025',
          category: 'Housing',
          budgeted: 1000,
          spent: 900,
          variance: -100,
          variancePct: -10,
        },
        {
          month: '2025-02',
          monthLabel: 'Feb 2025',
          category: 'Housing',
          budgeted: 1000,
          spent: 950,
          variance: -50,
          variancePct: -5,
        },
      ],
    })

    render(<BudgetHeatmap />)
    expect(screen.getByText('Jan 2025')).toBeInTheDocument()
    expect(screen.getByText('Feb 2025')).toBeInTheDocument()
  })

  it('renders heatmap cells with variance percentages', () => {
    mockUseEnvelopes.mockReturnValue({
      envelopes: [],
      isLoading: false,
      error: null,
    })
    mockUseBudgetVariance.mockReturnValue({
      months: ['2025-01'],
      categories: ['Housing'],
      cells: [
        {
          month: '2025-01',
          monthLabel: 'Jan 2025',
          category: 'Housing',
          budgeted: 1000,
          spent: 900,
          variance: -100,
          variancePct: -10,
        },
      ],
    })

    render(<BudgetHeatmap />)
    // Variance percentage should be displayed (-10%)
    expect(screen.getByText(/-10/)).toBeInTheDocument()
  })

  it('passes monthsBack prop to useBudgetVariance', () => {
    mockUseEnvelopes.mockReturnValue({
      envelopes: [],
      isLoading: false,
      error: null,
    })
    mockUseBudgetVariance.mockReturnValue({
      months: [],
      categories: [],
      cells: [],
    })

    render(<BudgetHeatmap monthsBack={6} />)
    expect(mockUseBudgetVariance).toHaveBeenCalledWith(expect.any(Array), 6)
  })

  it('renders legend with status labels', () => {
    mockUseEnvelopes.mockReturnValue({
      envelopes: [],
      isLoading: false,
      error: null,
    })
    mockUseBudgetVariance.mockReturnValue({
      months: ['2025-01'],
      categories: ['Housing'],
      cells: [
        {
          month: '2025-01',
          monthLabel: 'Jan 2025',
          category: 'Housing',
          budgeted: 1000,
          spent: 900,
          variance: -100,
          variancePct: -10,
        },
      ],
    })

    render(<BudgetHeatmap />)
    // Legend should contain status indicators
    expect(screen.getByText(/Under Budget|En dessous du budget/i)).toBeInTheDocument()
  })
})
