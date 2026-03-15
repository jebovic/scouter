import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { StockBadge } from './StockBadge'

describe('StockBadge', () => {
  it('renders "En stock" for in_stock', () => {
    render(<StockBadge status="in_stock" />)
    expect(screen.getByText(/en stock/i)).toBeInTheDocument()
  })

  it('renders "Stock limité" for limited', () => {
    render(<StockBadge status="limited" />)
    expect(screen.getByText(/stock limit/i)).toBeInTheDocument()
  })

  it('renders "Rupture" for out_of_stock', () => {
    render(<StockBadge status="out_of_stock" />)
    expect(screen.getByText(/rupture/i)).toBeInTheDocument()
  })

  it('renders a loading indicator when status is loading', () => {
    render(<StockBadge status="loading" />)
    // Should render something — a spinner or placeholder text
    const badge = screen.getByRole('status')
    expect(badge).toBeInTheDocument()
  })

  it('has an accessible role for each non-loading status', () => {
    const { rerender } = render(<StockBadge status="in_stock" />)
    expect(screen.getByRole('status')).toBeInTheDocument()

    rerender(<StockBadge status="limited" />)
    expect(screen.getByRole('status')).toBeInTheDocument()

    rerender(<StockBadge status="out_of_stock" />)
    expect(screen.getByRole('status')).toBeInTheDocument()
  })
})
