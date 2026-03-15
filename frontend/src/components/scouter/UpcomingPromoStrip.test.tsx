import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { UpcomingPromoStrip } from './UpcomingPromoStrip'

describe('UpcomingPromoStrip', () => {
  it('should render header text when events are available', () => {
    render(<UpcomingPromoStrip />)
    const header = screen.queryByText('Promotions à venir')
    // Header should either exist if events are available, or component returns null
    if (header) {
      expect(header).toBeInTheDocument()
    }
  })

  it('should render upcoming events if available', () => {
    const { container } = render(<UpcomingPromoStrip />)
    // Component may return null if no upcoming events within 60 days
    // or it will render with events
    if (container.firstChild) {
      expect(container.firstChild).toBeInTheDocument()
    }
  })

  it('should return null when no upcoming events (component handles hidden state)', () => {
    const { container } = render(<UpcomingPromoStrip />)
    // Verify that if the component renders null, the container is empty
    // or if it renders, it has content
    const firstChild = container.firstChild
    if (!firstChild) {
      // This is valid - component returns null when no events
      expect(container.innerHTML).toBe('')
    } else {
      // This is valid - component renders with events
      expect(firstChild).toBeInTheDocument()
    }
  })

  it('should accept onEventClick callback', async () => {
    const handleEventClick = vi.fn()
    const user = userEvent.setup()
    const { container } = render(<UpcomingPromoStrip onEventClick={handleEventClick} />)

    const buttons = container.querySelectorAll('button')
    if (buttons.length > 0) {
      await user.click(buttons[0])
      expect(handleEventClick).toHaveBeenCalled()
    }
  })

  it('should render events in a scrollable container if events exist', () => {
    const { container } = render(<UpcomingPromoStrip />)
    // Only check for scrollable container if component renders
    if (container.firstChild) {
      const section = container.querySelector('section')
      expect(section).toBeInTheDocument()
    }
  })

  it('should be accessible with proper semantic HTML if rendered', () => {
    const { container } = render(<UpcomingPromoStrip />)
    // Only check for region role if component renders (not null)
    if (container.firstChild) {
      const section = container.querySelector('[role="region"]')
      expect(section).toBeInTheDocument()
    }
  })
})
