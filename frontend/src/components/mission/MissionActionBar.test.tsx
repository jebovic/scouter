import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { MissionActionBar } from './MissionActionBar'

const defaultProps = {
  mission: { id: 'abc', slug: 'test-mission', name: 'Test Mission' },
  onEdit: vi.fn(),
  onArchive: vi.fn(),
  onDelete: vi.fn(),
}

describe('MissionActionBar', () => {
  it('renders Edit, Archive, and Delete buttons', () => {
    render(<MissionActionBar {...defaultProps} />)
    expect(screen.getByRole('button', { name: /edit/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /archive/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /delete/i })).toBeInTheDocument()
  })

  it('calls onEdit when Edit is clicked', () => {
    render(<MissionActionBar {...defaultProps} />)
    fireEvent.click(screen.getByRole('button', { name: /edit/i }))
    expect(defaultProps.onEdit).toHaveBeenCalled()
  })

  it('shows archive confirmation dialog when Archive is clicked', () => {
    render(<MissionActionBar {...defaultProps} />)
    fireEvent.click(screen.getByRole('button', { name: /archive/i }))
    expect(screen.getByText(/archive this mission/i)).toBeInTheDocument()
  })

  it('calls onArchive when archive is confirmed', async () => {
    render(<MissionActionBar {...defaultProps} />)
    fireEvent.click(screen.getByRole('button', { name: /archive/i }))
    fireEvent.click(screen.getByRole('button', { name: /confirm/i }))
    await waitFor(() => expect(defaultProps.onArchive).toHaveBeenCalled())
  })

  it('shows delete confirmation with name input when Delete is clicked', () => {
    render(<MissionActionBar {...defaultProps} />)
    fireEvent.click(screen.getByRole('button', { name: /delete/i }))
    expect(screen.getByPlaceholderText(/test mission/i)).toBeInTheDocument()
  })

  it('keeps delete confirm button disabled until name is typed correctly', () => {
    render(<MissionActionBar {...defaultProps} />)
    fireEvent.click(screen.getByRole('button', { name: /delete/i }))
    const confirmBtn = screen.getByRole('button', { name: /confirm delete/i })
    expect(confirmBtn).toBeDisabled()
    fireEvent.change(screen.getByPlaceholderText(/test mission/i), {
      target: { value: 'Test Mission' },
    })
    expect(confirmBtn).not.toBeDisabled()
  })

  it('calls onDelete when delete is confirmed with correct name', async () => {
    render(<MissionActionBar {...defaultProps} />)
    fireEvent.click(screen.getByRole('button', { name: /delete/i }))
    fireEvent.change(screen.getByPlaceholderText(/test mission/i), {
      target: { value: 'Test Mission' },
    })
    fireEvent.click(screen.getByRole('button', { name: /confirm delete/i }))
    await waitFor(() => expect(defaultProps.onDelete).toHaveBeenCalled())
  })
})
