import { describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { renderWithProviders } from '../../test/utils'
import { WishListItem } from './WishListItem'
import type { WishListItem as WishListItemType } from '../../api/wishlist'

const baseItem: WishListItemType = {
  id: 'test-uuid-1234',
  name: 'Sony WH-1000XM5',
  url: 'https://example.com/sony',
  targetPrice: 299.99,
  currency: 'EUR',
  notes: 'Wait for sale',
  alertEnabled: false,
  lastCheckedAt: null,
  lastCheckedPrice: null,
  createdAt: '2026-01-15T10:00:00Z',
  updatedAt: '2026-01-15T10:00:00Z',
}

describe('WishListItem', () => {
  it('renders item name', () => {
    renderWithProviders(<WishListItem item={baseItem} onDelete={vi.fn()} />)
    expect(screen.getByText('Sony WH-1000XM5')).toBeInTheDocument()
  })

  it('renders clickable URL link when url is provided', () => {
    renderWithProviders(<WishListItem item={baseItem} onDelete={vi.fn()} />)
    const link = screen.getByRole('link')
    expect(link).toHaveAttribute('href', 'https://example.com/sony')
    expect(link).toHaveAttribute('target', '_blank')
    expect(link).toHaveAttribute('rel', 'noopener noreferrer')
  })

  it('does not render link when url is null', () => {
    renderWithProviders(<WishListItem item={{ ...baseItem, url: null }} onDelete={vi.fn()} />)
    expect(screen.queryByRole('link')).not.toBeInTheDocument()
  })

  it('renders formatted target price', () => {
    renderWithProviders(<WishListItem item={baseItem} onDelete={vi.fn()} />)
    expect(screen.getByText(/299/)).toBeInTheDocument()
  })

  it('does not render price section when targetPrice is null', () => {
    renderWithProviders(<WishListItem item={{ ...baseItem, targetPrice: null }} onDelete={vi.fn()} />)
    expect(screen.queryByText(/299/)).not.toBeInTheDocument()
  })

  it('renders notes when provided', () => {
    renderWithProviders(<WishListItem item={baseItem} onDelete={vi.fn()} />)
    expect(screen.getByText('Wait for sale')).toBeInTheDocument()
  })

  it('does not render notes section when notes is null', () => {
    renderWithProviders(<WishListItem item={{ ...baseItem, notes: null }} onDelete={vi.fn()} />)
    expect(screen.queryByText('Wait for sale')).not.toBeInTheDocument()
  })

  it('renders delete button', () => {
    renderWithProviders(<WishListItem item={baseItem} onDelete={vi.fn()} />)
    expect(screen.getByRole('button', { name: /delete/i })).toBeInTheDocument()
  })

  it('calls onDelete with item id when delete button is clicked', async () => {
    const onDelete = vi.fn()
    renderWithProviders(<WishListItem item={baseItem} onDelete={onDelete} />)
    await userEvent.click(screen.getByRole('button', { name: /delete/i }))
    expect(onDelete).toHaveBeenCalledWith('test-uuid-1234')
  })

  it('renders item with all nullable fields null', () => {
    const minimalItem: WishListItemType = {
      ...baseItem,
      url: null,
      targetPrice: null,
      notes: null,
    }
    renderWithProviders(<WishListItem item={minimalItem} onDelete={vi.fn()} />)
    expect(screen.getByText('Sony WH-1000XM5')).toBeInTheDocument()
    expect(screen.queryByRole('link')).not.toBeInTheDocument()
  })

  it('displays currency with price', () => {
    renderWithProviders(<WishListItem item={{ ...baseItem, currency: 'USD', targetPrice: 150 }} onDelete={vi.fn()} />)
    expect(screen.getByText(/150/)).toBeInTheDocument()
  })
})
