import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { MemoryRouter } from 'react-router-dom'
import { MissionCard } from './MissionCard'
import type { Mission } from '../../types'

vi.mock('../../hooks/useMission', () => ({
  useArchiveMission: () => ({ archiveMission: vi.fn(), isPending: false }),
  useUnarchiveMission: () => ({ unarchiveMission: vi.fn(), isPending: false }),
  useDeleteMission: () => ({ deleteMission: vi.fn(), isPending: false }),
  useDuplicateMission: () => ({ duplicateMission: vi.fn(), isPending: false }),
  useCloneMission: () => ({ cloneMission: vi.fn(), isPending: false }),
}))

const mockMission: Partial<Mission> = {
  id: 'abc',
  slug: 'test-mission',
  name: 'Test Mission',
  phase: 'researching',
  budget: 2500,
  currency: 'EUR',
  category: 'computing',
  constraints: [],
  archivedAt: null,
}

function renderCard(props: Partial<{ mission: Mission }> = {}) {
  return render(
    <MemoryRouter>
      <MissionCard mission={(props.mission ?? mockMission) as Mission} />
    </MemoryRouter>
  )
}

describe('MissionCard ⋯ menu', () => {
  it('renders the overflow menu button', () => {
    renderCard()
    expect(screen.getByRole('button', { name: /more options/i })).toBeInTheDocument()
  })

  it('opens menu on click showing Archive and Delete buttons', () => {
    renderCard()
    fireEvent.click(screen.getByRole('button', { name: /more options/i }))
    // Archive and Delete should be present as clickable buttons in the menu
    expect(screen.getByRole('button', { name: /^archive$/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /^delete$/i })).toBeInTheDocument()
  })

  it('shows Unarchive instead of Archive when mission is archived', () => {
    renderCard({ mission: { ...mockMission, archivedAt: '2026-01-01T00:00:00Z' } as Mission })
    fireEvent.click(screen.getByRole('button', { name: /more options/i }))
    expect(screen.getByRole('button', { name: /unarchive/i })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /^archive$/i })).not.toBeInTheDocument()
  })

  it('shows delete confirmation dialog when Delete is clicked', () => {
    renderCard()
    fireEvent.click(screen.getByRole('button', { name: /more options/i }))
    fireEvent.click(screen.getByRole('button', { name: /^delete$/i }))
    expect(screen.getByRole('textbox')).toBeInTheDocument()
  })

  it('enables confirm button after typing the mission name', () => {
    renderCard()
    fireEvent.click(screen.getByRole('button', { name: /more options/i }))
    fireEvent.click(screen.getByRole('button', { name: /^delete$/i }))
    const input = screen.getByRole('textbox')
    // Before typing: confirm button is disabled
    const allButtons = screen.getAllByRole('button')
    const confirmBtnBefore = allButtons.find(
      (b) => b.textContent?.toLowerCase() === 'delete'
    ) as HTMLButtonElement | undefined
    expect(confirmBtnBefore?.disabled).toBe(true)
    // Type the correct mission name
    fireEvent.change(input, { target: { value: 'Test Mission' } })
    // After typing: confirm button is enabled
    const allButtonsAfter = screen.getAllByRole('button')
    const confirmBtnAfter = allButtonsAfter.find(
      (b) => b.textContent?.toLowerCase() === 'delete'
    ) as HTMLButtonElement | undefined
    expect(confirmBtnAfter?.disabled).toBe(false)
  })
})
