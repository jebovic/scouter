import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import React, { ReactNode } from 'react'
import { SubstituteSuggestions } from './SubstituteSuggestions'
import * as substitutesApi from '../../api/substitutes'

vi.mock('../../api/substitutes')

const threeSubstitutes: substitutesApi.Substitute[] = [
  {
    name: 'Anker Soundcore Q45',
    brand: 'Anker',
    price: 59.99,
    currency: 'EUR',
    retailer: 'Amazon France',
    reason: 'Excellent rapport qualité-prix',
    advantage: 'cheaper',
    savingsPercent: 46,
    pros: ['Autonomie 60h', 'ANC efficace', 'Léger'],
    cons: ['Son moins précis', 'Pas de multipoint'],
    whyConsider: 'La meilleure alternative budget pour 90% des usages quotidiens.',
  },
  {
    name: 'JBL Tune 760NC',
    brand: 'JBL',
    price: 89.99,
    currency: 'EUR',
    retailer: 'Fnac',
    reason: 'Bonne réduction de bruit pour le prix',
    advantage: 'cheaper',
    savingsPercent: 27,
    pros: ['Basses puissantes', "50h d'autonomie"],
    cons: ['Micro médiocre'],
    whyConsider: 'Idéal pour les amateurs de basses.',
  },
  {
    name: 'Sony WH-CH720N',
    brand: 'Sony',
    price: 129.99,
    currency: 'EUR',
    retailer: 'Darty',
    reason: 'Meilleur ANC dans cette gamme de prix',
    advantage: 'better_rated',
    savingsPercent: 8,
    pros: ['ANC top', 'Léger 192g', 'Appel multipoint'],
    cons: ['Basses légères'],
    whyConsider: 'Le meilleur choix si l\'ANC est prioritaire.',
  },
]

const mockResponse: substitutesApi.SubstituteResponse = {
  optionId: 'opt-111',
  productName: 'Sony WH-1000XM5',
  originalPrice: 350,
  substitutes: threeSubstitutes,
  generatedAt: new Date().toISOString(),
  cachedAt: 0,
}

function makeWrapper() {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return ({ children }: { children: ReactNode }) =>
    React.createElement(QueryClientProvider, { client: qc }, children)
}

describe('SubstituteSuggestions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the trigger button initially collapsed', () => {
    render(<SubstituteSuggestions optionId="opt-111" />, { wrapper: makeWrapper() })
    expect(
      screen.getByRole('button', { name: /alternatives moins chères/i }),
    ).toBeInTheDocument()
    expect(screen.queryByText('Anker Soundcore Q45')).not.toBeInTheDocument()
  })

  it('calls fetchSubstitutes after clicking the trigger button', async () => {
    vi.spyOn(substitutesApi, 'fetchSubstitutes').mockReturnValue(new Promise(() => {}))

    render(<SubstituteSuggestions optionId="opt-111" />, { wrapper: makeWrapper() })
    await userEvent.click(screen.getByRole('button', { name: /alternatives moins chères/i }))

    expect(substitutesApi.fetchSubstitutes).toHaveBeenCalledWith('opt-111')
  })

  it('renders 3 substitute names after successful fetch', async () => {
    vi.spyOn(substitutesApi, 'fetchSubstitutes').mockResolvedValue(mockResponse)

    render(<SubstituteSuggestions optionId="opt-111" />, { wrapper: makeWrapper() })
    await userEvent.click(screen.getByRole('button', { name: /alternatives moins chères/i }))

    await waitFor(() => {
      expect(screen.getByText('Anker Soundcore Q45')).toBeInTheDocument()
      expect(screen.getByText('JBL Tune 760NC')).toBeInTheDocument()
      expect(screen.getByText('Sony WH-CH720N')).toBeInTheDocument()
    })
  })

  it('shows savings badge with percentage (e.g. -46%)', async () => {
    vi.spyOn(substitutesApi, 'fetchSubstitutes').mockResolvedValue(mockResponse)

    render(<SubstituteSuggestions optionId="opt-111" />, { wrapper: makeWrapper() })
    await userEvent.click(screen.getByRole('button', { name: /alternatives moins chères/i }))

    await waitFor(() => {
      expect(screen.getByText('-46%')).toBeInTheDocument()
    })
  })

  it('shows prices formatted in French locale (gold color)', async () => {
    vi.spyOn(substitutesApi, 'fetchSubstitutes').mockResolvedValue(mockResponse)

    render(<SubstituteSuggestions optionId="opt-111" />, { wrapper: makeWrapper() })
    await userEvent.click(screen.getByRole('button', { name: /alternatives moins chères/i }))

    await waitFor(() => {
      const priceEls = screen.getAllByText(/€/)
      expect(priceEls.length).toBeGreaterThan(0)
    })
  })

  it('shows pros list with green bullets', async () => {
    vi.spyOn(substitutesApi, 'fetchSubstitutes').mockResolvedValue(mockResponse)

    render(<SubstituteSuggestions optionId="opt-111" />, { wrapper: makeWrapper() })
    await userEvent.click(screen.getByRole('button', { name: /alternatives moins chères/i }))

    await waitFor(() => {
      expect(screen.getByText('Autonomie 60h')).toBeInTheDocument()
      expect(screen.getByText('ANC efficace')).toBeInTheDocument()
    })
  })

  it('shows cons list with coral bullets', async () => {
    vi.spyOn(substitutesApi, 'fetchSubstitutes').mockResolvedValue(mockResponse)

    render(<SubstituteSuggestions optionId="opt-111" />, { wrapper: makeWrapper() })
    await userEvent.click(screen.getByRole('button', { name: /alternatives moins chères/i }))

    await waitFor(() => {
      expect(screen.getByText('Son moins précis')).toBeInTheDocument()
    })
  })

  it('shows whyConsider italic text', async () => {
    vi.spyOn(substitutesApi, 'fetchSubstitutes').mockResolvedValue(mockResponse)

    render(<SubstituteSuggestions optionId="opt-111" />, { wrapper: makeWrapper() })
    await userEvent.click(screen.getByRole('button', { name: /alternatives moins chères/i }))

    await waitFor(() => {
      expect(
        screen.getByText(/la meilleure alternative budget/i),
      ).toBeInTheDocument()
    })
  })

  it('shows Google Shopping link hint', async () => {
    vi.spyOn(substitutesApi, 'fetchSubstitutes').mockResolvedValue(mockResponse)

    render(<SubstituteSuggestions optionId="opt-111" />, { wrapper: makeWrapper() })
    await userEvent.click(screen.getByRole('button', { name: /alternatives moins chères/i }))

    await waitFor(() => {
      const links = screen.getAllByRole('link')
      const googleLinks = links.filter((l) => l.getAttribute('href')?.includes('google'))
      expect(googleLinks.length).toBeGreaterThan(0)
    })
  })

  it('shows average savings context when data is loaded', async () => {
    vi.spyOn(substitutesApi, 'fetchSubstitutes').mockResolvedValue(mockResponse)

    render(<SubstituteSuggestions optionId="opt-111" />, { wrapper: makeWrapper() })
    await userEvent.click(screen.getByRole('button', { name: /alternatives moins chères/i }))

    await waitFor(() => {
      // Should show total/average savings context mentioning "économiseraient" or "économies"
      expect(screen.getByText(/économi/i)).toBeInTheDocument()
    })
  })

  it('shows empty state when substitutes array is empty', async () => {
    vi.spyOn(substitutesApi, 'fetchSubstitutes').mockResolvedValue({
      ...mockResponse,
      substitutes: [],
    })

    render(<SubstituteSuggestions optionId="opt-111" />, { wrapper: makeWrapper() })
    await userEvent.click(screen.getByRole('button', { name: /alternatives moins chères/i }))

    await waitFor(() => {
      expect(screen.getByText(/aucune alternative/i)).toBeInTheDocument()
    })
  })
})
