import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { OptionEditModal } from './OptionEditModal'
import type { Option } from '../../types'

const mockOption: Partial<Option> = {
  id: 'opt-1',
  missionId: 'mission-1',
  name: 'MacBook Pro M4',
  badge: 'recommended',
  priceRange: { min: 2000, max: 2500, best: 2200 },
  notes: 'Great laptop',
  warnings: 'Expensive',
  attributes: [],
}

describe('OptionEditModal', () => {
  it('renders with option name pre-filled', () => {
    render(
      <OptionEditModal option={mockOption as Option} onSave={vi.fn()} onClose={vi.fn()} />
    )
    expect(screen.getByDisplayValue('MacBook Pro M4')).toBeInTheDocument()
  })

  it('shows badge dropdown with current badge selected', () => {
    render(
      <OptionEditModal option={mockOption as Option} onSave={vi.fn()} onClose={vi.fn()} />
    )
    const select = screen.getByRole('combobox')
    expect((select as HTMLSelectElement).value).toBe('recommended')
  })

  it('calls onClose when Cancel is clicked', () => {
    const onClose = vi.fn()
    render(
      <OptionEditModal option={mockOption as Option} onSave={vi.fn()} onClose={onClose} />
    )
    fireEvent.click(screen.getByRole('button', { name: /cancel/i }))
    expect(onClose).toHaveBeenCalled()
  })

  it('calls onSave with updated fields on submit', () => {
    const onSave = vi.fn()
    render(
      <OptionEditModal option={mockOption as Option} onSave={onSave} onClose={vi.fn()} />
    )
    fireEvent.change(screen.getByDisplayValue('MacBook Pro M4'), {
      target: { value: 'MacBook Pro M4 Pro' },
    })
    fireEvent.submit(screen.getByRole('form'))
    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({ name: 'MacBook Pro M4 Pro' })
    )
  })
})
