import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { MemoryRouter } from 'react-router-dom'
import { ShortlistPanel } from './ShortlistPanel'
import type { Option } from '../../types'

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}))

const pinnedOption: Partial<Option> = {
  id: 'opt-1',
  name: 'MacBook Pro',
  badge: 'recommended',
  priceRange: { min: 2000, max: 2500, best: 2200 },
  pinned: true,
}

function renderPanel(options: Partial<Option>[] = [], onSelect = vi.fn()) {
  return render(
    <MemoryRouter>
      <ShortlistPanel
        options={options as Option[]}
        missionSlug="test-mission"
        onSelect={onSelect}
      />
    </MemoryRouter>
  )
}

describe('ShortlistPanel', () => {
  it('shows the panel title', () => {
    renderPanel([pinnedOption])
    expect(screen.getByText('shortlist.title')).toBeInTheDocument()
  })

  it('renders pinned options as cards', () => {
    renderPanel([pinnedOption])
    expect(screen.getByText('MacBook Pro')).toBeInTheDocument()
  })

  it('renders a Select button per option', () => {
    renderPanel([pinnedOption])
    expect(screen.getByRole('button', { name: 'shortlist.select' })).toBeInTheDocument()
  })

  it('calls onSelect with the option when Select is clicked', () => {
    const onSelect = vi.fn()
    renderPanel([pinnedOption], onSelect)
    fireEvent.click(screen.getByRole('button', { name: 'shortlist.select' }))
    expect(onSelect).toHaveBeenCalledWith(pinnedOption)
  })

  it('shows empty state when no options are pinned', () => {
    renderPanel([])
    expect(screen.getByText('shortlist.empty')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'shortlist.emptyLink' })).toBeInTheDocument()
  })
})
