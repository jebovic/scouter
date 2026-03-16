import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { OptionCard } from './OptionCard'
import type { Option } from '../../types'

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en' } }),
}))

vi.mock('../../hooks/useOptionImages', () => ({
  useOptionImages: () => ({ data: [], isLoading: false }),
}))

vi.mock('../../hooks/useTranslatedOption', () => ({
  useTranslatedOption: (option: Option) => option,
}))

vi.mock('./VoteButtons', () => ({
  VoteButtons: () => null,
}))

vi.mock('./ReviewSummaryCard', () => ({
  ReviewSummaryCard: () => null,
}))

vi.mock('./PriceComparisonPanel', () => ({
  PriceComparisonPanel: () => null,
}))

vi.mock('./SubstituteSuggestions', () => ({
  SubstituteSuggestions: () => null,
}))

vi.mock('./RetailerLinks', () => ({
  RetailerLinks: () => null,
}))

vi.mock('./OptionNote', () => ({
  OptionNote: () => null,
}))

vi.mock('./ImageStrip', () => ({
  ImageStrip: () => null,
}))

vi.mock('../scouter', () => ({
  Card: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  Badge: () => null,
}))

const mockOption: Partial<Option> = {
  id: 'opt-1',
  missionId: 'mission-1',
  name: 'MacBook Pro',
  badge: 'recommended',
  priceRange: { min: 2000, max: 2500, best: 2200 },
  attributes: [],
  pinned: false,
  rejected: false,
  warnings: [],
}

describe('OptionCard overflow menu', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the overflow menu button', () => {
    render(
      <OptionCard
        option={mockOption as Option}
        missionId="mission-1"
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />
    )
    expect(screen.getByRole('button', { name: 'more options' })).toBeInTheDocument()
  })

  it('opens menu showing Edit and Delete items when button clicked', () => {
    render(
      <OptionCard
        option={mockOption as Option}
        missionId="mission-1"
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'more options' }))
    expect(screen.getByRole('menuitem', { name: /option\.actions\.edit/i })).toBeInTheDocument()
    expect(screen.getByRole('menuitem', { name: /option\.actions\.delete/i })).toBeInTheDocument()
  })

  it('calls onEdit with optionId when Edit is clicked', () => {
    const onEdit = vi.fn()
    render(
      <OptionCard
        option={mockOption as Option}
        missionId="mission-1"
        onEdit={onEdit}
        onDelete={vi.fn()}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'more options' }))
    fireEvent.click(screen.getByRole('menuitem', { name: /option\.actions\.edit/i }))
    expect(onEdit).toHaveBeenCalledWith('opt-1')
  })

  it('shows delete confirmation text when Delete is clicked', () => {
    render(
      <OptionCard
        option={mockOption as Option}
        missionId="mission-1"
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'more options' }))
    fireEvent.click(screen.getByRole('menuitem', { name: /option\.actions\.delete/i }))
    expect(screen.getByText('option.actions.deleteConfirm')).toBeInTheDocument()
  })

  it('calls onDelete with optionId when confirm button is clicked', () => {
    const onDelete = vi.fn()
    render(
      <OptionCard
        option={mockOption as Option}
        missionId="mission-1"
        onEdit={vi.fn()}
        onDelete={onDelete}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'more options' }))
    fireEvent.click(screen.getByRole('menuitem', { name: /option\.actions\.delete/i }))
    const confirmBtn = screen.getByRole('button', { name: /common\.delete/i })
    fireEvent.click(confirmBtn)
    expect(onDelete).toHaveBeenCalledWith('opt-1')
  })
})
