import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ComparisonPanel } from './ComparisonPanel'
import type { Option } from '../../types'

function makeOption(overrides: Partial<Option> & { id: string; name: string }): Option {
  return {
    id: overrides.id,
    name: overrides.name,
    missionId: 'mission-1',
    category: 'Electronics',
    badge: 'recommended',
    attributes: overrides.attributes ?? [],
    priceRange: overrides.priceRange,
    warnings: [],
    pinned: false,
    rejected: false,
    createdAt: '2024-01-01T00:00:00Z',
    ...overrides,
  }
}

const optionA = makeOption({
  id: 'opt-1',
  name: 'Option Alpha',
  priceRange: { min: 100, max: 200, best: 150 },
  attributes: [
    { key: 'ram', label: 'RAM', value: '16GB', type: 'text' },
    { key: 'score_val', label: 'Score', value: 85, type: 'score', max: 100 },
  ],
})

const optionB = makeOption({
  id: 'opt-2',
  name: 'Option Beta',
  priceRange: { min: 80, max: 180, best: 90 },
  attributes: [
    { key: 'ram', label: 'RAM', value: '8GB', type: 'text' },
    { key: 'score_val', label: 'Score', value: 70, type: 'score', max: 100 },
  ],
})

describe('ComparisonPanel', () => {
  const noop = vi.fn()

  it('renders option names in header', () => {
    render(<ComparisonPanel options={[optionA, optionB]} onClose={noop} />)
    expect(screen.getByText('Option Alpha')).toBeInTheDocument()
    expect(screen.getByText('Option Beta')).toBeInTheDocument()
  })

  it('renders a close button', () => {
    render(<ComparisonPanel options={[optionA, optionB]} onClose={noop} />)
    expect(screen.getByRole('button', { name: /close|fermer|×/i })).toBeInTheDocument()
  })

  it('calls onClose when close button is clicked', async () => {
    const onClose = vi.fn()
    render(<ComparisonPanel options={[optionA, optionB]} onClose={onClose} />)
    await userEvent.click(screen.getByRole('button', { name: /close|fermer|×/i }))
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('renders price row', () => {
    render(<ComparisonPanel options={[optionA, optionB]} onClose={noop} />)
    // Should have a Price row label
    expect(screen.getByText(/prix|price/i)).toBeInTheDocument()
  })

  it('highlights the lowest price cell', () => {
    render(<ComparisonPanel options={[optionA, optionB]} onClose={noop} />)
    // optionB has lower best price (90 vs 150)
    // The panel uses data-testid or a class for best-value
    const bestCells = document.querySelectorAll('[data-best="true"]')
    expect(bestCells.length).toBeGreaterThan(0)
  })

  it('renders attribute rows', () => {
    render(<ComparisonPanel options={[optionA, optionB]} onClose={noop} />)
    expect(screen.getByText('RAM')).toBeInTheDocument()
  })

  it('renders option without priceRange gracefully', () => {
    const noPrice = makeOption({ id: 'opt-3', name: 'No Price Option' })
    render(<ComparisonPanel options={[optionA, noPrice]} onClose={noop} />)
    expect(screen.getByText('No Price Option')).toBeInTheDocument()
    // Should render '—' for missing price
    expect(screen.getAllByText('—').length).toBeGreaterThan(0)
  })

  it('renders nothing for empty options array', () => {
    const { container } = render(<ComparisonPanel options={[]} onClose={noop} />)
    expect(container.firstChild).toBeNull()
  })

  it('renders up to 4 options side by side', () => {
    const options = ['Alpha', 'Beta', 'Gamma', 'Delta'].map((name, i) =>
      makeOption({ id: `opt-${i}`, name }),
    )
    render(<ComparisonPanel options={options} onClose={noop} />)
    options.forEach((o) => expect(screen.getByText(o.name)).toBeInTheDocument())
  })
})
