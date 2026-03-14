import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Skeleton } from './Skeleton'

describe('Skeleton', () => {
  it('renders card variant', () => {
    render(<Skeleton variant="card" data-testid="sk" />)
    const el = screen.getByTestId('sk')
    expect(el).toBeInTheDocument()
    expect(el.className).toContain('card')
  })

  it('renders row variant', () => {
    render(<Skeleton variant="row" data-testid="sk" />)
    const el = screen.getByTestId('sk')
    expect(el.className).toContain('row')
  })

  it('renders chart variant', () => {
    render(<Skeleton variant="chart" data-testid="sk" />)
    const el = screen.getByTestId('sk')
    expect(el.className).toContain('chart')
  })

  it('renders text variant', () => {
    render(<Skeleton variant="text" data-testid="sk" />)
    const el = screen.getByTestId('sk')
    expect(el.className).toContain('text')
  })

  it('applies custom width and height', () => {
    render(<Skeleton variant="card" width="200px" height="100px" data-testid="sk" />)
    const el = screen.getByTestId('sk')
    expect(el.style.width).toBe('200px')
    expect(el.style.height).toBe('100px')
  })

  it('accepts additional className', () => {
    render(<Skeleton variant="row" className="extra" data-testid="sk" />)
    const el = screen.getByTestId('sk')
    expect(el.className).toContain('extra')
  })

  it('has shimmer class', () => {
    render(<Skeleton variant="card" data-testid="sk" />)
    const el = screen.getByTestId('sk')
    expect(el.className).toContain('shimmer')
  })
})
