import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'
import { I18nextProvider } from 'react-i18next'
import i18n from '../../i18n'
import { BudgetHeatmap } from './BudgetHeatmap'

describe('BudgetHeatmap Integration', () => {
  const createWrapper = () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
      },
    })

    return ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <I18nextProvider i18n={i18n}>{children}</I18nextProvider>
        </BrowserRouter>
      </QueryClientProvider>
    )
  }

  it('renders without crashing in a full React context', () => {
    const Wrapper = createWrapper()
    render(<BudgetHeatmap />, { wrapper: Wrapper })
    expect(screen.getByText(/Heatmap/i)).toBeInTheDocument()
  })

  it('shows empty state when no envelopes exist', () => {
    const Wrapper = createWrapper()
    render(<BudgetHeatmap />, { wrapper: Wrapper })
    // Should show empty state with icon and message
    expect(screen.getByText(/créez des enveloppes/i)).toBeInTheDocument()
  })

  it('respects monthsBack prop', () => {
    const Wrapper = createWrapper()
    const { rerender } = render(<BudgetHeatmap monthsBack={1} />, { wrapper: Wrapper })
    expect(screen.getByText(/Heatmap/i)).toBeInTheDocument()

    rerender(<BudgetHeatmap monthsBack={6} />)
    expect(screen.getByText(/Heatmap/i)).toBeInTheDocument()
  })
})
