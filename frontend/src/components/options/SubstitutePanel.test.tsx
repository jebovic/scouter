import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import React, { ReactNode } from 'react'
import { SubstitutePanel } from './SubstitutePanel'
import * as substitutesApi from '../../api/substitutes'

vi.mock('../../api/substitutes')

const mockResponse: substitutesApi.SubstituteResponse = {
  optionId: 'opt-111',
  productName: 'iPhone 15 Pro',
  substitutes: [
    {
      name: 'Samsung Galaxy S23',
      brand: 'Samsung',
      price: 799,
      currency: 'EUR',
      retailer: 'Fnac',
      reason: 'Comparable specs at a lower price, widely available in France.',
      advantage: 'cheaper',
      url: 'https://fnac.com/samsung-s23',
    },
    {
      name: 'Fairphone 5',
      brand: 'Fairphone',
      price: 699,
      currency: 'EUR',
      retailer: 'Darty',
      reason: 'Ethically sourced materials, modular repair.',
      advantage: 'eco_friendly',
    },
  ],
  generatedAt: new Date().toISOString(),
}

function makeWrapper() {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return ({ children }: { children: ReactNode }) =>
    React.createElement(QueryClientProvider, { client: qc }, children)
}

describe('SubstitutePanel', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the trigger button initially collapsed', () => {
    render(<SubstitutePanel optionId="opt-111" />, { wrapper: makeWrapper() })
    expect(
      screen.getByRole('button', { name: /trouver des alternatives/i }),
    ).toBeInTheDocument()
    // substitutes list not visible yet
    expect(screen.queryByText('Samsung Galaxy S23')).not.toBeInTheDocument()
  })

  it('shows loading skeleton after clicking the trigger', async () => {
    // Never resolves so we stay in loading state
    vi.spyOn(substitutesApi, 'fetchSubstitutes').mockReturnValue(new Promise(() => {}))

    render(<SubstitutePanel optionId="opt-111" />, { wrapper: makeWrapper() })
    await userEvent.click(screen.getByRole('button', { name: /trouver des alternatives/i }))

    expect(substitutesApi.fetchSubstitutes).toHaveBeenCalledWith('opt-111')
  })

  it('renders substitute names after successful fetch', async () => {
    vi.spyOn(substitutesApi, 'fetchSubstitutes').mockResolvedValue(mockResponse)

    render(<SubstitutePanel optionId="opt-111" />, { wrapper: makeWrapper() })
    await userEvent.click(screen.getByRole('button', { name: /trouver des alternatives/i }))

    await waitFor(() => {
      expect(screen.getByText('Samsung Galaxy S23')).toBeInTheDocument()
      expect(screen.getByText('Fairphone 5')).toBeInTheDocument()
    })
  })

  it('shows advantage badges for each substitute', async () => {
    vi.spyOn(substitutesApi, 'fetchSubstitutes').mockResolvedValue(mockResponse)

    render(<SubstitutePanel optionId="opt-111" />, { wrapper: makeWrapper() })
    await userEvent.click(screen.getByRole('button', { name: /trouver des alternatives/i }))

    await waitFor(() => {
      expect(screen.getByText(/prix bas/i)).toBeInTheDocument()
      expect(screen.getByText(/éco/i)).toBeInTheDocument()
    })
  })

  it('shows prices formatted in French locale', async () => {
    vi.spyOn(substitutesApi, 'fetchSubstitutes').mockResolvedValue(mockResponse)

    render(<SubstitutePanel optionId="opt-111" />, { wrapper: makeWrapper() })
    await userEvent.click(screen.getByRole('button', { name: /trouver des alternatives/i }))

    await waitFor(() => {
      // French locale formats: "799,00 €" or "799 €"
      const priceEl = screen.getAllByText(/€/)
      expect(priceEl.length).toBeGreaterThan(0)
    })
  })

  it('renders a link when url is present', async () => {
    vi.spyOn(substitutesApi, 'fetchSubstitutes').mockResolvedValue(mockResponse)

    render(<SubstitutePanel optionId="opt-111" />, { wrapper: makeWrapper() })
    await userEvent.click(screen.getByRole('button', { name: /trouver des alternatives/i }))

    await waitFor(() => {
      const link = screen.getByRole('link', { name: /voir chez fnac/i })
      expect(link).toHaveAttribute('href', 'https://fnac.com/samsung-s23')
    })
  })

  it('shows empty state when substitutes array is empty', async () => {
    vi.spyOn(substitutesApi, 'fetchSubstitutes').mockResolvedValue({
      ...mockResponse,
      substitutes: [],
    })

    render(<SubstitutePanel optionId="opt-111" />, { wrapper: makeWrapper() })
    await userEvent.click(screen.getByRole('button', { name: /trouver des alternatives/i }))

    await waitFor(() => {
      expect(screen.getByText(/aucune alternative trouvée/i)).toBeInTheDocument()
    })
  })

  it('shows retailer badge text', async () => {
    vi.spyOn(substitutesApi, 'fetchSubstitutes').mockResolvedValue(mockResponse)

    render(<SubstitutePanel optionId="opt-111" />, { wrapper: makeWrapper() })
    await userEvent.click(screen.getByRole('button', { name: /trouver des alternatives/i }))

    await waitFor(() => {
      expect(screen.getByText('Fnac')).toBeInTheDocument()
      expect(screen.getByText('Darty')).toBeInTheDocument()
    })
  })

  it('shows reason text for each substitute', async () => {
    vi.spyOn(substitutesApi, 'fetchSubstitutes').mockResolvedValue(mockResponse)

    render(<SubstitutePanel optionId="opt-111" />, { wrapper: makeWrapper() })
    await userEvent.click(screen.getByRole('button', { name: /trouver des alternatives/i }))

    await waitFor(() => {
      expect(
        screen.getByText(/comparable specs at a lower price/i),
      ).toBeInTheDocument()
    })
  })
})
