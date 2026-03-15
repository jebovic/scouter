import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { EnvelopeSelector } from './EnvelopeSelector'

// Mock the useEnvelopes hook
vi.mock('../../hooks/useEnvelopes', () => ({
  useEnvelopes: vi.fn(),
}))

import { useEnvelopes } from '../../hooks/useEnvelopes'
const mockUseEnvelopes = vi.mocked(useEnvelopes)

// Valid RFC 4122 v4 UUIDs for tests
const UUID_1 = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'
const UUID_2 = 'b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a22'

const mockEnvelopes = [
  {
    id: UUID_1,
    name: 'Alimentation',
    monthlyAmount: 300,
    currency: 'EUR',
    createdAt: '2026-01-01T00:00:00Z',
  },
  {
    id: UUID_2,
    name: 'Loisirs',
    monthlyAmount: 200,
    currency: 'EUR',
    createdAt: '2026-01-01T00:00:00Z',
  },
]

beforeEach(() => {
  mockUseEnvelopes.mockReturnValue({ envelopes: mockEnvelopes, isLoading: false, error: null })
})

describe('EnvelopeSelector', () => {
  it('renders a select with a no-envelope option first', () => {
    render(<EnvelopeSelector value={null} onChange={() => {}} />)
    const select = screen.getByRole('combobox')
    expect(select).toBeTruthy()
    const options = screen.getAllByRole('option')
    expect(options[0].textContent).toMatch(/no envelope/i)
  })

  it('renders all envelopes as options', () => {
    render(<EnvelopeSelector value={null} onChange={() => {}} />)
    expect(screen.getByText('Alimentation')).toBeTruthy()
    expect(screen.getByText('Loisirs')).toBeTruthy()
  })

  it('shows the selected envelope', () => {
    render(<EnvelopeSelector value={UUID_2} onChange={() => {}} />)
    const select = screen.getByRole('combobox') as HTMLSelectElement
    expect(select.value).toBe(UUID_2)
  })

  it('calls onChange with null when "no envelope" is selected', () => {
    const onChange = vi.fn()
    render(<EnvelopeSelector value={UUID_1} onChange={onChange} />)
    const select = screen.getByRole('combobox')
    fireEvent.change(select, { target: { value: '' } })
    expect(onChange).toHaveBeenCalledWith(null)
  })

  it('calls onChange with envelope id when an envelope is selected', () => {
    const onChange = vi.fn()
    render(<EnvelopeSelector value={null} onChange={onChange} />)
    const select = screen.getByRole('combobox')
    fireEvent.change(select, { target: { value: UUID_1 } })
    expect(onChange).toHaveBeenCalledWith(UUID_1)
  })

  it('renders skeleton/disabled state when loading', () => {
    mockUseEnvelopes.mockReturnValue({ envelopes: [], isLoading: true, error: null })
    render(<EnvelopeSelector value={null} onChange={() => {}} />)
    const select = screen.getByRole('combobox') as HTMLSelectElement
    expect(select.disabled).toBe(true)
  })

  it('shows empty list with only no-envelope option when envelopes is empty', () => {
    mockUseEnvelopes.mockReturnValue({ envelopes: [], isLoading: false, error: null })
    render(<EnvelopeSelector value={null} onChange={() => {}} />)
    const options = screen.getAllByRole('option')
    expect(options).toHaveLength(1)
    expect(options[0].textContent).toMatch(/no envelope/i)
  })
})
