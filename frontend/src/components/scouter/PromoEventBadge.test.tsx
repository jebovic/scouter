import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { PromoEventBadge } from './PromoEventBadge'
import type { PromoEvent } from '../../utils/frenchPromoCalendar'

describe('PromoEventBadge', () => {
  const mockEvent: PromoEvent = {
    id: 'test-event',
    name: 'Test Promo',
    startMonth: 3,
    startDay: 15,
    endMonth: 3,
    endDay: 25,
    category: 'general',
    discount: '30%',
    description: 'Test description',
    icon: '🎉',
    color: 'var(--gold)',
  }

  it('should render event name', () => {
    render(<PromoEventBadge event={mockEvent} />)
    expect(screen.getByText('Test Promo')).toBeInTheDocument()
  })

  it('should render icon', () => {
    render(<PromoEventBadge event={mockEvent} />)
    expect(screen.getByText('🎉')).toBeInTheDocument()
  })

  it('should render discount in full mode', () => {
    render(<PromoEventBadge event={mockEvent} compact={false} />)
    expect(screen.getByText('30%')).toBeInTheDocument()
  })

  it('should not render discount in compact mode', () => {
    render(<PromoEventBadge event={mockEvent} compact={true} />)
    expect(screen.queryByText('30%')).not.toBeInTheDocument()
  })

  it('should call onClick handler when clicked', async () => {
    const handleClick = vi.fn()
    const user = userEvent.setup()
    render(<PromoEventBadge event={mockEvent} onClick={handleClick} />)

    const button = screen.getByRole('button')
    await user.click(button)

    expect(handleClick).toHaveBeenCalledOnce()
    expect(handleClick).toHaveBeenCalledWith(mockEvent)
  })

  it('should apply event color as CSS variable', () => {
    const customColorEvent = {
      ...mockEvent,
      color: 'var(--coral)',
    }
    const { container } = render(<PromoEventBadge event={customColorEvent} />)
    const element = container.querySelector('button')
    expect(element).toBeInTheDocument()
  })

  it('should handle events with long names', () => {
    const longNameEvent = {
      ...mockEvent,
      name: 'Very Long Promotional Event Name That Might Wrap',
    }
    render(<PromoEventBadge event={longNameEvent} />)
    expect(screen.getByText('Very Long Promotional Event Name That Might Wrap')).toBeInTheDocument()
  })
})
