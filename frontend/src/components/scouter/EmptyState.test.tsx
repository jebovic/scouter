import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { EmptyState } from './EmptyState'

describe('EmptyState', () => {
  it('renders icon, title, and description', () => {
    render(
      <EmptyState
        icon="📭"
        title="No missions yet"
        description="Create your first mission to get started"
      />
    )
    expect(screen.getByText('📭')).toBeInTheDocument()
    expect(screen.getByText('No missions yet')).toBeInTheDocument()
    expect(screen.getByText('Create your first mission to get started')).toBeInTheDocument()
  })

  it('does not render CTA button when onAction is not provided', () => {
    render(
      <EmptyState icon="📭" title="Empty" description="Nothing here" />
    )
    expect(screen.queryByRole('button')).not.toBeInTheDocument()
  })

  it('renders CTA button when actionLabel and onAction are provided', () => {
    render(
      <EmptyState
        icon="📭"
        title="Empty"
        description="Nothing here"
        actionLabel="Create Mission"
        onAction={vi.fn()}
      />
    )
    expect(screen.getByRole('button', { name: 'Create Mission' })).toBeInTheDocument()
  })

  it('calls onAction when CTA button is clicked', async () => {
    const onAction = vi.fn()
    render(
      <EmptyState
        icon="📭"
        title="Empty"
        description="Nothing here"
        actionLabel="Do It"
        onAction={onAction}
      />
    )
    await userEvent.click(screen.getByRole('button', { name: 'Do It' }))
    expect(onAction).toHaveBeenCalledOnce()
  })
})
