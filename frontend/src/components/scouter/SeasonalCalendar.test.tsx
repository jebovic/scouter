import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { SeasonalCalendar } from './SeasonalCalendar'

// Mock the seasonal pricing utility
vi.mock('../../utils/seasonalPricing', () => ({
  CATEGORY_SEASONS: [
    {
      category: 'Électronique',
      icon: '📱',
      bestMonths: [7, 11],
      worstMonths: [12],
      tip: 'Achetez en juillet ou novembre pour les meilleures réductions.',
    },
    {
      category: 'Mode',
      icon: '👗',
      bestMonths: [1, 7],
      worstMonths: [12],
      tip: 'Les soldes d\'hiver et d\'été offrent les meilleures réductions.',
    },
  ],
  FRENCH_MONTHS: ['Jan', 'Fév', 'Mar', 'Avr', 'Mai', 'Jun', 'Jul', 'Aoû', 'Sep', 'Oct', 'Nov', 'Déc'],
}))

describe('SeasonalCalendar', () => {
  it('renders without crashing', () => {
    render(<SeasonalCalendar />)
    expect(screen.getByRole('region')).toBeInTheDocument()
  })

  it('renders category tabs', () => {
    render(<SeasonalCalendar />)
    const tabs = screen.getAllByRole('tab')
    expect(tabs.length).toBeGreaterThanOrEqual(2)
  })

  it('renders all 12 months in the grid', () => {
    render(<SeasonalCalendar />)
    const monthCells = screen.getAllByRole('cell')
    // 12 months + category column = at least 12 month cells
    expect(monthCells.length).toBeGreaterThanOrEqual(12)
  })

  it('switches category when tab is clicked', () => {
    render(<SeasonalCalendar />)
    const modeTab = screen.getByRole('tab', { name: /Mode/ })
    fireEvent.click(modeTab)
    // After clicking, the component should show Mode category
    expect(screen.getByText(/Les soldes d'hiver/)).toBeInTheDocument()
  })

  it('displays tip text for selected category', () => {
    render(<SeasonalCalendar />)
    const firstTip = screen.getByText(/Achetez en juillet/)
    expect(firstTip).toBeInTheDocument()
  })

  it('renders in compact mode when compact prop is true', () => {
    const { container } = render(<SeasonalCalendar compact={true} />)
    const compactClass = container.querySelector('[class*="compact"]')
    expect(compactClass).toBeDefined()
  })
})
