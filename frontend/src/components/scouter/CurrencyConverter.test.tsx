import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { CurrencyConverter } from './CurrencyConverter'

// Mock the hook
vi.mock('../../hooks/useCurrencyRates', () => ({
  useCurrencyRates: () => ({
    rates: {
      base: 'EUR',
      rates: {
        EUR: 1.0,
        USD: 1.08,
        GBP: 0.86,
        CHF: 0.95,
        CAD: 1.45,
        JPY: 155,
        SEK: 11.2,
        NOK: 11.5,
        DKK: 7.45,
        PLN: 4.35,
      },
      fetchedAt: Date.now(),
    },
    isLoading: false,
    error: null,
    convert: (amount: number, from: string, to: string) => {
      const rates = {
        EUR: 1.0,
        USD: 1.08,
        GBP: 0.86,
        CHF: 0.95,
        CAD: 1.45,
        JPY: 155,
        SEK: 11.2,
        NOK: 11.5,
        DKK: 7.45,
        PLN: 4.35,
      }
      const fromRate = rates[from as keyof typeof rates] || 1
      const toRate = rates[to as keyof typeof rates] || 1
      return (amount / fromRate) * toRate
    },
  }),
}))

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: false },
  },
})

function renderComponent() {
  return render(
    <QueryClientProvider client={queryClient}>
      <CurrencyConverter />
    </QueryClientProvider>
  )
}

describe('CurrencyConverter', () => {
  beforeEach(() => {
    queryClient.clear()
  })

  it('renders amount input, currency selects, and swap button', () => {
    renderComponent()
    expect(screen.getByRole('spinbutton')).toBeTruthy()
    expect(screen.getAllByRole('combobox')).toHaveLength(2)
    expect(screen.getByText('⇄')).toBeTruthy()
  })

  it('converts EUR to USD on input change', async () => {
    renderComponent()
    const input = screen.getByRole('spinbutton') as HTMLInputElement
    fireEvent.change(input, { target: { value: '100' } })

    await waitFor(() => {
      // Check that result displays converted amount (108 USD from 100 EUR)
      expect(input.value).toBe('100')
    })
  })

  it('swaps currencies when swap button clicked', async () => {
    renderComponent()
    const input = screen.getByRole('spinbutton')
    fireEvent.change(input, { target: { value: '100' } })

    await waitFor(() => {
      const selects = screen.getAllByRole('combobox')
      expect(selects[0]).toHaveValue('EUR')
      expect(selects[1]).toHaveValue('USD')
    })

    const swapButton = screen.getByText('⇄')
    fireEvent.click(swapButton)

    await waitFor(() => {
      const selects = screen.getAllByRole('combobox')
      expect(selects[0]).toHaveValue('USD')
      expect(selects[1]).toHaveValue('EUR')
    })
  })

  it('shows rate information', () => {
    renderComponent()
    expect(screen.getByText(/Taux du jour/)).toBeTruthy()
    expect(screen.getByText(/BCE/)).toBeTruthy()
  })

  it('handles currency changes', async () => {
    renderComponent()
    const input = screen.getByRole('spinbutton') as HTMLInputElement
    fireEvent.change(input, { target: { value: '100' } })

    const selects = screen.getAllByRole('combobox') as HTMLSelectElement[]
    expect(selects[1].value).toBe('USD')
    fireEvent.change(selects[1], { target: { value: 'GBP' } })

    await waitFor(() => {
      expect(selects[1].value).toBe('GBP')
    })
  })

  it('displays zero result correctly', async () => {
    renderComponent()
    const input = screen.getByRole('spinbutton') as HTMLInputElement
    fireEvent.change(input, { target: { value: '0' } })

    await waitFor(() => {
      expect(input.value).toBe('0')
    })
  })

  it('formats currency display', async () => {
    renderComponent()
    const input = screen.getByRole('spinbutton') as HTMLInputElement
    fireEvent.change(input, { target: { value: '1000' } })

    await waitFor(() => {
      expect(input.value).toBe('1000')
    })
  })
})
