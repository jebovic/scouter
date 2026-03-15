import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { PriceAlertBadge } from './PriceAlertBadge'

describe('PriceAlertBadge', () => {
  it('returns null when no targetPrice is provided', () => {
    const { container } = render(
      <PriceAlertBadge currentPrice={100} targetPrice={undefined} currency="USD" />
    )
    expect(container.firstChild).toBeNull()
  })

  it('returns null when targetPrice is null', () => {
    const { container } = render(
      <PriceAlertBadge currentPrice={100} targetPrice={null} currency="USD" />
    )
    expect(container.firstChild).toBeNull()
  })

  it('shows "Objectif atteint" badge (green) when currentPrice equals targetPrice', () => {
    render(
      <PriceAlertBadge currentPrice={100} targetPrice={100} currency="USD" />
    )
    const badge = screen.getByText('🎯 Objectif atteint!')
    expect(badge).toBeInTheDocument()
    expect(badge).toHaveClass('achieved')
  })

  it('shows "Objectif atteint" badge (green) when currentPrice is below targetPrice', () => {
    render(
      <PriceAlertBadge currentPrice={95} targetPrice={100} currency="USD" />
    )
    const badge = screen.getByText('🎯 Objectif atteint!')
    expect(badge).toBeInTheDocument()
    expect(badge).toHaveClass('achieved')
  })

  it('shows "Proche de l\'objectif" badge (gold) when currentPrice is 1-10% above targetPrice', () => {
    render(
      <PriceAlertBadge currentPrice={105} targetPrice={100} currency="USD" />
    )
    const badge = screen.getByText('📉 Proche de l\'objectif')
    expect(badge).toBeInTheDocument()
    expect(badge).toHaveClass('close')
  })

  it('shows "Proche de l\'objectif" badge (gold) when currentPrice is exactly 10% above targetPrice', () => {
    render(
      <PriceAlertBadge currentPrice={110} targetPrice={100} currency="USD" />
    )
    const badge = screen.getByText('📉 Proche de l\'objectif')
    expect(badge).toBeInTheDocument()
    expect(badge).toHaveClass('close')
  })

  it('shows target price badge (dim) when currentPrice is more than 10% above targetPrice', () => {
    render(
      <PriceAlertBadge currentPrice={112} targetPrice={100} currency="USD" />
    )
    const badge = screen.getByText(/🎯 Objectif:/)
    expect(badge).toBeInTheDocument()
    expect(badge).toHaveClass('set')
    expect(badge).toHaveTextContent('100')
  })

  it('formats currency correctly (USD)', () => {
    render(
      <PriceAlertBadge currentPrice={150} targetPrice={100} currency="USD" />
    )
    const badge = screen.getByText(/🎯 Objectif:/)
    expect(badge).toBeInTheDocument()
    expect(badge).toHaveTextContent('100')
  })

  it('formats currency correctly (EUR)', () => {
    render(
      <PriceAlertBadge currentPrice={150} targetPrice={100} currency="EUR" />
    )
    const badge = screen.getByText(/🎯 Objectif:/)
    expect(badge).toBeInTheDocument()
    expect(badge).toHaveTextContent('100')
  })

  it('formats currency correctly (fr-FR decimal format)', () => {
    render(
      <PriceAlertBadge currentPrice={150} targetPrice={99.99} currency="EUR" />
    )
    const badge = screen.getByText(/🎯 Objectif:/)
    expect(badge).toBeInTheDocument()
    expect(badge).toHaveTextContent('99,99')
  })

  it('has correct aria-label for achieved state', () => {
    render(
      <PriceAlertBadge currentPrice={50} targetPrice={100} currency="USD" />
    )
    const badge = screen.getByText('🎯 Objectif atteint!')
    const label = badge.getAttribute('aria-label')
    expect(label).toBeTruthy()
    expect(label).toMatch(/Target price achieved:/)
    expect(label).toMatch(/100/)
  })

  it('has correct aria-label for close state', () => {
    render(
      <PriceAlertBadge currentPrice={105} targetPrice={100} currency="USD" />
    )
    const badge = screen.getByText('📉 Proche de l\'objectif')
    const label = badge.getAttribute('aria-label')
    expect(label).toBeTruthy()
    expect(label).toMatch(/Close to target:/)
    expect(label).toMatch(/100/)
  })

  it('has correct aria-label for set state', () => {
    render(
      <PriceAlertBadge currentPrice={150} targetPrice={100} currency="USD" />
    )
    const badge = screen.getByText(/🎯 Objectif:/)
    const label = badge.getAttribute('aria-label')
    expect(label).toBeTruthy()
    expect(label).toMatch(/Target price set:/)
    expect(label).toMatch(/100/)
  })

  it('handles edge case: currentPrice at exactly 110% of targetPrice (boundary)', () => {
    render(
      <PriceAlertBadge currentPrice={110} targetPrice={100} currency="USD" />
    )
    const badge = screen.getByText('📉 Proche de l\'objectif')
    expect(badge).toBeInTheDocument()
  })

  it('handles edge case: currentPrice at 110.01% of targetPrice (just over boundary)', () => {
    render(
      <PriceAlertBadge currentPrice={110.01} targetPrice={100} currency="USD" />
    )
    const badge = screen.getByText(/🎯 Objectif:/)
    expect(badge).toBeInTheDocument()
  })

  it('handles zero targetPrice correctly', () => {
    render(
      <PriceAlertBadge currentPrice={50} targetPrice={0} currency="USD" />
    )
    const badge = screen.getByText(/🎯 Objectif:/)
    expect(badge).toBeInTheDocument()
  })

  it('handles very small targetPrice values', () => {
    render(
      <PriceAlertBadge currentPrice={0.5} targetPrice={0.49} currency="USD" />
    )
    const badge = screen.getByText('📉 Proche de l\'objectif')
    expect(badge).toBeInTheDocument()
  })

  it('handles very large targetPrice values', () => {
    render(
      <PriceAlertBadge currentPrice={10000} targetPrice={9500} currency="USD" />
    )
    const badge = screen.getByText('📉 Proche de l\'objectif')
    expect(badge).toBeInTheDocument()
  })
})
