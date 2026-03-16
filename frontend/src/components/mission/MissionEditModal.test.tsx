import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { MissionEditModal } from './MissionEditModal'
import type { Mission, MissionCreateRequest } from '../../types'

vi.mock('./MissionForm', () => ({
  MissionForm: ({ initialValues, onSubmit, onCancel }: {
    initialValues?: Partial<MissionCreateRequest & { envelopeId?: string | null }>
    onSubmit: (req: MissionCreateRequest & { envelopeId?: string | null }) => void
    onCancel: () => void
  }) => (
    <form aria-label="mission form" onSubmit={(e) => { e.preventDefault(); onSubmit({ name: initialValues?.name ?? '', icon: '🎯', category: 'custom', budget: 0, currency: 'USD', locale: 'en', constraints: [], costCategories: [], envelopeId: null }) }}>
      <input name="name" defaultValue={initialValues?.name} />
      <button type="button" onClick={onCancel}>Cancel</button>
      <button type="submit">Save</button>
    </form>
  )
}))

const mockMission: Partial<Mission> = {
  id: 'abc',
  slug: 'test-mission',
  name: 'Test Mission',
  budget: 2500,
  currency: 'EUR',
  category: 'computing',
  constraints: [],
}

describe('MissionEditModal', () => {
  it('renders with mission name visible', () => {
    render(
      <MissionEditModal
        mission={mockMission as Mission}
        onSave={vi.fn()}
        onClose={vi.fn()}
      />
    )
    expect(screen.getByDisplayValue('Test Mission')).toBeTruthy()
  })

  it('calls onClose when Cancel is clicked', () => {
    const onClose = vi.fn()
    render(
      <MissionEditModal
        mission={mockMission as Mission}
        onSave={vi.fn()}
        onClose={onClose}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: /cancel/i }))
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('calls onSave when form is submitted', () => {
    const onSave = vi.fn()
    render(
      <MissionEditModal
        mission={mockMission as Mission}
        onSave={onSave}
        onClose={vi.fn()}
      />
    )
    fireEvent.submit(screen.getByRole('form'))
    expect(onSave).toHaveBeenCalledOnce()
    expect(onSave).toHaveBeenCalledWith(expect.objectContaining({ name: 'Test Mission' }))
  })
})
