import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { BudgetRollupWidget } from './BudgetRollupWidget'
import type { BudgetRollup } from '../../hooks/useBudgetRollup'

// Mock useBudgetRollup
vi.mock('../../hooks/useBudgetRollup', () => ({
  useBudgetRollup: vi.fn(),
}))

import { useBudgetRollup } from '../../hooks/useBudgetRollup'

const mockUseBudgetRollup = useBudgetRollup as ReturnType<typeof vi.fn>

function makeRollup(overrides: Partial<BudgetRollup> = {}): BudgetRollup {
  return {
    totalBudget: 3500,
    byCategory: { electronics: 2000, travel: 1500 },
    byPhase: { researching: 2000, comparing: 1500 },
    activeMissions: 2,
    purchasedTotal: 500,
    entries: [
      {
        missionId: 'uuid-1',
        missionName: 'Laptop Pro',
        missionSlug: 'laptop-pro',
        category: 'electronics',
        budget: 2000,
        phase: 'researching',
        currency: 'EUR',
      },
      {
        missionId: 'uuid-2',
        missionName: 'Tokyo Trip',
        missionSlug: 'tokyo-trip',
        category: 'travel',
        budget: 1500,
        phase: 'comparing',
        currency: 'EUR',
      },
    ],
    ...overrides,
  }
}

function renderWidget(rollup?: BudgetRollup) {
  mockUseBudgetRollup.mockReturnValue({
    rollup: rollup ?? makeRollup(),
    isLoading: false,
  })
  return render(
    <MemoryRouter>
      <BudgetRollupWidget />
    </MemoryRouter>
  )
}

describe('BudgetRollupWidget', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders nothing when there are no entries', () => {
    mockUseBudgetRollup.mockReturnValue({
      rollup: makeRollup({ entries: [], totalBudget: 0, activeMissions: 0 }),
      isLoading: false,
    })
    const { container } = render(
      <MemoryRouter>
        <BudgetRollupWidget />
      </MemoryRouter>
    )
    expect(container.firstChild).toBeNull()
  })

  it('renders the header with active budget label', () => {
    renderWidget()
    expect(screen.getByText(/Budget Total Actif/i)).toBeInTheDocument()
  })

  it('renders the total budget as formatted EUR amount', () => {
    renderWidget()
    // 3 500 € formatted in fr-FR style
    expect(screen.getByText(/3\s*500/)).toBeInTheDocument()
  })

  it('renders the count of active missions', () => {
    renderWidget()
    // missionCount span contains "2 missions actives" across text nodes
    const missionCountEl = document.querySelector('[class*="missionCount"]')
    expect(missionCountEl?.textContent?.trim()).toMatch(/2/)
  })

  it('renders a progress bar element', () => {
    renderWidget()
    const progressBar = document.querySelector('[data-testid="budget-progress-fill"]')
    expect(progressBar).toBeInTheDocument()
  })

  it('renders category breakdown bars', () => {
    renderWidget()
    // category names appear in bars section and mission rows
    expect(screen.getAllByText('electronics').length).toBeGreaterThan(0)
    expect(screen.getAllByText('travel').length).toBeGreaterThan(0)
  })

  it('renders phase funnel badges', () => {
    renderWidget()
    // Phase labels should be visible (may appear multiple times due to badge + mission row)
    expect(screen.getAllByText(/Recherche/i).length).toBeGreaterThan(0)
    expect(screen.getAllByText(/Comparaison/i).length).toBeGreaterThan(0)
  })

  it('renders mission list with names', () => {
    renderWidget()
    expect(screen.getByText('Laptop Pro')).toBeInTheDocument()
    expect(screen.getByText('Tokyo Trip')).toBeInTheDocument()
  })

  it('collapses and expands on header click', () => {
    renderWidget()
    // Initially expanded - missions are visible
    expect(screen.getByText('Laptop Pro')).toBeInTheDocument()

    // Click header to collapse
    const header = screen.getByRole('button', { name: /Budget Total Actif/i })
    fireEvent.click(header)

    // After collapse, mission list should not be visible
    expect(screen.queryByText('Laptop Pro')).not.toBeInTheDocument()

    // Click again to expand
    fireEvent.click(header)
    expect(screen.getByText('Laptop Pro')).toBeInTheDocument()
  })

  it('renders purchased total when present', () => {
    renderWidget(
      makeRollup({
        purchasedTotal: 500,
        byPhase: { researching: 2000, comparing: 1500, done: 500 },
        entries: [
          {
            missionId: 'uuid-1',
            missionName: 'Laptop Pro',
            missionSlug: 'laptop-pro',
            category: 'electronics',
            budget: 2000,
            phase: 'researching',
            currency: 'EUR',
          },
          {
            missionId: 'uuid-3',
            missionName: 'Old Camera',
            missionSlug: 'old-camera',
            category: 'electronics',
            budget: 500,
            phase: 'done',
            currency: 'EUR',
          },
        ],
      })
    )
    expect(screen.getAllByText(/Acheté/i).length).toBeGreaterThan(0)
  })
})
